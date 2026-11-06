package manejadores

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"asistencia/internal/academico"
	"asistencia/internal/alumnos"
	"asistencia/internal/asistencia"
	"asistencia/internal/plataforma/fechas"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func detalleClaseDePrueba(t *testing.T) asistencia.ClaseDetalle {
	t.Helper()
	marcacion := &asistencia.Marcacion{
		AlumnoID: 10,
		Hora:     time.Date(2026, time.November, 6, 12, 42, 0, 0, time.UTC),
		Estado:   asistencia.MarcacionPuntual,
	}
	registro := &asistencia.RegistroDocente{
		ID: 1, AlumnoID: 10, Estado: asistencia.EstadoPresente,
		CreadoEn: time.Date(2026, time.November, 6, 13, 0, 0, 0, time.UTC),
	}
	return asistencia.ClaseDetalle{
		Clase: academico.ClaseProgramada{
			SeccionID: 5, CursoID: 2, Bloque: 1, Turno: academico.TurnoManana,
			HoraInicio: mustHora(t, "07:45"), HoraFin: mustHora(t, "08:30"),
			Seccion: "3.° A (2026)", Curso: "Matemática",
		},
		Fecha: fechas.Hoy(),
		Alumnos: []asistencia.AlumnoClase{
			{
				Alumno:     alumnos.AlumnoEnClase{ID: 10, Identificador: "80001111", Nombres: "Ana", Apellidos: "Ríos"},
				Registro:   registro,
				Marcacion:  marcacion,
				Diferencia: false,
			},
			{
				Alumno:     alumnos.AlumnoEnClase{ID: 11, Identificador: "80002222", Nombres: "Luis", Apellidos: "Pérez"},
				Diferencia: false,
			},
		},
	}
}

func TestAsistenciaClase_SeleccionYLista(t *testing.T) {
	falso := &asistenciaFalsa{
		clases: []asistencia.ClaseAsignada{{
			AsignacionDocente: asistencia.AsignacionDocente{ID: 1},
			Curso:             "Matemática",
			Seccion:           "3.° A (2026)",
			BloqueNumero:      1,
			HoraInicio:        mustHora(t, "07:45"),
			HoraFin:           mustHora(t, "08:30"),
		}},
		detalle: detalleClaseDePrueba(t),
	}
	manejador, sesiones := nuevoManejadorAsistencia(t, falso)

	peticion := peticionConSesion(t, sesiones, http.MethodGet, "/asistencia", nil, "", "docente")
	respuesta := ejecutarAsistencia(manejador, peticion)
	assert.Equal(t, http.StatusOK, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "Matemática")

	parametros := url.Values{
		"fecha":      {fechas.FormatearISO(fechas.Hoy())},
		"seccion_id": {"5"},
		"curso_id":   {"2"},
		"bloque":     {"1"},
	}
	peticion = peticionConSesion(t, sesiones, http.MethodGet, "/asistencia/clase?"+parametros.Encode(), nil, "", "docente")
	respuesta = ejecutarAsistencia(manejador, peticion)
	assert.Equal(t, http.StatusOK, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "Ríos, Ana")
	assert.Contains(t, respuesta.Body.String(), "QR a las 07:42")
}

func TestAsistenciaClase_GuardaYMarchaTodos(t *testing.T) {
	falso := &asistenciaFalsa{detalle: detalleClaseDePrueba(t), marcarCreados: 2}
	manejador, sesiones := nuevoManejadorAsistencia(t, falso)

	valores := url.Values{
		"fecha":          {fechas.FormatearISO(fechas.Hoy())},
		"seccion_id":     {"5"},
		"curso_id":       {"2"},
		"bloque":         {"1"},
		"estado_10":      {asistencia.EstadoFalta},
		"observacion_10": {"Sin justificar"},
	}
	peticion := peticionPostFormulario(t, sesiones, "/asistencia/clase", valores, "docente")
	respuesta := ejecutarAsistencia(manejador, peticion)
	require.Equal(t, http.StatusSeeOther, respuesta.Code)
	require.Len(t, falso.guardados, 1)
	assert.Equal(t, int64(10), falso.guardados[0].AlumnoID)
	assert.Equal(t, asistencia.EstadoFalta, falso.guardados[0].Estado)
	assert.Equal(t, "Sin justificar", falso.guardados[0].Observacion)
	assert.Equal(t, 1, falso.bloqueRecibido)

	peticion = peticionPostFormulario(t, sesiones, "/asistencia/clase/todos", url.Values{
		"fecha":      {fechas.FormatearISO(fechas.Hoy())},
		"seccion_id": {"5"},
		"curso_id":   {"2"},
		"bloque":     {"1"},
	}, "docente")
	respuesta = ejecutarAsistencia(manejador, peticion)
	require.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Contains(t, respuesta.Header().Get("Location"), "ok=todos")
}

func TestAsistenciaClase_EnviaElBloqueIndicado(t *testing.T) {
	falso := &asistenciaFalsa{detalle: detalleClaseDePrueba(t)}
	manejador, sesiones := nuevoManejadorAsistencia(t, falso)

	peticion := peticionPostFormulario(t, sesiones, "/asistencia/clase", url.Values{
		"fecha":      {fechas.FormatearISO(fechas.Hoy())},
		"seccion_id": {"5"},
		"curso_id":   {"2"},
		"bloque":     {"2"},
		"estado_10":  {asistencia.EstadoPresente},
	}, "docente")
	respuesta := ejecutarAsistencia(manejador, peticion)

	require.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Equal(t, 2, falso.bloqueRecibido)
	assert.Contains(t, respuesta.Header().Get("Location"), "bloque=2")
}

func TestAsistenciaClase_DirectorNoAccedeAlRegistro(t *testing.T) {
	manejador, sesiones := nuevoManejadorAsistencia(t, &asistenciaFalsa{})
	peticion := peticionConSesion(t, sesiones, http.MethodGet, "/asistencia", nil, "", "director")

	respuesta := ejecutarAsistencia(manejador, peticion)

	assert.Equal(t, http.StatusForbidden, respuesta.Code)
}
