package alumnos

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// longitudTokenQR son los bytes aleatorios del token. En base64 URL sin relleno ocupan 43 caracteres.
const longitudTokenQR = 32

// GenerarTokenQR crea un token opaco aleatorio. No contiene ningún dato del alumno.
func GenerarTokenQR() (string, error) {
	bytes := make([]byte, longitudTokenQR)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generando el token del QR: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
