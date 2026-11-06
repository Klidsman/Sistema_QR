package academico

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// ErrClaseNoProgramada indica que la celda del horario no tiene ese curso en el día y bloque.
var ErrClaseNoProgramada = errors.New("la clase no corresponde al horario de la sección")

// ResolverClase valida y devuelve una hora de clase para una fecha.
func (s *ServicioConfiguracion) ResolverClase(ctx context.Context, seccionID, cursoID int64, bloque int, fecha time.Time) (ClaseProgramada, error) {
	seccion, err := s.repositorio.BuscarSeccionPorID(ctx, seccionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ClaseProgramada{}, ErrSeccionNoEncontrada
	}
	if err != nil {
		return ClaseProgramada{}, fmt.Errorf("buscando la sección: %w", err)
	}

	turno, err := s.repositorio.TurnoVigenteSeccion(ctx, seccionID, fecha)
	if errors.Is(err, pgx.ErrNoRows) {
		return ClaseProgramada{}, ErrSeccionSinTurno
	}
	if err != nil {
		return ClaseProgramada{}, fmt.Errorf("buscando el turno de la sección: %w", err)
	}

	horario, err := s.repositorio.HorarioTurnoVigente(ctx, turno, fecha)
	if errors.Is(err, pgx.ErrNoRows) {
		return ClaseProgramada{}, ErrHorarioNoEncontrado
	}
	if err != nil {
		return ClaseProgramada{}, fmt.Errorf("buscando el horario del turno: %w", err)
	}

	bloqueClase, err := s.repositorio.BloqueVigentePorNumero(ctx, turno, bloque, fecha)
	if errors.Is(err, pgx.ErrNoRows) {
		return ClaseProgramada{}, ErrBloqueNoEncontrado
	}
	if err != nil {
		return ClaseProgramada{}, fmt.Errorf("buscando el bloque de clase: %w", err)
	}
	if err := ValidarBloque(bloqueClase, horario, nil); err != nil {
		return ClaseProgramada{}, err
	}

	celda, err := s.repositorio.HorarioSeccionCelda(ctx, seccionID, int(fecha.Weekday()), bloque, fecha)
	if errors.Is(err, pgx.ErrNoRows) {
		return ClaseProgramada{}, ErrClaseNoProgramada
	}
	if err != nil {
		return ClaseProgramada{}, fmt.Errorf("buscando la celda del horario: %w", err)
	}
	if celda.CursoID == nil || *celda.CursoID != cursoID {
		return ClaseProgramada{}, ErrClaseNoProgramada
	}
	vinculado, err := s.repositorio.CursoVinculadoAGrado(ctx, cursoID, seccion.GradoID)
	if err != nil {
		return ClaseProgramada{}, fmt.Errorf("comprobando el curso de la sección: %w", err)
	}
	if !vinculado {
		return ClaseProgramada{}, ErrCursoNoVinculado
	}

	curso, err := s.repositorio.BuscarCursoPorID(ctx, cursoID)
	if err != nil {
		return ClaseProgramada{}, fmt.Errorf("buscando el curso: %w", err)
	}

	return ClaseProgramada{
		SeccionID:  seccionID,
		CursoID:    cursoID,
		Bloque:     bloque,
		Turno:      turno,
		HoraInicio: bloqueClase.HoraInicio,
		HoraFin:    bloqueClase.HoraFin,
		Seccion:    seccion.Etiqueta(),
		Curso:      curso.Nombre,
	}, nil
}

// CursoVinculadoASeccion indica si el curso está activo y vinculado al grado de la sección.
func (s *ServicioConfiguracion) CursoVinculadoASeccion(ctx context.Context, cursoID, seccionID int64) (bool, error) {
	gradoID, err := s.repositorio.GradoDeSeccion(ctx, seccionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, ErrSeccionNoEncontrada
	}
	if err != nil {
		return false, fmt.Errorf("buscando el grado de la sección: %w", err)
	}
	return s.repositorio.CursoVinculadoAGrado(ctx, cursoID, gradoID)
}

