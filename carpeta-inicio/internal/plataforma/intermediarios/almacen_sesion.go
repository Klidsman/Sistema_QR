package intermediarios

import (
	"context"
	"errors"
	"time"

	"asistencia/internal/plataforma/intermediarios/generado"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AlmacenSesion guarda las sesiones del navegador en la tabla sesiones de PostgreSQL.
type AlmacenSesion struct {
	consultas *generado.Queries
}

// NuevoAlmacenSesion crea el almacén de sesiones sobre el pool indicado.
func NuevoAlmacenSesion(pool *pgxpool.Pool) *AlmacenSesion {
	return &AlmacenSesion{consultas: generado.New(pool)}
}

// Find devuelve los datos de la sesión si existe y no ha expirado.
func (a *AlmacenSesion) Find(token string) ([]byte, bool, error) {
	fila, err := a.consultas.BuscarSesion(context.Background(), token)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	if !fila.ExpiraEn.Valid || !fila.ExpiraEn.Time.After(time.Now()) {
		return nil, false, nil
	}
	return fila.Datos, true, nil
}

// Commit guarda los datos de la sesión con su fecha de expiración.
func (a *AlmacenSesion) Commit(token string, datos []byte, expira time.Time) error {
	return a.consultas.GuardarSesion(context.Background(), generado.GuardarSesionParams{
		Token:    token,
		Datos:    datos,
		ExpiraEn: pgtype.Timestamptz{Time: expira, Valid: true},
	})
}

// Delete elimina la sesión indicada.
func (a *AlmacenSesion) Delete(token string) error {
	return a.consultas.EliminarSesion(context.Background(), token)
}
