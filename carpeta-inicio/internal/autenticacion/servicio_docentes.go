package autenticacion

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"asistencia/internal/autenticacion/generado"
	"asistencia/internal/plataforma/auditoria"

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

// CrearDocente registra una cuenta con rol docente y la audita.
func (s *Servicio) CrearDocente(ctx context.Context, usuarioID int64, datos DatosDocente) (int64, error) {
	datos.Correo = normalizarCorreo(datos.Correo)
	datos.Nombre = strings.TrimSpace(datos.Nombre)
	if datos.Correo == "" || !strings.Contains(datos.Correo, "@") || datos.Nombre == "" {
		return 0, ErrDatosDocenteInvalidos
	}
	if len(datos.Contrasenia) < 8 {
		return 0, ErrDatosDocenteInvalidos
	}
	if _, err := s.repositorio.BuscarPorCorreo(ctx, datos.Correo); err == nil {
		return 0, ErrDocenteExistente
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("comprobando el correo del docente: %w", err)
	}

	hash, err := HashearContrasenia(datos.Contrasenia)
	if err != nil {
		return 0, err
	}
	id, err := s.repositorio.Crear(ctx, generado.CrearUsuarioParams{
		Correo:          datos.Correo,
		ContraseniaHash: hash,
		NombreCompleto:  datos.Nombre,
		Rol:             "docente",
	})
	if err != nil {
		if esCorreoDuplicado(err) {
			return 0, ErrDocenteExistente
		}
		return 0, fmt.Errorf("creando el docente: %w", err)
	}
	if err := auditoria.Registrar(ctx, s.db, auditoria.Evento{
		UsuarioID:  usuarioID,
		Entidad:    "docente",
		EntidadID:  id,
		Accion:     "crear_docente",
		ValorNuevo: map[string]any{"correo": datos.Correo, "nombre": datos.Nombre},
	}); err != nil {
		return 0, err
	}
	return id, nil
}

// ListarDocentes devuelve las cuentas con rol docente.
func (s *Servicio) ListarDocentes(ctx context.Context) ([]Docente, error) {
	filas, err := s.repositorio.ListarDocentes(ctx)
	if err != nil {
		return nil, fmt.Errorf("listando los docentes: %w", err)
	}
	docentes := make([]Docente, 0, len(filas))
	for _, fila := range filas {
		docentes = append(docentes, Docente{
			ID:     fila.ID,
			Correo: fila.Correo,
			Nombre: fila.NombreCompleto,
			Activo: fila.Activo,
			Rol:    fila.Rol,
		})
	}
	return docentes, nil
}

// DesactivarDocente desactiva la cuenta sin borrarla.
func (s *Servicio) DesactivarDocente(ctx context.Context, usuarioID, docenteID int64) error {
	filas, err := s.repositorio.DesactivarDocente(ctx, docenteID)
	if err != nil {
		return fmt.Errorf("desactivando el docente: %w", err)
	}
	if filas == 0 {
		return ErrDocenteNoEncontrado
	}
	return auditoria.Registrar(ctx, s.db, auditoria.Evento{
		UsuarioID:     usuarioID,
		Entidad:       "docente",
		EntidadID:     docenteID,
		Accion:        "desactivar_docente",
		ValorAnterior: map[string]any{"activo": true},
		ValorNuevo:    map[string]any{"activo": false},
	})
}

// RestablecerClave cambia la contraseña de un docente.
func (s *Servicio) RestablecerClave(ctx context.Context, usuarioID, docenteID int64, clave string) error {
	if len(clave) < 8 {
		return ErrDatosDocenteInvalidos
	}
	hash, err := HashearContrasenia(clave)
	if err != nil {
		return err
	}
	filas, err := s.repositorio.ActualizarClaveDocente(ctx, generado.ActualizarClaveDocenteParams{
		ID:              docenteID,
		ContraseniaHash: hash,
	})
	if err != nil {
		return fmt.Errorf("actualizando la clave del docente: %w", err)
	}
	if filas == 0 {
		return ErrDocenteNoEncontrado
	}
	return auditoria.Registrar(ctx, s.db, auditoria.Evento{
		UsuarioID:  usuarioID,
		Entidad:    "docente",
		EntidadID:  docenteID,
		Accion:     "restablecer_clave",
		ValorNuevo: map[string]any{"clave_restablecida": true},
	})
}

func esCorreoDuplicado(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
