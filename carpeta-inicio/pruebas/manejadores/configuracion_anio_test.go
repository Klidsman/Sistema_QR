package manejadores

import (
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"asistencia/internal/academico"

	"github.com/alexedwards/scs/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func nuevoManejadorConfiguracionConFalso(t *testing.T, falso configuracionFalsa) (*academico.Manejador, *scs.SessionManager) {
	t.Helper()
	sesiones := scs.New()
	manejador := academico.NuevoManejador(academico.DependenciasConfiguracion{
		Configuracion: falso,
		Sesiones:      sesiones,
		Registro:      slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	return manejador, sesiones
}

func peticionFormulario(t *testing.T, sesiones *scs.SessionManager, metodo, ruta string, valores url.Values, rol string) *http.Request {
	t.Helper()
	cuerpo := strings.NewReader(valores.Encode())
	peticion := peticionConSesion(t, sesiones, metodo, ruta, cuerpo, "application/x-www-form-urlencoded", rol)
	return peticion
}

func anioActivoDePrueba() *academico.AnioEscolar {
	return &academico.AnioEscolar{
		ID:          7,
		Nombre:      "2026",
		FechaInicio: time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC),
		FechaFin:    time.Date(2026, time.December, 15, 0, 0, 0, 0, time.UTC),
		Estado:      academico.EstadoAnioActivo,
	}
}

func TestConfiguracionAnio_MuestraFormularioSinAnioActivo(t *testing.T) {
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{})
	peticion := peticionConSesion(t, sesiones, http.MethodGet, "/configuracion/anio", nil, "", "administrador")

	respuesta := ejecutarConfiguracion(manejador, peticion)

	assert.Equal(t, http.StatusOK, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "Crear año escolar")
	assert.Contains(t, respuesta.Body.String(), "Fecha de inicio")
}

func TestConfiguracionAnio_MuestraElAnioActivo(t *testing.T) {
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen: academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
	})
	peticion := peticionConSesion(t, sesiones, http.MethodGet, "/configuracion/anio", nil, "", "director")

	respuesta := ejecutarConfiguracion(manejador, peticion)

	assert.Equal(t, http.StatusOK, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "Año activo")
	assert.Contains(t, respuesta.Body.String(), "Cerrar año escolar")
}

func TestConfiguracionAnio_CreaConDatosNormalizados(t *testing.T) {
	var recibido academico.DatosAnio
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		crearAnio: func(datos academico.DatosAnio) (int64, error) {
			recibido = datos
			return 9, nil
		},
	})
	peticion := peticionFormulario(t, sesiones, http.MethodPost, "/configuracion/anio", url.Values{
		"nombre":       {"  2027  "},
		"fecha_inicio": {"2027-03-01"},
		"fecha_fin":    {"2027-12-15"},
	}, "administrador")

	respuesta := ejecutarConfiguracion(manejador, peticion)

	assert.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Equal(t, "/configuracion/anio?ok=creado", respuesta.Header().Get("Location"))
	assert.Equal(t, "  2027  ", recibido.Nombre, "el manejador entrega el valor; el servicio lo normaliza")
	assert.Equal(t, 2027, recibido.FechaInicio.Year())
	assert.Equal(t, 12, int(recibido.FechaFin.Month()))
}

func TestConfiguracionAnio_RechazaFechasInvalidas(t *testing.T) {
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		crearAnio: func(academico.DatosAnio) (int64, error) {
			return 0, academico.ErrFechasAnioInvalidas
		},
	})
	peticion := peticionFormulario(t, sesiones, http.MethodPost, "/configuracion/anio", url.Values{
		"nombre":       {"2027"},
		"fecha_inicio": {"2027-12-15"},
		"fecha_fin":    {"2027-03-01"},
	}, "administrador")

	respuesta := ejecutarConfiguracion(manejador, peticion)

	assert.Equal(t, http.StatusUnprocessableEntity, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "La fecha de fin debe ser posterior a la fecha de inicio")
}

