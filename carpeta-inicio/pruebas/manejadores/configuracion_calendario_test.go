package manejadores

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"asistencia/internal/academico"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func excepcionDePrueba() academico.ExcepcionCalendario {
	return academico.ExcepcionCalendario{
		ID:            3,
		AnioEscolarID: 1,
		Tipo:          academico.TipoFeriado,
		FechaInicio:   time.Date(2026, time.November, 6, 0, 0, 0, 0, time.UTC),
		FechaFin:      time.Date(2026, time.November, 6, 0, 0, 0, 0, time.UTC),
		Motivo:        "Fiestas Patrias",
	}
}

func valoresExcepcion() url.Values {
	return url.Values{
		"tipo":         {"feriado"},
		"fecha_inicio": {"2026-11-06"},
		"fecha_fin":    {"2026-11-06"},
		"motivo":       {"Fiestas Patrias"},
	}
}

func TestConfiguracionCalendario_MuestraLasExcepciones(t *testing.T) {
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen:     academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
		grados:      []academico.Grado{{ID: 3, Nombre: "3.°"}},
		excepciones: []academico.ExcepcionCalendario{excepcionDePrueba()},
	})
	peticion := peticionConSesion(t, sesiones, http.MethodGet, "/configuracion/calendario", nil, "", "director")

	respuesta := ejecutarConfiguracion(manejador, peticion)

	assert.Equal(t, http.StatusOK, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "Fiestas Patrias")
	assert.Contains(t, respuesta.Body.String(), "Feriado")
	assert.Contains(t, respuesta.Body.String(), "Todo el colegio")
}

func TestConfiguracionCalendario_AdministradorRegistra(t *testing.T) {
	var recibido academico.DatosExcepcion
	var confirmado bool
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen: academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
		grados:  []academico.Grado{{ID: 3, Nombre: "3.°"}},
		registrarExcepcion: func(datos academico.DatosExcepcion, confirmar bool) (int64, error) {
			recibido = datos
			confirmado = confirmar
			return 5, nil
		},
	})
	valores := valoresExcepcion()
	valores.Add("grados", "3")
	peticion := peticionFormulario(t, sesiones,
		"/configuracion/calendario", valores, "administrador")

	respuesta := ejecutarConfiguracion(manejador, peticion)

	require.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Equal(t, "/configuracion/calendario?ok=excepcion", respuesta.Header().Get("Location"))
	assert.Equal(t, academico.TipoFeriado, recibido.Tipo)
	assert.Equal(t, "Fiestas Patrias", recibido.Motivo)
	assert.Equal(t, []int64{3}, recibido.Grados)
	assert.False(t, confirmado)
}

func TestConfiguracionCalendario_PideConfirmacionAnteSolapamiento(t *testing.T) {
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen: academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
		grados:  []academico.Grado{{ID: 3, Nombre: "3.°"}},
		registrarExcepcion: func(academico.DatosExcepcion, bool) (int64, error) {
			return 0, academico.ErrExcepcionSolapada
		},
	})
	peticion := peticionFormulario(t, sesiones,
		"/configuracion/calendario", valoresExcepcion(), "administrador")

	respuesta := ejecutarConfiguracion(manejador, peticion)

	assert.Equal(t, http.StatusOK, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "Confirmar y guardar")
	assert.Contains(t, respuesta.Body.String(), "Fiestas Patrias")
}

func TestConfiguracionCalendario_DirectorNoPuedeRegistrar(t *testing.T) {
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen: academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
	})
	peticion := peticionFormulario(t, sesiones,
		"/configuracion/calendario", valoresExcepcion(), "director")

	respuesta := ejecutarConfiguracion(manejador, peticion)

	assert.Equal(t, http.StatusForbidden, respuesta.Code)
}
