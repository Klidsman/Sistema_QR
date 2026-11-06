package vista

// DocenteOpcion es un docente seleccionable al asignar.
type DocenteOpcion struct {
	ID     int64
	Nombre string
}

// SeccionOpcion es una sección seleccionable.
type SeccionOpcion struct {
	ID       int64
	Etiqueta string
}

// CursoOpcion es un curso seleccionable.
type CursoOpcion struct {
	ID     int64
	Nombre string
}

// AsignacionVista es una asignación preparada para mostrarse.
type AsignacionVista struct {
	ID      int64
	Docente string
	Curso   string
	Seccion string
	Inicio  string
	Fin     string
	Tipo    string
	Estado  string
}
