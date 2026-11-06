package alumnos

import (
	"context"
	"fmt"
	"time"

	"asistencia/internal/alumnos/generado"

	"github.com/jackc/pgx/v5/pgtype"
)

// ListarAlumnosDeSeccionEnFecha devuelve los alumnos activos de la sección en la fecha.
func (s *Servicio) ListarAlumnosDeSeccionEnFecha(ctx context.Context, seccionID int64, fecha time.Time) ([]AlumnoEnClase, error) {
	filas, err := s.repositorio.consultas.ListarAlumnosDeSeccionEnFecha(ctx, generado.ListarAlumnosDeSeccionEnFechaParams{
		SeccionID:   seccionID,
		FechaInicio: pgtype.Date{Time: fecha, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("listando los alumnos de la sección: %w", err)
	}
	alumnos := make([]AlumnoEnClase, 0, len(filas))
	for _, fila := range filas {
		alumnos = append(alumnos, AlumnoEnClase{
			ID:            fila.ID,
			Identificador: fila.Identificador,
			Nombres:       fila.Nombres,
			Apellidos:     fila.Apellidos,
		})
	}
	return alumnos, nil
}

// AlumnosPorIDs devuelve los datos mínimos de los alumnos indicados.
func (s *Servicio) AlumnosPorIDs(ctx context.Context, ids []int64) ([]AlumnoEnClase, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	filas, err := s.repositorio.consultas.ListarAlumnosPorIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("listando los alumnos: %w", err)
	}
	alumnos := make([]AlumnoEnClase, 0, len(filas))
	for _, fila := range filas {
		alumnos = append(alumnos, AlumnoEnClase{
			ID:            fila.ID,
			Identificador: fila.Identificador,
			Nombres:       fila.Nombres,
			Apellidos:     fila.Apellidos,
		})
	}
	return alumnos, nil
}
