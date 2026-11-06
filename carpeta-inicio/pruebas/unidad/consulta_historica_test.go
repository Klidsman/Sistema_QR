package unidad

import (
	"testing"
	"time"

	"asistencia/internal/academico"

	"github.com/stretchr/testify/assert"
)

func TestAnioCubreFecha_IncluyeLosBordes(t *testing.T) {
	anio := academico.AnioEscolar{
		ID:          1,
		Nombre:      "2027",
		FechaInicio: time.Date(2027, time.March, 1, 0, 0, 0, 0, time.UTC),
		FechaFin:    time.Date(2027, time.December, 15, 0, 0, 0, 0, time.UTC),
	}
	zonaLima := time.FixedZone("America/Lima", -5*60*60)

	assert.True(t, academico.AnioCubreFecha(anio, time.Date(2027, time.March, 1, 0, 0, 0, 0, zonaLima)))
	assert.True(t, academico.AnioCubreFecha(anio, time.Date(2027, time.December, 15, 0, 0, 0, 0, zonaLima)))
	assert.True(t, academico.AnioCubreFecha(anio, time.Date(2027, time.June, 15, 0, 0, 0, 0, zonaLima)))
	assert.False(t, academico.AnioCubreFecha(anio, time.Date(2027, time.February, 28, 0, 0, 0, 0, zonaLima)))
	assert.False(t, academico.AnioCubreFecha(anio, time.Date(2027, time.December, 16, 0, 0, 0, 0, zonaLima)))
}

func TestVersionVigente_EligeLaMasRecienteHastaLaFecha(t *testing.T) {
	type version struct {
		nombre string
		desde  time.Time
	}
	versiones := []version{
		{"enero", time.Date(2027, time.January, 1, 0, 0, 0, 0, time.UTC)},
		{"marzo", time.Date(2027, time.March, 1, 0, 0, 0, 0, time.UTC)},
		{"junio", time.Date(2027, time.June, 1, 0, 0, 0, 0, time.UTC)},
	}
	desde := func(v version) time.Time { return v.desde }

	elegida, aplica := academico.VersionVigente(versiones,
		time.Date(2027, time.March, 15, 0, 0, 0, 0, time.UTC), desde)
	assert.True(t, aplica)
	assert.Equal(t, "marzo", elegida.nombre)

	elegida, aplica = academico.VersionVigente(versiones,
		time.Date(2028, time.January, 1, 0, 0, 0, 0, time.UTC), desde)
	assert.True(t, aplica)
	assert.Equal(t, "junio", elegida.nombre)

	_, aplica = academico.VersionVigente(versiones,
		time.Date(2026, time.December, 31, 0, 0, 0, 0, time.UTC), desde)
	assert.False(t, aplica)
}
