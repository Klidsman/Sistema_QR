package intermediarios

import (
	"context"
	"fmt"
	"net/http"

	"github.com/alexedwards/scs/v2"
)

// Roles de la aplicación.
const (
	RolAdministrador = "administrador"
	RolDirector      = "director"
	RolDocente       = "docente"
	RolPorteria      = "porteria"
)

const (
	claveUsuarioID = "usuario_id"
	claveRol       = "rol"
	claveNombre    = "nombre_completo"
	claveRetorno   = "url_retorno"
)

type contextoUsuario struct {
	id  int64
	rol string
}

type claveContextoUsuario struct{}

// IniciarSesionUsuario regenera la sesión desde cero y guarda la identidad.
// Primero limpia cualquier dato de una sesión anterior y luego renueva el token.
func IniciarSesionUsuario(ctx context.Context, sesiones *scs.SessionManager, usuarioID int64, nombre, rol string) error {
	if err := sesiones.Clear(ctx); err != nil {
		return err
	}
	if err := sesiones.RenewToken(ctx); err != nil {
		return err
	}
	sesiones.Put(ctx, claveUsuarioID, usuarioID)
	sesiones.Put(ctx, claveNombre, nombre)
	sesiones.Put(ctx, claveRol, rol)
	return nil
}

// GuardarUrlRetorno recuerda una ruta interna para usarla después del ingreso.
func GuardarUrlRetorno(ctx context.Context, sesiones *scs.SessionManager, ruta string) {
	if !rutaInterna(ruta) || ruta == "/ingreso" || ruta == "/salir" {
		return
	}
	sesiones.Put(ctx, claveRetorno, ruta)
}

// UrlRetorno extrae y borra la ruta de retorno guardada.
func UrlRetorno(ctx context.Context, sesiones *scs.SessionManager) string {
	return sesiones.PopString(ctx, claveRetorno)
}

// NombreUsuario devuelve el nombre completo guardado en la sesión.
func NombreUsuario(ctx context.Context, sesiones *scs.SessionManager) string {
	return sesiones.GetString(ctx, claveNombre)
}

// CerrarSesionUsuario destruye la sesión completa, incluida la URL de retorno.
func CerrarSesionUsuario(ctx context.Context, sesiones *scs.SessionManager) error {
	sesiones.Remove(ctx, claveRetorno)
	return sesiones.Destroy(ctx)
}

// ConUsuario guarda la identidad ya validada en el contexto de la petición.
func ConUsuario(ctx context.Context, usuarioID int64, rol string) context.Context {
	return context.WithValue(ctx, claveContextoUsuario{}, contextoUsuario{id: usuarioID, rol: rol})
}

// UsuarioID devuelve el identificador del usuario autenticado, o cero si no hay sesión.
func UsuarioID(ctx context.Context) int64 {
	usuario, _ := ctx.Value(claveContextoUsuario{}).(contextoUsuario)
	return usuario.id
}

// RolUsuario devuelve el rol del usuario autenticado, o vacío si no hay sesión.
func RolUsuario(ctx context.Context) string {
	usuario, _ := ctx.Value(claveContextoUsuario{}).(contextoUsuario)
	return usuario.rol
}

// RequerirRol exige una sesión iniciada con uno de los roles indicados.
// El director solo puede usar métodos de lectura, sin importar la ruta.
func RequerirRol(sesiones *scs.SessionManager, roles ...string) func(http.Handler) http.Handler {
	return requerirRol(sesiones, false, roles...)
}

// RequerirRolConEscritura permite al director usar métodos que cambian estado.
// Solo debe aplicarse a las rutas que la especificación autoriza expresamente.
func RequerirRolConEscritura(sesiones *scs.SessionManager, roles ...string) func(http.Handler) http.Handler {
	return requerirRol(sesiones, true, roles...)
}

func requerirRol(sesiones *scs.SessionManager, permitirEscrituraDirector bool, roles ...string) func(http.Handler) http.Handler {
	permitidos := make(map[string]bool, len(roles))
	for _, rol := range roles {
		permitidos[rol] = true
	}

	return func(siguiente http.Handler) http.Handler {
		return http.HandlerFunc(func(respuesta http.ResponseWriter, peticion *http.Request) {
			contexto := peticion.Context()
			usuarioID := sesiones.GetInt64(contexto, claveUsuarioID)
			if usuarioID == 0 {
				if peticion.Method == http.MethodGet {
					GuardarUrlRetorno(contexto, sesiones, peticion.URL.RequestURI())
				}
				http.Redirect(respuesta, peticion, "/ingreso", http.StatusSeeOther)
				return
			}

			rol := sesiones.GetString(contexto, claveRol)
			if !permitidos[rol] {
				responderAccesoDenegado(respuesta, rol)
				return
			}
			if rol == RolDirector && !permitirEscrituraDirector && !metodoSeguro(peticion.Method) {
				responderAccesoDenegado(respuesta, rol)
				return
			}
			contexto = ConUsuario(contexto, usuarioID, rol)
			siguiente.ServeHTTP(respuesta, peticion.WithContext(contexto))
		})
	}
}

// responderAccesoDenegado muestra una página con el enlace al inicio del rol.
func responderAccesoDenegado(respuesta http.ResponseWriter, rol string) {
	inicio := RutaInicioPorRol(rol)
	respuesta.Header().Set("Content-Type", "text/html; charset=utf-8")
	respuesta.WriteHeader(http.StatusForbidden)
	_, _ = fmt.Fprintf(respuesta, `<!DOCTYPE html>
<html lang="es">
<head><meta charset="utf-8"/><meta name="viewport" content="width=device-width, initial-scale=1"/>
<title>Acceso no permitido · Asistencia</title>
<link rel="stylesheet" href="/estaticos/css/app.css"/></head>
<body class="min-h-screen bg-slate-50 text-slate-900 antialiased">
<main class="mx-auto flex min-h-screen w-full max-w-lg flex-col items-center justify-center gap-4 p-6 text-center">
<h1 class="text-2xl font-semibold">Acceso no permitido</h1>
<p class="text-slate-600">No tiene permiso para esta acción con su rol actual.</p>
<a href="%s" class="rounded bg-blue-700 px-4 py-2 font-semibold text-white">Volver a mi inicio</a>
</main>
</body>
</html>`, inicio)
}
