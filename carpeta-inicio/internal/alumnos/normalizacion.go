package alumnos

import "strings"

var reemplazosAcentos = map[rune]rune{
	'á': 'a', 'à': 'a', 'ä': 'a', 'â': 'a', 'ã': 'a',
	'é': 'e', 'è': 'e', 'ë': 'e', 'ê': 'e',
	'í': 'i', 'ì': 'i', 'ï': 'i', 'î': 'i',
	'ó': 'o', 'ò': 'o', 'ö': 'o', 'ô': 'o', 'õ': 'o',
	'ú': 'u', 'ù': 'u', 'ü': 'u', 'û': 'u',
	'ñ': 'n', 'ç': 'c', 'ý': 'y',
}

var caracteresEliminados = map[rune]bool{
	'°': true, 'º': true, 'ª': true,
}

// Normalizar prepara un texto para búsqueda: lo pasa a minúsculas, quita tildes y
// diéresis, convierte la ñ en n y reduce los espacios internos a uno solo.
func Normalizar(texto string) string {
	var constructor strings.Builder
	espacioPendiente := false

	for _, caracter := range strings.ToLower(strings.TrimSpace(texto)) {
		if caracter == ' ' || caracter == '\t' || caracter == '\n' || caracter == '\r' {
			espacioPendiente = constructor.Len() > 0
			continue
		}
		if caracteresEliminados[caracter] {
			continue
		}
		if espacioPendiente {
			constructor.WriteRune(' ')
			espacioPendiente = false
		}
		if reemplazo, existe := reemplazosAcentos[caracter]; existe {
			constructor.WriteRune(reemplazo)
			continue
		}
		constructor.WriteRune(caracter)
	}
	return constructor.String()
}

// LimpiarIdentificador quita los espacios de los extremos y conserva los ceros a la izquierda.
func LimpiarIdentificador(identificador string) string {
	return strings.TrimSpace(identificador)
}
