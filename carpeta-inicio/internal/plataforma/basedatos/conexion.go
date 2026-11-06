// Package basedatos abre el pool de conexiones a PostgreSQL.
package basedatos

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Abrir crea un pool de conexiones y verifica que la base de datos responde.
func Abrir(ctx context.Context, urlBaseDatos string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, urlBaseDatos)
	if err != nil {
		return nil, fmt.Errorf("creando el pool de conexiones: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("conectando con la base de datos: %w", err)
	}
	return pool, nil
}
