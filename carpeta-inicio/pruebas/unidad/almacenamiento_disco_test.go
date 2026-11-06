package unidad

import (
	"bytes"
	"context"
	"io"
	"testing"

	"asistencia/internal/plataforma/almacenamiento"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAlmacenDisco_GuardaYObtieneElObjeto(t *testing.T) {
	almacen := almacenamiento.NuevoAlmacenDisco(t.TempDir())
	ctx := context.Background()

	require.NoError(t, almacen.Guardar(ctx, "alumnos/12/foto.jpg", bytes.NewReader([]byte("imagen"))))

	lectura, err := almacen.Obtener(ctx, "alumnos/12/foto.jpg")
	require.NoError(t, err)
	defer func() { _ = lectura.Close() }()
	contenido, err := io.ReadAll(lectura)
	require.NoError(t, err)
	assert.Equal(t, "imagen", string(contenido))
}

func TestAlmacenDisco_RechazaClavesQueSalenDeLaRaiz(t *testing.T) {
	almacen := almacenamiento.NuevoAlmacenDisco(t.TempDir())
	ctx := context.Background()

	for _, clave := range []string{"../secreto.txt", "alumnos/../../secreto.txt", "/etc/passwd", "."} {
		err := almacen.Guardar(ctx, clave, bytes.NewReader(nil))
		assert.Error(t, err, "la clave %q debe rechazarse", clave)
	}
}

func TestAlmacenDisco_ObjetoInexistente(t *testing.T) {
	almacen := almacenamiento.NuevoAlmacenDisco(t.TempDir())

	_, err := almacen.Obtener(context.Background(), "alumnos/99/foto.jpg")

	assert.ErrorIs(t, err, almacenamiento.ErrNoEncontrado)
}
