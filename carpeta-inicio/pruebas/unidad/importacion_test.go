package unidad

import (
	"testing"

	"asistencia/internal/alumnos"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func filaImportada(numero int, identificador, nombres, apellidos, grado, seccion string) alumnos.FilaImportada {
	return alumnos.FilaImportada{
		Numero:        numero,
		Identificador: identificador,
		Nombres:       nombres,
		Apellidos:     apellidos,
		Grado:         grado,
		Seccion:       seccion,
	}
}

func TestClasificarFilas_RechazaTodasLasOcurrenciasDuplicadas(t *testing.T) {
	validas, rechazadas := alumnos.ClasificarFilas([]alumnos.FilaImportada{
		filaImportada(2, "70000001", "Ana", "Ríos", "3.°", "A"),
		filaImportada(3, "70000002", "Luis", "Pérez", "3.°", "A"),
		filaImportada(4, "70000001", "Ana María", "Ríos", "3.°", "A"),
	})

	require.Len(t, validas, 1)
	assert.Equal(t, 3, validas[0].Numero)
	require.Len(t, rechazadas, 2)
	assert.Equal(t, alumnos.MotivoIdentificadorRepetidoEnArchivo, rechazadas[0].Motivo)
	assert.Equal(t, alumnos.MotivoIdentificadorRepetidoEnArchivo, rechazadas[1].Motivo)
	assert.Equal(t, 2, rechazadas[0].NumeroFila)
	assert.Equal(t, 4, rechazadas[1].NumeroFila)
	assert.Equal(t, "70000001", rechazadas[0].Identificador)
}

func TestClasificarFilas_RechazaCamposObligatoriosVacios(t *testing.T) {
	validas, rechazadas := alumnos.ClasificarFilas([]alumnos.FilaImportada{
		filaImportada(2, "70000001", "", "Ríos", "3.°", "A"),
		filaImportada(3, "70000002", "Luis", "Pérez", "", "A"),
		filaImportada(4, "70000003", "Elena", "Ñuño", "3.°", "A"),
	})

	require.Len(t, validas, 1)
	assert.Equal(t, 4, validas[0].Numero)
	require.Len(t, rechazadas, 2)
	assert.Equal(t, alumnos.MotivoCampoObligatorioVacio, rechazadas[0].Motivo)
	assert.Equal(t, alumnos.MotivoCampoObligatorioVacio, rechazadas[1].Motivo)
}

func TestClasificarFilas_RechazaIdentificadorConEspacios(t *testing.T) {
	validas, rechazadas := alumnos.ClasificarFilas([]alumnos.FilaImportada{
		filaImportada(2, "7000 0001", "Ana", "Ríos", "3.°", "A"),
	})

	assert.Empty(t, validas)
	require.Len(t, rechazadas, 1)
	assert.Equal(t, alumnos.MotivoFormatoInvalido, rechazadas[0].Motivo)
}

func TestClasificarFilas_ConservaTildesYNyLimpiaEspacios(t *testing.T) {
	validas, rechazadas := alumnos.ClasificarFilas([]alumnos.FilaImportada{
		{
			Numero:        2,
			Identificador: " 70000001 ",
			Nombres:       "  José María ",
			Apellidos:     " Núñez Ríos ",
			Grado:         " 3.° ",
			Seccion:       " A ",
		},
	})

	assert.Empty(t, rechazadas)
	require.Len(t, validas, 1)
	assert.Equal(t, "70000001", validas[0].Identificador)
	assert.Equal(t, "José María", validas[0].Nombres)
	assert.Equal(t, "Núñez Ríos", validas[0].Apellidos)
	assert.Equal(t, "3.°", validas[0].Grado)
	assert.Equal(t, "A", validas[0].Seccion)
}

func TestMotivoImportacionLegible(t *testing.T) {
	assert.Equal(t, "Campo obligatorio vacío", alumnos.MotivoImportacionLegible(alumnos.MotivoCampoObligatorioVacio))
	assert.Equal(t, "Identificador repetido en el archivo", alumnos.MotivoImportacionLegible(alumnos.MotivoIdentificadorRepetidoEnArchivo))
	assert.Equal(t, "El identificador ya existe", alumnos.MotivoImportacionLegible(alumnos.MotivoIdentificadorExistente))
	assert.Equal(t, "Grado o sección inexistente", alumnos.MotivoImportacionLegible(alumnos.MotivoSeccionInexistente))
	assert.Equal(t, "Formato inválido", alumnos.MotivoImportacionLegible(alumnos.MotivoFormatoInvalido))
}
