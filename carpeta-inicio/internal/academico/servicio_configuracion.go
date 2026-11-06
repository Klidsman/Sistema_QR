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
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// codigoViolacionUnica es el código de PostgreSQL para un índice único duplicado.
const codigoViolacionUnica = "23505"

// Errores de negocio de la configuración académica.
var (
	ErrAnioActivoExistente = errors.New("ya existe un año escolar activo")
	ErrAnioNombreExistente = errors.New("ya existe un año escolar con ese nombre")
	ErrAnioNoEncontrado    = errors.New("el año escolar no existe")
	ErrAnioCerrado         = errors.New("el año escolar está cerrado y no admite cambios")
	ErrGradoNoEncontrado   = errors.New("el grado no existe")
	ErrGradoExistente      = errors.New("ya existe un grado con ese nombre en el año")
	ErrSeccionExistente    = errors.New("ya existe una sección con ese nombre en el grado")
)

// Reloj devuelve la fecha y hora actuales. Se inyecta en los servicios y en las pruebas.
type Reloj func() time.Time

// RelojDelSistema devuelve un reloj que usa la fecha de hoy en la zona del colegio.
func RelojDelSistema() Reloj {
	return fechas.Hoy
}

// DatosAnio son los datos para crear un año escolar.
type DatosAnio struct {
	Nombre      string
	FechaInicio time.Time
	FechaFin    time.Time
}

// DatosGrado son los datos para crear un grado.
type DatosGrado struct {
	AnioEscolarID int64
	Nombre        string
}

// DatosSeccion son los datos para crear una sección con su turno inicial.
type DatosSeccion struct {
	GradoID      int64
	Nombre       string
	Turno        string
	VigenteDesde time.Time
}

// ResumenConfiguracion es el estado que muestra el panel de configuración.
type ResumenConfiguracion struct {
	Anio      *AnioEscolar
	Grados    int
	Secciones int
}

// ServicioConfiguracion aplica las reglas de negocio de la configuración académica.
type ServicioConfiguracion struct {
	repositorio *RepositorioConfiguracion
	reloj       Reloj
}

// NuevoServicioConfiguracion crea el servicio sobre el pool indicado.
func NuevoServicioConfiguracion(pool *pgxpool.Pool) *ServicioConfiguracion {
	return &ServicioConfiguracion{
		repositorio: NuevoRepositorioConfiguracion(pool),
		reloj:       RelojDelSistema(),
	}
}

// Resumen devuelve el año activo y sus cantidades, o un resumen vacío si no hay año activo.
func (s *ServicioConfiguracion) Resumen(ctx context.Context) (ResumenConfiguracion, error) {
	anio, err := s.AnioActivo(ctx)
	if err != nil {
		return ResumenConfiguracion{}, err
	}
	if anio == nil {
		return ResumenConfiguracion{}, nil
	}
	grados, err := s.repositorio.ListarGradosPorAnio(ctx, anio.ID)
	if err != nil {
		return ResumenConfiguracion{}, fmt.Errorf("listando los grados: %w", err)
	}
	secciones, err := s.repositorio.ListarSeccionesPorAnio(ctx, anio.ID)
	if err != nil {
		return ResumenConfiguracion{}, fmt.Errorf("listando las secciones: %w", err)
	}
	return ResumenConfiguracion{
		Anio:      anio,
		Grados:    len(grados),
		Secciones: len(secciones),
	}, nil
}

// AnioActivo devuelve el año escolar activo, o nil si no existe.
func (s *ServicioConfiguracion) AnioActivo(ctx context.Context) (*AnioEscolar, error) {
	anio, err := s.repositorio.BuscarAnioActivo(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("buscando el año escolar activo: %w", err)
	}
	return &anio, nil
}

// ListarAnios devuelve todos los años escolares.
func (s *ServicioConfiguracion) ListarAnios(ctx context.Context) ([]AnioEscolar, error) {
	anios, err := s.repositorio.ListarAnios(ctx)
	if err != nil {
		return nil, fmt.Errorf("listando los años escolares: %w", err)
	}
	return anios, nil
}

// ListarGrados devuelve los grados del año indicado.
func (s *ServicioConfiguracion) ListarGrados(ctx context.Context, anioID int64) ([]Grado, error) {
	grados, err := s.repositorio.ListarGradosPorAnio(ctx, anioID)
	if err != nil {
		return nil, fmt.Errorf("listando los grados: %w", err)
	}
	return grados, nil
}

// ListarSecciones devuelve las secciones del año indicado.
func (s *ServicioConfiguracion) ListarSecciones(ctx context.Context, anioID int64) ([]Seccion, error) {
	secciones, err := s.repositorio.ListarSeccionesPorAnio(ctx, anioID)
	if err != nil {
		return nil, fmt.Errorf("listando las secciones: %w", err)
	}
	return secciones, nil
}

