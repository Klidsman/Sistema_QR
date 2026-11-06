package intermediarios

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/alexedwards/scs/v2"
)

const (
	claveCSRF         = "token_csrf"
	campoCSRF         = "token_csrf"
	encabezadoCSRF    = "X-CSRF-Token"
	longitudTokenCSRF = 32
)

// TokenCSRF devuelve el token CSRF de la sesión y lo crea si todavía no existe.
func TokenCSRF(ctx context.Context, sesiones *scs.SessionManager) (string, error) {
	token := sesiones.GetString(ctx, claveCSRF)
	if token != "" {
		return token, nil
	}
	nuevo, err := generarTokenCSRF()
	if err != nil {
		return "", err
	}
	sesiones.Put(ctx, claveCSRF, nuevo)
	return nuevo, nil
}

// ProteccionCSRF exige el token CSRF en los métodos que cambian estado.
// Debe ejecutarse dentro de LoadAndSave del administrador de sesiones.
func ProteccionCSRF(sesiones *scs.SessionManager) func(http.Handler) http.Handler {
	return func(siguiente http.Handler) http.Handler {
		return http.HandlerFunc(func(respuesta http.ResponseWriter, peticion *http.Request) {
			token, err := TokenCSRF(peticion.Context(), sesiones)
			if err != nil {
				http.Error(respuesta, "No se pudo preparar la sesión", http.StatusInternalServerError)
				return
			}
			if metodoSeguro(peticion.Method) {
				siguiente.ServeHTTP(respuesta, peticion)
				return
			}

			enviado := peticion.Header.Get(encabezadoCSRF)
			if enviado == "" {
				if strings.HasPrefix(peticion.Header.Get("Content-Type"), "multipart/form-data") {
					if err := peticion.ParseMultipartForm(32 << 20); err != nil {
						http.Error(respuesta, "Solicitud no válida. Recargue la página e intente de nuevo.", http.StatusForbidden)
						return
					}
				} else {
					_ = peticion.ParseForm()
				}
				enviado = peticion.PostFormValue(campoCSRF)
			}
			if subtle.ConstantTimeCompare([]byte(enviado), []byte(token)) != 1 {
				http.Error(respuesta, "Solicitud no válida. Recargue la página e intente de nuevo.", http.StatusForbidden)
				return
			}
			siguiente.ServeHTTP(respuesta, peticion)
		})
	}
}

func metodoSeguro(metodo string) bool {
	return metodo == http.MethodGet || metodo == http.MethodHead || metodo == http.MethodOptions
}

func generarTokenCSRF() (string, error) {
	bytes := make([]byte, longitudTokenCSRF)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
