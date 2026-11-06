package unidad

import (
	"testing"
	"time"

	"asistencia/internal/asistencia"

	"github.com/stretchr/testify/assert"
)

func TestDentroDelPlazo_UsaLaHoraDeCreacion(t *testing.T) {
	creado := time.Date(2026, time.November, 6, 8, 0, 0, 0, time.UTC)

	assert.True(t, asistencia.DentroDelPlazo(creado, creado.Add(23*time.Hour), 24))
	assert.True(t, asistencia.DentroDelPlazo(creado, creado.Add(24*time.Hour), 24))
	assert.False(t, asistencia.DentroDelPlazo(creado, creado.Add(24*time.Hour+time.Second), 24))
	assert.False(t, asistencia.DentroDelPlazo(creado, creado.Add(48*time.Hour), 24))
}

func TestValidarFechaNoFutura_UsaElDiaCivil(t *testing.T) {
	hoy := time.Date(2026, time.November, 6, 0, 0, 0, 0, time.UTC)
	zonaLima := time.FixedZone("America/Lima", -5*60*60)

	assert.NoError(t, asistencia.ValidarFechaNoFutura(hoy, hoy))
	assert.NoError(t, asistencia.ValidarFechaNoFutura(
		time.Date(2026, time.November, 6, 0, 0, 0, 0, zonaLima), hoy))
	assert.ErrorIs(t, asistencia.ValidarFechaNoFutura(
		time.Date(2026, time.November, 7, 0, 0, 0, 0, zonaLima), hoy),
		asistencia.ErrFechaFutura)
}

func TestValidarPlazoHoras_AceptaDeUnoACientoSesentaYOcho(t *testing.T) {
	for _, horas := range []int{1, 24, 168} {
		assert.NoError(t, asistencia.ValidarPlazoHoras(horas))
	}
	for _, horas := range []int{0, -1, 169} {
		assert.ErrorIs(t, asistencia.ValidarPlazoHoras(horas), asistencia.ErrPlazoFueraDeRango)
	}
}
