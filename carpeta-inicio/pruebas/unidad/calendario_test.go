package unidad

import (
	"testing"
	"time"

	"asistencia/internal/academico"

	"github.com/stretchr/testify/assert"
)

func excepcion(fechaInicio, fechaFin time.Time, grados ...int64) academico.ExcepcionCalendario {
	return academico.ExcepcionCalendario{
		FechaInicio: fechaInicio,
		FechaFin:    fechaFin,
		Grados:      grados,
	}
}

func TestSeSolapan_DetectaCrucesInclusoEnLosBordes(t *testing.T) {
	base := time.Date(2027, time.July, 28, 0, 0, 0, 0, time.UTC)
	primera := excepcion(base, base.AddDate(0, 0, 2))

	assert.True(t, academico.SeSolapan(primera, excepcion(base.AddDate(0, 0, 2), base.AddDate(0, 0, 4))))
	assert.True(t, academico.SeSolapan(primera, excepcion(base.AddDate(0, 0, 1), base.AddDate(0, 0, 1))))
	assert.False(t, academico.SeSolapan(primera, excepcion(base.AddDate(0, 0, 3), base.AddDate(0, 0, 4))))
	assert.False(t, academico.SeSolapan(primera, excepcion(base.AddDate(0, 0, -2), base.AddDate(0, 0, -1))))
}

func TestMismoAlcance_ComparaGlobalesYPorGrado(t *testing.T) {
	base := time.Date(2027, time.July, 28, 0, 0, 0, 0, time.UTC)

	global := excepcion(base, base)
	global2 := excepcion(base, base)
	assert.True(t, academico.MismoAlcance(global, global2))

	grado3 := excepcion(base, base, 3)
	assert.False(t, academico.MismoAlcance(global, grado3))

	grado3y4 := excepcion(base, base, 3, 4)
	assert.True(t, academico.MismoAlcance(grado3, grado3y4))

	grado5 := excepcion(base, base, 5)
	assert.False(t, academico.MismoAlcance(grado3, grado5))
}

func TestValidarExcepcion_AceptaFinesDeSemana(t *testing.T) {
	sabado := time.Date(2027, time.July, 31, 0, 0, 0, 0, time.UTC)

	assert.NoError(t, academico.ValidarExcepcion(academico.TipoFeriado, "Feriado de fin de semana", sabado, sabado))
}
