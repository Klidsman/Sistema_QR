package manejadores

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"asistencia/internal/academico"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type configuracionFalsa struct {
	resumen academico.ResumenConfiguracion
	err     error
}

func (c configuracionFalsa) Resumen(_ context.Context) (academico.ResumenConfiguracion, error) {
	return c.resumen, c.err
}

func nuevoManejadorConfiguracion(t *testing.T) (*academico.Manejador, *scs.SessionManager) {
	t.Helper()
	sesiones := scs.New()
	resumen := academico.ResumenConfiguracion{
		Anio: &academico.AnioEscolar{
			ID:     1,
			Nombre: "2026",
			Estado: academico.EstadoAnioActivo,
		},
		Grados:    5,
		Secciones: 10,
	}
	manejador := academico.NuevoManejador(academico.DependenciasConfiguracion{
		Configuracion: configuracionFalsa{resumen: resumen},
		Sesiones:      sesiones,
		Registro:      slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	return manejador, sesiones
}

func ejecutarConfiguracion(manejador *academico.Manejador, peticion *http.Request) *httptest.ResponseRecorder {
	enrutador := chi.NewRouter()
	manejador.Montar(enrutador)
	respuesta := httptest.NewRecorder()
	enrutador.ServeHTTP(respuesta, peticion)
	return respuesta
}

func TestConfiguracion_PanelVisibleParaAdministradorYDirector(t *testing.T) {
	manejador, sesiones := nuevoManejadorConfiguracion(t)

	for _, rol := range []string{"administrador", "director"} {
		t.Run(rol, func(t *testing.T) {
			peticion := peticionConSesion(t, sesiones, http.MethodGet, "/configuracion", nil, "", rol)
			respuesta := ejecutarConfiguracion(manejador, peticion)

			assert.Equal(t, http.StatusOK, respuesta.Code)
			assert.Contains(t, respuesta.Body.String(), "Configuración académica")
			assert.Contains(t, respuesta.Body.String(), "Año 2026")
		})
	}
}

func TestConfiguracion_OtrosRolesNoAccedenAlPanel(t *testing.T) {
	manejador, sesiones := nuevoManejadorConfiguracion(t)

	for _, rol := range []string{"docente", "porteria"} {
		t.Run(rol, func(t *testing.T) {
			peticion := peticionConSesion(t, sesiones, http.MethodGet, "/configuracion", nil, "", rol)
			respuesta := ejecutarConfiguracion(manejador, peticion)

			assert.Equal(t, http.StatusForbidden, respuesta.Code)
		})
	}
}

func TestConfiguracion_SinSesionRedirigeAlIngreso(t *testing.T) {
	manejador, sesiones := nuevoManejadorConfiguracion(t)
	peticion := peticionConSesion(t, sesiones, http.MethodGet, "/configuracion", nil, "", "")
	sesiones.Remove(peticion.Context(), "usuario_id")

	respuesta := ejecutarConfiguracion(manejador, peticion)

	assert.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Equal(t, "/ingreso", respuesta.Header().Get("Location"))
}

func TestConfiguracion_ErrorDelServicioResponde500(t *testing.T) {
	sesiones := scs.New()
	manejador := academico.NuevoManejador(academico.DependenciasConfiguracion{
		Configuracion: configuracionFalsa{err: io.ErrUnexpectedEOF},
		Sesiones:      sesiones,
		Registro:      slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	peticion := peticionConSesion(t, sesiones, http.MethodGet, "/configuracion", nil, "", "administrador")

	respuesta := ejecutarConfiguracion(manejador, peticion)

	require.Equal(t, http.StatusInternalServerError, respuesta.Code)
}
