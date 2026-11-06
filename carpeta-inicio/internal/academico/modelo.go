package academico

// Seccion es una sección de un grado con los datos que la interfaz necesita para mostrarla.
type Seccion struct {
	ID            int64
	Nombre        string
	Turno         string
	GradoID       int64
	Grado         string
	AnioEscolarID int64
	AnioEscolar   string
}

// Etiqueta devuelve el texto que se muestra en los formularios.
func (s Seccion) Etiqueta() string {
	return s.Grado + " " + s.Nombre + " (" + s.AnioEscolar + ")"
}
