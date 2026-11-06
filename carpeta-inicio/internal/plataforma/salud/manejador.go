// Package salud expone el endpoint de verificación de estado.
package salud

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Verificador comprueba que una dependencia responde.
type Verificador interface {
	Ping(ctx context.Context) error
}

// Manejador responde el estado de la aplicación y de la base de datos.
func Manejador(verificador Verificador) http.HandlerFunc {
	return func(respuesta http.ResponseWriter, peticion *http.Request) {
		ctx, cancelar := context.WithTimeout(peticion.Context(), 2*time.Second)
		defer cancelar()

		respuesta.Header().Set("Content-Type", "application/json")
		if err := verificador.Ping(ctx); err != nil {
			respuesta.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(respuesta).Encode(map[string]string{
				"estado":     "error",
				"base_datos": "no disponible",
			})
			return
		}
		_ = json.NewEncoder(respuesta).Encode(map[string]string{
			"estado":     "correcto",
			"base_datos": "correcta",
		})
	}
}
