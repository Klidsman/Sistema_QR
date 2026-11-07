package manejadores

import (
	"net/http"
	"net/url"
	"testing"

	"asistencia/internal/academico"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfiguracionGrados_AdministradorElimina(t *testing.T) {
	var eliminado int64
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen: academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
		eliminarGrado: func(gradoID int64) error {
			eliminado = gradoID
			return nil
		},
	})

	peticion := peticionFormulario(t, sesiones, "/configuracion/grados/4/eliminacion", url.Values{}, "administrador")
	respuesta := ejecutarConfiguracion(manejador, peticion)

	require.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Equal(t, "/configuracion/grados?ok=eliminado", respuesta.Header().Get("Location"))
	assert.Equal(t, int64(4), eliminado)
}

func TestConfiguracionGrados_NoEliminaConDependencias(t *testing.T) {
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen: academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
		eliminarGrado: func(int64) error {
			return academico.ErrGradoConDependencias
		},
	})

	peticion := peticionFormulario(t, sesiones, "/configuracion/grados/4/eliminacion", url.Values{}, "administrador")
	respuesta := ejecutarConfiguracion(manejador, peticion)

	require.Equal(t, http.StatusUnprocessableEntity, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "No se puede eliminar el grado")
}

func TestConfiguracionSecciones_AdministradorElimina(t *testing.T) {
	var eliminada int64
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen: academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
		eliminarSeccion: func(seccionID int64) error {
			eliminada = seccionID
			return nil
		},
	})

	peticion := peticionFormulario(t, sesiones, "/configuracion/secciones/9/eliminacion", url.Values{}, "administrador")
	respuesta := ejecutarConfiguracion(manejador, peticion)

	require.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Equal(t, "/configuracion/secciones?ok=eliminada", respuesta.Header().Get("Location"))
	assert.Equal(t, int64(9), eliminada)
}

func TestConfiguracionSecciones_DirectorNoElimina(t *testing.T) {
	var llamada bool
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen: academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
		eliminarSeccion: func(int64) error {
			llamada = true
			return nil
		},
	})

	peticion := peticionFormulario(t, sesiones, "/configuracion/secciones/9/eliminacion", url.Values{}, "director")
	respuesta := ejecutarConfiguracion(manejador, peticion)

	assert.Equal(t, http.StatusForbidden, respuesta.Code)
	assert.False(t, llamada)
}

func TestConfiguracionSecciones_NoEliminaConRegistros(t *testing.T) {
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen: academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
		eliminarSeccion: func(int64) error {
			return academico.ErrSeccionConDependencias
		},
	})

	peticion := peticionFormulario(t, sesiones, "/configuracion/secciones/9/eliminacion", url.Values{}, "administrador")
	respuesta := ejecutarConfiguracion(manejador, peticion)

	require.Equal(t, http.StatusUnprocessableEntity, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "No se puede eliminar la sección")
}
