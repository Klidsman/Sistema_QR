package asistencia

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"asistencia/internal/asistencia/generado"
	"asistencia/internal/plataforma/auditoria"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// ConsultaRegistro es un registro enriquecido para la consulta del director y del administrador.
type ConsultaRegistro struct {
	Registro   RegistroDocente
	Alumno     string
	Curso      string
	Seccion    string
	Docente    string
	Marcacion  *Marcacion
	Diferencia bool
}

// PlazoEdicion devuelve el plazo configurado.
func (s *Servicio) PlazoEdicion(ctx context.Context) (ParametrosAsistencia, error) {
	return s.repositorio.ObtenerParametros(ctx)
}

// ConfigurarPlazo guarda el plazo de edición y lo audita.
func (s *Servicio) ConfigurarPlazo(ctx context.Context, usuarioID int64, horas int) error {
	if err := ValidarPlazoHoras(horas); err != nil {
		return err
	}
	return s.repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		anterior, err := consultas.ObtenerParametrosAsistencia(ctx)
		if err != nil {
			return fmt.Errorf("leyendo el plazo anterior: %w", err)
		}
		if _, err := consultas.ActualizarPlazoEdicion(ctx, generado.ActualizarPlazoEdicionParams{
			PlazoEdicionHoras: int32(horas),
			ActualizadoPor:    pgtypeInt8(usuarioID),
		}); err != nil {
			return fmt.Errorf("guardando el plazo de edición: %w", err)
		}
		return auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID:     usuarioID,
			Entidad:       EntidadParametroAsistencia,
			EntidadID:     1,
			Accion:        AccionConfigurarPlazo,
			ValorAnterior: map[string]any{"plazo_edicion_horas": int(anterior.PlazoEdicionHoras)},
			ValorNuevo:    map[string]any{"plazo_edicion_horas": horas},
		})
	})
}

// CorregirRegistro permite al administrador corregir un registro con motivo obligatorio.
func (s *Servicio) CorregirRegistro(ctx context.Context, usuarioID, registroID int64, estado, observacion, motivo string) error {
	if !EstadoValido(estado) {
		return ErrEstadoInvalido
	}
	if err := ValidarObservacion(observacion); err != nil {
		return err
	}
	motivo = strings.TrimSpace(motivo)
	if motivo == "" || len([]rune(motivo)) > 300 {
		return ErrMotivoRequerido
	}

	return s.repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		anterior, err := consultas.BuscarRegistroPorID(ctx, registroID)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrRegistroNoEncontrado
		}
		if err != nil {
			return fmt.Errorf("buscando el registro: %w", err)
		}
		if _, err := consultas.CorregirRegistro(ctx, generado.CorregirRegistroParams{
			ID:          registroID,
			Estado:      estado,
			Observacion: textoParametro(observacion),
			DocenteID:   usuarioID,
		}); err != nil {
			return fmt.Errorf("corrigiendo el registro: %w", err)
		}
		return auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID: usuarioID,
			Entidad:   EntidadRegistroDocente,
			EntidadID: registroID,
			Accion:    AccionCorregirRegistro,
			ValorAnterior: map[string]any{
				"estado":      anterior.Estado,
				"observacion": anterior.Observacion.String,
			},
			ValorNuevo: map[string]any{
				"estado":      estado,
				"observacion": observacion,
			},
			Motivo: motivo,
		})
	})
}

// ConsultaRegistros devuelve los registros de una sección en una fecha con datos para mostrar.
func (s *Servicio) ConsultaRegistros(ctx context.Context, fecha time.Time, seccionID int64) ([]ConsultaRegistro, error) {
	registros, err := s.repositorio.ListarRegistrosConsulta(ctx, fecha, seccionID)
	if err != nil {
		return nil, fmt.Errorf("listando los registros: %w", err)
	}
	if len(registros) == 0 {
		return nil, nil
	}
	ids := make([]int64, 0, len(registros))
	for _, registro := range registros {
		ids = append(ids, registro.AlumnoID)
	}
	alumnos, err := s.alumnos.AlumnosPorIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	nombresAlumno := make(map[int64]string, len(alumnos))
	for _, alumno := range alumnos {
		nombresAlumno[alumno.ID] = alumno.NombreCompleto()
	}
	marcaciones, err := s.repositorio.ListarMarcacionesPorFechaYAlumnos(ctx, fecha, ids)
	if err != nil {
		return nil, fmt.Errorf("listando las marcaciones: %w", err)
	}
	marcacionPorAlumno := make(map[int64]*Marcacion, len(marcaciones))
	for i := range marcaciones {
		marcacionPorAlumno[marcaciones[i].AlumnoID] = &marcaciones[i]
	}
	seccion, err := s.academico.EtiquetaSeccion(ctx, seccionID)
	if err != nil {
		return nil, err
	}

	resultado := make([]ConsultaRegistro, 0, len(registros))
	cursos := make(map[int64]string)
	docentes := make(map[int64]string)
	for _, registro := range registros {
		curso, existe := cursos[registro.CursoID]
		if !existe {
			curso, err = s.academico.NombreCurso(ctx, registro.CursoID)
			if err != nil {
				return nil, err
			}
			cursos[registro.CursoID] = curso
		}
		docente, existe := docentes[registro.DocenteID]
		if !existe {
			docente, err = s.docentes.NombreDocente(ctx, registro.DocenteID)
			if err != nil {
				return nil, err
			}
			docentes[registro.DocenteID] = docente
		}
		marcacion := marcacionPorAlumno[registro.AlumnoID]
		resultado = append(resultado, ConsultaRegistro{
			Registro:   registro,
			Alumno:     nombresAlumno[registro.AlumnoID],
			Curso:      curso,
			Seccion:    seccion,
			Docente:    docente,
			Marcacion:  marcacion,
			Diferencia: HayDiferencia(marcacion, registro.Estado),
		})
	}
	return resultado, nil
}

func pgtypeInt8(valor int64) pgtype.Int8 {
	return pgtype.Int8{Int64: valor, Valid: true}
}
