package alumnos

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Foto es una imagen validada lista para guardarse.
type Foto struct {
	Contenido []byte
	Extension string
	Tipo      string
}

// LeerFoto valida que el archivo sea JPG o PNG y que no supere el tamaño máximo en megabytes.
func LeerFoto(archivo io.Reader, maximoMB int) (Foto, error) {
	limite := int64(maximoMB) * 1024 * 1024
	contenido, err := io.ReadAll(io.LimitReader(archivo, limite+1))
	if err != nil {
		return Foto{}, fmt.Errorf("leyendo la foto: %w", err)
	}
	if int64(len(contenido)) > limite {
		return Foto{}, ErrFotoMuyGrande
	}

	tipo := http.DetectContentType(contenido)
	var extension string
	switch {
	case strings.HasPrefix(tipo, "image/jpeg"):
		extension = ".jpg"
	case strings.HasPrefix(tipo, "image/png"):
		extension = ".png"
	default:
		return Foto{}, ErrFotoInvalida
	}

	return Foto{Contenido: contenido, Extension: extension, Tipo: tipo}, nil
}

// ExtensionDesdeClave devuelve la extensión del archivo a partir de su clave.
func ExtensionDesdeClave(clave string) string {
	if strings.HasSuffix(strings.ToLower(clave), ".png") {
		return ".png"
	}
	return ".jpg"
}

// ClaveFoto arma la clave de almacenamiento de la foto de un alumno.
func ClaveFoto(alumnoID int64, variante string, extension string) string {
	return fmt.Sprintf("alumnos/%d/foto_%s%s", alumnoID, variante, extension)
}

// EsFotoValida indica si el contenido corresponde a una imagen JPG o PNG.
func EsFotoValida(contenido []byte) bool {
	tipo := http.DetectContentType(contenido)
	return strings.HasPrefix(tipo, "image/jpeg") || strings.HasPrefix(tipo, "image/png")
}
