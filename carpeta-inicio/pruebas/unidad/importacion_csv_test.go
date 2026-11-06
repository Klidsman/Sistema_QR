package unidad

import (
	"errors"
	"testing"

	"asistencia/internal/alumnos"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLeerCSV_ConBOMYSeparadorPuntoYComa(t *testing.T) {
	contenido := append([]byte{0xEF, 0xBB, 0xBF},
		[]byte("identificador;nombres;apellidos;grado;seccion;apoderado_nombre\n"+
			"70000001;José;Núñez;3.°;A;María Núñez\n"+
			"70000002;Luis;Pérez;3.°;B;\n")...)

	filas, err := alumnos.LeerCSV(contenido, 100)

	require.NoError(t, err)
	require.Len(t, filas, 2)
	assert.Equal(t, 2, filas[0].Numero)
	assert.Equal(t, "70000001", filas[0].Identificador)
	assert.Equal(t, "José", filas[0].Nombres)
	assert.Equal(t, "Núñez", filas[0].Apellidos)
	assert.Equal(t, "3.°", filas[0].Grado)
	assert.Equal(t, "A", filas[0].Seccion)
	assert.Equal(t, "María Núñez", filas[0].ApoderadoNombre)
	assert.Equal(t, 3, filas[1].Numero)
}

func TestLeerCSV_ConSeparadorComa(t *testing.T) {
	contenido := []byte("identificador,nombres,apellidos,grado,seccion\n" +
		"70000001,Ana,Ríos,3.°,A\n")

	filas, err := alumnos.LeerCSV(contenido, 100)

	require.NoError(t, err)
	require.Len(t, filas, 1)
	assert.Equal(t, "Ana", filas[0].Nombres)
}

func TestLeerCSV_IgnoraFilasVacias(t *testing.T) {
	contenido := []byte("identificador,nombres,apellidos,grado,seccion\n" +
		"70000001,Ana,Ríos,3.°,A\n" +
		",,,,\n")

	filas, err := alumnos.LeerCSV(contenido, 100)

	require.NoError(t, err)
	assert.Len(t, filas, 1)
}

func TestLeerCSV_RechazaColumnasFaltantes(t *testing.T) {
	contenido := []byte("identificador,nombres,apellidos\n" +
		"70000001,Ana,Ríos\n")

	_, err := alumnos.LeerCSV(contenido, 100)

	var faltantes *alumnos.ErrorColumnasFaltantes
	require.ErrorAs(t, err, &faltantes)
	assert.Contains(t, faltantes.Columnas, "grado")
	assert.Contains(t, faltantes.Columnas, "seccion")
}

func TestLeerCSV_RechazaDemasiadasFilas(t *testing.T) {
	contenido := []byte("identificador,nombres,apellidos,grado,seccion\n" +
		"70000001,Ana,Ríos,3.°,A\n" +
		"70000002,Luis,Pérez,3.°,A\n")

	_, err := alumnos.LeerCSV(contenido, 1)

	assert.True(t, errors.Is(err, alumnos.ErrDemasiadasFilas))
}
