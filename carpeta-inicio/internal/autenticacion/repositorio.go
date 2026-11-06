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

// ListarDocentes devuelve las cuentas con rol docente.
func (r *Repositorio) ListarDocentes(ctx context.Context) ([]generado.ListarDocentesRow, error) {
	return r.consultas.ListarDocentes(ctx)
}

// DesactivarDocente desactiva una cuenta de docente y devuelve cuántas filas cambió.
func (r *Repositorio) DesactivarDocente(ctx context.Context, id int64) (int64, error) {
	return r.consultas.DesactivarDocente(ctx, id)
}

// ActualizarClaveDocente cambia el hash de contraseña de un docente.
func (r *Repositorio) ActualizarClaveDocente(ctx context.Context, parametros generado.ActualizarClaveDocenteParams) (int64, error) {
	return r.consultas.ActualizarClaveDocente(ctx, parametros)
}

// Crear inserta un usuario y devuelve su identificador.
func (r *Repositorio) Crear(ctx context.Context, parametros generado.CrearUsuarioParams) (int64, error) {
	return r.consultas.CrearUsuario(ctx, parametros)
}
