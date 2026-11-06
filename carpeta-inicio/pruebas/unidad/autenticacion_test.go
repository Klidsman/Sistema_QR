package unidad

import (
	"strings"
	"testing"

	"asistencia/internal/autenticacion"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashearContrasenia_VerificaLaMismaContrasenia(t *testing.T) {
	hash, err := autenticacion.HashearContrasenia("clave-segura-123")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(hash, "$argon2id$"))

	ok, err := autenticacion.VerificarContrasenia(hash, "clave-segura-123")
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestVerificarContrasenia_RechazaOtraContrasenia(t *testing.T) {
	hash, err := autenticacion.HashearContrasenia("clave-segura-123")
	require.NoError(t, err)

	ok, err := autenticacion.VerificarContrasenia(hash, "clave-incorrecta")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestHashearContrasenia_UsaSalAleatoria(t *testing.T) {
	primero, err := autenticacion.HashearContrasenia("misma-clave")
	require.NoError(t, err)
	segundo, err := autenticacion.HashearContrasenia("misma-clave")
	require.NoError(t, err)

	assert.NotEqual(t, primero, segundo)
}

func TestVerificarContrasenia_HashInvalidoDevuelveError(t *testing.T) {
	_, err := autenticacion.VerificarContrasenia("no-es-un-hash", "clave")

	assert.Error(t, err)
}
