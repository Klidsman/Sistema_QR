package autenticacion

import (
	"context"
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

// ServicioCuentas son las operaciones de cuentas docentes y de portería que usa el manejador.
type ServicioCuentas interface {
	ListarCuentasPorRol(ctx context.Context, rol string) ([]Cuenta, error)
	CrearCuenta(ctx context.Context, usuarioID int64, rol string, datos DatosCuenta) (int64, error)
	DesactivarCuenta(ctx context.Context, usuarioID, cuentaID int64, rol string) error
	RestablecerClaveCuenta(ctx context.Context, usuarioID, cuentaID int64, rol, clave string) error
}

// Manejador atiende las pantallas y acciones de ingreso, salida y cuentas.
type Manejador struct {
	servicio *Servicio
	cuentas  ServicioCuentas
	sesiones *scs.SessionManager
	limite   *intermediarios.LimiteIntentos
	registro *slog.Logger
}

// NuevoManejador crea el manejador de ingreso con sus dependencias.
func NuevoManejador(servicio *Servicio, sesiones *scs.SessionManager, limite *intermediarios.LimiteIntentos, registro *slog.Logger) *Manejador {
	return &Manejador{servicio: servicio, cuentas: servicio, sesiones: sesiones, limite: limite, registro: registro}
}

// NuevoManejadorConCuentas permite inyectar el servicio de cuentas en las pruebas.
func NuevoManejadorConCuentas(servicio *Servicio, cuentas ServicioCuentas, sesiones *scs.SessionManager, limite *intermediarios.LimiteIntentos, registro *slog.Logger) *Manejador {
	manejador := NuevoManejador(servicio, sesiones, limite, registro)
	manejador.cuentas = cuentas
	return manejador
}

// NuevoManejadorConDocentes se mantiene por compatibilidad con las pruebas existentes.
func NuevoManejadorConDocentes(servicio *Servicio, cuentas ServicioCuentas, sesiones *scs.SessionManager, limite *intermediarios.LimiteIntentos, registro *slog.Logger) *Manejador {
	return NuevoManejadorConCuentas(servicio, cuentas, sesiones, limite, registro)
}

// Montar registra las rutas de ingreso, salida y gestión de cuentas en el enrutador.
func (m *Manejador) Montar(enrutador chi.Router) {
	enrutador.Get("/", m.inicio)
	enrutador.Get("/ingreso", m.mostrarIngreso)
	enrutador.Post("/ingreso", m.procesarIngreso)
	enrutador.Post("/salir", m.salir)

	enrutador.Group(func(lectura chi.Router) {
		lectura.Use(intermediarios.RequerirRol(m.sesiones,
			intermediarios.RolAdministrador, intermediarios.RolDirector))
		lectura.Get("/docentes", m.listarDocentes)
	})

	enrutador.Group(func(escritura chi.Router) {
		escritura.Use(intermediarios.RequerirRol(m.sesiones, intermediarios.RolAdministrador))
		escritura.Post("/docentes", m.crearDocente)
		escritura.Post("/docentes/{id}/clave", m.restablecerClaveDocente)
		escritura.Post("/docentes/{id}/desactivacion", m.desactivarDocente)
	})

	enrutador.Group(func(porteros chi.Router) {
		porteros.Use(intermediarios.RequerirRol(m.sesiones, intermediarios.RolAdministrador))
		porteros.Get("/porteros", m.listarPorteros)
		porteros.Post("/porteros", m.crearPortero)
		porteros.Post("/porteros/{id}/clave", m.restablecerClavePortero)
		porteros.Post("/porteros/{id}/desactivacion", m.desactivarPortero)
	})
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
	destino := intermediarios.RutaInicioPorRol(usuario.Rol)
	if retorno := intermediarios.UrlRetorno(peticion.Context(), m.sesiones); retorno != "" &&
		intermediarios.RutaPermitidaParaRol(usuario.Rol, retorno) {
		destino = retorno
	}
	http.Redirect(respuesta, peticion, destino, http.StatusSeeOther)
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
