package unidad

import (
	"testing"
	"time"

	"asistencia/internal/alertas"

	"github.com/stretchr/testify/assert"
)

func TestRangoDelMes_CubreElMesCompleto(t *testing.T) {
	casos := []struct {
		nombre    string
		fecha     time.Time
		inicio    string
		fin       string
		ultimoDia int
	}{
		{"febrero bisiesto", time.Date(2024, time.February, 15, 0, 0, 0, 0, time.UTC), "2024-02-01", "2024-02-29", 29},
		{"febrero comun", time.Date(2025, time.February, 1, 0, 0, 0, 0, time.UTC), "2025-02-01", "2025-02-28", 28},
		{"diciembre", time.Date(2026, time.December, 31, 0, 0, 0, 0, time.UTC), "2026-12-01", "2026-12-31", 31},
		{"abril", time.Date(2026, time.April, 30, 0, 0, 0, 0, time.UTC), "2026-04-01", "2026-04-30", 30},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			inicio, fin := alertas.RangoDelMes(caso.fecha)
			assert.Equal(t, caso.inicio, inicio.Format("2006-01-02"))
			assert.Equal(t, caso.fin, fin.Format("2006-01-02"))
			assert.Equal(t, caso.ultimoDia, fin.Day())
		})
	}
}

func TestDiaCuentaParaRacha_ExcluyeFinesDeSemana(t *testing.T) {
	assert.False(t, alertas.DiaCuentaParaRacha(time.Saturday))
	assert.False(t, alertas.DiaCuentaParaRacha(time.Sunday))
	for _, dia := range []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday} {
		assert.True(t, alertas.DiaCuentaParaRacha(dia), dia.String())
	}
}

func TestRachaAlcanzaElUmbral(t *testing.T) {
	assert.True(t, alertas.AlcanzaUmbral(2, 2))
	assert.False(t, alertas.AlcanzaUmbral(1, 2))
	assert.False(t, alertas.AlcanzaUmbral(3, 2), "una racha mayor no vuelve a alertar")
}
