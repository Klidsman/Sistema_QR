package asistencia

import (
	"context"
	"time"

	"asistencia/internal/academico"
	"asistencia/internal/alumnos"
)

// ResolutorClase resuelve una hora de clase y el calendario sin que asistencia consulte
// las tablas del módulo académico.
type ResolutorClase interface {
	ResolverClase(ctx context.Context, seccionID, cursoID int64, bloque int, fecha time.Time) (academico.ClaseProgramada, error)
	EsDiaNoLaborable(ctx context.Context, seccionID int64, fecha time.Time) (bool, error)
	EtiquetaSeccion(ctx context.Context, seccionID int64) (string, error)
	NombreCurso(ctx context.Context, cursoID int64) (string, error)
}

// ListadorAlumnos devuelve los alumnos activos de una sección en la fecha.
type ListadorAlumnos interface {
	ListarAlumnosDeSeccionEnFecha(ctx context.Context, seccionID int64, fecha time.Time) ([]alumnos.AlumnoEnClase, error)
}

// BuscadorDocente valida el rol y el estado de un docente.
type BuscadorDocente interface {
	DocenteActivo(ctx context.Context, docenteID int64) (bool, string, error)
}
