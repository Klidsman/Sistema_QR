package asistencia

import (
	"fmt"
	"time"

	"asistencia/internal/plataforma/fechas"
)

// EstadoValido indica si el estado pertenece al registro del docente.
func EstadoValido(estado string) bool {
	for _, valido := range EstadosRegistro() {
		if estado == valido {
			return true
		}
	}
	return false
}

// ValidarObservacion exige un máximo de 300 caracteres.
func ValidarObservacion(texto string) error {
	if len([]rune(texto)) > 300 {
		return ErrObservacionLarga
	}
	return nil
}

// ValidarFechaNoFutura rechaza fechas posteriores a hoy en la zona del colegio.
func ValidarFechaNoFutura(fecha, hoy time.Time) error {
	if fechas.CompararDia(fecha, hoy) > 0 {
		return ErrFechaFutura
	}
	return nil
}

// DentroDelPlazo indica si todavía se puede corregir un registro creado en ese momento.
func DentroDelPlazo(creadoEn, ahora time.Time, plazoHoras int) bool {
	return !ahora.After(creadoEn.Add(time.Duration(plazoHoras) * time.Hour))
}

// HayDiferencia compara la presencia del QR con el estado del docente.
// Es una discrepancia informativa: el QR sigue siendo la asistencia oficial.
func HayDiferencia(marcacion *Marcacion, estadoDocente string) bool {
	if marcacion != nil {
		return estadoDocente == EstadoFalta || estadoDocente == EstadoFaltaJustificada
	}
	return estadoDocente == EstadoPresente || estadoDocente == EstadoTardanza
}

// ValidarRangoExportacion exige un rango de fechas ordenado.
func ValidarRangoExportacion(desde, hasta time.Time) error {
	if fechas.CompararDia(desde, hasta) > 0 {
		return ErrRangoFechasInvalido
	}
	return nil
}

// ValidarFechasAsignacion exige que la fecha de fin no sea anterior a la de inicio.
func ValidarFechasAsignacion(inicio time.Time, fin *time.Time) error {
	if fin != nil && fechas.CompararDia(*fin, inicio) < 0 {
		return ErrAsignacionFueraDeRango
	}
	return nil
}

// AsignacionVigenteEnFecha indica si la asignación cubre la fecha.
func AsignacionVigenteEnFecha(asignacion AsignacionDocente, fecha time.Time) bool {
	if fechas.CompararDia(asignacion.FechaInicio, fecha) > 0 {
		return false
	}
	return asignacion.FechaFin == nil || fechas.CompararDia(fecha, *asignacion.FechaFin) <= 0
}

// ClaveClase identifica una hora de clase por fecha, sección, curso y bloque.
func ClaveClase(fecha time.Time, seccionID, cursoID int64, bloque int) string {
	return fmt.Sprintf("%s|%d|%d|%d", fechas.FormatearISO(fecha), seccionID, cursoID, bloque)
}

// ValidarPlazoHoras exige un plazo entre 1 y 168 horas.
func ValidarPlazoHoras(horas int) error {
	if horas < 1 || horas > 168 {
		return ErrPlazoFueraDeRango
	}
	return nil
}

// ValidarPlazoJustificacion exige un plazo entre 1 y 60 días.
func ValidarPlazoJustificacion(dias int) error {
	if dias < 1 || dias > 60 {
		return ErrPlazoJustificacionFueraDeRango
	}
	return nil
}
