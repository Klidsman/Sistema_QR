package manejadores

import (
	"net/http"
	"testing"
	"time"

	"asistencia/internal/alumnos"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAlumnosManejador_CambiarSeccion(t *testing.T) {
	servicio := &alumnosFalso{alumno: alumnoDePrueba()}
	manejador, sesiones := nuevoManejadorDePrueba(t, servicio, &qrFalso{})

	cuerpo, tipo := cuerpoMultipart(map[string]string{
		"seccion_id": "2",
		"fecha":      "2026-05-15",
	})
	respuesta := ejecutar(manejador, peticionConSesion(t, sesiones, http.MethodPost, "/alumnos/7/seccion", cuerpo, tipo, "administrador"))

	require.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Equal(t, "/alumnos/7?ok=seccion_cambiada", respuesta.Header().Get("Location"))
	assert.Equal(t, int64(2), servicio.cambiarSeccionID)
	assert.Equal(t, "2026-05-15", servicio.cambiarSeccionFecha.Format("2006-01-02"))
}

func TestAlumnosManejador_RegistrarSalida(t *testing.T) {
	servicio := &alumnosFalso{alumno: alumnoDePrueba()}
	manejador, sesiones := nuevoManejadorDePrueba(t, servicio, &qrFalso{})

	cuerpo, tipo := cuerpoMultipart(map[string]string{
		"tipo":  "trasladado",
		"fecha": "2026-06-30",
	})
	respuesta := ejecutar(manejador, peticionConSesion(t, sesiones, http.MethodPost, "/alumnos/7/salida", cuerpo, tipo, "administrador"))

	require.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Equal(t, "/alumnos/7?ok=salida_registrada", respuesta.Header().Get("Location"))
	assert.Equal(t, "trasladado", servicio.salidaTipo)
	assert.Equal(t, "2026-06-30", servicio.salidaFecha.Format("2006-01-02"))
}

func TestAlumnosManejador_DirectorNoPuedeCambiarSeccionNiSalir(t *testing.T) {
	servicio := &alumnosFalso{alumno: alumnoDePrueba()}
	manejador, sesiones := nuevoManejadorDePrueba(t, servicio, &qrFalso{})

	cuerpo, tipo := cuerpoMultipart(map[string]string{"seccion_id": "2", "fecha": "2026-05-15"})
	cambio := ejecutar(manejador, peticionConSesion(t, sesiones, http.MethodPost, "/alumnos/7/seccion", cuerpo, tipo, "director"))
	assert.Equal(t, http.StatusForbidden, cambio.Code)

	cuerpo, tipo = cuerpoMultipart(map[string]string{"tipo": "retirado", "fecha": "2026-06-30"})
	salida := ejecutar(manejador, peticionConSesion(t, sesiones, http.MethodPost, "/alumnos/7/salida", cuerpo, tipo, "director"))
	assert.Equal(t, http.StatusForbidden, salida.Code)

	assert.Zero(t, servicio.cambiarSeccionID)
	assert.Empty(t, servicio.salidaTipo)
}

func TestAlumnosManejador_SeccionInvalida(t *testing.T) {
	servicio := &alumnosFalso{alumno: alumnoDePrueba()}
	manejador, sesiones := nuevoManejadorDePrueba(t, servicio, &qrFalso{})

	cuerpo, tipo := cuerpoMultipart(map[string]string{"seccion_id": "0", "fecha": "2026-05-15"})
	respuesta := ejecutar(manejador, peticionConSesion(t, sesiones, http.MethodPost, "/alumnos/7/seccion", cuerpo, tipo, "administrador"))

	require.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Equal(t, "/alumnos/7?error=seccion_invalida", respuesta.Header().Get("Location"))
}

func TestAlumnosManejador_MismaSeccion(t *testing.T) {
	servicio := &alumnosFalso{alumno: alumnoDePrueba(), cambiarSeccionErr: alumnos.ErrMismaSeccion}
	manejador, sesiones := nuevoManejadorDePrueba(t, servicio, &qrFalso{})

	cuerpo, tipo := cuerpoMultipart(map[string]string{"seccion_id": "1", "fecha": "2026-05-15"})
	respuesta := ejecutar(manejador, peticionConSesion(t, sesiones, http.MethodPost, "/alumnos/7/seccion", cuerpo, tipo, "administrador"))

	require.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Equal(t, "/alumnos/7?error=misma_seccion", respuesta.Header().Get("Location"))
}

func TestAlumnosManejador_SalidaConTipoInvalido(t *testing.T) {
	servicio := &alumnosFalso{alumno: alumnoDePrueba(), salidaErr: alumnos.ErrTipoSalidaInvalido}
	manejador, sesiones := nuevoManejadorDePrueba(t, servicio, &qrFalso{})

	cuerpo, tipo := cuerpoMultipart(map[string]string{"tipo": "activo", "fecha": "2026-06-30"})
	respuesta := ejecutar(manejador, peticionConSesion(t, sesiones, http.MethodPost, "/alumnos/7/salida", cuerpo, tipo, "administrador"))

	require.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Equal(t, "/alumnos/7?error=tipo_salida_invalido", respuesta.Header().Get("Location"))
}

func TestAlumnosManejador_FichaMuestraHistorialDeSecciones(t *testing.T) {
	cierre := time.Date(2026, time.May, 15, 0, 0, 0, 0, time.UTC)
	servicio := &alumnosFalso{
		alumno: alumnoDePrueba(),
		historialSecciones: []alumnos.HistorialSeccion{
			{ID: 2, SeccionID: 1, Grado: "3.°", Seccion: "A", AnioEscolar: "2026", FechaInicio: cierre},
			{ID: 1, SeccionID: 1, Grado: "2.°", Seccion: "A", AnioEscolar: "2026", FechaInicio: cierre.AddDate(0, -3, 0), FechaFin: &cierre},
		},
	}
	manejador, sesiones := nuevoManejadorDePrueba(t, servicio, &qrFalso{})

	respuesta := ejecutar(manejador, peticionConSesion(t, sesiones, http.MethodGet, "/alumnos/7", nil, "", "administrador"))

	require.Equal(t, http.StatusOK, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "Historial de secciones")
	assert.Contains(t, respuesta.Body.String(), "Cambiar de sección")
	assert.Contains(t, respuesta.Body.String(), "Registrar salida")
}

func TestAlumnosManejador_DirectorNoVeAccionesDeSalida(t *testing.T) {
	servicio := &alumnosFalso{alumno: alumnoDePrueba()}
	manejador, sesiones := nuevoManejadorDePrueba(t, servicio, &qrFalso{})

	respuesta := ejecutar(manejador, peticionConSesion(t, sesiones, http.MethodGet, "/alumnos/7", nil, "", "director"))

	require.Equal(t, http.StatusOK, respuesta.Code)
	assert.NotContains(t, respuesta.Body.String(), "Cambiar de sección")
	assert.NotContains(t, respuesta.Body.String(), "Registrar salida")
}
