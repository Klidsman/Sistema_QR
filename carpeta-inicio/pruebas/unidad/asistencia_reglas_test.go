package unidad

import (
	"strings"
	"testing"
	"time"

	"asistencia/internal/asistencia"

	"github.com/stretchr/testify/assert"
)

func TestEstadoValido_SoloAceptaLosCuatroEstados(t *testing.T) {
	for _, estado := range []string{"presente", "tardanza", "falta", "falta_justificada"} {
		assert.True(t, asistencia.EstadoValido(estado), estado)
	}
	for _, estado := range []string{"", "ausente", "puntual", "PRESENTE"} {
		assert.False(t, asistencia.EstadoValido(estado), estado)
	}
}

func TestValidarObservacion_AceptaHastaTrescientosCaracteres(t *testing.T) {
	assert.NoError(t, asistencia.ValidarObservacion(""))
	assert.NoError(t, asistencia.ValidarObservacion(strings.Repeat("a", 300)))
	assert.ErrorIs(t, asistencia.ValidarObservacion(strings.Repeat("a", 301)),
		asistencia.ErrObservacionLarga)
}

func TestHayDiferencia_ComparaPresenciaNoClasificacion(t *testing.T) {
	marcacion := &asistencia.Marcacion{Estado: asistencia.MarcacionPuntual}

	assert.True(t, asistencia.HayDiferencia(marcacion, asistencia.EstadoFalta))
	assert.True(t, asistencia.HayDiferencia(marcacion, asistencia.EstadoFaltaJustificada))
	assert.False(t, asistencia.HayDiferencia(marcacion, asistencia.EstadoPresente))
	assert.False(t, asistencia.HayDiferencia(marcacion, asistencia.EstadoTardanza))

	assert.True(t, asistencia.HayDiferencia(nil, asistencia.EstadoPresente))
	assert.True(t, asistencia.HayDiferencia(nil, asistencia.EstadoTardanza))
	assert.False(t, asistencia.HayDiferencia(nil, asistencia.EstadoFalta))
	assert.False(t, asistencia.HayDiferencia(nil, asistencia.EstadoFaltaJustificada))
}

func TestValidarRangoExportacion_ExigeOrden(t *testing.T) {
	desde := time.Date(2026, time.November, 1, 0, 0, 0, 0, time.UTC)
	assert.NoError(t, asistencia.ValidarRangoExportacion(desde, desde))
	assert.NoError(t, asistencia.ValidarRangoExportacion(desde, desde.AddDate(0, 0, 7)))
	assert.ErrorIs(t, asistencia.ValidarRangoExportacion(desde.AddDate(0, 0, 7), desde),
		asistencia.ErrRangoFechasInvalido)
}
