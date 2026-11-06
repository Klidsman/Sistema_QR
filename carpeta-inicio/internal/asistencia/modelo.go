package asistencia

import "time"

// AsignacionDocente vincula un docente con un curso y una sección en un rango de fechas.
type AsignacionDocente struct {
	ID          int64
	DocenteID   int64
	CursoID     int64
	SeccionID   int64
	FechaInicio time.Time
	FechaFin    *time.Time
	Titular     bool
	CreadoPor   int64
	CreadoEn    time.Time
}

// RegistroDocente es el estado personal que el docente asigna a un alumno en una clase.
type RegistroDocente struct {
	ID            int64
	AlumnoID      int64
	Fecha         time.Time
	CursoID       int64
	SeccionID     int64
	BloqueNumero  int
	Estado        string
	Observacion   string
	DocenteID     int64
	CreadoEn      time.Time
	ActualizadoEn time.Time
}

// Marcacion es el ingreso oficial por QR. Esta feature solo la lee.
type Marcacion struct {
	ID       int64
	AlumnoID int64
	Fecha    time.Time
	Hora     time.Time
	Estado   string
	CreadaEn time.Time
}

// ParametrosAsistencia contiene el plazo de edición del docente.
type ParametrosAsistencia struct {
	PlazoEdicionHoras int
	ActualizadoPor    *int64
	ActualizadoEn     time.Time
}

// ClaseAsignada es una clase que el docente puede abrir en una fecha.
type ClaseAsignada struct {
	AsignacionDocente
	Curso        string
	Seccion      string
	Turno        string
	BloqueNumero int
	HoraInicio   time.Time
	HoraFin      time.Time
	Fecha        time.Time
}