// ListarTurnosSeccion devuelve la historia de turnos de una sección.
func (s *ServicioConfiguracion) ListarTurnosSeccion(ctx context.Context, seccionID int64) ([]TurnoSeccion, error) {
	turnos, err := s.repositorio.ListarTurnosSeccion(ctx, seccionID)
	if err != nil {
		return nil, fmt.Errorf("listando los turnos de la sección: %w", err)
	}
	return turnos, nil
}

// CrearAnio registra un año escolar activo y lo audita.
func (s *ServicioConfiguracion) CrearAnio(ctx context.Context, usuarioID int64, datos DatosAnio) (int64, error) {
	datos.Nombre = NormalizarNombre(datos.Nombre)
	if err := ValidarNombre(datos.Nombre); err != nil {
		return 0, err
	}
	if err := ValidarFechasAnio(datos.FechaInicio, datos.FechaFin); err != nil {
		return 0, err
	}

	var id int64
	err := s.repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		if _, err := consultas.BuscarAnioActivo(ctx); err == nil {
			return ErrAnioActivoExistente
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("comprobando el año activo: %w", err)
		}

		creado, err := consultas.CrearAnioEscolar(ctx, generado.CrearAnioEscolarParams{
			Nombre:      datos.Nombre,
			FechaInicio: parametroFecha(datos.FechaInicio),
			FechaFin:    parametroFecha(datos.FechaFin),
		})
		if err != nil {
			if esViolacionUnica(err) {
				return ErrAnioNombreExistente
			}
			return fmt.Errorf("creando el año escolar: %w", err)
		}
		if err := auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID: usuarioID,
			Entidad:   EntidadAnioEscolar,
			EntidadID: creado,
			Accion:    AccionCrear,
			ValorNuevo: map[string]any{
				"nombre":       datos.Nombre,
				"fecha_inicio": fechas.FormatearISO(datos.FechaInicio),
				"fecha_fin":    fechas.FormatearISO(datos.FechaFin),
				"estado":       EstadoAnioActivo,
			},
		}); err != nil {
			return err
		}
		id = creado
		return nil
	})
	return id, err
}

// CerrarAnio cierra el año escolar y lo deja de solo lectura.
func (s *ServicioConfiguracion) CerrarAnio(ctx context.Context, usuarioID, anioID int64) error {
	return s.repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		anio, err := consultas.BuscarAnioEscolarPorID(ctx, anioID)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrAnioNoEncontrado
		}
		if err != nil {
			return fmt.Errorf("buscando el año escolar: %w", err)
		}
		if anio.Estado == EstadoAnioCerrado {
			return ErrAnioCerrado
		}
		filas, err := consultas.CerrarAnioEscolar(ctx, anioID)
		if err != nil {
			return fmt.Errorf("cerrando el año escolar: %w", err)
		}
		if filas == 0 {
			return ErrAnioCerrado
		}
		return auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID:     usuarioID,
			Entidad:       EntidadAnioEscolar,
			EntidadID:     anioID,
			Accion:        AccionCerrar,
			ValorAnterior: map[string]any{"estado": anio.Estado},
			ValorNuevo:    map[string]any{"estado": EstadoAnioCerrado},
		})
	})
}

// CrearGrado registra un grado en el año indicado.
func (s *ServicioConfiguracion) CrearGrado(ctx context.Context, usuarioID int64, datos DatosGrado) (int64, error) {
	datos.Nombre = NormalizarNombre(datos.Nombre)
	if err := ValidarNombre(datos.Nombre); err != nil {
		return 0, err
	}

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

		if _, err := consultas.BuscarGradoPorNombre(ctx, generado.BuscarGradoPorNombreParams{
			AnioEscolarID: datos.AnioEscolarID,
			Nombre:        datos.Nombre,
		}); err == nil {
			return ErrGradoExistente
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("comprobando el grado: %w", err)
		}

		creado, err := consultas.CrearGrado(ctx, generado.CrearGradoParams{
			AnioEscolarID: datos.AnioEscolarID,
			Nombre:        datos.Nombre,
		})
		if err != nil {
			if esViolacionUnica(err) {
				return ErrGradoExistente
			}
			return fmt.Errorf("creando el grado: %w", err)
		}
		if err := auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID:  usuarioID,
			Entidad:    EntidadGrado,
			EntidadID:  creado,
			Accion:     AccionCrear,
			ValorNuevo: map[string]any{"anio_escolar_id": datos.AnioEscolarID, "nombre": datos.Nombre},
		}); err != nil {
			return err
		}
		id = creado
		return nil
	})
	return id, err
}

