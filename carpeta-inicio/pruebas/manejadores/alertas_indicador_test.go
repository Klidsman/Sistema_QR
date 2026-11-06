package manejadores

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAlertasIndicador_SoloRolesAutorizados(t *testing.T) {
	servicio := &servicioAlertasFalso{nuevas: 2}
	manejador, sesiones := nuevoManejadorAlertas(t, servicio)

	peticion := peticionConSesion(t, sesiones, http.MethodGet, "/alertas/indicador", nil, "", "docente")
	respuesta := ejecutarAlertas(manejador, sesiones, peticion, false)
	require.Equal(t, http.StatusOK, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "indicador-alertas")
	assert.Contains(t, respuesta.Body.String(), "2")

	peticion = peticionConSesion(t, sesiones, http.MethodGet, "/alertas/indicador", nil, "", "director")
	respuesta = ejecutarAlertas(manejador, sesiones, peticion, false)
	assert.Equal(t, http.StatusForbidden, respuesta.Code)
}
