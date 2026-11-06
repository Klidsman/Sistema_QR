package asistencia

import (
	"context"
	"errors"
	"fmt"
	"time"

	"asistencia/internal/academico"
	"asistencia/internal/alumnos"
	"asistencia/internal/asistencia/generado"
	"asistencia/internal/plataforma/auditoria"
	"asistencia/internal/plataforma/fechas"

	"github.com/jackc/pgx/v5/pgtype"
)

// AlumnoClase reúne el alumno, su registro docente y su marcación QR del día.
type AlumnoClase struct {
	Alumno     alumnos.AlumnoEnClase
	Registro   *RegistroDocente
	Marcacion  *Marcacion
	Diferencia bool
}

// ClaseDetalle es la lista completa de una hora de clase.
type ClaseDetalle struct {
	Clase   academico.ClaseProgramada
	Fecha   time.Time
	Alumnos []AlumnoClase
}

// CambioRegistro es el estado que el docente envía para un alumno.
type CambioRegistro struct {
	AlumnoID    int64
	Estado      string
	Observacion string
}

// ListarClase devuelve la lista de alumnos con su registro y su marcación QR.
func (s *Servicio) ListarClase(ctx context.Context, usuarioID int64, administrador bool, fecha time.Time, seccionID, cursoID int64, bloque int) (ClaseDetalle, error) {
	if err := ValidarFechaNoFutura(fecha, s.hoy()); err != nil {
		return ClaseDetalle{}, err
	}
	if !administrador {
		asignado, err := s.repositorio.DocenteAsignadoAClase(ctx, usuarioID, cursoID, seccionID, fecha)
		if err != nil {
			return ClaseDetalle{}, fmt.Errorf("comprobando la asignación: %w", err)
		}
		if !asignado {
			return ClaseDetalle{}, ErrDocenteSinAsignacion
		}
	}
	clase, err := s.academico.ResolverClase(ctx, seccionID, cursoID, bloque, fecha)
	if err != nil {
		return ClaseDetalle{}, mapearErrorAcademico(err)
	}
	noLaborable, err := s.academico.EsDiaNoLaborable(ctx, seccionID, fecha)
	if err != nil {
		return ClaseDetalle{}, err
	}
	if noLaborable {
		return ClaseDetalle{}, ErrDiaNoLaborable
	}

	alumnosClase, err := s.alumnos.ListarAlumnosDeSeccionEnFecha(ctx, seccionID, fecha)
	if err != nil {
		return ClaseDetalle{}, err
	}
	registros, err := s.repositorio.ListarRegistrosClase(ctx, fecha, seccionID, cursoID, bloque)
	if err != nil {
		return ClaseDetalle{}, fmt.Errorf("listando los registros de la clase: %w", err)
	}
	ids := make([]int64, 0, len(alumnosClase))
	for _, alumno := range alumnosClase {
		ids = append(ids, alumno.ID)
	}
	marcaciones, err := s.repositorio.ListarMarcacionesPorFechaYAlumnos(ctx, fecha, ids)
	if err != nil {
		return ClaseDetalle{}, fmt.Errorf("listando las marcaciones: %w", err)
	}

	porAlumno := make(map[int64]*RegistroDocente, len(registros))
	for i := range registros {
		porAlumno[registros[i].AlumnoID] = &registros[i]
	}
	marcacionPorAlumno := make(map[int64]*Marcacion, len(marcaciones))
	for i := range marcaciones {
		marcacionPorAlumno[marcaciones[i].AlumnoID] = &marcaciones[i]
	}

	detalle := ClaseDetalle{Clase: clase, Fecha: fecha, Alumnos: make([]AlumnoClase, 0, len(alumnosClase))}
	for _, alumno := range alumnosClase {
		registro := porAlumno[alumno.ID]
		marcacion := marcacionPorAlumno[alumno.ID]
		diferencia := false
		if registro != nil {
			diferencia = HayDiferencia(marcacion, registro.Estado)
		}
		detalle.Alumnos = append(detalle.Alumnos, AlumnoClase{
			Alumno:     alumno,
			Registro:   registro,
			Marcacion:  marcacion,
			Diferencia: diferencia,
		})
	}
	return detalle, nil
}

