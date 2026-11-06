// Package fechas presenta las fechas y horas en la zona horaria del colegio (America/Lima).
package fechas

import "time"

// zonaLima es UTC-5 todo el año; Perú no aplica horario de verano.
var zonaLima = time.FixedZone("America/Lima", -5*60*60)

// Formatear devuelve la fecha y hora en formato dd/mm/aaaa hh:mm.
func Formatear(fecha time.Time) string {
	if fecha.IsZero() {
		return ""
	}
	return fecha.In(zonaLima).Format("02/01/2006 15:04")
}

// FormatearDia devuelve solo la fecha en formato dd/mm/aaaa.
func FormatearDia(fecha time.Time) string {
	if fecha.IsZero() {
		return ""
	}
	return fecha.In(zonaLima).Format("02/01/2006")
}
