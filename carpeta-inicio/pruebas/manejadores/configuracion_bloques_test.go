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

func bloqueDePrueba(t *testing.T) academico.BloqueClase {
	t.Helper()
	inicio, err := fechas.ParsearHora("07:45")
	require.NoError(t, err)
	fin, err := fechas.ParsearHora("08:30")
	require.NoError(t, err)
	return academico.BloqueClase{
		ID:           4,
		Turno:        academico.TurnoManana,
		Numero:       1,
		HoraInicio:   inicio,
		HoraFin:      fin,
		VigenteDesde: time.Date(2026, time.November, 6, 0, 0, 0, 0, time.UTC),
	}
}

func valoresBloque() url.Values {
	return url.Values{
		"turno":         {"manana"},
		"numero":        {"2"},
		"hora_inicio":   {"08:30"},
		"hora_fin":      {"09:15"},
		"vigente_desde": {"2026-11-06"},
	}
}

func TestConfiguracionBloques_MuestraLosBloquesVigentes(t *testing.T) {
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen: academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
		bloques: map[string][]academico.BloqueClase{
			academico.TurnoManana: {bloqueDePrueba(t)},
		},
	})
	peticion := peticionConSesion(t, sesiones, http.MethodGet,
		"/configuracion/bloques?turno=manana", nil, "", "director")

	respuesta := ejecutarConfiguracion(manejador, peticion)

	assert.Equal(t, http.StatusOK, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "Bloques de clase")
	assert.Contains(t, respuesta.Body.String(), "07:45")
	assert.Contains(t, respuesta.Body.String(), "Ajustar")
}

func TestConfiguracionBloques_AdministradorCrea(t *testing.T) {
	var recibido academico.BloqueClase
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen: academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
		crearBloque: func(bloque academico.BloqueClase) (int64, error) {
			recibido = bloque
			return 8, nil
		},
	})
	peticion := peticionFormulario(t, sesiones,
		"/configuracion/bloques", valoresBloque(), "administrador")

	respuesta := ejecutarConfiguracion(manejador, peticion)

	require.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Equal(t, "/configuracion/bloques?turno=manana&ok=creado", respuesta.Header().Get("Location"))
	assert.Equal(t, 2, recibido.Numero)
	assert.Equal(t, academico.TurnoManana, recibido.Turno)
}

func TestConfiguracionBloques_DirectorNoPuedeCrear(t *testing.T) {
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen: academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
	})
	peticion := peticionFormulario(t, sesiones,
		"/configuracion/bloques", valoresBloque(), "director")

	respuesta := ejecutarConfiguracion(manejador, peticion)

	assert.Equal(t, http.StatusForbidden, respuesta.Code)
}

func TestConfiguracionBloques_DirectorPuedeAjustarHoras(t *testing.T) {
	var bloqueID int64
	var inicio time.Time
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen: academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
		ajustarBloque: func(id int64, inicioRecibido, _, _ time.Time) error {
			bloqueID = id
			inicio = inicioRecibido
			return nil
		},
	})
	peticion := peticionFormulario(t, sesiones,
		"/configuracion/bloques/4/horas", url.Values{
			"turno":         {"manana"},
			"hora_inicio":   {"07:50"},
			"hora_fin":      {"08:25"},
			"vigente_desde": {"2026-11-06"},
		}, "director")

	respuesta := ejecutarConfiguracion(manejador, peticion)

	require.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Equal(t, "/configuracion/bloques?turno=manana&ok=ajustado", respuesta.Header().Get("Location"))
	assert.Equal(t, int64(4), bloqueID)
	assert.Equal(t, "07:50", fechas.FormatearHora(inicio))
}

func TestConfiguracionBloques_RechazaSolapamiento(t *testing.T) {
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen: academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
		crearBloque: func(academico.BloqueClase) (int64, error) {
			return 0, academico.ErrBloqueSolapado
		},
	})
	peticion := peticionFormulario(t, sesiones,
		"/configuracion/bloques", valoresBloque(), "administrador")

	respuesta := ejecutarConfiguracion(manejador, peticion)

	assert.Equal(t, http.StatusUnprocessableEntity, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "El bloque se superpone con otro bloque del mismo turno")
}