// GuardarRegistros guarda los estados de los alumnos de una clase.
func (s *Servicio) GuardarRegistros(ctx context.Context, usuarioID int64, administrador bool, fecha time.Time, seccionID, cursoID int64, bloque int, cambios []CambioRegistro) error {
	detalle, err := s.ListarClase(ctx, usuarioID, administrador, fecha, seccionID, cursoID, bloque)
	if err != nil {
		return err
	}
	permitidos := make(map[int64]bool, len(detalle.Alumnos))
	existentes := make(map[int64]*RegistroDocente, len(detalle.Alumnos))
	for _, alumno := range detalle.Alumnos {
		permitidos[alumno.Alumno.ID] = true
		existentes[alumno.Alumno.ID] = alumno.Registro
	}

	plazo, err := s.repositorio.ObtenerParametros(ctx)
	if err != nil {
		return fmt.Errorf("leyendo el plazo de edición: %w", err)
	}

	for _, cambio := range cambios {
		if !permitidos[cambio.AlumnoID] {
			return ErrAlumnoFueraDeSeccion
		}
		if !EstadoValido(cambio.Estado) {
			return ErrEstadoInvalido
		}
		if err := ValidarObservacion(cambio.Observacion); err != nil {
			return err
		}
		if existente := existentes[cambio.AlumnoID]; existente != nil && !administrador {
			if !DentroDelPlazo(existente.CreadoEn, s.hoy(), plazo.PlazoEdicionHoras) {
				return ErrPlazoVencido
			}
		}
	}

	return s.repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		if _, err := db.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtext($1))",
			ClaveClase(fecha, seccionID, cursoID, bloque)); err != nil {
			return fmt.Errorf("bloqueando la clase: %w", err)
		}
		for _, cambio := range cambios {
			existente := existentes[cambio.AlumnoID]
			if existente != nil && existente.Estado == cambio.Estado && existente.Observacion == cambio.Observacion {
				continue
			}
			_, err := consultas.GuardarRegistro(ctx, generado.GuardarRegistroParams{
				AlumnoID:     cambio.AlumnoID,
				Fecha:        fechaParametro(fecha),
				CursoID:      cursoID,
				SeccionID:    seccionID,
				BloqueNumero: int32(bloque),
				Estado:       cambio.Estado,
				Observacion:  textoParametro(cambio.Observacion),
				DocenteID:    usuarioID,
			})
			if err != nil {
				return fmt.Errorf("guardando el registro del alumno: %w", err)
			}
			accion := AccionCorregirRegistro
			valorAnterior := any(nil)
			if existente != nil {
				valorAnterior = mapaRegistro(*existente)
			} else {
				accion = AccionCrearRegistro
			}
			if err := auditoria.Registrar(ctx, db, auditoria.Evento{
				UsuarioID:     usuarioID,
				Entidad:       EntidadRegistroDocente,
				EntidadID:     cambio.AlumnoID,
				Accion:        accion,
				ValorAnterior: valorAnterior,
				ValorNuevo: map[string]any{
					"fecha":         fechas.FormatearISO(fecha),
					"curso_id":      cursoID,
					"seccion_id":    seccionID,
					"bloque_numero": bloque,
					"estado":        cambio.Estado,
					"observacion":   cambio.Observacion,
				},
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

// MarcarTodosPresentes crea un presente para los alumnos que aún no tienen registro.
func (s *Servicio) MarcarTodosPresentes(ctx context.Context, usuarioID int64, administrador bool, fecha time.Time, seccionID, cursoID int64, bloque int) (int, error) {
	detalle, err := s.ListarClase(ctx, usuarioID, administrador, fecha, seccionID, cursoID, bloque)
	if err != nil {
		return 0, err
	}
	if !administrador {
		if err := s.verificarPlazoNuevos(ctx, detalle); err != nil {
			return 0, err
		}
	}

	creados := 0
	err = s.repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		if _, err := db.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtext($1))",
			ClaveClase(fecha, seccionID, cursoID, bloque)); err != nil {
			return fmt.Errorf("bloqueando la clase: %w", err)
		}
		for _, alumno := range detalle.Alumnos {
			if alumno.Registro != nil {
				continue
			}
			filas, err := consultas.InsertarRegistroSiFalta(ctx, generado.InsertarRegistroSiFaltaParams{
				AlumnoID:     alumno.Alumno.ID,
				Fecha:        fechaParametro(fecha),
				CursoID:      cursoID,
				SeccionID:    seccionID,
				BloqueNumero: int32(bloque),
				DocenteID:    usuarioID,
			})
			if err != nil {
				return fmt.Errorf("marcando al alumno como presente: %w", err)
			}
			creados += int(filas)
		}
		if creados > 0 {
			return auditoria.Registrar(ctx, db, auditoria.Evento{
				UsuarioID: usuarioID,
				Entidad:   EntidadRegistroDocente,
				EntidadID: seccionID,
				Accion:    AccionMarcarTodos,
				ValorNuevo: map[string]any{
					"fecha":         fechas.FormatearISO(fecha),
					"curso_id":      cursoID,
					"seccion_id":    seccionID,
					"bloque_numero": bloque,
					"creados":       creados,
				},
			})
		}
		return nil
	})
	return creados, err
}

func (s *Servicio) verificarPlazoNuevos(ctx context.Context, detalle ClaseDetalle) error {
	plazo, err := s.repositorio.ObtenerParametros(ctx)
	if err != nil {
		return err
	}
	for _, alumno := range detalle.Alumnos {
		if alumno.Registro != nil && !DentroDelPlazo(alumno.Registro.CreadoEn, s.hoy(), plazo.PlazoEdicionHoras) {
			return ErrPlazoVencido
		}
	}
	return nil
}

func mapearErrorAcademico(err error) error {
	switch {
	case errors.Is(err, academico.ErrClaseNoProgramada):
		return ErrClaseNoProgramada
	case errors.Is(err, academico.ErrBloqueNoEncontrado):
		return ErrBloqueNoEncontrado
	case errors.Is(err, academico.ErrSeccionSinTurno), errors.Is(err, academico.ErrSeccionNoEncontrada):
		return ErrClaseNoProgramada
	case errors.Is(err, academico.ErrCursoNoVinculado):
		return ErrCursoNoVinculado
	default:
		return err
	}
}

func mapaRegistro(registro RegistroDocente) map[string]any {
	return map[string]any{
		"alumno_id":     registro.AlumnoID,
		"fecha":         fechas.FormatearISO(registro.Fecha),
		"estado":        registro.Estado,
		"observacion":   registro.Observacion,
		"docente_id":    registro.DocenteID,
		"bloque_numero": registro.BloqueNumero,
	}
}

func textoParametro(texto string) pgtype.Text {
	return pgtype.Text{String: texto, Valid: texto != ""}
}
