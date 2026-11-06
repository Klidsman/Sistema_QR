package academico

import (
	"context"
	"errors"
	"fmt"
	"time"

	"asistencia/internal/academico/generado"
	"asistencia/internal/plataforma/auditoria"
	"asistencia/internal/plataforma/fechas"

	"github.com/jackc/pgx/v5"
)

// ErrExcepcionSolapada indica que la excepción se cruza con otra del mismo alcance.
var ErrExcepcionSolapada = errors.New("la excepción se superpone con otra del mismo alcance")

// DatosExcepcion son los datos para registrar una excepción de calendario.
type DatosExcepcion struct {
	AnioEscolarID int64
	Tipo          string
	FechaInicio   time.Time
	FechaFin      time.Time
	Motivo        string
	Grados        []int64
}

// SeSolapan indica si dos excepciones comparten al menos un día.
func SeSolapan(primera, segunda ExcepcionCalendario) bool {
	return fechas.CompararDia(primera.FechaFin, segunda.FechaInicio) >= 0 &&
		fechas.CompararDia(segunda.FechaFin, primera.FechaInicio) >= 0
}

// MismoAlcance indica si dos excepciones alcanzan exactamente el mismo conjunto de grados.
// Dos excepciones globales tienen el mismo alcance; una global y una por grado, no.
func MismoAlcance(primera, segunda ExcepcionCalendario) bool {
	if len(primera.Grados) == 0 && len(segunda.Grados) == 0 {
		return true
	}
	if len(primera.Grados) == 0 || len(segunda.Grados) == 0 {
		return false
	}
	grados := make(map[int64]bool, len(primera.Grados))
	for _, gradoID := range primera.Grados {
		grados[gradoID] = true
	}
	for _, gradoID := range segunda.Grados {
		if grados[gradoID] {
			return true
		}
	}
	return false
}

// RegistrarExcepcion registra un feriado, día no laborable o suspensión y lo audita.
func (s *ServicioConfiguracion) RegistrarExcepcion(ctx context.Context, usuarioID int64, datos DatosExcepcion, confirmar bool) (int64, error) {
	datos.Motivo = NormalizarNombre(datos.Motivo)
	if err := ValidarExcepcion(datos.Tipo, datos.Motivo, datos.FechaInicio, datos.FechaFin); err != nil {
		return 0, err
	}
	grados := gradosUnicos(datos.Grados)

	var id int64
	err := s.repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		anio, err := consultas.BuscarAnioEscolarPorID(ctx, datos.AnioEscolarID)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrAnioNoEncontrado
		}
		if err != nil {
			return fmt.Errorf("buscando el año escolar: %w", err)
		}
		if anio.Estado == EstadoAnioCerrado {
			return ErrAnioCerrado
		}

		filasGrados, err := consultas.ListarGradosPorAnio(ctx, datos.AnioEscolarID)
		if err != nil {
			return fmt.Errorf("listando los grados del año: %w", err)
		}
		permitidos := make(map[int64]bool, len(filasGrados))
		for _, grado := range filasGrados {
			permitidos[grado.ID] = true
		}
		for _, gradoID := range grados {
			if !permitidos[gradoID] {
				return ErrGradoNoEncontrado
			}
		}

		existentes, err := cargarExcepciones(ctx, consultas, datos.AnioEscolarID)
		if err != nil {
			return err
		}
		nueva := ExcepcionCalendario{
			Tipo:        datos.Tipo,
			FechaInicio: datos.FechaInicio,
			FechaFin:    datos.FechaFin,
			Grados:      grados,
		}
		if len(excepcionesSolapadas(existentes, nueva)) > 0 && !confirmar {
			return ErrExcepcionSolapada
		}

		creado, err := consultas.CrearExcepcionCalendario(ctx, generado.CrearExcepcionCalendarioParams{
			AnioEscolarID: datos.AnioEscolarID,
			Tipo:          datos.Tipo,
			FechaInicio:   parametroFecha(datos.FechaInicio),
			FechaFin:      parametroFecha(datos.FechaFin),
			Motivo:        datos.Motivo,
			CreadoPor:     usuarioID,
		})
		if err != nil {
			return fmt.Errorf("creando la excepción de calendario: %w", err)
		}
		for _, gradoID := range grados {
			if err := consultas.CrearExcepcionGrado(ctx, generado.CrearExcepcionGradoParams{
				ExcepcionID: creado,
				GradoID:     gradoID,
			}); err != nil {
				return fmt.Errorf("vinculando la excepción con el grado: %w", err)
			}
		}
		if err := auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID: usuarioID,
			Entidad:   EntidadCalendario,
			EntidadID: creado,
			Accion:    AccionRegistrarExcepcion,
			ValorNuevo: map[string]any{
				"tipo":         datos.Tipo,
				"fecha_inicio": fechas.FormatearISO(datos.FechaInicio),
				"fecha_fin":    fechas.FormatearISO(datos.FechaFin),
				"motivo":       datos.Motivo,
				"grados":       grados,
			},
		}); err != nil {
			return err
		}
		id = creado
		return nil
	})
	return id, err
}

