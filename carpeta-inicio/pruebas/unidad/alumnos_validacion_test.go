package unidad

import (
	"testing"

	"asistencia/internal/alumnos"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func datosValidos() alumnos.DatosAlumno {
	return alumnos.DatosAlumno{
		Identificador: "00123456",
		Nombres:       "José",
		Apellidos:     "Núñez",
		SeccionID:     1,
	}
}

func TestValidar_DatosCompletosNoTienenErrores(t *testing.T) {
	assert.Nil(t, alumnos.Validar(datosValidos()))
}

func TestValidar_ExigeCamposObligatorios(t *testing.T) {
	datos := alumnos.DatosAlumno{}

	errores := alumnos.Validar(datos)
	require.NotNil(t, errores)

	assert.Contains(t, errores.Campos, "identificador")
	assert.Contains(t, errores.Campos, "nombres")
	assert.Contains(t, errores.Campos, "apellidos")
	assert.Contains(t, errores.Campos, "seccion_id")
}

func TestValidar_RechazaIdentificadorConEspaciosInternos(t *testing.T) {
	datos := datosValidos()
	datos.Identificador = "0012 3456"

	errores := alumnos.Validar(datos)

	require.NotNil(t, errores)
	assert.Contains(t, errores.Campos, "identificador")
}

func TestValidar_AceptaIdentificadorConEspaciosEnLosExtremos(t *testing.T) {
	datos := datosValidos()
	datos.Identificador = "  00123456  "

	assert.Nil(t, alumnos.Validar(datos))
}

func TestValidar_RechazaNombresEnBlanco(t *testing.T) {
	datos := datosValidos()
	datos.Nombres = "   "

	errores := alumnos.Validar(datos)

	require.NotNil(t, errores)
	assert.Contains(t, errores.Campos, "nombres")
}
