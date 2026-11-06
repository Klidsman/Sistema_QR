package unidad

import (
	"testing"

	"asistencia/internal/academico"

	"github.com/stretchr/testify/assert"
)

func TestValidarBloque_AceptaBordesDelHorario(t *testing.T) {
	horario := horarioValido(t)
	horario.HoraCierre = hora(t, "12:30")
	horario.HoraIngreso = hora(t, "07:30")

	bloque := academico.BloqueClase{
		Turno:      academico.TurnoManana,
		Numero:     1,
		HoraInicio: hora(t, "07:30"),
		HoraFin:    hora(t, "12:30"),
	}

	assert.NoError(t, academico.ValidarBloque(bloque, horario, nil))
}

func TestValidarBloque_IgnoraBloquesDeOtroTurno(t *testing.T) {
	horario := horarioValido(t)
	horario.HoraCierre = hora(t, "12:30")

	otros := []academico.BloqueClase{{
		Turno:      academico.TurnoTarde,
		Numero:     9,
		HoraInicio: hora(t, "08:00"),
		HoraFin:    hora(t, "09:00"),
	}}

	assert.NoError(t, academico.ValidarBloque(bloqueValido(t), horario, otros))
}

func TestValidarBloque_RechazaNumeroFueraDeRango(t *testing.T) {
	horario := horarioValido(t)
	horario.HoraCierre = hora(t, "12:30")

	for _, numero := range []int{0, 21} {
		bloque := bloqueValido(t)
		bloque.Numero = numero
		assert.ErrorIs(t, academico.ValidarBloque(bloque, horario, nil),
			academico.ErrBloqueNumeroInvalido, "número %d", numero)
	}
}
