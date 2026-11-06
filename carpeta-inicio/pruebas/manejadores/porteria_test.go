package manejadores

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"asistencia/internal/plataforma/intermediarios"
	"asistencia/internal/porteria"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPorteria_PantallaYRoles(t *testing.T) {
	manejador, sesiones := nuevoManejadorPorteria(&servicioPorteriaFalso{})

	for _, rol := range []string{"porteria", "administrador"} {
		peticion := peticionConSesion(t, sesiones, http.MethodGet, "/porteria", nil, "", rol)
		respuesta := ejecutarPorteria(manejador, sesiones, peticion, false)
		assert.Equal(t, http.StatusOK, respuesta.Code, rol)
		assert.Contains(t, respuesta.Body.String(), "Registro de ingreso")
	}

	for _, rol := range []string{"director", "docente"} {
		peticion := peticionConSesion(t, sesiones, http.MethodGet, "/porteria", nil, "", rol)
		respuesta := ejecutarPorteria(manejador, sesiones, peticion, false)
		assert.Equal(t, http.StatusForbidden, respuesta.Code, rol)
	}
}

func TestPorteria_RegistraYExplicaRechazos(t *testing.T) {
	servicio := &servicioPorteriaFalso{resultado: porteria.Resultado{
		Registrado: true,
		Alumno:     "Perez, Ana",
		Seccion:    "1.° A",
		Hora:       "07:45",
		Estado:     "Tardanza",
	}}
	manejador, sesiones := nuevoManejadorPorteria(servicio)

	peticion := peticionFormularioPorteria(t, sesiones, "token-valido", "porteria")
	sesiones.Put(peticion.Context(), "token_csrf", "token-correcto")
	respuesta := ejecutarPorteria(manejador, sesiones, peticion, true)
	require.Equal(t, http.StatusOK, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "Perez, Ana")
	assert.Contains(t, respuesta.Body.String(), "Tardanza")

	servicio.error = porteria.ErrTokenInvalido
	peticion = peticionFormularioPorteria(t, sesiones, "no-existe", "porteria")
	sesiones.Put(peticion.Context(), "token_csrf", "token-correcto")
	respuesta = ejecutarPorteria(manejador, sesiones, peticion, true)
	require.Equal(t, http.StatusOK, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "El código QR no es válido.")
}

func TestPorteria_ExigeCSRF(t *testing.T) {
	manejador, sesiones := nuevoManejadorPorteria(&servicioPorteriaFalso{})

	peticion := peticionFormularioPorteria(t, sesiones, "token-valido", "porteria")
	respuesta := ejecutarPorteria(manejador, sesiones, peticion, true)

	assert.Equal(t, http.StatusForbidden, respuesta.Code)
}

func nuevoManejadorPorteria(servicio porteria.ServicioPorteria) (*porteria.Manejador, *scs.SessionManager) {
	sesiones := scs.New()
	return porteria.NuevoManejador(porteria.DependenciasManejador{
		Servicio: servicio,
		Sesiones: sesiones,
		Registro: registroDePrueba(),
	}), sesiones
}

func peticionFormularioPorteria(t *testing.T, sesiones *scs.SessionManager, token, rol string) *http.Request {
	t.Helper()
	cuerpo := url.Values{"token_csrf": {"token-correcto"}, "token": {token}}.Encode()
	return peticionConSesion(t, sesiones, http.MethodPost, "/porteria/ingreso",
		strings.NewReader(cuerpo), "application/x-www-form-urlencoded", rol)
}

func ejecutarPorteria(manejador *porteria.Manejador, sesiones *scs.SessionManager, peticion *http.Request, conCSRF bool) *httptest.ResponseRecorder {
	enrutador := chi.NewRouter()
	if conCSRF {
		enrutador.Use(intermediarios.ProteccionCSRF(sesiones))
	}
	manejador.Montar(enrutador)
	respuesta := httptest.NewRecorder()
	enrutador.ServeHTTP(respuesta, peticion)
	return respuesta
}

type servicioPorteriaFalso struct {
	resultado porteria.Resultado
	error     error
}

func (s *servicioPorteriaFalso) Registrar(context.Context, int64, string) (porteria.Resultado, error) {
	return s.resultado, s.error
}
