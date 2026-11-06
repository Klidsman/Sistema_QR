package unidad

import (
	"testing"

	"asistencia/internal/alumnos"

	"github.com/stretchr/testify/assert"
)

func TestMotivoRevocacion_Validos(t *testing.T) {
	for _, motivo := range []string{"perdida", "robo", "desactivacion", "retiro", "otro"} {
		assert.True(t, alumnos.MotivoRevocacionValido(motivo), "motivo %q", motivo)
	}
}

func TestMotivoRevocacion_RechazaVacioYDesconocido(t *testing.T) {
	assert.False(t, alumnos.MotivoRevocacionValido(""))
	assert.False(t, alumnos.MotivoRevocacionValido("porque si"))
	assert.False(t, alumnos.MotivoRevocacionValido("PERDIDA"))
}
