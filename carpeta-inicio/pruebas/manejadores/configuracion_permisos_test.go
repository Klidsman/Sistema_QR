package manejadores

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"asistencia/internal/academico"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type configuracionFalsa struct {
	resumen           academico.ResumenConfiguracion
	err               error
	grados            []academico.Grado
	secciones         []academico.Seccion
	crearAnio         func(academico.DatosAnio) (int64, error)
	crearGrado        func(academico.DatosGrado) (int64, error)
	crearSeccion      func(academico.DatosSeccion) (int64, error)
	cambiarTurno      func(int64, string, time.Time) error
	cerrarAnio        func() error
	horarios          map[string]academico.HorarioTurno
	historial         map[string][]academico.HorarioTurno
	configurarHorario func(academico.HorarioTurno) error
}

func (c configuracionFalsa) Resumen(_ context.Context) (academico.ResumenConfiguracion, error) {
	return c.resumen, c.err
}

func (c configuracionFalsa) ListarAnios(_ context.Context) ([]academico.AnioEscolar, error) {
	return nil, c.err
}

func (c configuracionFalsa) AnioActivo(_ context.Context) (*academico.AnioEscolar, error) {
	return c.resumen.Anio, c.err
}

func (c configuracionFalsa) ListarGrados(_ context.Context, _ int64) ([]academico.Grado, error) {
	return c.grados, c.err
}

func (c configuracionFalsa) ListarSecciones(_ context.Context, _ int64) ([]academico.Seccion, error) {
	return c.secciones, c.err
}

func (c configuracionFalsa) ListarTurnosSeccion(_ context.Context, _ int64) ([]academico.TurnoSeccion, error) {
	return nil, c.err
}

func (c configuracionFalsa) HorarioVigente(_ context.Context, turno string, _ time.Time) (academico.HorarioTurno, error) {
	if horario, existe := c.horarios[turno]; existe {
		return horario, nil
	}
	return academico.HorarioTurno{}, academico.ErrHorarioNoEncontrado
}

func (c configuracionFalsa) ListarHorarios(_ context.Context, turno string) ([]academico.HorarioTurno, error) {
	return c.historial[turno], c.err
}

func (c configuracionFalsa) ConfigurarHorario(_ context.Context, _ int64, horario academico.HorarioTurno) error {
	if c.configurarHorario != nil {
		return c.configurarHorario(horario)
	}
	return c.err
}

func (c configuracionFalsa) CrearAnio(_ context.Context, _ int64, datos academico.DatosAnio) (int64, error) {
	if c.crearAnio != nil {
		return c.crearAnio(datos)
	}
	return 1, c.err
}

func (c configuracionFalsa) CerrarAnio(_ context.Context, _, _ int64) error {
	if c.cerrarAnio != nil {
		return c.cerrarAnio()
	}
	return c.err
}

func (c configuracionFalsa) CrearGrado(_ context.Context, _ int64, datos academico.DatosGrado) (int64, error) {
	if c.crearGrado != nil {
		return c.crearGrado(datos)
	}
	return 1, c.err
}

func (c configuracionFalsa) CrearSeccion(_ context.Context, _ int64, datos academico.DatosSeccion) (int64, error) {
	if c.crearSeccion != nil {
		return c.crearSeccion(datos)
	}
	return 1, c.err
}

func (c configuracionFalsa) CambiarTurno(_ context.Context, _ int64, seccionID int64, turno string, vigenteDesde time.Time) error {
	if c.cambiarTurno != nil {
		return c.cambiarTurno(seccionID, turno, vigenteDesde)
	}
	return c.err
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
