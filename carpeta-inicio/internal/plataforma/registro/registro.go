// Package registro crea el registrador estructurado de la aplicación.
package registro

import (
	"io"
	"log/slog"
)

// Nuevo devuelve un registrador que escribe eventos en formato JSON.
func Nuevo(salida io.Writer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(salida, &slog.HandlerOptions{Level: slog.LevelInfo}))
}
