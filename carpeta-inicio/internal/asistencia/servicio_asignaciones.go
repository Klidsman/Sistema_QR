package asistencia

import (
	"context"
	"errors"
	"fmt"
	"time"

	"asistencia/internal/academico"
	"asistencia/internal/asistencia/generado"
	"asistencia/internal/autenticacion"
	"asistencia/internal/plataforma/auditoria"
	"asistencia/internal/plataforma/fechas"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// DatosAsignacion son los datos para crear una asignación docente-curso-sección.
type DatosAsignacion struct {
	DocenteID   int64
	CursoID     int64
	SeccionID   int64
	FechaInicio time.Time
	FechaFin    *time.Time
	Titular     bool
}

// AsignacionVista añade nombres a una asignación para mostrarla.
type AsignacionVista struct {
	AsignacionDocente
	Docente string
	Curso   string
	Seccion string
	Vigente bool
}

// CrearAsignacion registra una asignación y la audita.
func (s *Servicio) CrearAsignacion(ctx context.Context, usuarioID int64, datos DatosAsignacion) (int64, error) {
	if err := ValidarFechasAsignacion(datos.FechaInicio, datos.FechaFin); err != nil {
		return 0, err
	}
	activo, rol, err := s.docentes.DocenteActivo(ctx, datos.DocenteID)
	if err != nil {
		return 0, fmt.Errorf("consultando el docente: %w", err)
	}
	if !activo || rol != "docente" {
		return 0, ErrDocenteNoEncontrado
	}
	vinculado, err := s.academico.CursoVinculadoASeccion(ctx, datos.CursoID, datos.SeccionID)
	if err != nil {
		return 0, fmt.Errorf("comprobando el curso de la sección: %w", err)
	}
	if !vinculado {
		return 0, ErrCursoNoVinculado
	}

	var id int64
	err = s.repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		creado, err := consultas.CrearAsignacion(ctx, generado.CrearAsignacionParams{
			DocenteID:   datos.DocenteID,
			CursoID:     datos.CursoID,
			SeccionID:   datos.SeccionID,
			FechaInicio: fechaParametro(datos.FechaInicio),
			FechaFin:    fechaParametroOpcional(datos.FechaFin),
			Titular:     datos.Titular,
			CreadoPor:   usuarioID,
		})
		if err != nil {
			return fmt.Errorf("creando la asignación: %w", err)
		}
		if err := auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID: usuarioID,
			Entidad:   EntidadAsignacionDocente,
			EntidadID: creado,
			Accion:    AccionCrearAsignacion,
			ValorNuevo: map[string]any{
				"docente_id":   datos.DocenteID,
				"curso_id":     datos.CursoID,
				"seccion_id":   datos.SeccionID,
				"fecha_inicio": fechas.FormatearISO(datos.FechaInicio),
				"fecha_fin":    fechaFinTexto(datos.FechaFin),
				"titular":      datos.Titular,
			},
		}); err != nil {
			return err
		}
		id = creado
		return nil
	})
	return id, err
}

// CerrarAsignacion fija la fecha de fin de una asignación.
func (s *Servicio) CerrarAsignacion(ctx context.Context, usuarioID, asignacionID int64, fecha time.Time) error {
	asignacion, err := s.repositorio.BuscarAsignacionPorID(ctx, asignacionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrAsignacionNoEncontrada
	}
	if err != nil {
		return fmt.Errorf("buscando la asignación: %w", err)
	}
	if fechas.CompararDia(fecha, asignacion.FechaInicio) < 0 {
		return ErrAsignacionFueraDeRango
	}

	return s.repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		filas, err := consultas.CerrarAsignacion(ctx, generado.CerrarAsignacionParams{
			ID:       asignacionID,
			FechaFin: fechaParametro(fecha),
		})
		if err != nil {
			return fmt.Errorf("cerrando la asignación: %w", err)
		}
		if filas == 0 {
			return ErrAsignacionNoEncontrada
		}
		return auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID:     usuarioID,
			Entidad:       EntidadAsignacionDocente,
			EntidadID:     asignacionID,
			Accion:        AccionCerrarAsignacion,
			ValorAnterior: map[string]any{"fecha_fin": fechaFinTexto(asignacion.FechaFin)},
			ValorNuevo:    map[string]any{"fecha_fin": fechas.FormatearISO(fecha)},
		})
	})
}

