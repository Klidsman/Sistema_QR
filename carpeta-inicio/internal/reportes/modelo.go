// Package reportes calcula reportes de asistencia sobre los registros existentes.
package reportes

import "time"

// FilaDiaria es un alumno en el reporte del día.
type FilaDiaria struct {
	AlumnoID      int64
	Identificador string
	Alumno        string
	Grado         string
	Seccion       string
	Turno         string
	Estado        string
	Hora          string
}

// PanelDelDia resume los totales de una fecha.
type PanelDelDia struct {
	Fecha      time.Time
	Presentes  int
	Tardanzas  int
	Ausentes   int
	Pendientes int
	Total      int
}

// ResumenAlumno es el acumulado de un alumno en una sección.
type ResumenAlumno struct {
	AlumnoID              int64
	Identificador         string
	Alumno                string
	Grado                 string
	Seccion               string
	Turno                 string
	Asistencias           int
	Tardanzas             int
	Faltas                int
	FaltasJustificadas    int
	TardanzasJustificadas int
	DiasLaborables        int
	Porcentaje            float64
	FaltasSeguidas        int
}

// Reiteracion es un caso que supera un umbral.
type Reiteracion struct {
	Alumno          string
	Seccion         string
	Tardanzas       int
	FaltasSeguidas  int
	UmbralTardanzas int
	UmbralFaltas    int
}

// Filtros comunes de los reportes.
type Filtros struct {
	Desde     time.Time
	Hasta     time.Time
	Turno     string
	Grado     string
	Seccion   string
	Alumno    string
	Estado    string
	UsuarioID int64
	Rol       string
}

// Estados visibles del reporte diario.
const (
	EstadoPresente            = "presente"
	EstadoTardanza            = "tardanza"
	EstadoFalta               = "falta"
	EstadoFaltaJustificada    = "falta_justificada"
	EstadoTardanzaJustificada = "tardanza_justificada"
	EstadoPendiente           = "pendiente"
)

// EtiquetaEstado devuelve el texto visible.
func EtiquetaEstado(estado string) string {
	switch estado {
	case EstadoPresente:
		return "Presente"
	case EstadoTardanza:
		return "Tardanza"
	case EstadoFalta:
		return "Falta"
	case EstadoFaltaJustificada:
		return "Falta justificada"
	case EstadoTardanzaJustificada:
		return "Tardanza justificada"
	case EstadoPendiente:
		return "Pendiente"
	default:
		return estado
	}
}
