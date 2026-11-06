package autenticacion

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
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
