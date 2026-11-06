package almacenamiento

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// AlmacenDisco guarda los objetos en una carpeta local. Se usa en desarrollo.
type AlmacenDisco struct {
	raiz string
}

// NuevoAlmacenDisco crea un almacén que escribe dentro de la carpeta raíz indicada.
func NuevoAlmacenDisco(raiz string) *AlmacenDisco {
	return &AlmacenDisco{raiz: raiz}
}

// Guardar escribe el contenido en la clave indicada, creando las carpetas necesarias.
func (a *AlmacenDisco) Guardar(ctx context.Context, clave string, contenido io.Reader) error {
	ruta, err := a.ruta(clave)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(ruta), 0o750); err != nil {
		return fmt.Errorf("creando la carpeta del objeto: %w", err)
	}
	archivo, err := os.Create(ruta)
	if err != nil {
		return fmt.Errorf("creando el objeto: %w", err)
	}
	defer archivo.Close()

	if _, err := io.Copy(archivo, contenido); err != nil {
		return fmt.Errorf("escribiendo el objeto: %w", err)
	}
	return nil
}

// Obtener abre el objeto para lectura.
func (a *AlmacenDisco) Obtener(ctx context.Context, clave string) (io.ReadCloser, error) {
	ruta, err := a.ruta(clave)
	if err != nil {
		return nil, err
	}
	archivo, err := os.Open(ruta)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, ErrNoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("abriendo el objeto: %w", err)
	}
	return archivo, nil
}

// Eliminar borra el objeto indicado.
func (a *AlmacenDisco) Eliminar(ctx context.Context, clave string) error {
	ruta, err := a.ruta(clave)
	if err != nil {
		return err
	}
	err = os.Remove(ruta)
	if errors.Is(err, fs.ErrNotExist) {
		return ErrNoEncontrado
	}
	if err != nil {
		return fmt.Errorf("eliminando el objeto: %w", err)
	}
	return nil
}

func (a *AlmacenDisco) ruta(clave string) (string, error) {
	limpia := filepath.Clean(filepath.FromSlash(clave))
	separador := string(filepath.Separator)
	esAbsoluta := filepath.IsAbs(limpia) || filepath.VolumeName(limpia) != "" || strings.HasPrefix(limpia, separador)
	saleDeLaRaiz := limpia == ".." || strings.HasPrefix(limpia, ".."+separador)
	if limpia == "." || esAbsoluta || saleDeLaRaiz {
		return "", fmt.Errorf("la clave %q no es válida", clave)
	}
	return filepath.Join(a.raiz, limpia), nil
}
