package unidad

import (
	"testing"
	"time"

	"asistencia/internal/plataforma/intermediarios"

	"github.com/stretchr/testify/assert"
)

func TestLimiteIntentos_BloqueaTrasCincoFallos(t *testing.T) {
	ahora := time.Date(2026, 11, 5, 9, 0, 0, 0, time.UTC)
	limite := intermediarios.NuevoLimiteIntentos(5, 15*time.Minute, func() time.Time { return ahora })
	clave := "10.0.0.1|correo@colegio.pe"

	for i := 0; i < 5; i++ {
		assert.True(t, limite.Permitir(clave))
		limite.RegistrarFallo(clave)
	}

	assert.False(t, limite.Permitir(clave))
}

func TestLimiteIntentos_LaVentanaExpira(t *testing.T) {
	ahora := time.Date(2026, 11, 5, 9, 0, 0, 0, time.UTC)
	reloj := &relojFalso{actual: ahora}
	limite := intermediarios.NuevoLimiteIntentos(2, 15*time.Minute, reloj.Ahora)
	clave := "10.0.0.2|correo@colegio.pe"

	limite.RegistrarFallo(clave)
	limite.RegistrarFallo(clave)
	assert.False(t, limite.Permitir(clave))

	reloj.actual = ahora.Add(16 * time.Minute)
	assert.True(t, limite.Permitir(clave))
}

func TestLimiteIntentos_LimpiarTrasIngresoCorrecto(t *testing.T) {
	ahora := time.Date(2026, 11, 5, 9, 0, 0, 0, time.UTC)
	limite := intermediarios.NuevoLimiteIntentos(2, 15*time.Minute, func() time.Time { return ahora })
	clave := "10.0.0.3|correo@colegio.pe"

	limite.RegistrarFallo(clave)
	limite.RegistrarFallo(clave)
	limite.Limpiar(clave)

	assert.True(t, limite.Permitir(clave))
}

type relojFalso struct {
	actual time.Time
}

func (r *relojFalso) Ahora() time.Time {
	return r.actual
}
