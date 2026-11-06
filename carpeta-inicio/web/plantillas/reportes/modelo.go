// Package reportes contiene las plantillas de los reportes de asistencia.
package reportes

import "fmt"

// Filtros son los valores del formulario.
type Filtros struct {
	Desde   string
	Hasta   string
	Turno   string
	Grado   string
	Seccion string
	Alumno  string
	Estado  string
}

// FilaDiaria es una fila del reporte diario.
type FilaDiaria struct {
	Turno         string
	Grado         string
	Seccion       string
	Identificador string
	Alumno        string
	Estado        string
	Hora          string
}

// Panel son los totales del día.
type Panel struct {
	Presentes  int
	Tardanzas  int
	Ausentes   int
	Pendientes int
	Total      int
}

// Diario son los datos de la pantalla diaria.
type Diario struct {
	Nombre    string
	Rol       string
	TokenCSRF string
	Filtros   Filtros
	Filas     []FilaDiaria
	Panel     Panel
	Fecha     string
}

// ResumenFila es una fila del reporte acumulado.
type ResumenFila struct {
	Grado                 string
	Seccion               string
	Identificador         string
	Alumno                string
	Asistencias           int
	Tardanzas             int
	Faltas                int
	FaltasJustificadas    int
	TardanzasJustificadas int
	Porcentaje            string
}

// Resumen son los datos del reporte acumulado.
type Resumen struct {
	Nombre    string
	Rol       string
	TokenCSRF string
	Filtros   Filtros
	Filas     []ResumenFila
}

// ReiteracionFila es un caso del reporte de reiteración.
type ReiteracionFila struct {
	Alumno          string
	Seccion         string
	Tardanzas       int
	FaltasSeguidas  int
	UmbralTardanzas int
	UmbralFaltas    int
}

// Reiteracion son los datos del reporte de reiteración.
type Reiteracion struct {
	Nombre    string
	Rol       string
	TokenCSRF string
	Filtros   Filtros
	Filas     []ReiteracionFila
}

func numero(valor int) string {
	return fmt.Sprintf("%d", valor)
}
