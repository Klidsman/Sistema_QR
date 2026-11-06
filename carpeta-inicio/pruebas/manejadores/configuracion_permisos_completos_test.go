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

func TestConfiguracion_DirectorNoPuedeEscribirEnNingunaRutaDeAdministracion(t *testing.T) {
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen: academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
	})

	rutas := []string{
		"/configuracion/anio",
		"/configuracion/anio/cierre",
		"/configuracion/grados",
		"/configuracion/secciones",
		"/configuracion/secciones/1/turno",
		"/configuracion/bloques",
		"/configuracion/cursos",
		"/configuracion/cursos/1/desactivacion",
		"/configuracion/horarios",
		"/configuracion/calendario",
		"/configuracion/umbrales",
	}

	for _, ruta := range rutas {
		t.Run(ruta, func(t *testing.T) {
			peticion := peticionFormulario(t, sesiones, ruta, url.Values{}, "director")
			respuesta := ejecutarConfiguracion(manejador, peticion)
			assert.Equal(t, http.StatusForbidden, respuesta.Code, ruta)
		})
	}
}

func TestConfiguracion_DirectorPuedeEditarSoloTurnosYHorasDeBloques(t *testing.T) {
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen: academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
		configurarHorario: func(academico.HorarioTurno) error {
			return nil
		},
		ajustarBloque: func(int64, time.Time, time.Time, time.Time) error {
			return nil
		},
	})

	peticion := peticionFormulario(t, sesiones,
		"/configuracion/turnos/manana", valoresHorario(), "director")
	respuesta := ejecutarConfiguracion(manejador, peticion)
	require.Equal(t, http.StatusSeeOther, respuesta.Code)

	peticion = peticionFormulario(t, sesiones,
		"/configuracion/bloques/1/horas", url.Values{
			"turno":         {"manana"},
			"hora_inicio":   {"07:50"},
			"hora_fin":      {"08:25"},
			"vigente_desde": {"2026-11-06"},
		}, "director")
	respuesta = ejecutarConfiguracion(manejador, peticion)
	assert.Equal(t, http.StatusSeeOther, respuesta.Code)
}

func TestConfiguracion_DirectorPuedeLeerTodasLasPantallas(t *testing.T) {
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen: academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
		grilla:  grillaDePrueba(t),
	})

	rutas := []string{
		"/configuracion",
		"/configuracion/anio",
		"/configuracion/grados",
		"/configuracion/secciones",
		"/configuracion/turnos",
		"/configuracion/bloques",
		"/configuracion/cursos",
		"/configuracion/horarios",
		"/configuracion/calendario",
		"/configuracion/umbrales",
		"/configuracion/consulta",
	}

	for _, ruta := range rutas {
		t.Run(ruta, func(t *testing.T) {
			peticion := peticionConSesion(t, sesiones, http.MethodGet, ruta, nil, "", "director")
			respuesta := ejecutarConfiguracion(manejador, peticion)
			assert.Equal(t, http.StatusOK, respuesta.Code, ruta)
		})
	}
}
