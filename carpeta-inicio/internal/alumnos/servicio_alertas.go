package alumnos

import (
	"context"
	"errors"
	"time"

	"asistencia/internal/alertas"
	"asistencia/internal/alumnos/generado"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// AlumnosActivosDeSeccionEnFecha implementa la fuente de alumnos para alertas.
func (s *Servicio) AlumnosActivosDeSeccionEnFecha(ctx context.Context, seccionID int64, fecha time.Time) ([]alertas.AlumnoEvento, error) {
	alumnos, err := s.ListarAlumnosDeSeccionEnFecha(ctx, seccionID, fecha)
	if err != nil {
		return nil, err
	}
	eventos := make([]alertas.AlumnoEvento, 0, len(alumnos))
	for _, alumno := range alumnos {
		eventos = append(eventos, alertas.AlumnoEvento{
			ID:            alumno.ID,
			Identificador: alumno.Identificador,
			Nombres:       alumno.Nombres,
			Apellidos:     alumno.Apellidos,
		})
	}
	return eventos, nil
}

// AlumnoPorID devuelve los datos mínimos de un alumno para las alertas.
func (s *Servicio) AlumnoPorID(ctx context.Context, alumnoID int64) (alertas.AlumnoEvento, error) {
	alumno, err := s.Obtener(ctx, alumnoID)
	if err != nil {
		return alertas.AlumnoEvento{}, err
	}
	return alertas.AlumnoEvento{
		ID:            alumno.ID,
		Identificador: alumno.Identificador,
		Nombres:       alumno.Nombres,
		Apellidos:     alumno.Apellidos,
	}, nil
}

// SeccionDeAlumnoEnFecha devuelve la sección del alumno activo en la fecha, o cero.
func (s *Servicio) SeccionDeAlumnoEnFecha(ctx context.Context, alumnoID int64, fecha time.Time) (int64, error) {
	seccionID, err := s.repositorio.consultas.SeccionDeAlumnoEnFecha(ctx, generado.SeccionDeAlumnoEnFechaParams{
		AlumnoID:    alumnoID,
		FechaInicio: pgtype.Date{Time: fecha, Valid: true},
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return seccionID, nil
}
