package autenticacion

import (
	"net/http"

	"asistencia/internal/plataforma/intermediarios"
	"asistencia/web/plantillas/paginas"
)

// inicio muestra la página de presentación del colegio y del sistema.
func (m *Manejador) inicio(respuesta http.ResponseWriter, peticion *http.Request) {
	ctx := peticion.Context()
	rol := intermediarios.RolEnSesion(ctx, m.sesiones)
	nombre := intermediarios.NombreUsuario(ctx, m.sesiones)

	respuesta.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := paginas.Inicio(nombre, rol, rol != "").Render(ctx, respuesta); err != nil {
		m.registro.Error("no se pudo renderizar la página de inicio", "error", err)
	}
}
