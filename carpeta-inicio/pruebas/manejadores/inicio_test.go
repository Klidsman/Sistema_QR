package manejadores

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInicio_PublicoPresentaElColegioYLosAccesos(t *testing.T) {
	manejador, sesiones := nuevoManejadorDocentes(t, &docentesFalso{})

	respuesta := ejecutarDocentes(manejador, peticionConSesion(t, sesiones, http.MethodGet, "/", nil, "", ""))

	require.Equal(t, http.StatusOK, respuesta.Code)
	cuerpo := respuesta.Body.String()
	assert.Contains(t, cuerpo, "María Parado de Bellido")
	assert.Contains(t, cuerpo, "Ayacucho")
	assert.Contains(t, cuerpo, "Cómo funciona")
	assert.Contains(t, cuerpo, `id="acceso-administrador"`)
	assert.Contains(t, cuerpo, `id="acceso-direccion"`)
	assert.Contains(t, cuerpo, `id="acceso-porteria"`)
	assert.Contains(t, cuerpo, `id="acceso-docente"`)
	assert.Contains(t, cuerpo, "Ingresar al sistema")
	assert.NotContains(t, cuerpo, "Ir a mi panel")
}

func TestInicio_ConSesionLlevaAlPanelDelRol(t *testing.T) {
	manejador, sesiones := nuevoManejadorDocentes(t, &docentesFalso{})

	respuesta := ejecutarDocentes(manejador, peticionConSesion(t, sesiones, http.MethodGet, "/", nil, "", "porteria"))

	require.Equal(t, http.StatusOK, respuesta.Code)
	cuerpo := respuesta.Body.String()
	assert.Contains(t, cuerpo, "Ir a mi panel")
	assert.Contains(t, cuerpo, `href="/porteria"`)
	assert.Contains(t, cuerpo, "Continuar como Admin de prueba")
}
