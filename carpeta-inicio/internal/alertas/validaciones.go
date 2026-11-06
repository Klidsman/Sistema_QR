package alertas

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
