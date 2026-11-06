package unidad

import (
	"testing"

	"asistencia/internal/academico"

	"github.com/stretchr/testify/assert"
)

func TestValidarGradosCurso_ExigeAlMenosUnGrado(t *testing.T) {
	assert.NoError(t, academico.ValidarGradosCurso([]int64{1}))
	assert.NoError(t, academico.ValidarGradosCurso([]int64{1, 2}))
	assert.ErrorIs(t, academico.ValidarGradosCurso(nil), academico.ErrCursoSinGrados)
	assert.ErrorIs(t, academico.ValidarGradosCurso([]int64{}), academico.ErrCursoSinGrados)
}

func TestValidarDiaSemana_SoloLunesAViernes(t *testing.T) {
	for dia := 1; dia <= 5; dia++ {
		assert.NoError(t, academico.ValidarDiaSemana(dia))
	}
	assert.ErrorIs(t, academico.ValidarDiaSemana(0), academico.ErrDiaFueraDeRango)
	assert.ErrorIs(t, academico.ValidarDiaSemana(6), academico.ErrDiaFueraDeRango)
}
