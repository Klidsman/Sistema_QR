package asistencia

import (
	"context"
	"time"

	"asistencia/internal/asistencia/generado"

	"github.com/jackc/pgx/v5/pgtype"
)

// ListarMarcacionesPorFechaYAlumnos devuelve las marcaciones del día para los alumnos indicados.
func (r *Repositorio) ListarMarcacionesPorFechaYAlumnos(ctx context.Context, fecha time.Time, alumnoIDs []int64) ([]Marcacion, error) {
	filas, err := r.consultas.ListarMarcacionesPorFechaYAlumnos(ctx, generado.ListarMarcacionesPorFechaYAlumnosParams{
		Fecha:   pgtype.Date{Time: fecha, Valid: true},
		Column2: alumnoIDs,
	})
	if err != nil {
		return nil, err
	}
	marcaciones := make([]Marcacion, 0, len(filas))
	for _, fila := range filas {
		marcaciones = append(marcaciones, Marcacion{
			ID:       fila.ID,
			AlumnoID: fila.AlumnoID,
			Fecha:    fila.Fecha.Time,
			Hora:     fila.Hora.Time,
			Estado:   fila.Estado,
			CreadaEn: fila.CreadaEn.Time,
		})
	}
	return marcaciones, nil
}
