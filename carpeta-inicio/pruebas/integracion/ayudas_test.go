package integracion

import (
	"context"
	"testing"
	"time"

	"asistencia/migraciones"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	postgresmod "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// AbrirBaseDePruebas levanta un PostgreSQL real en un contenedor, aplica las migraciones y
// devuelve un pool conectado. El contenedor se elimina al terminar la prueba.
func AbrirBaseDePruebas(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx, cancelar := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancelar()

	contenedor, err := postgresmod.Run(ctx,
		"postgres:16",
		postgresmod.WithDatabase("asistencia_pruebas"),
		postgresmod.WithUsername("pruebas"),
		postgresmod.WithPassword("pruebas"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(time.Minute),
		),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = testcontainers.TerminateContainer(contenedor)
	})

	url, err := contenedor.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)
	require.NoError(t, migraciones.Subir(url))

	pool, err := pgxpool.New(ctx, url)
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	return pool
}
