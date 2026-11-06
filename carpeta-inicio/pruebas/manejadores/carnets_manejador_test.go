package manejadores

import (
	"net/http"
	"testing"

	"asistencia/internal/alumnos"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func carnetsDePrueba() []alumnos.AlumnoCarnet {
	return []alumnos.AlumnoCarnet{
		{
			ID: 1, Identificador: "001", Nombres: "Ana", Apellidos: "Ríos",
			GradoID: 1, Grado: "3.°", Seccion: "A", AnioEscolar: "2026",
			FotoClave: "alumnos/1/foto.jpg", TokenQR: "token-ana",
		},
		{
			ID: 2, Identificador: "002", Nombres: "Luis", Apellidos: "Pérez",
			GradoID: 1, Grado: "3.°", Seccion: "A", AnioEscolar: "2026",
			FotoClave: "alumnos/2/foto.jpg", TokenQR: "token-luis",
		},
		{
			ID: 3, Identificador: "003", Nombres: "Elena", Apellidos: "Ñuño",
			GradoID: 1, Grado: "3.°", Seccion: "A", AnioEscolar: "2026",
			TokenQR: "token-elena",
		},
		{
			ID: 4, Identificador: "004", Nombres: "Mario", Apellidos: "Quispe",
			GradoID: 1, Grado: "3.°", Seccion: "A", AnioEscolar: "2026",
			FotoClave: "alumnos/4/foto.jpg",
		},
	}
}

func TestCarnetsManejador_SeleccionMuestraGradosYSecciones(t *testing.T) {
	servicio := &alumnosFalso{}
	manejador, sesiones := nuevoManejadorDePrueba(t, servicio, &qrFalso{})

	respuesta := ejecutar(manejador, peticionConSesion(t, sesiones, http.MethodGet, "/carnets", nil, "", "administrador"))

	require.Equal(t, http.StatusOK, respuesta.Code)
	cuerpo := respuesta.Body.String()
	assert.Contains(t, cuerpo, "Generar carnés")
	assert.Contains(t, cuerpo, "3.° A (2026)")
	assert.Contains(t, cuerpo, "value=\"1\"")
}

func TestCarnetsManejador_LoteMarcaSinFotoYAvisaSinQR(t *testing.T) {
	servicio := &alumnosFalso{carnets: carnetsDePrueba()}
	manejador, sesiones := nuevoManejadorDePrueba(t, servicio, &qrFalso{})

	respuesta := ejecutar(manejador, peticionConSesion(t, sesiones, http.MethodGet, "/carnets/lote?seccion_id=1", nil, "", "administrador"))

	require.Equal(t, http.StatusOK, respuesta.Code)
	require.Equal(t, int64(1), servicio.carnetSeleccion.SeccionID)

	cuerpo := respuesta.Body.String()
	assert.Contains(t, cuerpo, "Ríos, Ana")
	assert.Contains(t, cuerpo, "Pérez, Luis")
	assert.Contains(t, cuerpo, "Ñuño, Elena")
	assert.Contains(t, cuerpo, "Sin foto")

	// Los alumnos con QR vigente muestran su imagen.
	assert.Contains(t, cuerpo, "/alumnos/1/qr/imagen")
	assert.Contains(t, cuerpo, "/alumnos/2/qr/imagen")
	assert.Contains(t, cuerpo, "/alumnos/3/qr/imagen")

	// El alumno sin QR vigente no recibe un carné con QR vacío.
	assert.NotContains(t, cuerpo, "/alumnos/4/qr/imagen")
	assert.Contains(t, cuerpo, "sin QR vigente")
	assert.Contains(t, cuerpo, "Quispe, Mario")

	// El aviso previo indica cuántos alumnos requieren atención.
	assert.Contains(t, cuerpo, "1 alumno sin foto")
	assert.Contains(t, cuerpo, "1 alumno sin QR vigente")
}

func TestCarnetsManejador_LoteDisponibleParaDirector(t *testing.T) {
	servicio := &alumnosFalso{carnets: carnetsDePrueba()}
	manejador, sesiones := nuevoManejadorDePrueba(t, servicio, &qrFalso{})

	respuesta := ejecutar(manejador, peticionConSesion(t, sesiones, http.MethodGet, "/carnets/lote?grado_id=1", nil, "", "director"))

	require.Equal(t, http.StatusOK, respuesta.Code)
	assert.Equal(t, int64(1), servicio.carnetSeleccion.GradoID)
}

func TestCarnetsManejador_LoteSinSeleccionRedirige(t *testing.T) {
	servicio := &alumnosFalso{}
	manejador, sesiones := nuevoManejadorDePrueba(t, servicio, &qrFalso{})

	respuesta := ejecutar(manejador, peticionConSesion(t, sesiones, http.MethodGet, "/carnets/lote", nil, "", "administrador"))

	require.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Equal(t, "/carnets?error=seleccion", respuesta.Header().Get("Location"))
	assert.Zero(t, servicio.carnetSeleccion.SeccionID)
}
