package manejadores

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"asistencia/internal/academico"
	"asistencia/internal/plataforma/fechas"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func horarioDePrueba(t *testing.T, turno, ingreso, limite, cierre string) academico.HorarioTurno {
	t.Helper()
	horaIngreso, err := fechas.ParsearHora(ingreso)
	require.NoError(t, err)
	horaLimite, err := fechas.ParsearHora(limite)
	require.NoError(t, err)
	horaCierre, err := fechas.ParsearHora(cierre)
	require.NoError(t, err)
	return academico.HorarioTurno{
		Turno:              turno,
		HoraIngreso:        horaIngreso,
		ToleranciaMinutos:  10,
		HoraLimiteTardanza: horaLimite,
		HoraCierre:         horaCierre,
		VigenteDesde:       time.Date(2026, time.November, 6, 0, 0, 0, 0, time.UTC),
	}
}

func valoresHorario() url.Values {
	return url.Values{
		"hora_ingreso":         {"07:30"},
		"tolerancia_minutos":   {"10"},
		"hora_limite_tardanza": {"08:00"},
		"hora_cierre":          {"08:30"},
		"vigente_desde":        {"2026-11-06"},
	}
}

func TestConfiguracionTurnos_MuestraHorariosVigentes(t *testing.T) {
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen: academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
		horarios: map[string]academico.HorarioTurno{
			academico.TurnoManana: horarioDePrueba(t, academico.TurnoManana, "07:30", "08:00", "08:30"),
			academico.TurnoTarde:  horarioDePrueba(t, academico.TurnoTarde, "12:45", "13:15", "13:45"),
		},
	})
	peticion := peticionConSesion(t, sesiones, http.MethodGet, "/configuracion/turnos", nil, "", "director")

	respuesta := ejecutarConfiguracion(manejador, peticion)

	assert.Equal(t, http.StatusOK, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "Turno de Mañana")
	assert.Contains(t, respuesta.Body.String(), "07:30")
	assert.Contains(t, respuesta.Body.String(), "Turno de Tarde")
}

func TestConfiguracionTurnos_DirectorPuedeGuardar(t *testing.T) {
	var recibido academico.HorarioTurno
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen: academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
		configurarHorario: func(horario academico.HorarioTurno) error {
			recibido = horario
			return nil
		},
	})
	peticion := peticionFormulario(t, sesiones,
		"/configuracion/turnos/manana", valoresHorario(), "director")

	respuesta := ejecutarConfiguracion(manejador, peticion)

	require.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Equal(t, "/configuracion/turnos?ok=horario", respuesta.Header().Get("Location"))
	assert.Equal(t, academico.TurnoManana, recibido.Turno)
	assert.Equal(t, 10, recibido.ToleranciaMinutos)
	assert.Equal(t, "08:30", fechas.FormatearHora(recibido.HoraCierre))
}

func TestConfiguracionTurnos_RechazaHorarioInvalido(t *testing.T) {
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen: academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
		configurarHorario: func(academico.HorarioTurno) error {
			return academico.ErrCierreNoPosterior
		},
	})
	peticion := peticionFormulario(t, sesiones,
		"/configuracion/turnos/manana", valoresHorario(), "administrador")

	respuesta := ejecutarConfiguracion(manejador, peticion)

	assert.Equal(t, http.StatusUnprocessableEntity, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "La hora de cierre debe ser posterior a la hora límite de tardanza")
}

func TestConfiguracionTurnos_OtrosRolesNoPuedenGuardar(t *testing.T) {
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen: academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
	})
	peticion := peticionFormulario(t, sesiones,
		"/configuracion/turnos/manana", valoresHorario(), "docente")

	respuesta := ejecutarConfiguracion(manejador, peticion)

	assert.Equal(t, http.StatusForbidden, respuesta.Code)
}
