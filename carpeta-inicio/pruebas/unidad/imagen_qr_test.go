package unidad

import (
	"bytes"
	"image"
	_ "image/png"
	"testing"

	"asistencia/internal/alumnos"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// decodificarQR devuelve el texto exacto contenido en el PNG del QR.
func decodificarQR(t *testing.T, imagenPNG []byte) string {
	t.Helper()

	imagen, _, err := image.Decode(bytes.NewReader(imagenPNG))
	require.NoError(t, err)
	bitmap, err := gozxing.NewBinaryBitmapFromImage(imagen)
	require.NoError(t, err)
	resultado, err := qrcode.NewQRCodeReader().Decode(bitmap, nil)
	require.NoError(t, err)
	return resultado.GetText()
}

func TestImagenPNG_ContieneSoloElToken(t *testing.T) {
	token, err := alumnos.GenerarTokenQR()
	require.NoError(t, err)

	imagen, err := alumnos.ImagenPNG(token)
	require.NoError(t, err)

	contenido := decodificarQR(t, imagen)
	assert.Equal(t, token, contenido)
	assert.NotContains(t, contenido, "00123456")
	assert.NotContains(t, contenido, "José")
	assert.NotContains(t, contenido, "Núñez")
	assert.NotContains(t, contenido, "3.°")
}
