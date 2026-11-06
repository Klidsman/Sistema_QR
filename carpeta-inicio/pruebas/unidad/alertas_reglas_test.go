package unidad

import (
	"testing"

	"asistencia/internal/alertas"

	"github.com/stretchr/testify/assert"
)

func TestDestinatarioTardanza_UsaPorteroOSuRespaldo(t *testing.T) {
	portero := int64(7)
	assert.Equal(t, int64(7), alertas.DestinatarioTardanza(&portero, 1))
	assert.Equal(t, int64(1), alertas.DestinatarioTardanza(nil, 1))
	vacio := int64(0)
	assert.Equal(t, int64(1), alertas.DestinatarioTardanza(&vacio, 1))
}

func TestAlcanzaUmbral_SoloEnElValorExacto(t *testing.T) {
	assert.True(t, alertas.AlcanzaUmbral(3, 3))
	assert.False(t, alertas.AlcanzaUmbral(4, 3))
	assert.False(t, alertas.AlcanzaUmbral(2, 3))
	assert.False(t, alertas.AlcanzaUmbral(3, 0))
}

func TestTipoYEstadoValidos(t *testing.T) {
	assert.True(t, alertas.TipoValido(alertas.TipoTardanza))
	assert.True(t, alertas.TipoValido(alertas.TipoReiteracionTardanza))
	assert.True(t, alertas.TipoValido(alertas.TipoReiteracionFalta))
	assert.False(t, alertas.TipoValido("otro"))

	assert.True(t, alertas.EstadoValido(alertas.EstadoNueva))
	assert.True(t, alertas.EstadoValido(alertas.EstadoVista))
	assert.False(t, alertas.EstadoValido("leida"))
}