func TestConfiguracionAnio_RechazaSegundoAnioActivo(t *testing.T) {
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		crearAnio: func(academico.DatosAnio) (int64, error) {
			return 0, academico.ErrAnioActivoExistente
		},
	})
	peticion := peticionFormulario(t, sesiones, http.MethodPost, "/configuracion/anio", url.Values{
		"nombre":       {"2027"},
		"fecha_inicio": {"2027-03-01"},
		"fecha_fin":    {"2027-12-15"},
	}, "administrador")

	respuesta := ejecutarConfiguracion(manejador, peticion)

	assert.Equal(t, http.StatusUnprocessableEntity, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "Ya existe un año escolar activo")
}

func TestConfiguracionAnio_CierraElAnioActivo(t *testing.T) {
	cerrado := false
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen: academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
		cerrarAnio: func() error {
			cerrado = true
			return nil
		},
	})
	peticion := peticionConSesion(t, sesiones, http.MethodPost, "/configuracion/anio/cierre", nil, "", "administrador")

	respuesta := ejecutarConfiguracion(manejador, peticion)

	assert.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Equal(t, "/configuracion/anio?ok=cerrado", respuesta.Header().Get("Location"))
	assert.True(t, cerrado)
}

func TestConfiguracionGrados_DirectorNoPuedeCrear(t *testing.T) {
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen: academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
	})
	peticion := peticionFormulario(t, sesiones, http.MethodPost, "/configuracion/grados", url.Values{
		"nombre": {"3.°"},
	}, "director")

	respuesta := ejecutarConfiguracion(manejador, peticion)

	assert.Equal(t, http.StatusForbidden, respuesta.Code)
}

func TestConfiguracionGrados_MuestraAvisoSinAnioActivo(t *testing.T) {
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{})
	peticion := peticionConSesion(t, sesiones, http.MethodGet, "/configuracion/grados", nil, "", "administrador")

	respuesta := ejecutarConfiguracion(manejador, peticion)

	assert.Equal(t, http.StatusOK, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "No hay un año escolar activo")
}

func TestConfiguracionSecciones_CreaYCambiaTurno(t *testing.T) {
	var seccionRecibida academico.DatosSeccion
	var turnoRecibido string
	var vigenciaRecibida time.Time
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen: academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
		grados:  []academico.Grado{{ID: 3, Nombre: "3.°"}},
		crearSeccion: func(datos academico.DatosSeccion) (int64, error) {
			seccionRecibida = datos
			return 15, nil
		},
		cambiarTurno: func(seccionID int64, turno string, vigenteDesde time.Time) error {
			turnoRecibido = turno
			vigenciaRecibida = vigenteDesde
			return nil
		},
	})

	peticion := peticionFormulario(t, sesiones, http.MethodPost, "/configuracion/secciones", url.Values{
		"grado_id":      {"3"},
		"nombre":        {"C"},
		"turno":         {"tarde"},
		"vigente_desde": {"2026-11-06"},
	}, "administrador")
	respuesta := ejecutarConfiguracion(manejador, peticion)

	require.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Equal(t, "/configuracion/secciones?ok=creado", respuesta.Header().Get("Location"))
	assert.Equal(t, int64(3), seccionRecibida.GradoID)
	assert.Equal(t, "C", seccionRecibida.Nombre)
	assert.Equal(t, academico.TurnoTarde, seccionRecibida.Turno)

	peticion = peticionFormulario(t, sesiones, http.MethodPost, "/configuracion/secciones/15/turno", url.Values{
		"turno":         {"manana"},
		"vigente_desde": {"2026-11-06"},
	}, "administrador")
	respuesta = ejecutarConfiguracion(manejador, peticion)

	require.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Equal(t, "/configuracion/secciones?ok=turno", respuesta.Header().Get("Location"))
	assert.Equal(t, academico.TurnoManana, turnoRecibido)
	assert.Equal(t, 2026, vigenciaRecibida.Year())
}
