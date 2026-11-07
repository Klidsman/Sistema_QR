package autenticacion

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"asistencia/internal/plataforma/intermediarios"
	vista "asistencia/web/plantillas/cuentas"

	"github.com/go-chi/chi/v5"
)

// configuracionCuenta describe cómo se presenta la gestión de un rol.
type configuracionCuenta struct {
	Rol        string
	Titulo     string
	RutaBase   string
	Singular   string
	BotonCrear string
}

var cuentaDocente = configuracionCuenta{
	Rol:        intermediarios.RolDocente,
	Titulo:     "Docentes",
	RutaBase:   "/docentes",
	Singular:   "docente",
	BotonCrear: "Crear docente",
}

var cuentaPortero = configuracionCuenta{
	Rol:        intermediarios.RolPorteria,
	Titulo:     "Porteros",
	RutaBase:   "/porteros",
	Singular:   "portero",
	BotonCrear: "Crear portero",
}

func (m *Manejador) listarDocentes(respuesta http.ResponseWriter, peticion *http.Request) {
	m.renderizarCuentas(respuesta, peticion, cuentaDocente, http.StatusOK, "", peticion.URL.Query().Get("ok"))
}

func (m *Manejador) crearDocente(respuesta http.ResponseWriter, peticion *http.Request) {
	m.crearCuenta(respuesta, peticion, cuentaDocente)
}

func (m *Manejador) restablecerClaveDocente(respuesta http.ResponseWriter, peticion *http.Request) {
	m.restablecerClaveCuenta(respuesta, peticion, cuentaDocente)
}

func (m *Manejador) desactivarDocente(respuesta http.ResponseWriter, peticion *http.Request) {
	m.desactivarCuenta(respuesta, peticion, cuentaDocente)
}

func (m *Manejador) listarPorteros(respuesta http.ResponseWriter, peticion *http.Request) {
	m.renderizarCuentas(respuesta, peticion, cuentaPortero, http.StatusOK, "", peticion.URL.Query().Get("ok"))
}

func (m *Manejador) crearPortero(respuesta http.ResponseWriter, peticion *http.Request) {
	m.crearCuenta(respuesta, peticion, cuentaPortero)
}

func (m *Manejador) restablecerClavePortero(respuesta http.ResponseWriter, peticion *http.Request) {
	m.restablecerClaveCuenta(respuesta, peticion, cuentaPortero)
}

func (m *Manejador) desactivarPortero(respuesta http.ResponseWriter, peticion *http.Request) {
	m.desactivarCuenta(respuesta, peticion, cuentaPortero)
}

func (m *Manejador) crearCuenta(respuesta http.ResponseWriter, peticion *http.Request, configuracion configuracionCuenta) {
	if err := peticion.ParseForm(); err != nil {
		m.renderizarCuentas(respuesta, peticion, configuracion, http.StatusBadRequest, "Solicitud no válida", "")
		return
	}
	datos := DatosCuenta{
		Nombre:      strings.TrimSpace(peticion.PostFormValue("nombre")),
		Correo:      peticion.PostFormValue("correo"),
		Contrasenia: peticion.PostFormValue("contrasenia"),
	}
	if _, err := m.cuentas.CrearCuenta(peticion.Context(), m.usuarioID(peticion.Context()), configuracion.Rol, datos); err != nil {
		m.renderizarCuentas(respuesta, peticion, configuracion, http.StatusUnprocessableEntity, mensajeErrorCuenta(err, configuracion), "")
		return
	}
	http.Redirect(respuesta, peticion, configuracion.RutaBase+"?ok=creado", http.StatusSeeOther)
}

func (m *Manejador) restablecerClaveCuenta(respuesta http.ResponseWriter, peticion *http.Request, configuracion configuracionCuenta) {
	cuentaID, err := strconv.ParseInt(chi.URLParam(peticion, "id"), 10, 64)
	if err != nil || cuentaID <= 0 {
		m.renderizarCuentas(respuesta, peticion, configuracion, http.StatusUnprocessableEntity,
			mensajeErrorCuenta(ErrCuentaNoEncontrada, configuracion), "")
		return
	}
	if err := m.cuentas.RestablecerClaveCuenta(peticion.Context(), m.usuarioID(peticion.Context()),
		cuentaID, configuracion.Rol, peticion.PostFormValue("contrasenia")); err != nil {
		m.renderizarCuentas(respuesta, peticion, configuracion, http.StatusUnprocessableEntity, mensajeErrorCuenta(err, configuracion), "")
		return
	}
	http.Redirect(respuesta, peticion, configuracion.RutaBase+"?ok=clave", http.StatusSeeOther)
}