// ListarAsignaciones devuelve las asignaciones con nombres.
func (s *Servicio) ListarAsignaciones(ctx context.Context) ([]AsignacionVista, error) {
	asignaciones, err := s.repositorio.ListarAsignaciones(ctx)
	if err != nil {
		return nil, fmt.Errorf("listando las asignaciones: %w", err)
	}
	hoy := s.hoy()
	vistas := make([]AsignacionVista, 0, len(asignaciones))
	for _, asignacion := range asignaciones {
		vista, err := s.asignacionVista(ctx, asignacion, hoy)
		if err != nil {
			return nil, err
		}
		vistas = append(vistas, vista)
	}
	return vistas, nil
}

// ClasesAsignadas devuelve las horas de clase que el docente puede abrir en la fecha.
func (s *Servicio) ClasesAsignadas(ctx context.Context, docenteID int64, fecha time.Time) ([]ClaseAsignada, error) {
	if err := ValidarFechaNoFutura(fecha, s.hoy()); err != nil {
		return nil, err
	}
	asignaciones, err := s.repositorio.ListarAsignacionesPorDocenteFecha(ctx, docenteID, fecha)
	if err != nil {
		return nil, fmt.Errorf("listando las asignaciones del docente: %w", err)
	}
	clases := make([]ClaseAsignada, 0)
	for _, asignacion := range asignaciones {
		noLaborable, err := s.academico.EsDiaNoLaborable(ctx, asignacion.SeccionID, fecha)
		if err != nil {
			return nil, err
		}
		if noLaborable {
			continue
		}
		bloques, err := s.academico.HorarioDelDia(ctx, asignacion.SeccionID, fecha)
		if err != nil {
			return nil, err
		}
		seccion, err := s.academico.EtiquetaSeccion(ctx, asignacion.SeccionID)
		if err != nil {
			return nil, err
		}
		curso, err := s.academico.NombreCurso(ctx, asignacion.CursoID)
		if err != nil {
			return nil, err
		}
		for _, bloque := range bloques {
			if bloque.CursoID != asignacion.CursoID {
				continue
			}
			clases = append(clases, ClaseAsignada{
				AsignacionDocente: asignacion,
				Curso:             curso,
				Seccion:           seccion,
				BloqueNumero:      bloque.Bloque,
				HoraInicio:        bloque.HoraInicio,
				HoraFin:           bloque.HoraFin,
				Fecha:             fecha,
			})
		}
	}
	return clases, nil
}

// DocentesDisponibles devuelve las cuentas de docente para asignar.
func (s *Servicio) DocentesDisponibles(ctx context.Context) ([]autenticacion.Docente, error) {
	return s.docentes.ListarDocentes(ctx)
}

// SeccionesDisponibles devuelve las secciones del año activo.
func (s *Servicio) SeccionesDisponibles(ctx context.Context) ([]academico.Seccion, error) {
	return s.academico.ListarSeccionesActivas(ctx)
}

// CursosDisponibles devuelve los cursos activos del año.
func (s *Servicio) CursosDisponibles(ctx context.Context) ([]academico.Curso, error) {
	return s.academico.ListarCursosActivos(ctx)
}

func (s *Servicio) asignacionVista(ctx context.Context, asignacion AsignacionDocente, hoy time.Time) (AsignacionVista, error) {
	docente, err := s.docentes.NombreDocente(ctx, asignacion.DocenteID)
	if err != nil {
		return AsignacionVista{}, err
	}
	curso, err := s.academico.NombreCurso(ctx, asignacion.CursoID)
	if err != nil {
		return AsignacionVista{}, err
	}
	seccion, err := s.academico.EtiquetaSeccion(ctx, asignacion.SeccionID)
	if err != nil {
		return AsignacionVista{}, err
	}
	vigente := AsignacionVigenteEnFecha(asignacion, hoy)
	return AsignacionVista{
		AsignacionDocente: asignacion,
		Docente:           docente,
		Curso:             curso,
		Seccion:           seccion,
		Vigente:           vigente,
	}, nil
}

func fechaParametro(fecha time.Time) pgtype.Date {
	return pgtype.Date{Time: fecha, Valid: true}
}

func fechaParametroOpcional(fecha *time.Time) pgtype.Date {
	if fecha == nil {
		return pgtype.Date{}
	}
	return pgtype.Date{Time: *fecha, Valid: true}
}

func fechaFinTexto(fecha *time.Time) string {
	if fecha == nil {
		return ""
	}
	return fechas.FormatearISO(*fecha)
}
