// Package vista contiene los modelos que reciben las plantillas de alumnos.
package vista

// Filtros son los valores del formulario de búsqueda del listado.
type Filtros struct {
	Texto     string
	Estado    string
	SeccionID int64
}

// DatosFormulario es el estado del formulario de alta o edición.
type DatosFormulario struct {
	EsEdicion         bool
	ID                int64
	Identificador     string
	Nombres           string
	Apellidos         string
	SeccionID         int64
	ApoderadoNombre   string
	ApoderadoTelefono string
	Errores           map[string]string
}

// TieneErrores indica si el formulario debe mostrar mensajes por campo.
func (f DatosFormulario) TieneErrores() bool {
	return len(f.Errores) > 0
}

// ErrorDe devuelve el mensaje del campo indicado.
func (f DatosFormulario) ErrorDe(campo string) string {
	return f.Errores[campo]
}

// Alumno es la ficha que muestra la plantilla.
type Alumno struct {
	ID                int64
	Identificador     string
	Nombres           string
	Apellidos         string
	Estado            string
	Grado             string
	Seccion           string
	Turno             string
	AnioEscolar       string
	FotoClave         string
	ApoderadoNombre   string
	ApoderadoTelefono string
	FechaSalida       string
}

// NombreCompleto devuelve apellidos y nombres en el orden en que se muestran.
func (a Alumno) NombreCompleto() string {
	return a.Apellidos + ", " + a.Nombres
}

// Activo indica si el alumno puede tener un QR vigente.
func (a Alumno) Activo() bool {
	return a.Estado == "activo"
}

// AlumnoListado es una fila del listado.
type AlumnoListado struct {
	ID            int64
	Identificador string
	Nombres       string
	Apellidos     string
	Estado        string
	Grado         string
	Seccion       string
	AnioEscolar   string
	TieneFoto     bool
}

// NombreCompleto devuelve apellidos y nombres en el orden en que se muestran.
func (a AlumnoListado) NombreCompleto() string {
	return a.Apellidos + ", " + a.Nombres
}

// Credencial es una credencial QR ya formateada para la plantilla.
type Credencial struct {
	Estado     string
	EmitidaEn  string
	RevocadaEn string
	Motivo     string
	Token      string
}

// HistorialSeccion es un tramo de sección ya formateado para la ficha.
type HistorialSeccion struct {
	Grado       string
	Seccion     string
	AnioEscolar string
	FechaInicio string
	FechaFin    string
	Actual      bool
}
