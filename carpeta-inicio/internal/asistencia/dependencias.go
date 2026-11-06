package asistencia

import (
	"context"
	"time"

	"asistencia/internal/academico"
	"asistencia/internal/alumnos"
	"asistencia/internal/autenticacion"
)

// ResolutorClase resuelve una hora de clase y el calendario sin que asistencia consulte
// las tablas del módulo académico.
type ResolutorClase interface {
	ResolverClase(ctx context.Context, seccionID, cursoID int64, bloque int, fecha time.Time) (academico.ClaseProgramada, error)
	EsDiaNoLaborable(ctx context.Context, seccionID int64, fecha time.Time) (bool, error)
	EtiquetaSeccion(ctx context.Context, seccionID int64) (string, error)
	NombreCurso(ctx context.Context, cursoID int64) (string, error)
	CursoVinculadoASeccion(ctx context.Context, cursoID, seccionID int64) (bool, error)
	HorarioDelDia(ctx context.Context, seccionID int64, fecha time.Time) ([]academico.BloqueCurso, error)
	ListarSeccionesActivas(ctx context.Context) ([]academico.Seccion, error)
	ListarCursosActivos(ctx context.Context) ([]academico.Curso, error)
}

// ListadorAlumnos devuelve los alumnos activos de una sección en la fecha.
type ListadorAlumnos interface {
	ListarAlumnosDeSeccionEnFecha(ctx context.Context, seccionID int64, fecha time.Time) ([]alumnos.AlumnoEnClase, error)
}

// BuscadorDocente valida el rol y el estado de un docente.
type BuscadorDocente interface {
	DocenteActivo(ctx context.Context, docenteID int64) (bool, string, error)
	NombreDocente(ctx context.Context, docenteID int64) (string, error)
	ListarDocentes(ctx context.Context) ([]autenticacion.Docente, error)
}
