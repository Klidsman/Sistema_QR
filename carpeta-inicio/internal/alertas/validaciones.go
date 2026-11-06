package alertas

import "time"

// RangoDelMes devuelve el primer y el último día del mes de la fecha.
func RangoDelMes(fecha time.Time) (time.Time, time.Time) {
	inicio := time.Date(fecha.Year(), fecha.Month(), 1, 0, 0, 0, 0, fecha.Location())
	return inicio, inicio.AddDate(0, 1, -1)
}

// DiaCuentaParaRacha indica si el día de la semana puede formar parte de una
// racha de faltas. Sábados y domingos no cuentan.
func DiaCuentaParaRacha(dia time.Weekday) bool {
	return dia != time.Saturday && dia != time.Sunday
}

// TipoValido indica si el tipo de alerta es conocido.
func TipoValido(tipo string) bool {
	return tipo == TipoTardanza || tipo == TipoReiteracionTardanza || tipo == TipoReiteracionFalta
}

// EstadoValido indica si el estado de lectura es conocido.
func EstadoValido(estado string) bool {
	return estado == EstadoNueva || estado == EstadoVista
}

// DestinatarioTardanza elige al portero o, si falta, al administrador.
func DestinatarioTardanza(registradaPor *int64, administradorID int64) int64 {
	if registradaPor != nil && *registradaPor > 0 {
		return *registradaPor
	}
	return administradorID
}

// AlcanzaUmbral indica si el conteo coincide exactamente con el umbral.
// Se usa "==" para generar una sola alerta por situación; los índices únicos
// refuerzan la deduplicación ante ejecuciones repetidas.
func AlcanzaUmbral(conteo, umbral int) bool {
	return umbral > 0 && conteo == umbral
}
