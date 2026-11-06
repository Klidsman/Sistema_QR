// Package porteria contiene las plantillas del registro de ingreso por QR.
package porteria

// Resultado es el resumen visible de un escaneo.
type Resultado struct {
	Registrado    bool
	Alumno        string
	Identificador string
	Seccion       string
	Hora          string
	Estado        string
	Motivo        string
}
