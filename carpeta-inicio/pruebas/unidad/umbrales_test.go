package unidad

import (
	"testing"

	"asistencia/internal/academico"

	"github.com/stretchr/testify/assert"
)

func TestValidarUmbral_AceptaValoresEntreUnoYCien(t *testing.T) {
	for _, valor := range []int{1, 3, 100} {
		assert.NoError(t, academico.ValidarUmbral(academico.TipoUmbralTardanzas, valor))
		assert.NoError(t, academico.ValidarUmbral(academico.TipoUmbralFaltas, valor))
		assert.NoError(t, academico.ValidarUmbral(academico.TipoUmbralFaltasSeguidas, valor))
	}
}

func TestValidarUmbral_RechazaValoresFueraDeRango(t *testing.T) {
	for _, valor := range []int{0, -1, 101} {
		assert.ErrorIs(t, academico.ValidarUmbral(academico.TipoUmbralTardanzas, valor),
			academico.ErrValorUmbralInvalido, "valor %d", valor)
	}
}

func TestValidarUmbral_RechazaTiposDesconocidos(t *testing.T) {
	assert.ErrorIs(t, academico.ValidarUmbral("otro", 3), academico.ErrTipoInvalido)
	assert.ErrorIs(t, academico.ValidarUmbral("", 3), academico.ErrTipoInvalido)
}
