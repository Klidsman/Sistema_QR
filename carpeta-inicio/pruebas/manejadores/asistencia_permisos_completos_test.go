package manejadores

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"asistencia/internal/asistencia"
	"asistencia/internal/plataforma/intermediarios"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsistenciaPermisos_ConsultasYParametros(t *testing.T) {
	falso := &asistenciaFalsa{detalle: detalleClaseDePrueba(t), exportados: exportacionDePrueba()}
	manejador, sesiones := nuevoManejadorAsistencia(t, falso)

	casos := []struct {
		rol    string
		ruta   string
		estado int
	}{
		{"director", "/asistencia/consulta", http.StatusOK},
		{"director", "/configuracion/asistencia", http.StatusOK},
		{"director", "/asistencia/exportar", http.StatusOK},
		{"docente", "/asistencia/consulta", http.StatusForbidden},
		{"docente", "/configuracion/asistencia", http.StatusForbidden},
		{"docente", "/asistencia/exportar", http.StatusOK},
		{"docente", "/asistencia/clase?fecha=2026-11-06&seccion_id=5&curso_id=2&bloque=1", http.StatusOK},
		{"porteria", "/asistencia/exportar", http.StatusForbidden},
	}
	for _, caso := range casos {
		t.Run(caso.rol+" "+caso.ruta, func(t *testing.T) {
			peticion := peticionConSesion(t, sesiones, http.MethodGet, caso.ruta, nil, "", caso.rol)
			respuesta := ejecutarAsistencia(manejador, peticion)
			assert.Equal(t, caso.estado, respuesta.Code)
		})
	}
}

func TestAsistenciaPermisos_DirectorNoPuedeEscribir(t *testing.T) {
	manejador, sesiones := nuevoManejadorAsistencia(t, &asistenciaFalsa{})
	for _, ruta := range []string{"/asistencia/consulta/correccion", "/configuracion/asistencia"} {
		peticion := peticionPostFormulario(t, sesiones, ruta, url.Values{}, "director")
		respuesta := ejecutarAsistencia(manejador, peticion)
		assert.Equal(t, http.StatusForbidden, respuesta.Code, ruta)
	}
}

func ejecutarAsistenciaConCSRF(manejador *asistencia.Manejador, sesiones *scs.SessionManager, peticion *http.Request) *httptest.ResponseRecorder {
	enrutador := chi.NewRouter()
	enrutador.Use(intermediarios.ProteccionCSRF(sesiones))
	manejador.Montar(enrutador)
	respuesta := httptest.NewRecorder()
	enrutador.ServeHTTP(respuesta, peticion)
	return respuesta
}

func TestAsistenciaCSRF_RechazaSinTokenYAceptaConToken(t *testing.T) {
	falso := &asistenciaFalsa{detalle: detalleClaseDePrueba(t)}
	manejador, sesiones := nuevoManejadorAsistencia(t, falso)

	peticion := peticionPostFormulario(t, sesiones, "/asistencia/clase", url.Values{
		"fecha":      {"2026-11-06"},
		"seccion_id": {"5"},
		"curso_id":   {"2"},
		"bloque":     {"1"},
		"estado_10":  {asistencia.EstadoPresente},
	}, "docente")
	respuesta := ejecutarAsistenciaConCSRF(manejador, sesiones, peticion)
	assert.Equal(t, http.StatusForbidden, respuesta.Code)

	valores := url.Values{
		"fecha":      {"2026-11-06"},
		"seccion_id": {"5"},
		"curso_id":   {"2"},
		"bloque":     {"1"},
		"estado_10":  {asistencia.EstadoPresente},
		"token_csrf": {"token-correcto"},
	}
	peticion = peticionPostFormulario(t, sesiones, "/asistencia/clase", valores, "docente")
	sesiones.Put(peticion.Context(), "token_csrf", "token-correcto")
	respuesta = ejecutarAsistenciaConCSRF(manejador, sesiones, peticion)
	require.Equal(t, http.StatusSeeOther, respuesta.Code)
}
