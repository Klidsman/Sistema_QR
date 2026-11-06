// Package academico consulta los años escolares, grados y secciones del colegio.
package academico

import (
	"context"
	"errors"
	"fmt"

	"asistencia/internal/academico/generado"

	"github.com/jackc/pgx/v5"
)

// ErrSeccionNoEncontrada indica que la sección no existe en el año escolar indicado.
var ErrSeccionNoEncontrada = errors.New("la sección no existe en el año escolar")

// BuscadorSeccion es la interfaz pública que otros módulos usan para resolver una sección.
// Así los módulos no consultan las tablas académicas directamente.
type BuscadorSeccion interface {
	BuscarSeccion(ctx context.Context, anioID int64, grado, seccion string) (int64, error)
	ListarSecciones(ctx context.Context) ([]Seccion, error)
	BuscarSeccionPorID(ctx context.Context, id int64) (Seccion, error)
}

// Servicio implementa BuscadorSeccion sobre la base de datos.
type Servicio struct {
	repositorio *Repositorio
}

// NuevoServicio crea el servicio académico sobre la conexión indicada.
func NuevoServicio(db generado.DBTX) *Servicio {
	return &Servicio{repositorio: NuevoRepositorio(db)}
}

// BuscarSeccion devuelve el identificador de la sección o ErrSeccionNoEncontrada.
func (s *Servicio) BuscarSeccion(ctx context.Context, anioID int64, grado, seccion string) (int64, error) {
	id, err := s.repositorio.BuscarSeccion(ctx, anioID, grado, seccion)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrSeccionNoEncontrada
	}
	if err != nil {
		return 0, fmt.Errorf("buscando la sección: %w", err)
	}
	return id, nil
}

// ListarSecciones devuelve todas las secciones para los formularios y filtros.
func (s *Servicio) ListarSecciones(ctx context.Context) ([]Seccion, error) {
	secciones, err := s.repositorio.ListarSecciones(ctx)
	if err != nil {
		return nil, fmt.Errorf("listando las secciones: %w", err)
	}
	return secciones, nil
}

// BuscarSeccionPorID devuelve la sección o ErrSeccionNoEncontrada.
func (s *Servicio) BuscarSeccionPorID(ctx context.Context, id int64) (Seccion, error) {
	seccion, err := s.repositorio.BuscarSeccionPorID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Seccion{}, ErrSeccionNoEncontrada
	}
	if err != nil {
		return Seccion{}, fmt.Errorf("buscando la sección: %w", err)
	}
	return seccion, nil
}
