package academico

import (
	"context"
	"io"
	"log/slog"
	"net/http"

	"asistencia/internal/plataforma/intermediarios"
	vista "asistencia/web/plantillas/configuracion"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
)

// ServicioAcademico son las operaciones de configuración que usa el manejador.
type ServicioAcademico interface {
	Resumen(ctx context.Context) (ResumenConfiguracion, error)
}

// DependenciasConfiguracion reúne lo que el manejador de configuración necesita.
type DependenciasConfiguracion struct {
	Configuracion ServicioAcademico
	Sesiones      *scs.SessionManager
	Registro      *slog.Logger
}

// Manejador atiende las pantallas de configuración académica.
type Manejador struct {
	configuracion ServicioAcademico
	sesiones      *scs.SessionManager
	registro      *slog.Logger
}

// NuevoManejador crea el manejador de configuración académica.
func NuevoManejador(dependencias DependenciasConfiguracion) *Manejador {
	return &Manejador{
		configuracion: dependencias.Configuracion,
		sesiones:      dependencias.Sesiones,
		registro:      dependencias.Registro,
	}
}

// Montar registra las rutas de configuración agrupadas por rol.
func (m *Manejador) Montar(enrutador chi.Router) {
	enrutador.Group(func(lectura chi.Router) {
		lectura.Use(intermediarios.RequerirRol(m.sesiones,
			intermediarios.RolAdministrador, intermediarios.RolDirector))
		lectura.Get("/configuracion", m.panel)
	})
}

func (m *Manejador) panel(respuesta http.ResponseWriter, peticion *http.Request) {
	resumen, err := m.configuracion.Resumen(peticion.Context())
	if err != nil {
		m.registro.Error("no se pudo cargar el resumen de configuración", "error", err)
		http.Error(respuesta, "No se pudo cargar la configuración", http.StatusInternalServerError)
		return
	}

	anio := ""
	estado := ""
	if resumen.Anio != nil {
		anio = resumen.Anio.Nombre
		estado = resumen.Anio.Estado
	}

	m.renderizar(respuesta, peticion, vista.Panel(
		m.nombreUsuario(peticion.Context()),
		intermediarios.RolUsuario(peticion.Context()),
		m.tokenCSRF(peticion.Context()),
		anio,
		estado,
		resumen.Grados,
		resumen.Secciones,
	))
}

func (m *Manejador) renderizar(respuesta http.ResponseWriter, peticion *http.Request, componente interface {
	Render(context.Context, io.Writer) error
}) {
	respuesta.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := componente.Render(peticion.Context(), respuesta); err != nil {
		m.registro.Error("no se pudo renderizar la vista de configuración", "error", err)
	}
}

func (m *Manejador) tokenCSRF(ctx context.Context) string {
	token, err := intermediarios.TokenCSRF(ctx, m.sesiones)
	if err != nil {
		m.registro.Error("no se pudo preparar el token CSRF", "error", err)
	}
	return token
}

func (m *Manejador) nombreUsuario(ctx context.Context) string {
	nombre := intermediarios.NombreUsuario(ctx, m.sesiones)
	if nombre == "" {
		return "Usuario"
	}
	return nombre
}
