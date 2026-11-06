// Package intermediarios agrupa la sesión, la protección CSRF, los roles y el límite de intentos.
package intermediarios

import (
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
)

// NuevaSesion crea el administrador de sesiones con cookies seguras.
func NuevaSesion(almacen scs.Store, duracion time.Duration) *scs.SessionManager {
	sesiones := scs.New()
	sesiones.Store = almacen
	sesiones.Lifetime = duracion
	sesiones.Cookie.Name = "sesion_asistencia"
	sesiones.Cookie.HttpOnly = true
	sesiones.Cookie.Secure = true
	sesiones.Cookie.SameSite = http.SameSiteStrictMode
	return sesiones
}
