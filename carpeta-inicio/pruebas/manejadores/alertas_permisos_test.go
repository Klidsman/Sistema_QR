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

func TestAlertasPermisos_RolesDelPanel(t *testing.T) {
	sesiones := scs.New()
	roles := []string{intermediarios.RolAdministrador, intermediarios.RolDocente, intermediarios.RolPorteria}

	enrutador := chi.NewRouter()
	enrutador.Use(intermediarios.RequerirRol(sesiones, roles...))
	enrutador.Get("/alertas", func(respuesta http.ResponseWriter, _ *http.Request) {
		respuesta.WriteHeader(http.StatusOK)
	})
	enrutador.Post("/alertas", func(respuesta http.ResponseWriter, _ *http.Request) {
		respuesta.WriteHeader(http.StatusOK)
	})

	for _, rol := range []string{"administrador", "docente", "porteria"} {
		peticion := peticionConSesion(t, sesiones, http.MethodGet, "/alertas", nil, "", rol)
		respuesta := httptest.NewRecorder()
		enrutador.ServeHTTP(respuesta, peticion)
		assert.Equal(t, http.StatusOK, respuesta.Code, rol)
	}

	peticion := peticionConSesion(t, sesiones, http.MethodGet, "/alertas", nil, "", "director")
	respuesta := httptest.NewRecorder()
	enrutador.ServeHTTP(respuesta, peticion)
	assert.Equal(t, http.StatusForbidden, respuesta.Code)
}
