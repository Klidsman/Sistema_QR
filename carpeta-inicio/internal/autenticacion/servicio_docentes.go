package autenticacion

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"asistencia/internal/autenticacion/generado"
	"asistencia/internal/plataforma/auditoria"
	"asistencia/internal/plataforma/intermediarios"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// DocenteActivo devuelve si el usuario existe, está activo y su rol.
func (s *Servicio) DocenteActivo(ctx context.Context, docenteID int64) (bool, string, error) {
	fila, err := s.repositorio.BuscarPorID(ctx, docenteID)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, "", nil
	}
	if err != nil {
		return false, "", fmt.Errorf("buscando el docente: %w", err)
	}
	return fila.Activo, fila.Rol, nil
}

// AdministradorActivo devuelve el identificador del administrador de respaldo.
func (s *Servicio) AdministradorActivo(ctx context.Context) (int64, error) {
	return s.repositorio.BuscarAdministradorActivo(ctx)
}

// NombreDocente devuelve el nombre completo de un docente.
func (s *Servicio) NombreDocente(ctx context.Context, docenteID int64) (string, error) {
	fila, err := s.repositorio.BuscarPorID(ctx, docenteID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("buscando el docente: %w", err)
	}
	return fila.NombreCompleto, nil
}

// CrearCuenta registra una cuenta con el rol indicado y la audita.
func (s *Servicio) CrearCuenta(ctx context.Context, usuarioID int64, rol string, datos DatosCuenta) (int64, error) {
	if rol != intermediarios.RolDocente && rol != intermediarios.RolPorteria {
		return 0, ErrRolNoPermitido
	}
	datos.Correo = normalizarCorreo(datos.Correo)
	datos.Nombre = strings.TrimSpace(datos.Nombre)
	if datos.Correo == "" || !strings.Contains(datos.Correo, "@") || datos.Nombre == "" {
		return 0, ErrDatosCuentaInvalidos
	}
	if len(datos.Contrasenia) < 8 {
		return 0, ErrDatosCuentaInvalidos
	}
	if _, err := s.repositorio.BuscarPorCorreo(ctx, datos.Correo); err == nil {
		return 0, ErrCuentaExistente
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("comprobando el correo de la cuenta: %w", err)
	}

	hash, err := HashearContrasenia(datos.Contrasenia)
	if err != nil {
		return 0, err
	}
	id, err := s.repositorio.Crear(ctx, generado.CrearUsuarioParams{
		Correo:          datos.Correo,
		ContraseniaHash: hash,
		NombreCompleto:  datos.Nombre,
		Rol:             rol,
	})
	if err != nil {
		if esCorreoDuplicado(err) {
			return 0, ErrCuentaExistente
		}
		return 0, fmt.Errorf("creando la cuenta: %w", err)
	}
	if err := auditoria.Registrar(ctx, s.db, auditoria.Evento{
		UsuarioID:  usuarioID,
		Entidad:    rol,
		EntidadID:  id,
		Accion:     "crear_" + rol,
		ValorNuevo: map[string]any{"correo": datos.Correo, "nombre": datos.Nombre, "rol": rol},
	}); err != nil {
		return 0, err
	}
	return id, nil
}

// CrearDocente registra una cuenta con rol docente y la audita.
func (s *Servicio) CrearDocente(ctx context.Context, usuarioID int64, datos DatosDocente) (int64, error) {
	return s.CrearCuenta(ctx, usuarioID, intermediarios.RolDocente, datos)
}

// ListarCuentasPorRol devuelve las cuentas con el rol indicado.
func (s *Servicio) ListarCuentasPorRol(ctx context.Context, rol string) ([]Cuenta, error) {
	if rol != intermediarios.RolDocente && rol != intermediarios.RolPorteria {
		return nil, ErrRolNoPermitido
	}
	filas, err := s.repositorio.ListarCuentasPorRol(ctx, rol)
	if err != nil {
		return nil, fmt.Errorf("listando las cuentas de %s: %w", rol, err)
	}
	cuentas := make([]Cuenta, 0, len(filas))
	for _, fila := range filas {
		cuentas = append(cuentas, Cuenta{
			ID:     fila.ID,
			Correo: fila.Correo,
			Nombre: fila.NombreCompleto,
			Activo: fila.Activo,
			Rol:    fila.Rol,
		})
	}
	return cuentas, nil
}

// ListarDocentes devuelve las cuentas con rol docente.
func (s *Servicio) ListarDocentes(ctx context.Context) ([]Docente, error) {
	return s.ListarCuentasPorRol(ctx, intermediarios.RolDocente)
}

// DesactivarCuenta desactiva la cuenta sin borrarla.
func (s *Servicio) DesactivarCuenta(ctx context.Context, usuarioID, cuentaID int64, rol string) error {
	if rol != intermediarios.RolDocente && rol != intermediarios.RolPorteria {
		return ErrRolNoPermitido
	}
	filas, err := s.repositorio.DesactivarCuenta(ctx, cuentaID, rol)
	if err != nil {
		return fmt.Errorf("desactivando la cuenta de %s: %w", rol, err)
	}
	if filas == 0 {
		return ErrCuentaNoEncontrada
	}
	return auditoria.Registrar(ctx, s.db, auditoria.Evento{
		UsuarioID:     usuarioID,
		Entidad:       rol,
		EntidadID:     cuentaID,
		Accion:        "desactivar_" + rol,
		ValorAnterior: map[string]any{"activo": true},
		ValorNuevo:    map[string]any{"activo": false},
	})
}

// DesactivarDocente desactiva una cuenta de docente sin borrarla.
func (s *Servicio) DesactivarDocente(ctx context.Context, usuarioID, docenteID int64) error {
	return s.DesactivarCuenta(ctx, usuarioID, docenteID, intermediarios.RolDocente)
}

// RestablecerClaveCuenta cambia la contraseña de una cuenta del rol indicado.
func (s *Servicio) RestablecerClaveCuenta(ctx context.Context, usuarioID, cuentaID int64, rol, clave string) error {
	if rol != intermediarios.RolDocente && rol != intermediarios.RolPorteria {
		return ErrRolNoPermitido
	}
	if len(clave) < 8 {
		return ErrDatosCuentaInvalidos
	}
	hash, err := HashearContrasenia(clave)
	if err != nil {
		return err
	}
	filas, err := s.repositorio.ActualizarClaveCuenta(ctx, generado.ActualizarClaveCuentaParams{
		ID:              cuentaID,
		ContraseniaHash: hash,
		Rol:             rol,
	})
	if err != nil {
		return fmt.Errorf("actualizando la clave de la cuenta: %w", err)
	}
	if filas == 0 {
		return ErrCuentaNoEncontrada
	}
	return auditoria.Registrar(ctx, s.db, auditoria.Evento{
		UsuarioID:  usuarioID,
		Entidad:    rol,
		EntidadID:  cuentaID,
		Accion:     "restablecer_clave",
		ValorNuevo: map[string]any{"clave_restablecida": true},
	})
}

// RestablecerClave cambia la contraseña de un docente.
func (s *Servicio) RestablecerClave(ctx context.Context, usuarioID, docenteID int64, clave string) error {
	return s.RestablecerClaveCuenta(ctx, usuarioID, docenteID, intermediarios.RolDocente, clave)
}

func esCorreoDuplicado(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
