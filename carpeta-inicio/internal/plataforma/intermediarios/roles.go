package intermediarios

import (
	"context"
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
)

type contextoUsuario struct {
	id  int64
	rol string
}

type claveContextoUsuario struct{}

// IniciarSesionUsuario renueva el token de sesión y guarda la identidad del usuario.
func IniciarSesionUsuario(ctx context.Context, sesiones *scs.SessionManager, usuarioID int64, nombre, rol string) error {
	if err := sesiones.RenewToken(ctx); err != nil {
		return err
	}
	sesiones.Put(ctx, claveUsuarioID, usuarioID)
	sesiones.Put(ctx, claveNombre, nombre)
	sesiones.Put(ctx, claveRol, rol)
	return nil
}

// NombreUsuario devuelve el nombre completo guardado en la sesión.
func NombreUsuario(ctx context.Context, sesiones *scs.SessionManager) string {
	return sesiones.GetString(ctx, claveNombre)
}

// CerrarSesionUsuario destruye la sesión actual.
func CerrarSesionUsuario(ctx context.Context, sesiones *scs.SessionManager) error {
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
	permitidos := make(map[string]bool, len(roles))
	for _, rol := range roles {
		permitidos[rol] = true
	}

	return func(siguiente http.Handler) http.Handler {
		return http.HandlerFunc(func(respuesta http.ResponseWriter, peticion *http.Request) {
			contexto := peticion.Context()
			usuarioID := sesiones.GetInt64(contexto, claveUsuarioID)
			if usuarioID == 0 {
				http.Redirect(respuesta, peticion, "/ingreso", http.StatusSeeOther)
				return
			}

			rol := sesiones.GetString(contexto, claveRol)
			if !permitidos[rol] {
				http.Error(respuesta, "No tiene permiso para esta acción", http.StatusForbidden)
				return
			}
			if rol == RolDirector && !metodoSeguro(peticion.Method) {
				http.Error(respuesta, "No tiene permiso para esta acción", http.StatusForbidden)
				return
			}
			contexto = ConUsuario(contexto, usuarioID, rol)
			siguiente.ServeHTTP(respuesta, peticion.WithContext(contexto))
		})
	}
}
