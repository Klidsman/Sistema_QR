package alumnos

import (
	"fmt"

	qrcode "github.com/skip2/go-qrcode"
)

// tamanoImagenQR es el lado del PNG en píxeles.
const tamanoImagenQR = 512

// ImagenPNG genera el PNG del QR. El contenido es únicamente el token; no incluye datos personales.
func ImagenPNG(token string) ([]byte, error) {
	imagen, err := qrcode.Encode(token, qrcode.Medium, tamanoImagenQR)
	if err != nil {
		return nil, fmt.Errorf("generando la imagen del QR: %w", err)
	}
	return imagen, nil
}