func (m *Manejador) desactivarCuenta(respuesta http.ResponseWriter, peticion *http.Request, configuracion configuracionCuenta) {
	cuentaID, err := strconv.ParseInt(chi.URLParam(peticion, "id"), 10, 64)
	if err != nil || cuentaID <= 0 {
		m.renderizarCuentas(respuesta, peticion, configuracion, http.StatusUnprocessableEntity,
			mensajeErrorCuenta(ErrCuentaNoEncontrada, configuracion), "")
		return
	}
	if err := m.cuentas.DesactivarCuenta(peticion.Context(), m.usuarioID(peticion.Context()),
		cuentaID, configuracion.Rol); err != nil {
		m.renderizarCuentas(respuesta, peticion, configuracion, http.StatusUnprocessableEntity, mensajeErrorCuenta(err, configuracion), "")
		return
	}
	http.Redirect(respuesta, peticion, configuracion.RutaBase+"?ok=desactivado", http.StatusSeeOther)
}

func (m *Manejador) renderizarCuentas(respuesta http.ResponseWriter, peticion *http.Request, configuracion configuracionCuenta, estado int, mensajeError, codigoCorrecto string) {
	cuentas, err := m.cuentas.ListarCuentasPorRol(peticion.Context(), configuracion.Rol)
	if err != nil {
		m.registro.Error("no se pudieron listar las cuentas", "rol", configuracion.Rol, "error", err)
		http.Error(respuesta, "No se pudieron cargar las cuentas", http.StatusInternalServerError)
		return
	}
	vistas := make([]vista.CuentaVista, 0, len(cuentas))
	for _, cuenta := range cuentas {
		vistas = append(vistas, vista.CuentaVista{
			ID: cuenta.ID, Nombre: cuenta.Nombre, Correo: cuenta.Correo, Activo: cuenta.Activo,
		})
	}

	rolSesion := intermediarios.RolUsuario(peticion.Context())
	respuesta.Header().Set("Content-Type", "text/html; charset=utf-8")
	respuesta.WriteHeader(estado)
	componente := vista.Listado(vista.Datos{
		Usuario:       m.nombreUsuario(peticion.Context()),
		RolSesion:     rolSesion,
		TokenCSRF:     m.tokenCSRF(peticion.Context()),
		Titulo:        configuracion.Titulo,
		RutaBase:      configuracion.RutaBase,
		Singular:      configuracion.Singular,
		BotonCrear:    configuracion.BotonCrear,
		Mensaje:       mensajeCorrectoCuenta(codigoCorrecto, configuracion),
		MensajeError:  mensajeError,
		Cuentas:       vistas,
		PuedeEscribir: rolSesion == intermediarios.RolAdministrador,
	})
	if err := componente.Render(peticion.Context(), respuesta); err != nil {
		m.registro.Error("no se pudo renderizar la lista de cuentas", "error", err)
	}
}

func (m *Manejador) usuarioID(ctx context.Context) int64 {
	return intermediarios.UsuarioID(ctx)
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

func mensajeErrorCuenta(err error, configuracion configuracionCuenta) string {
	switch {
	case errors.Is(err, ErrCuentaExistente):
		return "Ya existe una cuenta con ese correo."
	case errors.Is(err, ErrCuentaNoEncontrada):
		return "El " + configuracion.Singular + " no existe."
	case errors.Is(err, ErrDatosCuentaInvalidos):
		return "Revise los datos: nombre, correo válido y contraseña de al menos 8 caracteres."
	default:
		return "No se pudo guardar el cambio. Intente de nuevo."
	}
}

func mensajeCorrectoCuenta(codigo string, configuracion configuracionCuenta) string {
	switch codigo {
	case "creado":
		return "La cuenta de " + configuracion.Singular + " se creó correctamente."
	case "clave":
		return "La contraseña se restableció correctamente."
	case "desactivado":
		return "La cuenta de " + configuracion.Singular + " se desactivó correctamente."
	default:
		return ""
	}
}