// HorarioDelDia devuelve los bloques con curso del horario semanal de la sección.
func (s *ServicioConfiguracion) HorarioDelDia(ctx context.Context, seccionID int64, fecha time.Time) ([]BloqueCurso, error) {
	_, err := s.repositorio.BuscarSeccionPorID(ctx, seccionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrSeccionNoEncontrada
	}
	if err != nil {
		return nil, fmt.Errorf("buscando la sección: %w", err)
	}
	turno, err := s.repositorio.TurnoVigenteSeccion(ctx, seccionID, fecha)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrSeccionSinTurno
	}
	if err != nil {
		return nil, fmt.Errorf("buscando el turno de la sección: %w", err)
	}
	bloques, err := s.repositorio.BloquesVigentesTurno(ctx, turno, fecha)
	if err != nil {
		return nil, fmt.Errorf("listando los bloques del turno: %w", err)
	}
	celdas, err := s.repositorio.HorarioSeccionVigente(ctx, seccionID, fecha)
	if err != nil {
		return nil, fmt.Errorf("listando el horario de la sección: %w", err)
	}
	horas := make(map[int]BloqueClase, len(bloques))
	for _, bloque := range bloques {
		horas[bloque.Numero] = bloque
	}
	dia := int(fecha.Weekday())
	resultado := make([]BloqueCurso, 0)
	for _, celda := range celdas {
		if celda.DiaSemana != dia || celda.CursoID == nil {
			continue
		}
		bloque, existe := horas[celda.BloqueNumero]
		if !existe {
			continue
		}
		resultado = append(resultado, BloqueCurso{
			CursoID:    *celda.CursoID,
			Bloque:     celda.BloqueNumero,
			HoraInicio: bloque.HoraInicio,
			HoraFin:    bloque.HoraFin,
		})
	}
	return resultado, nil
}

// ListarSeccionesActivas devuelve las secciones del año activo.
func (s *ServicioConfiguracion) ListarSeccionesActivas(ctx context.Context) ([]Seccion, error) {
	anio, err := s.AnioActivo(ctx)
	if err != nil {
		return nil, err
	}
	if anio == nil {
		return nil, nil
	}
	return s.repositorio.ListarSeccionesPorAnio(ctx, anio.ID)
}

// ListarCursosActivos devuelve los cursos del año activo.
func (s *ServicioConfiguracion) ListarCursosActivos(ctx context.Context) ([]Curso, error) {
	anio, err := s.AnioActivo(ctx)
	if err != nil {
		return nil, err
	}
	if anio == nil {
		return nil, nil
	}
	cursos, err := s.ListarCursos(ctx, anio.ID)
	if err != nil {
		return nil, err
	}
	activos := make([]Curso, 0, len(cursos))
	for _, curso := range cursos {
		if curso.Activo {
			activos = append(activos, curso)
		}
	}
	return activos, nil
}

// EsDiaNoLaborable indica si una excepción de calendario alcanza a la sección en la fecha.
func (s *ServicioConfiguracion) EsDiaNoLaborable(ctx context.Context, seccionID int64, fecha time.Time) (bool, error) {
	excepciones, err := s.repositorio.ExcepcionesDeSeccion(ctx, seccionID, fecha)
	if err != nil {
		return false, fmt.Errorf("buscando las excepciones de la sección: %w", err)
	}
	return len(excepciones) > 0, nil
}

// EtiquetaSeccion devuelve el texto visible de una sección.
func (s *ServicioConfiguracion) EtiquetaSeccion(ctx context.Context, seccionID int64) (string, error) {
	seccion, err := s.repositorio.BuscarSeccionPorID(ctx, seccionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrSeccionNoEncontrada
	}
	if err != nil {
		return "", fmt.Errorf("buscando la sección: %w", err)
	}
	return seccion.Etiqueta(), nil
}

// NombreCurso devuelve el nombre de un curso.
func (s *ServicioConfiguracion) NombreCurso(ctx context.Context, cursoID int64) (string, error) {
	curso, err := s.repositorio.BuscarCursoPorID(ctx, cursoID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrCursoNoEncontrado
	}
	if err != nil {
		return "", fmt.Errorf("buscando el curso: %w", err)
	}
	return curso.Nombre, nil
}
