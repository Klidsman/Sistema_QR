package academico

import (
	"context"
	"errors"
	"fmt"
	"time"

	"asistencia/internal/academico/generado"
	"asistencia/internal/plataforma/auditoria"
	"asistencia/internal/plataforma/fechas"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// GrillaHorario reúne los datos para mostrar el horario semanal de una sección.
type GrillaHorario struct {
	Seccion Seccion
	Turno   string
	Bloques []BloqueClase
	Cursos  []Curso
	Celdas  []HorarioSeccion
}

// CambioCelda es una celda del horario que el usuario quiere guardar.
type CambioCelda struct {
	DiaSemana    int
	BloqueNumero int
	CursoID      *int64
}

// ObtenerGrilla devuelve el horario semanal vigente de una sección en una fecha.
func (s *ServicioConfiguracion) ObtenerGrilla(ctx context.Context, seccionID int64, fecha time.Time) (GrillaHorario, error) {
	seccion, err := s.repositorio.BuscarSeccionPorID(ctx, seccionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return GrillaHorario{}, ErrSeccionNoEncontrada
	}
	if err != nil {
		return GrillaHorario{}, fmt.Errorf("buscando la sección: %w", err)
	}

	turno, err := s.repositorio.TurnoVigenteSeccion(ctx, seccionID, fecha)
	if errors.Is(err, pgx.ErrNoRows) {
		return GrillaHorario{}, ErrSeccionSinTurno
	}
	if err != nil {
		return GrillaHorario{}, fmt.Errorf("buscando el turno de la sección: %w", err)
	}

	bloques, err := s.repositorio.BloquesVigentesTurno(ctx, turno, fecha)
	if err != nil {
		return GrillaHorario{}, fmt.Errorf("listando los bloques del turno: %w", err)
	}
	cursos, err := s.repositorio.ListarCursosPorAnio(ctx, seccion.AnioEscolarID)
	if err != nil {
		return GrillaHorario{}, fmt.Errorf("listando los cursos: %w", err)
	}
	celdas, err := s.repositorio.HorarioSeccionVigente(ctx, seccionID, fecha)
	if err != nil {
		return GrillaHorario{}, fmt.Errorf("listando el horario de la sección: %w", err)
	}

	return GrillaHorario{
		Seccion: seccion,
		Turno:   turno,
		Bloques: bloques,
		Cursos:  cursos,
		Celdas:  celdas,
	}, nil
}

// GuardarCeldas guarda los cambios del horario semanal y los audita.
func (s *ServicioConfiguracion) GuardarCeldas(ctx context.Context, usuarioID, seccionID int64, fecha time.Time, cambios []CambioCelda, confirmar bool) error {
	if err := ValidarVigencia(fecha, s.reloj()); err != nil {
		return err
	}
	if len(cambios) == 0 {
		return nil
	}

	return s.repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		if _, err := db.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtext($1))",
			fmt.Sprintf("horario_%d", seccionID)); err != nil {
			return fmt.Errorf("bloqueando el horario de la sección: %w", err)
		}

		seccion, err := consultas.BuscarSeccionPorID(ctx, seccionID)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrSeccionNoEncontrada
		}
		if err != nil {
			return fmt.Errorf("buscando la sección: %w", err)
		}
		anio, err := consultas.BuscarAnioEscolarPorID(ctx, seccion.AnioEscolarID)
		if err != nil {
			return fmt.Errorf("buscando el año escolar: %w", err)
		}
		if anio.Estado == EstadoAnioCerrado {
			return ErrAnioCerrado
		}

		turno, err := consultas.TurnoVigenteSeccion(ctx, generado.TurnoVigenteSeccionParams{
			SeccionID:    seccionID,
			VigenteDesde: parametroFecha(fecha),
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrSeccionSinTurno
		}
		if err != nil {
			return fmt.Errorf("buscando el turno de la sección: %w", err)
		}

		filasBloques, err := consultas.BloquesVigentesTurno(ctx, generado.BloquesVigentesTurnoParams{
			Turno:        turno,
			VigenteDesde: parametroFecha(fecha),
		})
		if err != nil {
			return fmt.Errorf("listando los bloques del turno: %w", err)
		}
		bloquesValidos := make(map[int]bool, len(filasBloques))
		for _, fila := range filasBloques {
			bloquesValidos[int(fila.Numero)] = true
		}

		aplicados := make([]CambioCelda, 0, len(cambios))
		for _, cambio := range cambios {
			if err := ValidarDiaSemana(cambio.DiaSemana); err != nil {
				return err
			}
			if !bloquesValidos[cambio.BloqueNumero] {
				return ErrBloqueNoEncontrado
			}
			if cambio.CursoID != nil {
				vinculado, err := consultas.CursoVinculadoAGrado(ctx, generado.CursoVinculadoAGradoParams{
					ID:      *cambio.CursoID,
					GradoID: seccion.GradoID,
				})
				if err != nil {
					return fmt.Errorf("comprobando el curso de la celda: %w", err)
				}
				if !vinculado {
					return ErrCursoNoVinculado
				}
			}

			actual, err := consultas.HorarioSeccionCelda(ctx, generado.HorarioSeccionCeldaParams{
				SeccionID:    seccionID,
				DiaSemana:    int16(cambio.DiaSemana),
				BloqueNumero: int32(cambio.BloqueNumero),
				VigenteDesde: parametroFecha(fecha),
			})
			if err == nil {
				if mismoCurso(actual.CursoID, cambio.CursoID) {
					continue
				}
				if actual.CursoID.Valid && !confirmar {
					return ErrRequiereConfirmacion
				}
			} else if !errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("buscando la celda del horario: %w", err)
			}

			var cursoID pgtype.Int8
			if cambio.CursoID != nil {
				cursoID = pgtype.Int8{Int64: *cambio.CursoID, Valid: true}
			}
			if err := consultas.UpsertHorarioSeccion(ctx, generado.UpsertHorarioSeccionParams{
				SeccionID:    seccionID,
				DiaSemana:    int16(cambio.DiaSemana),
				BloqueNumero: int32(cambio.BloqueNumero),
				CursoID:      cursoID,
				VigenteDesde: parametroFecha(fecha),
			}); err != nil {
				return fmt.Errorf("guardando la celda del horario: %w", err)
			}
			aplicados = append(aplicados, cambio)
		}

		if len(aplicados) == 0 {
			return nil
		}
		return auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID: usuarioID,
			Entidad:   EntidadHorarioSeccion,
			EntidadID: seccionID,
			Accion:    AccionAsignarHorario,
			ValorNuevo: map[string]any{
				"vigente_desde": fechas.FormatearISO(fecha),
				"celdas":        cambiosParaAuditoria(aplicados),
			},
		})
	})
}

// CursoDeBloque devuelve el curso vigente de una celda, o nil si no tiene clase.
func (s *ServicioConfiguracion) CursoDeBloque(ctx context.Context, seccionID int64, dia, bloque int, fecha time.Time) (*int64, error) {
	celda, err := s.repositorio.HorarioSeccionCelda(ctx, seccionID, dia, bloque, fecha)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("buscando la celda del horario: %w", err)
	}
	return celda.CursoID, nil
}

func mismoCurso(actual pgtype.Int8, nuevo *int64) bool {
	if !actual.Valid && nuevo == nil {
		return true
	}
	if actual.Valid && nuevo != nil {
		return actual.Int64 == *nuevo
	}
	return false
}

func cambiosParaAuditoria(cambios []CambioCelda) []map[string]any {
	resultado := make([]map[string]any, 0, len(cambios))
	for _, cambio := range cambios {
		curso := any(nil)
		if cambio.CursoID != nil {
			curso = *cambio.CursoID
		}
		resultado = append(resultado, map[string]any{
			"dia_semana":    cambio.DiaSemana,
			"bloque_numero": cambio.BloqueNumero,
			"curso_id":      curso,
		})
	}
	return resultado
}
