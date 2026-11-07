package manejadores

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"asistencia/internal/alumnos"
	"asistencia/internal/justificaciones"
	"asistencia/internal/plataforma/intermediarios"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJustificaciones_PermisosDelPanel(t *testing.T) {
	manejador, sesiones := nuevoManejadorJustificaciones(&servicioJustificacionesFalso{})

	for _, rol := range []string{"administrador", "director", "docente"} {
		peticion := peticionConSesion(t, sesiones, http.MethodGet, "/justificaciones", nil, "", rol)
		respuesta := ejecutarJustificaciones(manejador, sesiones, peticion, false)
		assert.Equal(t, http.StatusOK, respuesta.Code, rol)
	}

	peticion := peticionConSesion(t, sesiones, http.MethodGet, "/justificaciones", nil, "", "porteria")
	respuesta := ejecutarJustificaciones(manejador, sesiones, peticion, false)
	assert.Equal(t, http.StatusForbidden, respuesta.Code)
}

func TestJustificaciones_SoloResolutoresAprobanyAnulan(t *testing.T) {
	servicio := &servicioJustificacionesFalso{
		justificacion: justificaciones.Justificacion{
			ID: 4, AlumnoID: 5, SeccionID: 2, Estado: justificaciones.EstadoPendiente,
			RegistranteID: 1, CreadaEn: time.Now(),
		},
	}
	manejador, sesiones := nuevoManejadorJustificaciones(servicio)

	peticion := peticionFormularioJustificacion(t, sesiones, "/justificaciones/4/aprobar", "director")
	sesiones.Put(peticion.Context(), "token_csrf", "token-correcto")
	respuesta := ejecutarJustificaciones(manejador, sesiones, peticion, true)
	t.Log(respuesta.Body.String())
	require.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Equal(t, int64(4), servicio.resuelta)

	peticion = peticionFormularioJustificacion(t, sesiones, "/justificaciones/4/aprobar", "docente")
	sesiones.Put(peticion.Context(), "token_csrf", "token-correcto")
	respuesta = ejecutarJustificaciones(manejador, sesiones, peticion, true)
	assert.Equal(t, http.StatusForbidden, respuesta.Code)

	peticion = peticionFormularioJustificacion(t, sesiones, "/justificaciones/4/anular", "director")
	sesiones.Put(peticion.Context(), "token_csrf", "token-correcto")
	respuesta = ejecutarJustificaciones(manejador, sesiones, peticion, true)
	assert.Equal(t, http.StatusForbidden, respuesta.Code, "el director no anula")

	peticion = peticionFormularioJustificacion(t, sesiones, "/justificaciones/4/anular", "administrador")
	sesiones.Put(peticion.Context(), "token_csrf", "token-correcto")
	respuesta = ejecutarJustificaciones(manejador, sesiones, peticion, true)
	assert.Equal(t, http.StatusSeeOther, respuesta.Code)
}

func TestJustificaciones_AdministradorPuedeResolverLaSuya(t *testing.T) {
	servicio := &servicioJustificacionesFalso{
		justificacion: justificaciones.Justificacion{
			ID: 4, AlumnoID: 5, SeccionID: 2, Estado: justificaciones.EstadoPendiente,
			RegistranteID: 9, CreadaEn: time.Now(),
		},
	}
	manejador, sesiones := nuevoManejadorJustificaciones(servicio)

	peticion := peticionConSesion(t, sesiones, http.MethodGet, "/justificaciones/4", nil, "", "administrador")
	respuesta := ejecutarJustificaciones(manejador, sesiones, peticion, false)

	require.Equal(t, http.StatusOK, respuesta.Code)
	cuerpo := respuesta.Body.String()
	assert.Contains(t, cuerpo, `action="/justificaciones/4/aprobar"`)
	assert.Contains(t, cuerpo, `action="/justificaciones/4/rechazar"`)
}

func TestJustificaciones_DirectorNoResuelveLaSuya(t *testing.T) {
	servicio := &servicioJustificacionesFalso{
		justificacion: justificaciones.Justificacion{
			ID: 4, AlumnoID: 5, SeccionID: 2, Estado: justificaciones.EstadoPendiente,
			RegistranteID: 9, CreadaEn: time.Now(),
		},
	}
	manejador, sesiones := nuevoManejadorJustificaciones(servicio)

	peticion := peticionConSesion(t, sesiones, http.MethodGet, "/justificaciones/4", nil, "", "director")
	respuesta := ejecutarJustificaciones(manejador, sesiones, peticion, false)

	require.Equal(t, http.StatusOK, respuesta.Code)
	assert.NotContains(t, respuesta.Body.String(), `action="/justificaciones/4/aprobar"`)
}

