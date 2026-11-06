package porteria

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"asistencia/internal/plataforma/intermediarios"
	vista "asistencia/web/plantillas/porteria"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
)

// ServicioPorteria son las operaciones que usa el manejador.
type ServicioPorteria interface {
	Registrar(ctx context.Context, usuarioID int64, token string) (Resultado, error)
}

// DependenciasManejador reúne lo que necesita el manejador de portería.
type DependenciasManejador struct {
	Servicio ServicioPorteria
	Sesiones *scs.SessionManager
	Registro *slog.Logger
}

// Manejador atiende la pantalla de escaneo de portería.
type Manejador struct {
	servicio ServicioPorteria
	sesiones *scs.SessionManager
	registro *slog.Logger
}

// NuevoManejador crea el manejador de portería.
func NuevoManejador(dependencias DependenciasManejador) *Manejador {
	return &Manejador{
		servicio: dependencias.Servicio,
		sesiones: dependencias.Sesiones,
		registro: dependencias.Registro,
	}
}

// Montar registra las rutas de portería para los roles autorizados.
func (m *Manejador) Montar(enrutador chi.Router) {
	enrutador.Group(func(autorizados chi.Router) {
		autorizados.Use(intermediarios.RequerirRol(m.sesiones,
			intermediarios.RolAdministrador, intermediarios.RolPorteria))
		autorizados.Get("/porteria", m.panel)
		autorizados.Post("/porteria/ingreso", m.registrar)
	})
}

func (m *Manejador) panel(respuesta http.ResponseWriter, peticion *http.Request) {
	m.renderizar(respuesta, peticion, nil)
}

func (m *Manejador) registrar(respuesta http.ResponseWriter, peticion *http.Request) {
	ctx := peticion.Context()
	resultado, err := m.servicio.Registrar(ctx, intermediarios.UsuarioID(ctx), peticion.PostFormValue("token"))
	if err != nil {
		motivo := mensajeDeError(err)
		if motivo == "" {
			m.registro.Error("registrando el ingreso por QR", "error", err)
			motivo = "No se pudo registrar el ingreso. Intente nuevamente."
		}
		m.renderizar(respuesta, peticion, &vista.Resultado{Motivo: motivo})
		return
	}
	m.renderizar(respuesta, peticion, &vista.Resultado{
		Registrado:    resultado.Registrado,
		Alumno:        resultado.Alumno,
		Identificador: resultado.Identificador,
		Seccion:       resultado.Seccion,
		Hora:          resultado.Hora,
		Estado:        resultado.Estado,
	})
}

func (m *Manejador) renderizar(respuesta http.ResponseWriter, peticion *http.Request, resultado *vista.Resultado) {
	ctx := peticion.Context()
	token, err := intermediarios.TokenCSRF(ctx, m.sesiones)
	if err != nil {
		http.Error(respuesta, "No se pudo preparar el formulario.", http.StatusInternalServerError)
		return
	}
	respuesta.Header().Set("Content-Type", "text/html; charset=utf-8")
	componente := vista.Porteria(m.nombreUsuario(ctx), intermediarios.RolUsuario(ctx), token, resultado)
	if err := componente.Render(ctx, respuesta); err != nil {
		m.registro.Error("renderizando la pantalla de portería", "error", err)
	}
}

func (m *Manejador) nombreUsuario(ctx context.Context) string {
	nombre := intermediarios.NombreUsuario(ctx, m.sesiones)
	if nombre == "" {
		return "Usuario"
	}
	return nombre
}

func mensajeDeError(err error) string {
	switch {
	case errors.Is(err, ErrTokenInvalido):
		return "El código QR no es válido."
	case errors.Is(err, ErrDiaNoLaborable):
		return "Hoy no es un día laborable para la sección del alumno."
	case errors.Is(err, ErrAntesDeIngreso):
		return "Todavía no comienza el registro del turno."
	case errors.Is(err, ErrRegistroCerrado):
		return "El registro del turno ya cerró."
	case errors.Is(err, ErrSinTurno):
		return "La sección no tiene un turno vigente."
	case errors.Is(err, ErrSinHorario):
		return "El turno no tiene un horario configurado."
	case errors.Is(err, ErrSinSeccion):
		return "El alumno no tiene una sección vigente."
	case errors.Is(err, ErrIngresoDuplicado):
		return "El ingreso del día ya fue registrado."
	default:
		return ""
	}
}