// DetectarSolapamiento devuelve las excepciones del año que se cruzan con la propuesta.
func (s *ServicioConfiguracion) DetectarSolapamiento(ctx context.Context, anioID int64, excepcion ExcepcionCalendario) ([]ExcepcionCalendario, error) {
	existentes, err := s.repositorio.ListarExcepcionesPorAnio(ctx, anioID)
	if err != nil {
		return nil, fmt.Errorf("listando las excepciones del año: %w", err)
	}
	return excepcionesSolapadas(existentes, excepcion), nil
}

// ConfigurarUmbral registra una versión del umbral de alerta y la audita.
func (s *ServicioConfiguracion) ConfigurarUmbral(ctx context.Context, usuarioID int64, umbral UmbralAlerta) error {
	if err := ValidarUmbral(umbral.Tipo, umbral.Valor); err != nil {
		return err
	}
	if err := ValidarVigencia(umbral.VigenteDesde, s.reloj()); err != nil {
		return err
	}

	return s.repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		anio, err := consultas.BuscarAnioEscolarPorID(ctx, umbral.AnioEscolarID)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrAnioNoEncontrado
		}
		if err != nil {
			return fmt.Errorf("buscando el año escolar: %w", err)
		}
		if anio.Estado == EstadoAnioCerrado {
			return ErrAnioCerrado
		}

		anterior := any(nil)
		filaAnterior, err := consultas.UmbralVigente(ctx, generado.UmbralVigenteParams{
			AnioEscolarID: umbral.AnioEscolarID,
			Tipo:          umbral.Tipo,
			VigenteDesde:  parametroFecha(umbral.VigenteDesde),
		})
		if err == nil {
			anterior = map[string]any{"valor": int(filaAnterior.Valor)}
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("buscando el umbral anterior: %w", err)
		}

		if _, err := consultas.GuardarUmbralAlerta(ctx, generado.GuardarUmbralAlertaParams{
			AnioEscolarID: umbral.AnioEscolarID,
			Tipo:          umbral.Tipo,
			Valor:         int32(umbral.Valor),
			VigenteDesde:  parametroFecha(umbral.VigenteDesde),
		}); err != nil {
			return fmt.Errorf("guardando el umbral de alerta: %w", err)
		}
		return auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID:     usuarioID,
			Entidad:       EntidadUmbral,
			EntidadID:     umbral.AnioEscolarID,
			Accion:        AccionConfigurarUmbral,
			ValorAnterior: anterior,
			ValorNuevo: map[string]any{
				"tipo":          umbral.Tipo,
				"valor":         umbral.Valor,
				"vigente_desde": fechas.FormatearISO(umbral.VigenteDesde),
			},
		})
	})
}

// UmbralesVigentes devuelve los umbrales de tardanzas y faltas vigentes en la fecha.
func (s *ServicioConfiguracion) UmbralesVigentes(ctx context.Context, anioID int64, fecha time.Time) ([]UmbralAlerta, error) {
	umbrales := make([]UmbralAlerta, 0, len(TiposUmbral()))
	for _, tipo := range TiposUmbral() {
		umbral, err := s.repositorio.UmbralVigente(ctx, anioID, tipo, fecha)
		if errors.Is(err, pgx.ErrNoRows) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("buscando el umbral vigente: %w", err)
		}
		umbrales = append(umbrales, umbral)
	}
	return umbrales, nil
}

