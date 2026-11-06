// Package fechas presenta las fechas y horas en la zona horaria del colegio (America/Lima).
package fechas

import (
	"fmt"
	"strings"
	"time"
)

// zonaLima es UTC-5 todo el año; Perú no aplica horario de verano.
var zonaLima = time.FixedZone("America/Lima", -5*60*60)

// Formatear devuelve la fecha y hora en formato dd/mm/aaaa hh:mm.
func Formatear(fecha time.Time) string {
	if fecha.IsZero() {
		return ""
	}
	return fecha.In(zonaLima).Format("02/01/2006 15:04")
}

// FormatearDia devuelve solo la fecha en formato dd/mm/aaaa. Se usa con valores de tipo
// date (sin hora): no se convierten de zona para no mover el día.
func FormatearDia(fecha time.Time) string {
	if fecha.IsZero() {
		return ""
	}
	return fecha.Format("02/01/2006")
}

// FormatearISO devuelve la fecha en formato aaaa-mm-dd, el que usan los campos de tipo fecha.
// Igual que FormatearDia, no convierte de zona porque recibe valores de tipo date.
func FormatearISO(fecha time.Time) string {
	if fecha.IsZero() {
		return ""
	}
	return fecha.Format("2006-01-02")
}

// Fecha recorta una marca de tiempo al día en la zona del colegio.
func Fecha(momento time.Time) time.Time {
	local := momento.In(zonaLima)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, zonaLima)
}

// Hoy devuelve la fecha actual en la zona del colegio, sin hora.
func Hoy() time.Time {
	return Fecha(time.Now())
}

// ParsearDia convierte aaaa-mm-dd a una fecha en la zona del colegio.
func ParsearDia(texto string) (time.Time, error) {
	limpio := strings.TrimSpace(texto)
	if limpio == "" {
		return time.Time{}, fmt.Errorf("la fecha está vacía")
	}
	fecha, err := time.ParseInLocation("2006-01-02", limpio, zonaLima)
	if err != nil {
		return time.Time{}, fmt.Errorf("la fecha no tiene el formato aaaa-mm-dd: %w", err)
	}
	return fecha, nil
}
