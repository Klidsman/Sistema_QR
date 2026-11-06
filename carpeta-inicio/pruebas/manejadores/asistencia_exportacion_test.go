package manejadores

import (
	"net/http"
	"testing"

	"asistencia/internal/asistencia"
	"asistencia/internal/plataforma/fechas"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func exportacionDePrueba() []asistencia.RegistroExportacion {
	return []asistencia.RegistroExportacion{{
		Fecha:         fechas.Hoy(),
		Curso:         "Matemática",
		Seccion:       "3.° A (2026)",
		Bloque:        1,
		Identificador: "80001111",
		Alumno:        "Ríos, Ana",
		Estado:        asistencia.EstadoPresente,
		HoraQR:        "07:42",
		Docente:       "Docente Uno",
	}}
}

func TestExportacion_FormularioYDescarga(t *testing.T) {
	falso := &asistenciaFalsa{exportados: exportacionDePrueba()}
	manejador, sesiones := nuevoManejadorAsistencia(t, falso)

	peticion := peticionConSesion(t, sesiones, http.MethodGet, "/asistencia/exportar", nil, "", "director")
	respuesta := ejecutarAsistencia(manejador, peticion)
	assert.Equal(t, http.StatusOK, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "Descargar CSV")

	peticion = peticionConSesion(t, sesiones, http.MethodGet,
		"/asistencia/exportar/archivo?desde=2026-11-01&hasta=2026-11-06&seccion_id=5", nil, "", "administrador")
	respuesta = ejecutarAsistencia(manejador, peticion)
	require.Equal(t, http.StatusOK, respuesta.Code)
	assert.Equal(t, "text/csv; charset=utf-8", respuesta.Header().Get("Content-Type"))
	assert.Contains(t, respuesta.Body.String(), "identificador")
	assert.Contains(t, respuesta.Body.String(), "Ríos, Ana")
	assert.Equal(t, int64(5), falso.exportarSeccionID)

	peticion = peticionConSesion(t, sesiones, http.MethodGet,
		"/asistencia/exportar/archivo?desde=2026-11-01&hasta=2026-11-06", nil, "", "docente")
	respuesta = ejecutarAsistencia(manejador, peticion)
	require.Equal(t, http.StatusOK, respuesta.Code)
	assert.Equal(t, int64(9), falso.exportarDocenteID)
}

func TestExportacion_DirectorNecesitaSeccionYDocenteNo(t *testing.T) {
	falso := &asistenciaFalsa{exportados: exportacionDePrueba()}
	manejador, sesiones := nuevoManejadorAsistencia(t, falso)

	peticion := peticionConSesion(t, sesiones, http.MethodGet,
		"/asistencia/exportar/archivo?desde=2026-11-01&hasta=2026-11-06", nil, "", "director")
	respuesta := ejecutarAsistencia(manejador, peticion)
	assert.Equal(t, http.StatusUnprocessableEntity, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "Seleccione una sección")
}

func TestExportacion_RangoInvalido(t *testing.T) {
	manejador, sesiones := nuevoManejadorAsistencia(t, &asistenciaFalsa{})
	peticion := peticionConSesion(t, sesiones, http.MethodGet,
		"/asistencia/exportar/archivo?desde=2026-11-06&hasta=2026-11-01&seccion_id=5", nil, "", "administrador")

	respuesta := ejecutarAsistencia(manejador, peticion)

	assert.Equal(t, http.StatusUnprocessableEntity, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "El rango de fechas no es válido")
}
