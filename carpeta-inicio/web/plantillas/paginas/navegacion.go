package paginas

import "asistencia/internal/plataforma/intermediarios"

// EnlaceInicio devuelve la pantalla de inicio según el rol.
func EnlaceInicio(rol string) string {
	return intermediarios.RutaInicioPorRol(rol)
}
