package asistencia

import "errors"

// Errores de negocio de la asistencia del docente.
var (
	ErrDocenteNoEncontrado    = errors.New("el docente no existe o no está activo")
	ErrDocenteSinAsignacion   = errors.New("el docente no tiene asignada esta clase en la fecha")
	ErrFechaFutura            = errors.New("no se puede registrar asistencia de una fecha futura")
	ErrDiaNoLaborable         = errors.New("la fecha no es laborable para esta sección")
	ErrClaseNoProgramada      = errors.New("la clase no corresponde al horario de la sección")
	ErrBloqueNoEncontrado     = errors.New("el bloque indicado no existe en el turno de la sección")
	ErrEstadoInvalido         = errors.New("el estado de asistencia no es válido")
	ErrObservacionLarga       = errors.New("la observación no puede superar los 300 caracteres")
	ErrRegistroNoEncontrado   = errors.New("el registro del docente no existe")
	ErrPlazoVencido           = errors.New("el plazo de edición terminó")
	ErrMotivoRequerido        = errors.New("indique el motivo de la corrección")
	ErrAlumnoFueraDeSeccion   = errors.New("el alumno no pertenece a la sección en la fecha indicada")
	ErrAsignacionNoEncontrada = errors.New("la asignación no existe")
	ErrRangoFechasInvalido    = errors.New("el rango de fechas no es válido")
	ErrPlazoFueraDeRango      = errors.New("el plazo de edición debe estar entre 1 y 168 horas")
	ErrCursoNoVinculado       = errors.New("el curso no está vinculado al grado de la sección")
	ErrAsignacionFueraDeRango = errors.New("las fechas de la asignación no son válidas")
)