// ListarUmbrales devuelve la historia de umbrales del año.
func (s *ServicioConfiguracion) ListarUmbrales(ctx context.Context, anioID int64) ([]UmbralAlerta, error) {
	umbrales, err := s.repositorio.ListarUmbralesPorAnio(ctx, anioID)
	if err != nil {
		return nil, fmt.Errorf("listando los umbrales del año: %w", err)
	}
	return umbrales, nil
}

// ListarExcepciones devuelve todas las excepciones del año.
func (s *ServicioConfiguracion) ListarExcepciones(ctx context.Context, anioID int64) ([]ExcepcionCalendario, error) {
	excepciones, err := s.repositorio.ListarExcepcionesPorAnio(ctx, anioID)
	if err != nil {
		return nil, fmt.Errorf("listando las excepciones del año: %w", err)
	}
	return excepciones, nil
}

// ExcepcionesVigentes devuelve las excepciones que cubren la fecha indicada.
func (s *ServicioConfiguracion) ExcepcionesVigentes(ctx context.Context, anioID int64, fecha time.Time) ([]ExcepcionCalendario, error) {
	excepciones, err := s.ListarExcepciones(ctx, anioID)
	if err != nil {
		return nil, err
	}
	vigentes := make([]ExcepcionCalendario, 0)
	for _, excepcion := range excepciones {
		if fechas.CompararDia(fecha, excepcion.FechaInicio) >= 0 &&
			fechas.CompararDia(fecha, excepcion.FechaFin) <= 0 {
			vigentes = append(vigentes, excepcion)
		}
	}
	return vigentes, nil
}

// ExcepcionesDeSeccion devuelve las excepciones que alcanzan una sección en la fecha.
func (s *ServicioConfiguracion) ExcepcionesDeSeccion(ctx context.Context, seccionID int64, fecha time.Time) ([]ExcepcionCalendario, error) {
	excepciones, err := s.repositorio.ExcepcionesDeSeccion(ctx, seccionID, fecha)
	if err != nil {
		return nil, fmt.Errorf("buscando las excepciones de la sección: %w", err)
	}
	return excepciones, nil
}

func cargarExcepciones(ctx context.Context, consultas *generado.Queries, anioID int64) ([]ExcepcionCalendario, error) {
	filas, err := consultas.ListarExcepcionesPorAnio(ctx, anioID)
	if err != nil {
		return nil, fmt.Errorf("listando las excepciones del año: %w", err)
	}
	grados, err := consultas.ListarGradosDeExcepciones(ctx, anioID)
	if err != nil {
		return nil, fmt.Errorf("listando los grados de las excepciones: %w", err)
	}
	porExcepcion := make(map[int64][]int64, len(filas))
	for _, fila := range grados {
		porExcepcion[fila.ExcepcionID] = append(porExcepcion[fila.ExcepcionID], fila.GradoID)
	}
	excepciones := make([]ExcepcionCalendario, 0, len(filas))
	for _, fila := range filas {
		excepciones = append(excepciones, ExcepcionCalendario{
			ID:            fila.ID,
			AnioEscolarID: fila.AnioEscolarID,
			Tipo:          fila.Tipo,
			FechaInicio:   fila.FechaInicio.Time,
			FechaFin:      fila.FechaFin.Time,
			Motivo:        fila.Motivo,
			Grados:        porExcepcion[fila.ID],
			CreadoPor:     fila.CreadoPor,
			CreadoEn:      fila.CreadoEn.Time,
		})
	}
	return excepciones, nil
}

func excepcionesSolapadas(existentes []ExcepcionCalendario, nueva ExcepcionCalendario) []ExcepcionCalendario {
	solapadas := make([]ExcepcionCalendario, 0)
	for _, existente := range existentes {
		if SeSolapan(existente, nueva) && MismoAlcance(existente, nueva) {
			solapadas = append(solapadas, existente)
		}
	}
	return solapadas
}
