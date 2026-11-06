package academico

import (
	"context"
	"errors"
	"sort"
	"time"

	"asistencia/internal/plataforma/fechas"

	"github.com/jackc/pgx/v5"
)

// AnioCubreFecha indica si la fecha está dentro del rango del año escolar.
func AnioCubreFecha(anio AnioEscolar, fecha time.Time) bool {
	return fechas.CompararDia(fecha, anio.FechaInicio) >= 0 &&
		fechas.CompararDia(fecha, anio.FechaFin) <= 0
}

// VersionVigente elige la versión más reciente cuya fecha de vigencia no supere la fecha.
// Devuelve el valor y verdadero si encontró alguna versión aplicable.
func VersionVigente[T any](versiones []T, fecha time.Time, vigenteDesde func(T) time.Time) (T, bool) {
	var elegida T
	encontrada := false
	for _, version := range versiones {
		inicio := vigenteDesde(version)
		if fechas.CompararDia(inicio, fecha) > 0 {
			continue
		}
		if !encontrada || fechas.CompararDia(inicio, vigenteDesde(elegida)) > 0 {
			elegida = version
			encontrada = true
		}
	}
	return elegida, encontrada
}

// ConfiguracionEnFecha devuelve la configuración vigente en una fecha determinada.
func (s *ServicioConfiguracion) ConfiguracionEnFecha(ctx context.Context, fecha time.Time) (ConfiguracionFecha, error) {
	configuracion := ConfiguracionFecha{Fecha: fecha}

	anio, err := s.repositorio.AnioQueCubreFecha(ctx, fecha)
	if errors.Is(err, pgx.ErrNoRows) {
		return configuracion, nil
	}
	if err != nil {
		return ConfiguracionFecha{}, err
	}
	if !AnioCubreFecha(anio, fecha) {
		return configuracion, nil
	}
	configuracion.Anio = &anio

	for _, turno := range Turnos() {
		historial, err := s.repositorio.ListarHorariosTurno(ctx, turno)
		if err != nil {
			return ConfiguracionFecha{}, err
		}
		if horario, aplica := VersionVigente(historial, fecha, func(h HorarioTurno) time.Time {
			return h.VigenteDesde
		}); aplica {
			configuracion.Horarios = append(configuracion.Horarios, horario)
		}

		bloquesTurno, err := s.repositorio.ListarBloquesTurno(ctx, turno)
		if err != nil {
			return ConfiguracionFecha{}, err
		}
		configuracion.Bloques = append(configuracion.Bloques, bloquesVigentes(bloquesTurno, fecha)...)
	}

	excepciones, err := s.ExcepcionesVigentes(ctx, anio.ID, fecha)
	if err != nil {
		return ConfiguracionFecha{}, err
	}
	configuracion.Excepciones = excepciones

	umbrales, err := s.repositorio.ListarUmbralesPorAnio(ctx, anio.ID)
	if err != nil {
		return ConfiguracionFecha{}, err
	}
	for _, tipo := range TiposUmbral() {
		delTipo := make([]UmbralAlerta, 0)
		for _, umbral := range umbrales {
			if umbral.Tipo == tipo {
				delTipo = append(delTipo, umbral)
			}
		}
		if umbral, aplica := VersionVigente(delTipo, fecha, func(u UmbralAlerta) time.Time {
			return u.VigenteDesde
		}); aplica {
			configuracion.Umbrales = append(configuracion.Umbrales, umbral)
		}
	}

	secciones, err := s.repositorio.ListarSeccionesPorAnio(ctx, anio.ID)
	if err != nil {
		return ConfiguracionFecha{}, err
	}
	for i := range secciones {
		turnos, err := s.repositorio.ListarTurnosSeccion(ctx, secciones[i].ID)
		if err != nil {
			return ConfiguracionFecha{}, err
		}
		if turno, aplica := VersionVigente(turnos, fecha, func(t TurnoSeccion) time.Time {
			return t.VigenteDesde
		}); aplica {
			secciones[i].Turno = turno.Turno
		} else {
			secciones[i].Turno = ""
		}
	}
	configuracion.Secciones = secciones

	return configuracion, nil
}

func bloquesVigentes(bloques []BloqueClase, fecha time.Time) []BloqueClase {
	porNumero := make(map[int][]BloqueClase)
	for _, bloque := range bloques {
		porNumero[bloque.Numero] = append(porNumero[bloque.Numero], bloque)
	}
	vigentes := make([]BloqueClase, 0, len(porNumero))
	for _, versiones := range porNumero {
		if bloque, aplica := VersionVigente(versiones, fecha, func(b BloqueClase) time.Time {
			return b.VigenteDesde
		}); aplica {
			vigentes = append(vigentes, bloque)
		}
	}
	sort.Slice(vigentes, func(i, j int) bool {
		if vigentes[i].Turno != vigentes[j].Turno {
			return vigentes[i].Turno < vigentes[j].Turno
		}
		return vigentes[i].Numero < vigentes[j].Numero
	})
	return vigentes
}
