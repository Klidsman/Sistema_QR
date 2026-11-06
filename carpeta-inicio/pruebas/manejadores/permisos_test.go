package manejadores

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// rutasPostDeAlumnos reúne todas las rutas de escritura que expone el módulo de alumnos.
var rutasPostDeAlumnos = []string{
	"/alumnos",
	"/alumnos/importar",
	"/alumnos/7",
	"/alumnos/7/desactivacion",
	"/alumnos/7/seccion",
	"/alumnos/7/salida",
	"/alumnos/7/foto",
	"/alumnos/7/qr",
	"/alumnos/7/qr/reemplazo",
}

func TestPermisos_DirectorRecibeProhibidoEnTodasLasRutasDeEscritura(t *testing.T) {
	manejador, sesiones := nuevoManejadorDePrueba(t, &alumnosFalso{}, &qrFalso{})

	for _, ruta := range rutasPostDeAlumnos {
		t.Run(ruta, func(t *testing.T) {
			peticion := peticionConSesion(t, sesiones, http.MethodPost, ruta, nil, "", "director")
			respuesta := ejecutar(manejador, peticion)

			assert.Equal(t, http.StatusForbidden, respuesta.Code,
				"el director no debe poder escribir en %s", ruta)
			assert.NotContains(t, respuesta.Body.String(), "no se pudo")
		})
	}
}

func TestPermisos_DirectorConservaAccesoDeLectura(t *testing.T) {
	manejador, sesiones := nuevoManejadorDePrueba(t, &alumnosFalso{}, &qrFalso{})

	for _, ruta := range []string{"/alumnos", "/alumnos/importaciones", "/carnets"} {
		t.Run(ruta, func(t *testing.T) {
			peticion := peticionConSesion(t, sesiones, http.MethodGet, ruta, nil, "", "director")
			respuesta := ejecutar(manejador, peticion)

			assert.Equal(t, http.StatusOK, respuesta.Code,
				"el director debe poder leer %s", ruta)
		})
	}
}

func TestPermisos_SinSesionRedirigeAlIngreso(t *testing.T) {
	manejador, sesiones := nuevoManejadorDePrueba(t, &alumnosFalso{}, &qrFalso{})
	peticion := httptest.NewRequest(http.MethodPost, "/alumnos", nil)
	ctx, err := sesiones.Load(peticion.Context(), "token-sin-usuario")
	require.NoError(t, err)
	sesiones.Put(ctx, "marcador", true)
	peticion = peticion.WithContext(ctx)

	respuesta := ejecutar(manejador, peticion)

	assert.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Equal(t, "/ingreso", respuesta.Header().Get("Location"))
}
