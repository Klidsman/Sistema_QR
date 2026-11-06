// Package almacenamiento guarda las fotos y los carnés detrás de una interfaz.
package almacenamiento

import (
	"context"
	"errors"
	"io"
)

// ErrNoEncontrado indica que el objeto solicitado no existe.
var ErrNoEncontrado = errors.New("el objeto no existe en el almacenamiento")

// AlmacenFotos abstrae el disco local y el almacenamiento de objetos compatible con S3.
type AlmacenFotos interface {
	Guardar(ctx context.Context, clave string, contenido io.Reader) error
	Obtener(ctx context.Context, clave string) (io.ReadCloser, error)
	Eliminar(ctx context.Context, clave string) error
}
