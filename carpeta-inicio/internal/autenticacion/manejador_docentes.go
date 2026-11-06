package autenticacion

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"asistencia/internal/plataforma/intermediarios"
	vista "asistencia/web/plantillas/docentes"

	"github.com/go-chi/chi/v5"
)

func (m *Manejador) listarDocentes(respuesta http.ResponseWriter, peticion *http.Request) {
	m.renderizarDocentes(respuesta, peticion, http.StatusOK, "", peticion.URL.Query().Get("ok"))
}

func (m *Manejador) crearDocente(respuesta http.ResponseWriter, peticion *http.Request) {
	if err := peticion.ParseForm(); err != nil {
		m.renderizarDocentes(respuesta, peticion, http.StatusBadRequest, "Solicitud no válida", "")
		return
	}
	datos := DatosDocente{
		Nombre:      strings.TrimSpace(peticion.PostFormValue("nombre")),
		Correo:      peticion.PostFormValue("correo"),
		Contrasenia: peticion.PostFormValue("contrasenia"),
	}
	if _, err := m.docentes.CrearDocente(peticion.Context(), m.usuarioID(peticion.Context()), datos); err != nil {
		m.renderizarDocentes(respuesta, peticion, http.StatusUnprocessableEntity, mensajeErrorDocente(err), "")
		return
	}
	http.Redirect(respuesta, peticion, "/docentes?ok=creado", http.StatusSeeOther)
}

func (m *Manejador) restablecerClaveDocente(respuesta http.ResponseWriter, peticion *http.Request) {
	docenteID, err := strconv.ParseInt(chi.URLParam(peticion, "id"), 10, 64)
	if err != nil || docenteID <= 0 {
		m.renderizarDocentes(respuesta, peticion, http.StatusUnprocessableEntity,
			mensajeErrorDocente(ErrDocenteNoEncontrado), "")
		return
	}
	if err := m.docentes.RestablecerClave(peticion.Context(), m.usuarioID(peticion.Context()),
		docenteID, peticion.PostFormValue("contrasenia")); err != nil {
		m.renderizarDocentes(respuesta, peticion, http.StatusUnprocessableEntity, mensajeErrorDocente(err), "")
		return
	}
	http.Redirect(respuesta, peticion, "/docentes?ok=clave", http.StatusSeeOther)
}

func (m *Manejador) desactivarDocente(respuesta http.ResponseWriter, peticion *http.Request) {
	docenteID, err := strconv.ParseInt(chi.URLParam(peticion, "id"), 10, 64)
	if err != nil || docenteID <= 0 {
		m.renderizarDocentes(respuesta, peticion, http.StatusUnprocessableEntity,
			mensajeErrorDocente(ErrDocenteNoEncontrado), "")
		return
	}
	if err := m.docentes.DesactivarDocente(peticion.Context(), m.usuarioID(peticion.Context()), docenteID); err != nil {
		m.renderizarDocentes(respuesta, peticion, http.StatusUnprocessableEntity, mensajeErrorDocente(err), "")
		return
	}
	http.Redirect(respuesta, peticion, "/docentes?ok=desactivado", http.StatusSeeOther)
}

func (m *Manejador) renderizarDocentes(respuesta http.ResponseWriter, peticion *http.Request, estado int, mensajeError, codigoCorrecto string) {
	docentes, err := m.docentes.ListarDocentes(peticion.Context())
	if err != nil {
		m.registro.Error("no se pudieron listar los docentes", "error", err)
		http.Error(respuesta, "No se pudieron cargar los docentes", http.StatusInternalServerError)
		return
	}
	vistas := make([]vista.DocenteVista, 0, len(docentes))
	for _, docente := range docentes {
		vistas = append(vistas, vista.DocenteVista{
			ID: docente.ID, Nombre: docente.Nombre, Correo: docente.Correo, Activo: docente.Activo,
		})
	}

	respuesta.Header().Set("Content-Type", "text/html; charset=utf-8")
	respuesta.WriteHeader(estado)
	componente := vista.Listado(
		m.nombreUsuario(peticion.Context()),
		intermediarios.RolUsuario(peticion.Context()),
		m.tokenCSRF(peticion.Context()),
		vistas,
		mensajeCorrectoDocente(codigoCorrecto),
		mensajeError,
	)
	if err := componente.Render(peticion.Context(), respuesta); err != nil {
		m.registro.Error("no se pudo renderizar la lista de docentes", "error", err)
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

func mensajeErrorDocente(err error) string {
	switch {
	case errors.Is(err, ErrDocenteExistente):
		return "Ya existe una cuenta con ese correo."
	case errors.Is(err, ErrDocenteNoEncontrado):
		return "El docente no existe."
	case errors.Is(err, ErrDatosDocenteInvalidos):
		return "Revise los datos: nombre, correo válido y contraseña de al menos 8 caracteres."
	default:
		return "No se pudo guardar el cambio. Intente de nuevo."
	}
}

func mensajeCorrectoDocente(codigo string) string {
	switch codigo {
	case "creado":
		return "La cuenta de docente se creó correctamente."
	case "clave":
		return "La contraseña se restableció correctamente."
	case "desactivado":
		return "La cuenta de docente se desactivó correctamente."
	default:
		return ""
	}
}
