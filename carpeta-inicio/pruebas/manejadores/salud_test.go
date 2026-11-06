package manejadores

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"asistencia/internal/plataforma/salud"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type verificadorFalso struct {
	err error
}

func (v verificadorFalso) Ping(ctx context.Context) error {
	return v.err
}

func TestSalud_RespondeCorrectoConBaseDisponible(t *testing.T) {
	manejador := salud.Manejador(verificadorFalso{})
	respuesta := httptest.NewRecorder()

	manejador.ServeHTTP(respuesta, httptest.NewRequest(http.MethodGet, "/salud", nil))

	require.Equal(t, http.StatusOK, respuesta.Code)
	var cuerpo map[string]string
	require.NoError(t, json.Unmarshal(respuesta.Body.Bytes(), &cuerpo))
	assert.Equal(t, "correcto", cuerpo["estado"])
	assert.Equal(t, "correcta", cuerpo["base_datos"])
}

func TestSalud_RespondeErrorSiLaBaseNoResponde(t *testing.T) {
	manejador := salud.Manejador(verificadorFalso{err: errors.New("sin conexión")})
	respuesta := httptest.NewRecorder()

	manejador.ServeHTTP(respuesta, httptest.NewRequest(http.MethodGet, "/salud", nil))

	assert.Equal(t, http.StatusServiceUnavailable, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "error")
}
