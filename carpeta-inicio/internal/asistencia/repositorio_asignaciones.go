package asistencia

import (
	"context"
	"time"

	"asistencia/internal/asistencia/generado"

	"github.com/jackc/pgx/v5/pgtype"
)

// CrearAsignacion registra una asignación de docente.
func (r *Repositorio) CrearAsignacion(ctx context.Context, asignacion AsignacionDocente) (int64, error) {
	return r.consultas.CrearAsignacion(ctx, generado.CrearAsignacionParams{
		DocenteID:   asignacion.DocenteID,
		CursoID:     asignacion.CursoID,
		SeccionID:   asignacion.SeccionID,
		FechaInicio: pgtype.Date{Time: asignacion.FechaInicio, Valid: true},
		FechaFin:    fechaOpcional(asignacion.FechaFin),
		Titular:     asignacion.Titular,
		CreadoPor:   asignacion.CreadoPor,
	})
}

// BuscarAsignacionPorID devuelve la asignación con ese identificador.
func (r *Repositorio) BuscarAsignacionPorID(ctx context.Context, id int64) (AsignacionDocente, error) {
	fila, err := r.consultas.BuscarAsignacionPorID(ctx, id)
	if err != nil {
		return AsignacionDocente{}, err
	}
	return aAsignacion(fila), nil
}

// CerrarAsignacion fija la fecha de fin y devuelve falso si ya estaba cerrada.
func (r *Repositorio) CerrarAsignacion(ctx context.Context, id int64, fecha time.Time) (bool, error) {
	filas, err := r.consultas.CerrarAsignacion(ctx, generado.CerrarAsignacionParams{
		ID:       id,
		FechaFin: pgtype.Date{Time: fecha, Valid: true},
	})
	if err != nil {
		return false, err
	}
	return filas > 0, nil
}

// ListarAsignacionesPorDocenteFecha devuelve las asignaciones vigentes del docente.
func (r *Repositorio) ListarAsignacionesPorDocenteFecha(ctx context.Context, docenteID int64, fecha time.Time) ([]AsignacionDocente, error) {
	filas, err := r.consultas.ListarAsignacionesPorDocenteFecha(ctx, generado.ListarAsignacionesPorDocenteFechaParams{
		DocenteID:   docenteID,
		FechaInicio: pgtype.Date{Time: fecha, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return aAsignaciones(filas), nil
}

// ListarAsignaciones devuelve todas las asignaciones.
func (r *Repositorio) ListarAsignaciones(ctx context.Context) ([]AsignacionDocente, error) {
	filas, err := r.consultas.ListarAsignaciones(ctx)
	if err != nil {
		return nil, err
	}
	return aAsignaciones(filas), nil
}

// ListarDocentesDeSeccionFecha devuelve los docentes asignados a la sección en la fecha.
func (r *Repositorio) ListarDocentesDeSeccionFecha(ctx context.Context, seccionID int64, fecha time.Time) ([]int64, error) {
	return r.consultas.ListarDocentesDeSeccionFecha(ctx, generado.ListarDocentesDeSeccionFechaParams{
		SeccionID:   seccionID,
		FechaInicio: pgtype.Date{Time: fecha, Valid: true},
	})
}

// DocenteAsignadoAClase indica si el docente puede registrar esa clase en la fecha.
func (r *Repositorio) DocenteAsignadoAClase(ctx context.Context, docenteID, cursoID, seccionID int64, fecha time.Time) (bool, error) {
	return r.consultas.DocenteAsignadoAClase(ctx, generado.DocenteAsignadoAClaseParams{
		DocenteID:   docenteID,
		CursoID:     cursoID,
		SeccionID:   seccionID,
		FechaInicio: pgtype.Date{Time: fecha, Valid: true},
	})
}

func aAsignacion(fila generado.AsignacionesDocente) AsignacionDocente {
	asignacion := AsignacionDocente{
		ID:          fila.ID,
		DocenteID:   fila.DocenteID,
		CursoID:     fila.CursoID,
		SeccionID:   fila.SeccionID,
		FechaInicio: fila.FechaInicio.Time,
		Titular:     fila.Titular,
		CreadoPor:   fila.CreadoPor,
		CreadoEn:    fila.CreadoEn.Time,
	}
	if fila.FechaFin.Valid {
		fin := fila.FechaFin.Time
		asignacion.FechaFin = &fin
	}
	return asignacion
}

func aAsignaciones(filas []generado.AsignacionesDocente) []AsignacionDocente {
	asignaciones := make([]AsignacionDocente, 0, len(filas))
	for _, fila := range filas {
		asignaciones = append(asignaciones, aAsignacion(fila))
	}
	return asignaciones
}

func fechaOpcional(fecha *time.Time) pgtype.Date {
	if fecha == nil {
		return pgtype.Date{}
	}
	return pgtype.Date{Time: *fecha, Valid: true}
}

// DocenteAsignadoASeccionFecha indica si el docente tiene asignación vigente a la sección.
func (r *Repositorio) DocenteAsignadoASeccionFecha(ctx context.Context, docenteID, seccionID int64, fecha time.Time) (bool, error) {
	return r.consultas.DocenteAsignadoASeccionFecha(ctx, generado.DocenteAsignadoASeccionFechaParams{
		DocenteID:   docenteID,
		SeccionID:   seccionID,
		FechaInicio: pgtype.Date{Time: fecha, Valid: true},
	})
}
