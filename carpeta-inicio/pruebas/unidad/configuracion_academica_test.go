package unidad

import (
	"testing"

	"asistencia/internal/academico"

	"github.com/stretchr/testify/assert"
)

func TestNormalizarNombre_RecortaYColapsaEspacios(t *testing.T) {
	assert.Equal(t, "3.° A", academico.NormalizarNombre("  3.°    A  "))
	assert.Equal(t, "Educación Física", academico.NormalizarNombre("Educación Física"))
	assert.Equal(t, "", academico.NormalizarNombre("   "))
}

func TestValidarNombre_ExigeLongitudMinimaYMaxima(t *testing.T) {
	assert.NoError(t, academico.ValidarNombre("3.°"))
	assert.NoError(t, academico.ValidarNombre("Educación Física"))

	assert.ErrorIs(t, academico.ValidarNombre(""), academico.ErrNombreInvalido)
	assert.ErrorIs(t, academico.ValidarNombre("   "), academico.ErrNombreInvalido)
	assert.ErrorIs(t, academico.ValidarNombre("A"), academico.ErrNombreInvalido)
}

func TestValidarNombreSeccion_AceptaUnaLetra(t *testing.T) {
	assert.NoError(t, academico.ValidarNombreSeccion("A"))
	assert.NoError(t, academico.ValidarNombreSeccion(" B "))
	assert.ErrorIs(t, academico.ValidarNombreSeccion(""), academico.ErrNombreSeccionInvalido)
	assert.ErrorIs(t, academico.ValidarNombreSeccion("SECCIONDEMASIADOLARGA"), academico.ErrNombreSeccionInvalido)
}

func TestValidarTurno_SoloAceptaMananaYTarde(t *testing.T) {
	assert.NoError(t, academico.ValidarTurno(academico.TurnoManana))
	assert.NoError(t, academico.ValidarTurno(academico.TurnoTarde))
	assert.ErrorIs(t, academico.ValidarTurno("noche"), academico.ErrTurnoInvalido)
	assert.ErrorIs(t, academico.ValidarTurno(""), academico.ErrTurnoInvalido)
}
