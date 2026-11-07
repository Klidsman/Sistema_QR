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

// ClaseSeleccionVista es una clase que el docente puede abrir.
type ClaseSeleccionVista struct {
	Curso   string
	Seccion string
	Horario string
	Bloque  int
	URL     string
}

// AsignacionSeleccionVista agrupa las clases de una asignación y explica si falta algo.
type AsignacionSeleccionVista struct {
	Curso   string
	Seccion string
	Mensaje string
	Clases  []ClaseSeleccionVista
}

// ClaseVista es la cabecera de una hora de clase.
type ClaseVista struct {
	SeccionID  int64
	CursoID    int64
	Seccion    string
	Curso      string
	Turno      string
	Fecha      string
	FechaISO   string
	Bloque     int
	HoraInicio string
	HoraFin    string
}

// AlumnoFilaVista es un alumno con su registro y su QR.
type AlumnoFilaVista struct {
	ID            int64
	Identificador string
	Nombre        string
	Estado        string
	Observacion   string
	QRTexto       string
	TieneQR       bool
	Diferencia    bool
}

// EstadoOpcion es un estado seleccionable.
type EstadoOpcion struct {
	Valor    string
	Etiqueta string
}

// ConsultaFilaVista es un registro de asistencia preparado para consulta.
type ConsultaFilaVista struct {
	ID          int64
	Alumno      string
	Curso       string
	Seccion     string
	Bloque      int
	Estado      string
	Observacion string
	QRTexto     string
	Docente     string
	Diferencia  bool
}

// ParametrosVista son los parámetros de asistencia para mostrar.
type ParametrosVista struct {
	Plazo              int
	PlazoJustificacion int
	Actualizado        string
	ActualizadoPor     string
}

// AsignacionVista es una asignación preparada para mostrarse.
type AsignacionVista struct {
	ID        int64
	SeccionID int64
	Docente   string
	Curso     string
	Seccion   string
	Inicio    string
	Fin       string
	Tipo      string
	Estado    string
}
