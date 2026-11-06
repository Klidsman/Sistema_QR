package asistencia

import (
	"context"
	"time"

	"asistencia/internal/alertas"
)

// MarcacionesDelRango implementa la fuente de eventos para alertas.
func (s *Servicio) MarcacionesDelRango(ctx context.Context, desde, hasta time.Time) ([]alertas.EventoMarcacion, error) {
	marcaciones, err := s.repositorio.ListarMarcacionesDelRango(ctx, desde, hasta)
	if err != nil {
		return nil, err
	}
	eventos := make([]alertas.EventoMarcacion, 0, len(marcaciones))
	for _, marcacion := range marcaciones {
		eventos = append(eventos, alertas.EventoMarcacion{
			MarcacionID:   marcacion.ID,
			AlumnoID:      marcacion.AlumnoID,
			Fecha:         marcacion.Fecha,
			Hora:          marcacion.Hora,
			Estado:        marcacion.Estado,
			RegistradaPor: marcacion.RegistradaPor,
		})
	}
	return eventos, nil
}

// ContarTardanzas implementa el conteo mensual para alertas.
func (s *Servicio) ContarTardanzas(ctx context.Context, alumnoID int64, desde, hasta time.Time) (int, error) {
	return s.repositorio.ContarTardanzas(ctx, alumnoID, desde, hasta)
}

// TieneMarcacion indica si el alumno tiene ingreso oficial en la fecha.
func (s *Servicio) TieneMarcacion(ctx context.Context, alumnoID int64, fecha time.Time) (bool, error) {
	return s.repositorio.TieneMarcacion(ctx, alumnoID, fecha)
}

// DocentesDeSeccionEnFecha devuelve los destinatarios docentes de una sección.
func (s *Servicio) DocentesDeSeccionEnFecha(ctx context.Context, seccionID int64, fecha time.Time) ([]int64, error) {
	return s.repositorio.ListarDocentesDeSeccionFecha(ctx, seccionID, fecha)
}