// CrearSeccion registra una sección con su turno inicial.
func (s *ServicioConfiguracion) CrearSeccion(ctx context.Context, usuarioID int64, datos DatosSeccion) (int64, error) {
	datos.Nombre = NormalizarNombre(datos.Nombre)
	if err := ValidarNombreSeccion(datos.Nombre); err != nil {
		return 0, err
	}
	if err := ValidarTurno(datos.Turno); err != nil {
		return 0, err
	}
	if err := ValidarVigencia(datos.VigenteDesde, s.reloj()); err != nil {
		return 0, err
	}

	var id int64
	err := s.repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		grado, err := consultas.BuscarGradoPorID(ctx, datos.GradoID)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrGradoNoEncontrado
		}
		if err != nil {
			return fmt.Errorf("buscando el grado: %w", err)
		}
		if grado.AnioEstado == EstadoAnioCerrado {
			return ErrAnioCerrado
		}

		if _, err := consultas.BuscarSeccionPorNombre(ctx, generado.BuscarSeccionPorNombreParams{
			GradoID: datos.GradoID,
			Nombre:  datos.Nombre,
		}); err == nil {
			return ErrSeccionExistente
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("comprobando la sección: %w", err)
		}

		creado, err := consultas.CrearSeccion(ctx, generado.CrearSeccionParams{
			GradoID: datos.GradoID,
			Nombre:  datos.Nombre,
		})
		if err != nil {
			if esViolacionUnica(err) {
				return ErrSeccionExistente
			}
			return fmt.Errorf("creando la sección: %w", err)
		}
		if _, err := consultas.CrearTurnoSeccion(ctx, generado.CrearTurnoSeccionParams{
			SeccionID:    creado,
			Turno:        datos.Turno,
			VigenteDesde: parametroFecha(datos.VigenteDesde),
		}); err != nil {
			return fmt.Errorf("asignando el turno de la sección: %w", err)
		}
		if err := auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID: usuarioID,
			Entidad:   EntidadSeccion,
			EntidadID: creado,
			Accion:    AccionCrear,
			ValorNuevo: map[string]any{
				"grado_id":      datos.GradoID,
				"nombre":        datos.Nombre,
				"turno":         datos.Turno,
				"vigente_desde": fechas.FormatearISO(datos.VigenteDesde),
			},
		}); err != nil {
			return err
		}
		id = creado
		return nil
	})
	return id, err
}

// CambiarTurno registra un turno nuevo para una sección desde una fecha.
func (s *ServicioConfiguracion) CambiarTurno(ctx context.Context, usuarioID, seccionID int64, turno string, vigenteDesde time.Time) error {
	if err := ValidarTurno(turno); err != nil {
		return err
	}
	if err := ValidarVigencia(vigenteDesde, s.reloj()); err != nil {
		return err
	}

	return s.repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		gradoID, err := consultas.GradoDeSeccion(ctx, seccionID)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrGradoNoEncontrado
		}
		if err != nil {
			return fmt.Errorf("buscando el grado de la sección: %w", err)
		}
		grado, err := consultas.BuscarGradoPorID(ctx, gradoID)
		if err != nil {
			return fmt.Errorf("buscando el grado: %w", err)
		}
		if grado.AnioEstado == EstadoAnioCerrado {
			return ErrAnioCerrado
		}

		anterior := ""
		turnoAnterior, err := consultas.TurnoVigenteSeccion(ctx, generado.TurnoVigenteSeccionParams{
			SeccionID:    seccionID,
			VigenteDesde: parametroFecha(s.reloj()),
		})
		if err == nil {
			anterior = turnoAnterior
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("buscando el turno anterior: %w", err)
		}

		if _, err := consultas.CrearTurnoSeccion(ctx, generado.CrearTurnoSeccionParams{
			SeccionID:    seccionID,
			Turno:        turno,
			VigenteDesde: parametroFecha(vigenteDesde),
		}); err != nil {
			return fmt.Errorf("cambiando el turno de la sección: %w", err)
		}
		return auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID:     usuarioID,
			Entidad:       EntidadTurno,
			EntidadID:     seccionID,
			Accion:        AccionCambiarTurno,
			ValorAnterior: map[string]any{"turno": anterior},
			ValorNuevo: map[string]any{
				"turno":         turno,
				"vigente_desde": fechas.FormatearISO(vigenteDesde),
			},
		})
	})
}

func esViolacionUnica(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == codigoViolacionUnica
}

// parametroFecha convierte una fecha a su parámetro de base de datos.
func parametroFecha(fecha time.Time) pgtype.Date {
	return pgtype.Date{Time: fecha, Valid: true}
}
