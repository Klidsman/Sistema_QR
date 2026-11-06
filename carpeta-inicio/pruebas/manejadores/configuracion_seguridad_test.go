package manejadores

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"asistencia/internal/academico"
	"asistencia/internal/plataforma/intermediarios"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ejecutarConCSRF(manejador *academico.Manejador, sesiones *scs.SessionManager, peticion *http.Request) *httptest.ResponseRecorder {
	enrutador := chi.NewRouter()
	enrutador.Use(intermediarios.ProteccionCSRF(sesiones))
	manejador.Montar(enrutador)
	respuesta := httptest.NewRecorder()
	enrutador.ServeHTTP(respuesta, peticion)
	return respuesta
}

func TestConfiguracion_RechazaPostSinTokenCSRF(t *testing.T) {
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen: academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
	})
	peticion := peticionFormulario(t, sesiones, "/configuracion/grados",
		url.Values{"nombre": {"3.°"}}, "administrador")

	respuesta := ejecutarConCSRF(manejador, sesiones, peticion)

	assert.Equal(t, http.StatusForbidden, respuesta.Code)
}

func TestConfiguracion_RechazaTokenCSRFIncorrecto(t *testing.T) {
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen: academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
	})
	peticion := peticionFormulario(t, sesiones, "/configuracion/grados",
		url.Values{"nombre": {"3.°"}, "token_csrf": {"token-falso"}}, "administrador")
	sesiones.Put(peticion.Context(), "token_csrf", "token-correcto")

	respuesta := ejecutarConCSRF(manejador, sesiones, peticion)

	assert.Equal(t, http.StatusForbidden, respuesta.Code)
}

func TestConfiguracion_AceptaTokenCSRFCorrecto(t *testing.T) {
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen: academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
		crearGrado: func(academico.DatosGrado) (int64, error) {
			return 1, nil
		},
	})
	peticion := peticionFormulario(t, sesiones, "/configuracion/grados",
		url.Values{"nombre": {"3.°"}, "token_csrf": {"token-correcto"}}, "administrador")
	sesiones.Put(peticion.Context(), "token_csrf", "token-correcto")

	respuesta := ejecutarConCSRF(manejador, sesiones, peticion)

	require.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Equal(t, "/configuracion/grados?ok=creado", respuesta.Header().Get("Location"))
}
