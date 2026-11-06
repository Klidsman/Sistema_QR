package asistencia

import (
	"context"
	"errors"
	"time"

	"asistencia/internal/asistencia/generado"

	"github.com/jackc/pgx/v5"
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

// ListarMarcacionesDelRango devuelve todas las marcaciones del rango.
func (r *Repositorio) ListarMarcacionesDelRango(ctx context.Context, desde, hasta time.Time) ([]Marcacion, error) {
	filas, err := r.consultas.ListarMarcacionesDelRango(ctx, generado.ListarMarcacionesDelRangoParams{
		Fecha:   pgtype.Date{Time: desde, Valid: true},
		Fecha_2: pgtype.Date{Time: hasta, Valid: true},
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

// ContarTardanzas devuelve las tardanzas del alumno en el rango.
func (r *Repositorio) ContarTardanzas(ctx context.Context, alumnoID int64, desde, hasta time.Time) (int, error) {
	cantidad, err := r.consultas.ContarTardanzasAlumnoRango(ctx, generado.ContarTardanzasAlumnoRangoParams{
		AlumnoID: alumnoID,
		Fecha:    pgtype.Date{Time: desde, Valid: true},
		Fecha_2:  pgtype.Date{Time: hasta, Valid: true},
	})
	return int(cantidad), err
}

// TieneMarcacion indica si el alumno tiene marcación oficial ese día.
func (r *Repositorio) TieneMarcacion(ctx context.Context, alumnoID int64, fecha time.Time) (bool, error) {
	return r.consultas.ExisteMarcacion(ctx, generado.ExisteMarcacionParams{
		AlumnoID: alumnoID,
		Fecha:    pgtype.Date{Time: fecha, Valid: true},
	})
}

// ListarMarcacionesPorRangoYAlumnos devuelve las marcaciones del rango para los alumnos.
func (r *Repositorio) ListarMarcacionesPorRangoYAlumnos(ctx context.Context, desde, hasta time.Time, alumnoIDs []int64) ([]Marcacion, error) {
	filas, err := r.consultas.ListarMarcacionesPorRangoYAlumnos(ctx, generado.ListarMarcacionesPorRangoYAlumnosParams{
		Fecha:   pgtype.Date{Time: desde, Valid: true},
		Fecha_2: pgtype.Date{Time: hasta, Valid: true},
		Column3: alumnoIDs,
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

// MarcacionDeAlumnoFecha devuelve la marcación del día, si existe.
func (r *Repositorio) MarcacionDeAlumnoFecha(ctx context.Context, alumnoID int64, fecha time.Time) (Marcacion, bool, error) {
	fila, err := r.consultas.MarcacionDeAlumnoFecha(ctx, generado.MarcacionDeAlumnoFechaParams{
		AlumnoID: alumnoID,
		Fecha:    pgtype.Date{Time: fecha, Valid: true},
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Marcacion{}, false, nil
	}
	if err != nil {
		return Marcacion{}, false, err
	}
	return Marcacion{
		ID:          fila.ID,
		AlumnoID:    fila.AlumnoID,
		Fecha:       fila.Fecha.Time,
		Hora:        fila.Hora.Time,
		Estado:      fila.Estado,
		Justificada: fila.Justificada,
	}, true, nil
}

// MarcarMarcacionJustificada cambia la marca de justificación de una tardanza.
func (r *Repositorio) MarcarMarcacionJustificada(ctx context.Context, alumnoID int64, fecha time.Time, justificada bool) error {
	return r.consultas.MarcarMarcacionJustificada(ctx, generado.MarcarMarcacionJustificadaParams{
		AlumnoID:    alumnoID,
		Fecha:       pgtype.Date{Time: fecha, Valid: true},
		Justificada: justificada,
	})
}
