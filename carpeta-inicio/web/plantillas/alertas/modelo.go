// Package alertas contiene las plantillas de las alertas internas.
package alertas

import "strconv"

// AlertaVista son los datos de una alerta listos para mostrar.
type AlertaVista struct {
	ID                int64
	Tipo              string
	TipoEtiqueta      string
	Alumno            string
	Seccion           string
	Fecha             string
	Estado            string
	Nueva             bool
	SinResponsable    bool
	SituacionCambiada bool
	MotivoCambio      string
}

// Panel son los datos de la pantalla de alertas.
type Panel struct {
	Nombre    string
	Rol       string
	TokenCSRF string
	Nuevas    int
	Alertas   []AlertaVista
	Desde     string
	Hasta     string
	SeccionID string
	Tipo      string
	Estado    string
	Mensaje   string
	EsAdmin   bool
}

// Indicador es el contador de alertas nuevas.
type Indicador struct {
	Nuevas int
}

func numero(valor int) string {
	return strconv.Itoa(valor)
}

func numero64(valor int64) string {
	return strconv.FormatInt(valor, 10)
}

func filaClase(alerta AlertaVista) string {
	if alerta.Nueva {
		return "bg-blue-50/50"
	}
	return ""
}
