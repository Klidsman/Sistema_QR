package manejadores

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"asistencia/internal/plataforma/intermediarios"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func ejecutarConRol(sesiones *scs.SessionManager, peticion *http.Request, roles ...string) *httptest.ResponseRecorder {
	enrutador := chi.NewRouter()
	enrutador.Use(intermediarios.RequerirRol(sesiones, roles...))
	enrutador.Get("/prueba", func(respuesta http.ResponseWriter, _ *http.Request) {
		respuesta.WriteHeader(http.StatusOK)
	})
	enrutador.Post("/prueba", func(respuesta http.ResponseWriter, _ *http.Request) {
		respuesta.WriteHeader(http.StatusOK)
	})
	respuesta := httptest.NewRecorder()
	enrutador.ServeHTTP(respuesta, peticion)
	return respuesta
}

func TestAsistenciaPermisos_DocentePuedeLeerYEscribir(t *testing.T) {
	sesiones := scs.New()
	roles := []string{intermediarios.RolDocente, intermediarios.RolAdministrador}

	for _, metodo := range []string{http.MethodGet, http.MethodPost} {
		peticion := peticionConSesion(t, sesiones, metodo, "/prueba", nil, "", "docente")
		respuesta := ejecutarConRol(sesiones, peticion, roles...)
		assert.Equal(t, http.StatusOK, respuesta.Code, metodo)
	}
}

func TestAsistenciaPermisos_DirectorSoloLee(t *testing.T) {
	sesiones := scs.New()
	roles := []string{intermediarios.RolDocente, intermediarios.RolAdministrador, intermediarios.RolDirector}

	peticion := peticionConSesion(t, sesiones, http.MethodGet, "/prueba", nil, "", "director")
	assert.Equal(t, http.StatusOK, ejecutarConRol(sesiones, peticion, roles...).Code)

	peticion = peticionConSesion(t, sesiones, http.MethodPost, "/prueba", nil, "", "director")
	assert.Equal(t, http.StatusForbidden, ejecutarConRol(sesiones, peticion, roles...).Code)
}

func TestAsistenciaPermisos_PorteriaSinAcceso(t *testing.T) {
	sesiones := scs.New()
	roles := []string{intermediarios.RolDocente, intermediarios.RolAdministrador, intermediarios.RolDirector}

	for _, metodo := range []string{http.MethodGet, http.MethodPost} {
		peticion := peticionConSesion(t, sesiones, metodo, "/prueba", nil, "", "porteria")
		respuesta := ejecutarConRol(sesiones, peticion, roles...)
		assert.Equal(t, http.StatusForbidden, respuesta.Code, metodo)
	}
}

func TestAsistenciaPermisos_SinSesionRedirige(t *testing.T) {
	sesiones := scs.New()
	peticion := peticionConSesion(t, sesiones, http.MethodGet, "/prueba", nil, "", "")
	sesiones.Remove(peticion.Context(), "usuario_id")

	respuesta := ejecutarConRol(sesiones, peticion, intermediarios.RolDocente)

	assert.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Equal(t, "/ingreso", respuesta.Header().Get("Location"))
}
