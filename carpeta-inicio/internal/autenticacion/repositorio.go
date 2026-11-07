package autenticacion

import (
	"context"

	"asistencia/internal/autenticacion/generado"
)

// Repositorio accede a la tabla usuarios mediante consultas sqlc.
type Repositorio struct {
	consultas *generado.Queries
}

// NuevoRepositorio crea el repositorio sobre la conexión o transacción indicada.
func NuevoRepositorio(db generado.DBTX) *Repositorio {
	return &Repositorio{consultas: generado.New(db)}
}

// BuscarPorCorreo devuelve el usuario con ese correo, con su hash de contraseña.
func (r *Repositorio) BuscarPorCorreo(ctx context.Context, correo string) (generado.BuscarUsuarioPorCorreoRow, error) {
	return r.consultas.BuscarUsuarioPorCorreo(ctx, correo)
}

// BuscarPorID devuelve el usuario con ese identificador.
func (r *Repositorio) BuscarPorID(ctx context.Context, id int64) (generado.BuscarUsuarioPorIDRow, error) {
	return r.consultas.BuscarUsuarioPorID(ctx, id)
}

// BuscarAdministradorActivo devuelve el primer administrador activo.
func (r *Repositorio) BuscarAdministradorActivo(ctx context.Context) (int64, error) {
	return r.consultas.BuscarAdministradorActivo(ctx)
}

// ListarCuentasPorRol devuelve las cuentas activas e inactivas de un rol.
func (r *Repositorio) ListarCuentasPorRol(ctx context.Context, rol string) ([]generado.ListarCuentasPorRolRow, error) {
	return r.consultas.ListarCuentasPorRol(ctx, rol)
}

// DesactivarCuenta desactiva una cuenta del rol indicado y devuelve cuántas filas cambió.
func (r *Repositorio) DesactivarCuenta(ctx context.Context, id int64, rol string) (int64, error) {
	return r.consultas.DesactivarCuenta(ctx, generado.DesactivarCuentaParams{ID: id, Rol: rol})
}

// ActualizarClaveCuenta cambia el hash de contraseña de una cuenta del rol indicado.
func (r *Repositorio) ActualizarClaveCuenta(ctx context.Context, parametros generado.ActualizarClaveCuentaParams) (int64, error) {
	return r.consultas.ActualizarClaveCuenta(ctx, parametros)
}

// Crear inserta un usuario y devuelve su identificador.
func (r *Repositorio) Crear(ctx context.Context, parametros generado.CrearUsuarioParams) (int64, error) {
	return r.consultas.CrearUsuario(ctx, parametros)
}
