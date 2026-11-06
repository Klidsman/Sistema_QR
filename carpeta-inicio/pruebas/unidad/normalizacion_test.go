package unidad

import (
	"testing"

	"asistencia/internal/alumnos"

	"github.com/stretchr/testify/assert"
)

func TestNormalizar_QuitaTildesYMinusculas(t *testing.T) {
	casos := map[string]string{
		"José":                 "jose",
		"Núñez":                "nunez",
		"MUÑOZ":                "munoz",
		"ÁÉÍÓÚÜ":               "aeiouu",
		"María de los Ángeles": "maria de los angeles",
		"  Ana   Lucía  ":      "ana lucia",
		"O'Brien":              "o'brien",
		"3.° A":                "3. a",
	}

	for entrada, esperado := range casos {
		assert.Equal(t, esperado, alumnos.Normalizar(entrada), "entrada %q", entrada)
	}
}

func TestLimpiarIdentificador_QuitaEspaciosDeLosExtremos(t *testing.T) {
	assert.Equal(t, "00123456", alumnos.LimpiarIdentificador("  00123456  "))
	assert.Equal(t, "00123456", alumnos.LimpiarIdentificador("00123456"))
}
