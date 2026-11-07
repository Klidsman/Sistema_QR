package academico

import (
	"context"
	"errors"
	"fmt"

	"asistencia/internal/academico/generado"
	"asistencia/internal/plataforma/auditoria"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// codigoViolacionForanea es el código de PostgreSQL para una referencia inexistente.
const codigoViolacionForanea = "23503"

// EliminarGrado borra un grado sin secciones, cursos ni excepciones asociados.
func (s *ServicioConfiguracion) EliminarGrado(ctx context.Context, usuarioID, gradoID int64) error {
	grado, estado, err := s.repositorio.BuscarGradoPorID(ctx, gradoID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrGradoNoEncontrado
	}
	if err != nil {
		return fmt.Errorf("buscando el grado: %w", err)
	}
	if estado == EstadoAnioCerrado {
		return ErrAnioCerrado
	}

	return s.repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		filas, err := consultas.EliminarGrado(ctx, gradoID)
		if err != nil {
			if esViolacionForanea(err) {
				return ErrGradoConDependencias
			}
			return fmt.Errorf("eliminando el grado: %w", err)
		}
		if filas == 0 {
			return ErrGradoNoEncontrado
		}
		return auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID:     usuarioID,
			Entidad:       EntidadGrado,
			EntidadID:     gradoID,
			Accion:        AccionEliminar,
			ValorAnterior: map[string]any{"nombre": grado.Nombre, "anio_escolar_id": grado.AnioEscolarID},
		})
	})
}

// EliminarSeccion borra una sección sin alumnos, asignaciones ni registros asociados.
func (s *ServicioConfiguracion) EliminarSeccion(ctx context.Context, usuarioID, seccionID int64) error {
	seccion, err := s.repositorio.BuscarSeccionPorID(ctx, seccionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrSeccionNoEncontrada
	}
	if err != nil {
		return fmt.Errorf("buscando la sección: %w", err)
	}
	anio, err := s.repositorio.BuscarAnioPorID(ctx, seccion.AnioEscolarID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrAnioNoEncontrado
	}
	if err != nil {
		return fmt.Errorf("buscando el año escolar de la sección: %w", err)
	}
	if anio.Estado == EstadoAnioCerrado {
		return ErrAnioCerrado
	}

	return s.repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		if err := consultas.EliminarHorarioSeccion(ctx, seccionID); err != nil {
			return fmt.Errorf("eliminando el horario de la sección: %w", err)
		}
		if err := consultas.EliminarTurnosSeccion(ctx, seccionID); err != nil {
			return fmt.Errorf("eliminando los turnos de la sección: %w", err)
		}
		filas, err := consultas.EliminarSeccion(ctx, seccionID)
		if err != nil {
			if esViolacionForanea(err) {
				return ErrSeccionConDependencias
			}
			return fmt.Errorf("eliminando la sección: %w", err)
		}
		if filas == 0 {
			return ErrSeccionNoEncontrada
		}
		return auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID:     usuarioID,
			Entidad:       EntidadSeccion,
			EntidadID:     seccionID,
			Accion:        AccionEliminar,
			ValorAnterior: map[string]any{"nombre": seccion.Nombre, "grado_id": seccion.GradoID},
		})
	})
}

func esViolacionForanea(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == codigoViolacionForanea
}
