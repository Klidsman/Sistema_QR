package manejadores

import (
	"net/http"
	"testing"
	"time"

	"asistencia/internal/academico"

	"github.com/stretchr/testify/assert"
)

func configuracionHistoricaDePrueba(t *testing.T) academico.ConfiguracionFecha {
	t.Helper()
	return academico.ConfiguracionFecha{
		Fecha: time.Date(2027, time.June, 15, 0, 0, 0, 0, time.UTC),
		Anio: &academico.AnioEscolar{
			ID:          1,
			Nombre:      "2027",
			FechaInicio: time.Date(2027, time.March, 1, 0, 0, 0, 0, time.UTC),
			FechaFin:    time.Date(2027, time.December, 15, 0, 0, 0, 0, time.UTC),
			Estado:      academico.EstadoAnioActivo,
		},
		Horarios: []academico.HorarioTurno{{
			Turno:              academico.TurnoManana,
			HoraIngreso:        mustHora(t, "07:30"),
			ToleranciaMinutos:  10,
			HoraLimiteTardanza: mustHora(t, "08:00"),
			HoraCierre:         mustHora(t, "12:30"),
			VigenteDesde:       time.Date(2027, time.March, 1, 0, 0, 0, 0, time.UTC),
		}},
		Secciones: []academico.Seccion{{
			ID: 5, Nombre: "A", Grado: "3.°", Turno: academico.TurnoManana,
			AnioEscolarID: 1, GradoID: 3, AnioEscolar: "2027",
		}},
		Umbrales: []academico.UmbralAlerta{{
			Tipo: academico.TipoUmbralTardanzas, Valor: 3,
			VigenteDesde: time.Date(2027, time.March, 1, 0, 0, 0, 0, time.UTC),
		}},
	}
}

func TestConfiguracionConsulta_MuestraLaConfiguracionDeLaFecha(t *testing.T) {
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		configuracionFecha: configuracionHistoricaDePrueba(t),
	})
	peticion := peticionConSesion(t, sesiones, http.MethodGet,
		"/configuracion/consulta?fecha=2027-06-15", nil, "", "director")

	respuesta := ejecutarConfiguracion(manejador, peticion)

	assert.Equal(t, http.StatusOK, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "Consulta histórica")
	assert.Contains(t, respuesta.Body.String(), "2027")
	assert.Contains(t, respuesta.Body.String(), "Turno de las secciones")
	assert.Contains(t, respuesta.Body.String(), "07:30")
}

func TestConfiguracionConsulta_SinConfiguracionVigente(t *testing.T) {
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{})
	peticion := peticionConSesion(t, sesiones, http.MethodGet,
		"/configuracion/consulta?fecha=2020-01-01", nil, "", "director")

	respuesta := ejecutarConfiguracion(manejador, peticion)

	assert.Equal(t, http.StatusOK, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "No hay configuración vigente para la fecha seleccionada")
}
