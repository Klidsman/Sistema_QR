package academico

import (
	"context"
	"errors"
	"fmt"

	"asistencia/internal/academico/generado"
	"asistencia/internal/plataforma/auditoria"

	"github.com/jackc/pgx/v5"
)

// Errores de negocio de cursos y horarios.
var (
	ErrCursoExistente       = errors.New("ya existe un curso con ese nombre en el año")
	ErrCursoNoEncontrado    = errors.New("el curso no existe")
	ErrCursoNoVinculado     = errors.New("el curso no está vinculado al grado de la sección")
	ErrSeccionSinTurno      = errors.New("la sección no tiene un turno vigente en la fecha")
	ErrRequiereConfirmacion = errors.New("la celda ya tiene un curso asignado; confirme el reemplazo")
)

// DatosCurso son los datos para crear un curso y sus grados.
type DatosCurso struct {
	AnioEscolarID int64
	Nombre        string
	Grados        []int64
}

// CrearCurso registra un curso vinculado a uno o más grados y lo audita.
func (s *ServicioConfiguracion) CrearCurso(ctx context.Context, usuarioID int64, datos DatosCurso) (int64, error) {
	datos.Nombre = NormalizarNombre(datos.Nombre)
	if err := ValidarNombre(datos.Nombre); err != nil {
		return 0, err
	}
	grados := gradosUnicos(datos.Grados)
	if err := ValidarGradosCurso(grados); err != nil {
		return 0, err
	}

	var id int64
	err := s.repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		anio, err := consultas.BuscarAnioEscolarPorID(ctx, datos.AnioEscolarID)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrAnioNoEncontrado
		}
		if err != nil {
			return fmt.Errorf("buscando el año escolar: %w", err)
		}
		if anio.Estado == EstadoAnioCerrado {
			return ErrAnioCerrado
		}

		existentes, err := consultas.ListarGradosPorAnio(ctx, datos.AnioEscolarID)
		if err != nil {
			return fmt.Errorf("listando los grados del año: %w", err)
		}
		permitidos := make(map[int64]bool, len(existentes))
		for _, grado := range existentes {
			permitidos[grado.ID] = true
		}
		for _, gradoID := range grados {
			if !permitidos[gradoID] {
				return ErrGradoNoEncontrado
			}
		}

		if _, err := consultas.BuscarCursoPorNombre(ctx, generado.BuscarCursoPorNombreParams{
			AnioEscolarID: datos.AnioEscolarID,
			Lower:         datos.Nombre,
		}); err == nil {
			return ErrCursoExistente
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("comprobando el curso: %w", err)
		}

		creado, err := consultas.CrearCurso(ctx, generado.CrearCursoParams{
			AnioEscolarID: datos.AnioEscolarID,
			Nombre:        datos.Nombre,
		})
		if err != nil {
			if esViolacionUnica(err) {
				return ErrCursoExistente
			}
			return fmt.Errorf("creando el curso: %w", err)
		}
		for _, gradoID := range grados {
			if err := consultas.CrearCursoGrado(ctx, generado.CrearCursoGradoParams{
				CursoID: creado,
				GradoID: gradoID,
			}); err != nil {
				return fmt.Errorf("vinculando el curso con el grado: %w", err)
			}
		}
		if err := auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID: usuarioID,
			Entidad:   EntidadCurso,
			EntidadID: creado,
			Accion:    AccionCrearCurso,
			ValorNuevo: map[string]any{
				"anio_escolar_id": datos.AnioEscolarID,
				"nombre":          datos.Nombre,
				"grados":          grados,
			},
		}); err != nil {
			return err
		}
		id = creado
		return nil
	})
	return id, err
}

// DesactivarCurso desactiva un curso sin borrar su historial.
func (s *ServicioConfiguracion) DesactivarCurso(ctx context.Context, usuarioID, cursoID int64) error {
	return s.repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		curso, err := consultas.BuscarCursoPorID(ctx, cursoID)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrCursoNoEncontrado
		}
		if err != nil {
			return fmt.Errorf("buscando el curso: %w", err)
		}
		anio, err := consultas.BuscarAnioEscolarPorID(ctx, curso.AnioEscolarID)
		if err != nil {
			return fmt.Errorf("buscando el año escolar: %w", err)
		}
		if anio.Estado == EstadoAnioCerrado {
			return ErrAnioCerrado
		}
		if !curso.Activo {
			return nil
		}
		if _, err := consultas.DesactivarCurso(ctx, cursoID); err != nil {
			return fmt.Errorf("desactivando el curso: %w", err)
		}
		return auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID:     usuarioID,
			Entidad:       EntidadCurso,
			EntidadID:     cursoID,
			Accion:        AccionDesactivarCurso,
			ValorAnterior: map[string]any{"activo": true},
			ValorNuevo:    map[string]any{"activo": false},
		})
	})
}

// ListarCursos devuelve los cursos del año con sus grados.
func (s *ServicioConfiguracion) ListarCursos(ctx context.Context, anioID int64) ([]Curso, error) {
	cursos, err := s.repositorio.ListarCursosPorAnio(ctx, anioID)
	if err != nil {
		return nil, fmt.Errorf("listando los cursos: %w", err)
	}
	return cursos, nil
}

func gradosUnicos(grados []int64) []int64 {
	vistos := make(map[int64]bool, len(grados))
	unicos := make([]int64, 0, len(grados))
	for _, gradoID := range grados {
		if !vistos[gradoID] {
			vistos[gradoID] = true
			unicos = append(unicos, gradoID)
		}
	}
	return unicos
}
