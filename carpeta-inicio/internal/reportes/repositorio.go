package reportes

import (
	"context"
	"time"

	"asistencia/internal/reportes/generado"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repositorio consulta los registros de asistencia para los reportes.
type Repositorio struct {
	consultas *generado.Queries
}

// NuevoRepositorio crea el repositorio.
func NuevoRepositorio(pool *pgxpool.Pool) *Repositorio {
	return &Repositorio{consultas: generado.New(pool)}
}

// AlumnosEnFecha devuelve los alumnos con sección vigente en la fecha.
func (r *Repositorio) AlumnosEnFecha(ctx context.Context, fecha time.Time) ([]generado.AlumnosEnFechaRow, error) {
	return r.consultas.AlumnosEnFecha(ctx, fechaParametro(fecha))
}

// MarcacionesDelDia devuelve las marcaciones oficiales del día.
func (r *Repositorio) MarcacionesDelDia(ctx context.Context, fecha time.Time) ([]generado.MarcacionesDelDiaRow, error) {
	return r.consultas.MarcacionesDelDia(ctx, fechaParametro(fecha))
}

// MarcacionesDelRango devuelve las marcaciones del periodo.
func (r *Repositorio) MarcacionesDelRango(ctx context.Context, desde, hasta time.Time) ([]generado.MarcacionesDelRangoRow, error) {
	return r.consultas.MarcacionesDelRango(ctx, generado.MarcacionesDelRangoParams{
		Fecha:   fechaParametro(desde),
		Fecha_2: fechaParametro(hasta),
	})
}

// JustificacionesAprobadas devuelve las fechas justificadas aprobadas del periodo.
func (r *Repositorio) JustificacionesAprobadas(ctx context.Context, desde, hasta time.Time) ([]generado.JustificacionesAprobadasDelRangoRow, error) {
	return r.consultas.JustificacionesAprobadasDelRango(ctx, generado.JustificacionesAprobadasDelRangoParams{
		FechaEvento:   fechaParametro(desde),
		FechaEvento_2: fechaParametro(hasta),
	})
}

// AlumnosConHistorial devuelve las secciones vigentes en el periodo.
func (r *Repositorio) AlumnosConHistorial(ctx context.Context, desde, hasta time.Time) ([]generado.AlumnosConHistorialEnRangoRow, error) {
	return r.consultas.AlumnosConHistorialEnRango(ctx, generado.AlumnosConHistorialEnRangoParams{
		FechaFin: fechaParametro(desde),
		Column2:  fechaParametro(hasta),
	})
}

// Excepciones devuelve feriados y suspensiones del periodo.
func (r *Repositorio) Excepciones(ctx context.Context, desde, hasta time.Time) ([]generado.ExcepcionesDelRangoRow, error) {
	return r.consultas.ExcepcionesDelRango(ctx, generado.ExcepcionesDelRangoParams{
		FechaFin:    fechaParametro(desde),
		FechaInicio: fechaParametro(hasta),
	})
}

// SeccionesDelDocente devuelve las secciones asignadas en el periodo.
func (r *Repositorio) SeccionesDelDocente(ctx context.Context, docenteID int64, desde, hasta time.Time) ([]int64, error) {
	return r.consultas.SeccionesDelDocenteEnRango(ctx, generado.SeccionesDelDocenteEnRangoParams{
		DocenteID:   docenteID,
		FechaFin:    fechaParametro(desde),
		FechaInicio: fechaParametro(hasta),
	})
}

func fechaParametro(fecha time.Time) pgtype.Date {
	return pgtype.Date{Time: fecha, Valid: true}
}
