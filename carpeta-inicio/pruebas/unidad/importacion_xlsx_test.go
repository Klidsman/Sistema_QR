package unidad

import (
	"bytes"
	"testing"

	"asistencia/internal/alumnos"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func TestLeerXLSX_LeeLaPrimeraHojaYConservaTildes(t *testing.T) {
	archivo := excelize.NewFile()
	defer func() { _ = archivo.Close() }()

	hoja := archivo.GetSheetName(0)
	require.NoError(t, archivo.SetSheetRow(hoja, "A1", &[]any{
		"identificador", "nombres", "apellidos", "grado", "seccion",
	}))
	require.NoError(t, archivo.SetSheetRow(hoja, "A2", &[]any{
		"70000001", "José María", "Núñez Ríos", "3.°", "A",
	}))

	var buffer bytes.Buffer
	require.NoError(t, archivo.Write(&buffer))

	filas, err := alumnos.LeerXLSX(buffer.Bytes(), 100)

	require.NoError(t, err)
	require.Len(t, filas, 1)
	assert.Equal(t, 2, filas[0].Numero)
	assert.Equal(t, "José María", filas[0].Nombres)
	assert.Equal(t, "Núñez Ríos", filas[0].Apellidos)
}
