package manejadores

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"asistencia/internal/alertas"
	"asistencia/internal/plataforma/intermediarios"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAlertasPanel_FiltraYAlcancePorRol(t *testing.T) {
	servicio := &servicioAlertasFalso{
		alertas: []alertas.Alerta{{
			ID:             1,
			Tipo:           alertas.TipoTardanza,
			AlumnoID:       5,
			SeccionID:      2,
			FechaEvento:    fechaDePrueba(),
			DestinatarioID: 9,
			Estado:         alertas.EstadoNueva,
		}},
	}
	manejador, sesiones := nuevoManejadorAlertas(t, servicio)

	peticion := peticionConSesion(t, sesiones, http.MethodGet,
		"/alertas?tipo=tardanza&estado=nueva&seccion=2", nil, "", "docente")
	respuesta := ejecutarAlertas(manejador, sesiones, peticion, false)

	require.Equal(t, http.StatusOK, respuesta.Code)
	assert.False(t, servicio.ultimosFiltros.Todos)
	assert.Equal(t, int64(9), servicio.ultimosFiltros.DestinatarioID)
	assert.Equal(t, "tardanza", servicio.ultimosFiltros.Tipo)
	assert.Equal(t, "nueva", servicio.ultimosFiltros.Estado)
	assert.Equal(t, int64(2), servicio.ultimosFiltros.SeccionID)
	assert.Contains(t, respuesta.Body.String(), "Perez, Ana")

	peticion = peticionConSesion(t, sesiones, http.MethodGet, "/alertas", nil, "", "administrador")
	respuesta = ejecutarAlertas(manejador, sesiones, peticion, false)
	require.Equal(t, http.StatusOK, respuesta.Code)
	assert.True(t, servicio.ultimosFiltros.Todos, "el administrador ve todas las alertas")
}

func TestAlertasPanel_MarcarVistaYCSRF(t *testing.T) {
	servicio := &servicioAlertasFalso{}
	manejador, sesiones := nuevoManejadorAlertas(t, servicio)

	peticion := peticionConSesion(t, sesiones, http.MethodPost, "/alertas/7/vista",
		strings.NewReader(url.Values{}.Encode()), "application/x-www-form-urlencoded", "docente")
	respuesta := ejecutarAlertas(manejador, sesiones, peticion, true)
	assert.Equal(t, http.StatusForbidden, respuesta.Code, "sin token CSRF debe rechazarse")

	peticion = peticionConSesion(t, sesiones, http.MethodPost, "/alertas/7/vista",
		strings.NewReader(url.Values{"token_csrf": {"token-correcto"}}.Encode()),
		"application/x-www-form-urlencoded", "docente")
	sesiones.Put(peticion.Context(), "token_csrf", "token-correcto")
	respuesta = ejecutarAlertas(manejador, sesiones, peticion, true)
	assert.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Equal(t, int64(7), servicio.ultimaVista)
}

func TestAlertasPanel_DirectorSinAccesoAEscritura(t *testing.T) {
	servicio := &servicioAlertasFalso{}
	manejador, sesiones := nuevoManejadorAlertas(t, servicio)

	peticion := peticionConSesion(t, sesiones, http.MethodPost, "/alertas/vistas",
		strings.NewReader(url.Values{"token_csrf": {"token-correcto"}}.Encode()),
		"application/x-www-form-urlencoded", "director")
	sesiones.Put(peticion.Context(), "token_csrf", "token-correcto")
	respuesta := ejecutarAlertas(manejador, sesiones, peticion, true)

	assert.Equal(t, http.StatusForbidden, respuesta.Code)
}

func TestAlertasIndicador_MuestraElContador(t *testing.T) {
	servicio := &servicioAlertasFalso{nuevas: 4}
	manejador, sesiones := nuevoManejadorAlertas(t, servicio)

	peticion := peticionConSesion(t, sesiones, http.MethodGet, "/alertas/indicador", nil, "", "porteria")
	respuesta := ejecutarAlertas(manejador, sesiones, peticion, false)

	require.Equal(t, http.StatusOK, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "4")
}

func nuevoManejadorAlertas(t *testing.T, servicio alertas.ServicioAlertas) (*alertas.Manejador, *scs.SessionManager) {
	t.Helper()
	sesiones := scs.New()
	return alertas.NuevoManejador(alertas.DependenciasManejador{
		Servicio: servicio,
		Sesiones: sesiones,
		Registro: registroDePrueba(),
	}), sesiones
}

func ejecutarAlertas(manejador *alertas.Manejador, sesiones *scs.SessionManager, peticion *http.Request, conCSRF bool) *httptest.ResponseRecorder {
	enrutador := chi.NewRouter()
	if conCSRF {
		enrutador.Use(intermediarios.ProteccionCSRF(sesiones))
	}
	manejador.Montar(enrutador)
	respuesta := httptest.NewRecorder()
	enrutador.ServeHTTP(respuesta, peticion)
	return respuesta
}

func fechaDePrueba() time.Time {
	return time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
}

func registroDePrueba() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

type servicioAlertasFalso struct {
	alertas        []alertas.Alerta
	nuevas         int
	ultimosFiltros alertas.FiltrosAlerta
	ultimaVista    int64
}

func (s *servicioAlertasFalso) ListarAlertas(_ context.Context, filtros alertas.FiltrosAlerta) ([]alertas.Alerta, error) {
	s.ultimosFiltros = filtros
	return s.alertas, nil
}

func (s *servicioAlertasFalso) Detallar(_ context.Context, lista []alertas.Alerta) ([]alertas.AlertaDetalle, error) {
	detalles := make([]alertas.AlertaDetalle, 0, len(lista))
	for _, alerta := range lista {
		detalles = append(detalles, alertas.AlertaDetalle{
			Alerta:          alerta,
			AlumnoNombre:    "Perez, Ana",
			SeccionEtiqueta: "1.° A",
		})
	}
	return detalles, nil
}

func (s *servicioAlertasFalso) ContarNuevas(context.Context, int64, bool) (int, error) {
	return s.nuevas, nil
}

func (s *servicioAlertasFalso) MarcarVista(_ context.Context, id, _ int64, _ bool) error {
	s.ultimaVista = id
	return nil
}

func (s *servicioAlertasFalso) MarcarVistas(context.Context, []int64, int64, bool) (int, error) {
	return 0, nil
}

func (s *servicioAlertasFalso) AsignarResponsable(context.Context, int64, int64, int64) error {
	return nil
}
