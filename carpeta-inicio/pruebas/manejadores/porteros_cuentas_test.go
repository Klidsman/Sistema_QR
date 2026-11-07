package manejadores

import (
	"net/http"
	"net/url"
	"testing"

	"asistencia/internal/autenticacion"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPorteros_AdministradorListaYCrea(t *testing.T) {
	falso := &docentesFalso{docentes: []autenticacion.Cuenta{{
		ID: 11, Nombre: "Portero Uno", Correo: "portero@colegio.pe", Activo: true, Rol: "porteria",
	}}}
	manejador, sesiones := nuevoManejadorDocentes(t, falso)

	peticion := peticionConSesion(t, sesiones, http.MethodGet, "/porteros", nil, "", "administrador")
	respuesta := ejecutarDocentes(manejador, peticion)
	require.Equal(t, http.StatusOK, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "Portero Uno")
	assert.Contains(t, respuesta.Body.String(), "Crear portero")

	peticion = peticionPostFormulario(t, sesiones, "/porteros", url.Values{
		"nombre": {"Portero Nuevo"}, "correo": {"nuevo.portero@colegio.pe"}, "contrasenia": {"clave-segura"},
	}, "administrador")
	respuesta = ejecutarDocentes(manejador, peticion)
	require.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Contains(t, respuesta.Header().Get("Location"), "/porteros?ok=creado")
	require.NotNil(t, falso.creado)
	assert.Equal(t, "porteria", falso.rolCreado)
	assert.Equal(t, "nuevo.portero@colegio.pe", falso.creado.Correo)
}

func TestPorteros_SoloAdministradorAccede(t *testing.T) {
	falso := &docentesFalso{}
	manejador, sesiones := nuevoManejadorDocentes(t, falso)

	for _, rol := range []string{"director", "docente", "porteria"} {
		peticion := peticionConSesion(t, sesiones, http.MethodGet, "/porteros", nil, "", rol)
		respuesta := ejecutarDocentes(manejador, peticion)
		assert.Equal(t, http.StatusForbidden, respuesta.Code, rol)
	}

	peticion := peticionPostFormulario(t, sesiones, "/porteros", url.Values{
		"nombre": {"Portero"}, "correo": {"portero@colegio.pe"}, "contrasenia": {"clave-segura"},
	}, "director")
	respuesta := ejecutarDocentes(manejador, peticion)
	assert.Equal(t, http.StatusForbidden, respuesta.Code)
	assert.Nil(t, falso.creado)
}

func TestPorteros_AdministradorRestableceYDesactiva(t *testing.T) {
	falso := &docentesFalso{}
	manejador, sesiones := nuevoManejadorDocentes(t, falso)

	peticion := peticionPostFormulario(t, sesiones, "/porteros/11/clave",
		url.Values{"contrasenia": {"nueva-clave"}}, "administrador")
	respuesta := ejecutarDocentes(manejador, peticion)
	require.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Equal(t, int64(11), falso.claveCambiada)

	peticion = peticionPostFormulario(t, sesiones, "/porteros/11/desactivacion",
		url.Values{}, "administrador")
	respuesta = ejecutarDocentes(manejador, peticion)
	require.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Equal(t, int64(11), falso.desactivado)
}
