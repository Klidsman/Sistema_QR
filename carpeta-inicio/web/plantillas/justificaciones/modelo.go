// Package justificaciones contiene las plantillas del flujo de justificaciones.
package justificaciones

import "strconv"

// FechaVista es una fecha cubierta.
type FechaVista struct {
	Fecha  string
	Tipo   string
	Estado string
}

// AdjuntoVista es un documento visible.
type AdjuntoVista struct {
	ID     int64
	Nombre string
	Tamano string
}

// JustificacionVista son los datos de una justificación para mostrar.
type JustificacionVista struct {
	ID              int64
	Alumno          string
	Seccion         string
	Motivo          string
	Estado          string
	Registrante     string
	CreadaEn        string
	Comentario      string
	MotivoAnulacion string
	Fechas          []FechaVista
	Adjuntos        []AdjuntoVista
	Pendiente       bool
	Aprobada        bool
	PuedeResolver   bool
	PuedeAnular     bool
}

// Panel es la lista de justificaciones.
type Panel struct {
	Nombre          string
	Rol             string
	TokenCSRF       string
	Justificaciones []JustificacionVista
	Mensaje         string
	Error           string
	Titulo          string
}

// Formulario es la pantalla de registro.
type Formulario struct {
	Nombre    string
	Rol       string
	TokenCSRF string
	AlumnoID  int64
	Alumno    string
	Seccion   string
	Desde     string
	Hasta     string
	Motivo    string
	Error     string
}

// Detalle es la ficha de una justificación.
type Detalle struct {
	Nombre        string
	Rol           string
	TokenCSRF     string
	Justificacion JustificacionVista
	Mensaje       string
	Error         string
}

func numero(valor int64) string {
	return strconv.FormatInt(valor, 10)
}