func TestJustificaciones_ExigeCSRF(t *testing.T) {
	manejador, sesiones := nuevoManejadorJustificaciones(&servicioJustificacionesFalso{})

	peticion := peticionFormularioJustificacion(t, sesiones, "/justificaciones/4/aprobar", "director")
	respuesta := ejecutarJustificaciones(manejador, sesiones, peticion, true)

	assert.Equal(t, http.StatusForbidden, respuesta.Code)
}

func nuevoManejadorJustificaciones(servicio justificaciones.ServicioJustificaciones) (*justificaciones.Manejador, *scs.SessionManager) {
	sesiones := scs.New()
	return justificaciones.NuevoManejador(justificaciones.DependenciasManejador{
		Servicio: servicio,
		Sesiones: sesiones,
		Registro: registroDePrueba(),
	}), sesiones
}

func peticionFormularioJustificacion(t *testing.T, sesiones *scs.SessionManager, ruta, rol string) *http.Request {
	t.Helper()
	cuerpo := url.Values{"token_csrf": {"token-correcto"}, "comentario": {"Aprobado"}, "motivo": {"Error"}}.Encode()
	return peticionConSesion(t, sesiones, http.MethodPost, ruta,
		strings.NewReader(cuerpo), "application/x-www-form-urlencoded", rol)
}

func ejecutarJustificaciones(manejador *justificaciones.Manejador, sesiones *scs.SessionManager, peticion *http.Request, conCSRF bool) *httptest.ResponseRecorder {
	enrutador := chi.NewRouter()
	if conCSRF {
		enrutador.Use(intermediarios.ProteccionCSRF(sesiones))
	}
	manejador.Montar(enrutador)
	respuesta := httptest.NewRecorder()
	enrutador.ServeHTTP(respuesta, peticion)
	return respuesta
}

type servicioJustificacionesFalso struct {
	justificacion justificaciones.Justificacion
	resuelta      int64
}

func (s *servicioJustificacionesFalso) Registrar(context.Context, int64, string, int64, string, time.Time, time.Time, []justificaciones.EntradaAdjunto) (justificaciones.ResultadoRegistro, error) {
	return justificaciones.ResultadoRegistro{}, nil
}

func (s *servicioJustificacionesFalso) Resolver(_ context.Context, _ int64, _ string, id int64, _ bool, _ string) error {
	s.resuelta = id
	return nil
}

func (s *servicioJustificacionesFalso) Anular(context.Context, int64, string, int64, string) error {
	return nil
}

func (s *servicioJustificacionesFalso) Listar(context.Context, int64, string) ([]justificaciones.Justificacion, error) {
	return nil, nil
}

func (s *servicioJustificacionesFalso) HistorialDeAlumno(context.Context, int64, string, int64) ([]justificaciones.Justificacion, error) {
	return nil, nil
}

func (s *servicioJustificacionesFalso) ObtenerConAcceso(context.Context, int64, string, int64) (justificaciones.Justificacion, error) {
	return s.justificacion, nil
}

func (s *servicioJustificacionesFalso) AbrirAdjunto(context.Context, int64, string, int64) (justificaciones.Adjunto, []byte, error) {
	return justificaciones.Adjunto{}, nil, nil
}

func (s *servicioJustificacionesFalso) Describir(_ context.Context, lista []justificaciones.Justificacion) ([]justificaciones.JustificacionDetalle, error) {
	resultado := make([]justificaciones.JustificacionDetalle, 0, len(lista))
	for _, justificacion := range lista {
		resultado = append(resultado, justificaciones.JustificacionDetalle{
			Justificacion: justificacion,
			Alumno:        "Perez, Ana",
			Seccion:       "1.° A",
		})
	}
	return resultado, nil
}

func (s *servicioJustificacionesFalso) Alumno(context.Context, int64) (alumnos.Alumno, error) {
	return alumnos.Alumno{ID: 5, Nombres: "Ana", Apellidos: "Perez", Grado: "1.°", Seccion: "A"}, nil
}
