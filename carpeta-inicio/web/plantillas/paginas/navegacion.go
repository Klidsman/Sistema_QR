package paginas

// EnlaceInicio devuelve la pantalla de inicio según el rol.
func EnlaceInicio(rol string) string {
	switch rol {
	case "porteria":
		return "/porteria"
	case "docente":
		return "/asistencia"
	default:
		return "/alumnos"
	}
}
