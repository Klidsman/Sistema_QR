package vista

import "time"

// AnioVista es un año escolar preparado para mostrarse en la interfaz.
type AnioVista struct {
	ID          int64
	Nombre      string
	FechaInicio time.Time
	FechaFin    time.Time
	Estado      string
}

// GradoVista es un grado preparado para mostrarse en la interfaz.
type GradoVista struct {
	ID     int64
	Nombre string
}

// SeccionVista es una sección preparada para mostrarse en la interfaz.
type SeccionVista struct {
	ID     int64
	Grado  string
	Nombre string
	Turno  string
}

// ExcepcionVista es una excepción de calendario preparada para mostrarse.
type ExcepcionVista struct {
	ID     int64
	Tipo   string
	Inicio string
	Fin    string
	Motivo string
	Grados string
}

// FormularioExcepcion conserva los datos del formulario de calendario.
type FormularioExcepcion struct {
	Tipo        string
	FechaInicio string
	FechaFin    string
	Motivo      string
	Grados      []int64
}

// UmbralVista es un umbral de alerta preparado para mostrarse.
type UmbralVista struct {
	Tipo         string
	Valor        int
	VigenteDesde time.Time
}

// FormularioUmbral conserva los datos del formulario de umbrales.
type FormularioUmbral struct {
	Tipo  string
	Valor string
	Fecha string
}

// CursoVista es un curso preparado para mostrarse en la interfaz.
type CursoVista struct {
	ID     int64
	Nombre string
	Grados string
	Activo bool
}

// GradoOpcion es un grado seleccionable al crear un curso.
type GradoOpcion struct {
	ID     int64
	Nombre string
}

// CursoOpcion es un curso seleccionable en una celda del horario.
type CursoOpcion struct {
	ID     int64
	Nombre string
}

// GrillaVista es el horario semanal de una sección preparado para mostrarse.
type GrillaVista struct {
	SeccionID int64
	Seccion   string
	Turno     string
	Bloques   []BloqueVista
	Cursos    []CursoOpcion
	Seleccion map[string]string
}

// BloqueVista es un bloque de hora de clase preparado para mostrarse en la interfaz.
type BloqueVista struct {
	ID           int64
	Turno        string
	Numero       int
	HoraInicio   string
	HoraFin      string
	VigenteDesde time.Time
}

// HorarioVista es un horario de turno preparado para mostrarse en la interfaz.
type HorarioVista struct {
	Turno              string
	HoraIngreso        string
	ToleranciaMinutos  int
	HoraLimiteTardanza string
	HoraCierre         string
	VigenteDesde       time.Time
}
