package unidad

import (
	"encoding/base64"
	"testing"

	"asistencia/internal/alumnos"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerarTokenQR_TieneCuarentaYTresCaracteres(t *testing.T) {
	token, err := alumnos.GenerarTokenQR()
	require.NoError(t, err)

	assert.Len(t, token, 43)
}

func TestGenerarTokenQR_EsBase64URLEstatico(t *testing.T) {
	token, err := alumnos.GenerarTokenQR()
	require.NoError(t, err)

	decodificado, err := base64.RawURLEncoding.DecodeString(token)
	require.NoError(t, err)
	assert.Len(t, decodificado, 32)
}

func TestGenerarTokenQR_NoSeRepite(t *testing.T) {
	vistos := make(map[string]bool, 100)

	for i := 0; i < 100; i++ {
		token, err := alumnos.GenerarTokenQR()
		require.NoError(t, err)
		assert.False(t, vistos[token], "el token se repitió")
		vistos[token] = true
	}
}
