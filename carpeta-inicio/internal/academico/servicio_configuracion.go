package academico

import (
	"time"

	"asistencia/internal/plataforma/fechas"
)

// Reloj devuelve la fecha y hora actuales. Se inyecta en los servicios y en las pruebas.
type Reloj func() time.Time

// RelojDelSistema devuelve un reloj que usa la fecha de hoy en la zona del colegio.
func RelojDelSistema() Reloj {
	return fechas.Hoy
}
