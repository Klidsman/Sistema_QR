package academico

import (
	"context"

	"asistencia/internal/academico/generado"
)

// CrearCurso registra un curso y lo vincula a sus grados.
func (r *RepositorioConfiguracion) CrearCurso(ctx context.Context, curso Curso) (int64, error) {
	var id int64
	err := r.EnTransaccion(ctx, func(consultas *generado.Queries, _ generado.DBTX) error {
		creado, err := consultas.CrearCurso(ctx, generado.CrearCursoParams{
			AnioEscolarID: curso.AnioEscolarID,
			Nombre:        curso.Nombre,
		})
		if err != nil {
			return err
		}
		for _, gradoID := range curso.Grados {
			if err := consultas.CrearCursoGrado(ctx, generado.CrearCursoGradoParams{
				CursoID: creado,
				GradoID: gradoID,
			}); err != nil {
				return err
			}
		}
		id = creado
		return nil
	})
	return id, err
}

// BuscarCursoPorID devuelve el curso con ese identificador.
func (r *RepositorioConfiguracion) BuscarCursoPorID(ctx context.Context, id int64) (Curso, error) {
	fila, err := r.consultas.BuscarCursoPorID(ctx, id)
	if err != nil {
		return Curso{}, err
	}
	return Curso{
		ID:            fila.ID,
		AnioEscolarID: fila.AnioEscolarID,
		Nombre:        fila.Nombre,
		Activo:        fila.Activo,
		CreadoEn:      fila.CreadoEn.Time,
	}, nil
}

// ListarCursosPorAnio devuelve los cursos del año con sus grados.
func (r *RepositorioConfiguracion) ListarCursosPorAnio(ctx context.Context, anioID int64) ([]Curso, error) {
	filas, err := r.consultas.ListarCursosPorAnio(ctx, anioID)
	if err != nil {
		return nil, err
	}
	grados, err := r.consultas.ListarCursosGradosPorAnio(ctx, anioID)
	if err != nil {
		return nil, err
	}
	porCurso := make(map[int64][]int64, len(filas))
	for _, fila := range grados {
		porCurso[fila.CursoID] = append(porCurso[fila.CursoID], fila.GradoID)
	}

	cursos := make([]Curso, 0, len(filas))
	for _, fila := range filas {
		cursos = append(cursos, Curso{
			ID:            fila.ID,
			AnioEscolarID: fila.AnioEscolarID,
			Nombre:        fila.Nombre,
			Activo:        fila.Activo,
			Grados:        porCurso[fila.ID],
			CreadoEn:      fila.CreadoEn.Time,
		})
	}
	return cursos, nil
}

// DesactivarCurso desactiva el curso y devuelve falso si ya estaba inactivo.
func (r *RepositorioConfiguracion) DesactivarCurso(ctx context.Context, id int64) (bool, error) {
	filas, err := r.consultas.DesactivarCurso(ctx, id)
	if err != nil {
		return false, err
	}
	return filas > 0, nil
}

// CursoVinculadoAGrado indica si el curso está activo y vinculado al grado.
func (r *RepositorioConfiguracion) CursoVinculadoAGrado(ctx context.Context, cursoID, gradoID int64) (bool, error) {
	return r.consultas.CursoVinculadoAGrado(ctx, generado.CursoVinculadoAGradoParams{
		ID:      cursoID,
		GradoID: gradoID,
	})
}

// GradoDeSeccion devuelve el grado al que pertenece la sección.
func (r *RepositorioConfiguracion) GradoDeSeccion(ctx context.Context, seccionID int64) (int64, error) {
	return r.consultas.GradoDeSeccion(ctx, seccionID)
}
