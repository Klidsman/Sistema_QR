package manejadores

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"asistencia/internal/asistencia"
	"asistencia/internal/plataforma/fechas"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConsultaAsistencia_MuestraRegistrosYCorrige(t *testing.T) {
	falso := &asistenciaFalsa{
		consulta: []asistencia.ConsultaRegistro{{
			Registro: asistencia.RegistroDocente{
				ID: 9, Estado: asistencia.EstadoFalta, Observacion: "Sin justificar", BloqueNumero: 1,
				Fecha: fechas.Hoy(),
			},
			Alumno: "Ríos, Ana", Curso: "Matemática", Seccion: "3.° A (2026)",
			Docente: "Docente Uno", Diferencia: true,
		}},
		plazo: asistencia.ParametrosAsistencia{PlazoEdicionHoras: 24},
	}
	manejador, sesiones := nuevoManejadorAsistencia(t, falso)

	peticion := peticionConSesion(t, sesiones, http.MethodGet,
		"/asistencia/consulta?fecha="+fechas.FormatearISO(fechas.Hoy())+"&seccion_id=5", nil, "", "director")
	respuesta := ejecutarAsistencia(manejador, peticion)
	assert.Equal(t, http.StatusOK, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "Ríos, Ana")
	assert.Contains(t, respuesta.Body.String(), "Matemática")

	peticion = peticionPostFormulario(t, sesiones, "/asistencia/consulta/correccion", url.Values{
		"registro_id": {"9"},
		"seccion_id":  {"5"},
		"fecha":       {fechas.FormatearISO(fechas.Hoy())},
		"estado":      {asistencia.EstadoPresente},
		"observacion": {"Corregido"},
		"motivo":      {"Error del docente"},
	}, "administrador")
	respuesta = ejecutarAsistencia(manejador, peticion)
	require.Equal(t, http.StatusSeeOther, respuesta.Code)
	require.NotNil(t, falso.corregido)
	assert.Equal(t, asistencia.EstadoPresente, falso.corregido.Registro.Estado)
}

func TestConsultaAsistencia_CorreccionSinMotivoFalla(t *testing.T) {
	falso := &asistenciaFalsa{correccionErr: asistencia.ErrMotivoRequerido}
	manejador, sesiones := nuevoManejadorAsistencia(t, falso)

	peticion := peticionPostFormulario(t, sesiones, "/asistencia/consulta/correccion", url.Values{
		"registro_id": {"9"},
		"seccion_id":  {"5"},
		"fecha":       {fechas.FormatearISO(fechas.Hoy())},
		"estado":      {asistencia.EstadoPresente},
		"motivo":      {""},
	}, "administrador")
	respuesta := ejecutarAsistencia(manejador, peticion)

	assert.Equal(t, http.StatusUnprocessableEntity, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "Indique el motivo de la corrección")
}

func TestParametrosAsistencia_MuestraYConfiguraPlazo(t *testing.T) {
	falso := &asistenciaFalsa{plazo: asistencia.ParametrosAsistencia{
		PlazoEdicionHoras: 24,
		ActualizadoEn:     time.Date(2026, time.November, 6, 0, 0, 0, 0, time.UTC),
	}}
	manejador, sesiones := nuevoManejadorAsistencia(t, falso)

	peticion := peticionConSesion(t, sesiones, http.MethodGet, "/configuracion/asistencia", nil, "", "director")
	respuesta := ejecutarAsistencia(manejador, peticion)
	assert.Equal(t, http.StatusOK, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "24 horas")

	peticion = peticionPostFormulario(t, sesiones, "/configuracion/asistencia",
		url.Values{"plazo_edicion_horas": {"12"}}, "administrador")
	respuesta = ejecutarAsistencia(manejador, peticion)
	require.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Equal(t, 12, falso.plazoNuevo)

	peticion = peticionPostFormulario(t, sesiones, "/configuracion/asistencia",
		url.Values{"plazo_edicion_horas": {"12"}}, "director")
	respuesta = ejecutarAsistencia(manejador, peticion)
	assert.Equal(t, http.StatusForbidden, respuesta.Code)
}
