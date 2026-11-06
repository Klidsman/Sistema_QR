package asistencia

import (
	"context"
	"fmt"

	"asistencia/internal/asistencia/generado"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repositorio accede a las tablas del módulo de asistencia.
type Repositorio struct {
	pool      *pgxpool.Pool
	consultas *generado.Queries
}

// NuevoRepositorio crea el repositorio sobre el pool indicado.
func NuevoRepositorio(pool *pgxpool.Pool) *Repositorio {
	return &Repositorio{pool: pool, consultas: generado.New(pool)}
}

// EnTransaccion ejecuta la función dentro de una transacción y confirma si no hay error.
func (r *Repositorio) EnTransaccion(ctx context.Context, funcion func(*generado.Queries, generado.DBTX) error) error {
	transaccion, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("iniciando la transacción: %w", err)
	}
	defer func() {
		_ = transaccion.Rollback(ctx)
	}()

	if err := funcion(generado.New(transaccion), transaccion); err != nil {
		return err
	}
	if err := transaccion.Commit(ctx); err != nil {
		return fmt.Errorf("confirmando la transacción: %w", err)
	}
	return nil
}
