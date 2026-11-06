package unidad

import (
	"testing"
	"time"

	"asistencia/internal/plataforma/fechas"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatearDia_NoMueveLaFechaDeUnValorDate(t *testing.T) {
	// pgx entrega los valores date como medianoche UTC.
	leidaDeBase := time.Date(2027, time.June, 30, 0, 0, 0, 0, time.UTC)

	assert.Equal(t, "30/06/2027", fechas.FormatearDia(leidaDeBase))
	assert.Equal(t, "2027-06-30", fechas.FormatearISO(leidaDeBase))
}

func TestParsearDia_ConservaElDiaEnLaZonaDelColegio(t *testing.T) {
	fecha, err := fechas.ParsearDia("2027-06-30")
	require.NoError(t, err)

	assert.Equal(t, 2027, fecha.Year())
	assert.Equal(t, time.June, fecha.Month())
	assert.Equal(t, 30, fecha.Day())
	assert.Equal(t, "2027-06-30", fechas.FormatearISO(fecha))
}

func TestParsearDia_RechazaFormatoInvalidoYVacio(t *testing.T) {
	_, err := fechas.ParsearDia("30/06/2027")
	assert.Error(t, err)

	_, err = fechas.ParsearDia("  ")
	assert.Error(t, err)
}
