package autenticacion

import (
	"log/slog"
	"net"
	"net/http"
	"strings"

	"asistencia/internal/plataforma/intermediarios"
	"asistencia/web/plantillas/paginas"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
)

const mensajeIngresoInvalido = "Correo o contraseña incorrectos."
const mensajeDemasiadosIntentos = "Demasiados intentos fallidos. Espere 15 minutos e intente de nuevo."

// Manejador atiende las pantallas y acciones de ingreso y salida.
type Manejador struct {
	servicio *Servicio
	sesiones *scs.SessionManager
	limite   *intermediarios.LimiteIntentos
	registro *slog.Logger
}

// NuevoManejador crea el manejador de ingreso con sus dependencias.
func NuevoManejador(servicio *Servicio, sesiones *scs.SessionManager, limite *intermediarios.LimiteIntentos, registro *slog.Logger) *Manejador {
	return &Manejador{servicio: servicio, sesiones: sesiones, limite: limite, registro: registro}
}

// Montar registra las rutas de ingreso y salida en el enrutador.
func (m *Manejador) Montar(enrutador chi.Router) {
	enrutador.Get("/ingreso", m.mostrarIngreso)
	enrutador.Post("/ingreso", m.procesarIngreso)
	enrutador.Post("/salir", m.salir)
}

func (m *Manejador) mostrarIngreso(respuesta http.ResponseWriter, peticion *http.Request) {
	m.renderizarIngreso(respuesta, peticion, http.StatusOK, "")
}

func (m *Manejador) procesarIngreso(respuesta http.ResponseWriter, peticion *http.Request) {
	if err := peticion.ParseForm(); err != nil {
		http.Error(respuesta, "Solicitud no válida", http.StatusBadRequest)
		return
	}
	correo := strings.ToLower(strings.TrimSpace(peticion.PostFormValue("correo")))
	clave := claveDeIntento(peticion, correo)

	if !m.limite.Permitir(clave) {
		m.renderizarIngreso(respuesta, peticion, http.StatusTooManyRequests, mensajeDemasiadosIntentos)
		return
	}

	usuario, err := m.servicio.Autenticar(peticion.Context(), correo, peticion.PostFormValue("contrasenia"))
	if err != nil {
		m.limite.RegistrarFallo(clave)
		m.registro.Warn("intento de ingreso fallido", "correo", correo)
		m.renderizarIngreso(respuesta, peticion, http.StatusUnauthorized, mensajeIngresoInvalido)
		return
	}

	m.limite.Limpiar(clave)
	if err := intermediarios.IniciarSesionUsuario(peticion.Context(), m.sesiones, usuario.ID, usuario.NombreCompleto, usuario.Rol); err != nil {
		m.registro.Error("no se pudo iniciar la sesión", "error", err)
		http.Error(respuesta, "No se pudo iniciar la sesión", http.StatusInternalServerError)
		return
	}
	m.registro.Info("ingreso correcto", "usuario_id", usuario.ID, "rol", usuario.Rol)
	http.Redirect(respuesta, peticion, "/alumnos", http.StatusSeeOther)
}

func (m *Manejador) salir(respuesta http.ResponseWriter, peticion *http.Request) {
	if err := intermediarios.CerrarSesionUsuario(peticion.Context(), m.sesiones); err != nil {
		m.registro.Error("no se pudo cerrar la sesión", "error", err)
	}
	http.Redirect(respuesta, peticion, "/ingreso", http.StatusSeeOther)
}

func (m *Manejador) renderizarIngreso(respuesta http.ResponseWriter, peticion *http.Request, estado int, mensaje string) {
	token, err := intermediarios.TokenCSRF(peticion.Context(), m.sesiones)
	if err != nil {
		http.Error(respuesta, "No se pudo preparar la página", http.StatusInternalServerError)
		return
	}
	respuesta.Header().Set("Content-Type", "text/html; charset=utf-8")
	respuesta.WriteHeader(estado)
	if err := paginas.Ingreso(token, mensaje).Render(peticion.Context(), respuesta); err != nil {
		m.registro.Error("no se pudo renderizar el ingreso", "error", err)
	}
}

// claveDeIntento agrupa los intentos por IP de origen y correo, detrás de un proxy usa X-Forwarded-For.
func claveDeIntento(peticion *http.Request, correo string) string {
	origen := peticion.RemoteAddr
	if host, _, err := net.SplitHostPort(origen); err == nil {
		origen = host
	}
	if reenviado := peticion.Header.Get("X-Forwarded-For"); reenviado != "" {
		origen = strings.TrimSpace(strings.Split(reenviado, ",")[0])
	}
	return origen + "|" + correo
}
