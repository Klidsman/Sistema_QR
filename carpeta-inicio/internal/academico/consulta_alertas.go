package academico

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// TurnoDeSeccion devuelve el turno vigente de una sección en la fecha.
func (s *ServicioConfiguracion) TurnoDeSeccion(ctx context.Context, seccionID int64, fecha time.Time) (string, error) {
	return s.repositorio.TurnoVigenteSeccion(ctx, seccionID, fecha)
}

// SeccionesActivas expone las secciones activas para el módulo de alertas.
func (s *ServicioConfiguracion) SeccionesActivas(ctx context.Context) ([]Seccion, error) {
	return s.ListarSeccionesActivas(ctx)
}

// UmbralVigente devuelve el valor del umbral de un tipo en la fecha.
func (s *ServicioConfiguracion) UmbralVigente(ctx context.Context, tipo string, fecha time.Time) (int, bool, error) {
	anio, err := s.AnioActivo(ctx)
	if err != nil {
		return 0, false, err
	}
	if anio == nil {
		return 0, false, nil
	}
	umbral, err := s.repositorio.UmbralVigente(ctx, anio.ID, tipo, fecha)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return umbral.Valor, true, nil
}
