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
