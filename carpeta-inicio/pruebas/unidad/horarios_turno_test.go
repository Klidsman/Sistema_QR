package unidad

import (
	"testing"
	"time"

	"asistencia/internal/academico"

	"github.com/stretchr/testify/assert"
)

func TestValidarHorarioConVigencia_AceptaHoyYFuturo(t *testing.T) {
	hoy := time.Date(2026, time.November, 6, 0, 0, 0, 0, time.UTC)

	horario := horarioValido(t)
	horario.VigenteDesde = hoy
	assert.NoError(t, academico.ValidarHorarioConVigencia(horario, hoy))

	horario.VigenteDesde = hoy.AddDate(0, 1, 0)
	assert.NoError(t, academico.ValidarHorarioConVigencia(horario, hoy))
}

func TestValidarHorarioConVigencia_RechazaFechaAnteriorAHoy(t *testing.T) {
	hoy := time.Date(2026, time.November, 6, 0, 0, 0, 0, time.UTC)

	horario := horarioValido(t)
	horario.VigenteDesde = hoy.AddDate(0, 0, -1)

	assert.ErrorIs(t, academico.ValidarHorarioConVigencia(horario, hoy),
		academico.ErrVigenciaPasada)
}

func TestValidarHorarioConVigencia_AplicaLasReglasDelHorario(t *testing.T) {
	hoy := time.Date(2026, time.November, 6, 0, 0, 0, 0, time.UTC)

	horario := horarioValido(t)
	horario.VigenteDesde = hoy
	horario.ToleranciaMinutos = 40

	assert.ErrorIs(t, academico.ValidarHorarioConVigencia(horario, hoy),
		academico.ErrToleranciaExcedida)
}
