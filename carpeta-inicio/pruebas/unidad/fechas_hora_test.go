package unidad

import (
	"testing"

	"asistencia/internal/plataforma/fechas"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsearHora_AceptaFormatoHHMM(t *testing.T) {
	hora, err := fechas.ParsearHora("07:30")
	require.NoError(t, err)

	assert.Equal(t, 7, hora.Hour())
	assert.Equal(t, 30, hora.Minute())
	assert.Equal(t, "07:30", fechas.FormatearHora(hora))
}

func TestParsearHora_RechazaFormatosInvalidos(t *testing.T) {
	for _, texto := range []string{"", "7:30", "0730", "07:60", "24:00", "07:30:00"} {
		_, err := fechas.ParsearHora(texto)
		assert.Error(t, err, "la hora %q debe rechazarse", texto)
	}
}

func TestSumarMinutos_NoCambiaDeDiaDentroDelHorario(t *testing.T) {
	hora, err := fechas.ParsearHora("07:30")
	require.NoError(t, err)

	assert.Equal(t, "07:40", fechas.FormatearHora(fechas.SumarMinutos(hora, 10)))
	assert.Equal(t, "12:30", fechas.FormatearHora(fechas.SumarMinutos(hora, 300)))
}

func TestCompararHoras_OrdenaDentroDelMismoDia(t *testing.T) {
	temprano, err := fechas.ParsearHora("07:30")
	require.NoError(t, err)
	igual, err := fechas.ParsearHora("07:30")
	require.NoError(t, err)
	tarde, err := fechas.ParsearHora("08:00")
	require.NoError(t, err)

	assert.Equal(t, -1, fechas.CompararHoras(temprano, tarde))
	assert.Equal(t, 0, fechas.CompararHoras(temprano, igual))
	assert.Equal(t, 1, fechas.CompararHoras(tarde, temprano))
}
