package intermediarios

import (
	"net/url"
	"strings"
)

// RutaInicioPorRol devuelve la pantalla de inicio de cada rol.
func RutaInicioPorRol(rol string) string {
	switch rol {
	case RolDocente:
		return "/asistencia"
	case RolPorteria:
		return "/porteria"
	default:
		return "/alumnos"
	}
}

// RutaPermitidaParaRol indica si una ruta interna es accesible para el rol.
// Se usa para decidir si una URL de retorno es segura después del ingreso.
func RutaPermitidaParaRol(rol, ruta string) bool {
	if !rutaInterna(ruta) {
		return false
	}
	destino := rutaSinConsulta(ruta)
	if destino == "/ingreso" || destino == "/salir" {
		return false
	}
	switch rol {
	case RolAdministrador:
		return true
	case RolDirector:
		return tienePrefijo(destino, "/alumnos", "/carnets", "/configuracion", "/asistencia",
			"/justificaciones", "/reportes", "/alertas", "/docentes")
	case RolDocente:
		return tienePrefijo(destino, "/asistencia", "/justificaciones", "/reportes", "/alertas")
	case RolPorteria:
		return tienePrefijo(destino, "/porteria", "/alertas")
	default:
		return false
	}
}

func rutaInterna(ruta string) bool {
	if ruta == "" || len(ruta) > 2048 || !strings.HasPrefix(ruta, "/") || strings.HasPrefix(ruta, "//") {
		return false
	}
	analizada, err := url.Parse(ruta)
	if err != nil {
		return false
	}
	return analizada.Scheme == "" && analizada.Host == "" && strings.HasPrefix(analizada.Path, "/")
}

func rutaSinConsulta(ruta string) string {
	if analizada, err := url.Parse(ruta); err == nil {
		return analizada.Path
	}
	return ruta
}

func tienePrefijo(ruta string, prefijos ...string) bool {
	for _, prefijo := range prefijos {
		if ruta == prefijo || strings.HasPrefix(ruta, prefijo+"/") || strings.HasPrefix(ruta, prefijo+"?") {
			return true
		}
	}
	return false
}
