package integracion

import (
	"context"
	"testing"

	"asistencia/migraciones"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigraciones_SubenYBajanTodasLasVersiones(t *testing.T) {
	pool, url := AbrirBaseDePruebasConURL(t)
	ctx := context.Background()

	assert.True(t, existeTabla(t, pool, "alumnos"), "las migraciones deben crear la tabla alumnos")

	for i := 0; i < 22; i++ {
		require.NoError(t, migraciones.Bajar(url), "revirtiendo la migración %d", 22-i)
	}
	assert.False(t, existeTabla(t, pool, "alumnos"), "al revertir todo no debe quedar la tabla alumnos")
	assert.False(t, existeTabla(t, pool, "anios_escolares"), "al revertir todo no debe quedar la tabla de años")

	require.NoError(t, migraciones.Subir(url))

	assert.True(t, existeTabla(t, pool, "alumnos"))
	assert.True(t, existeTabla(t, pool, "horario_seccion"))
	assert.True(t, existeTabla(t, pool, "umbrales_alerta"))
	assert.True(t, existeTabla(t, pool, "parametros_asistencia"))
	assert.True(t, existeTabla(t, pool, "registros_docente"))

	var version int64
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT max(version_id) FROM goose_db_version`).Scan(&version))
	assert.Equal(t, int64(22), version)
}

func existeTabla(t *testing.T, pool *pgxpool.Pool, nombre string) bool {
	t.Helper()
	var existe bool
	err := pool.QueryRow(context.Background(),
		`SELECT to_regclass($1) IS NOT NULL`, "public."+nombre).Scan(&existe)
	require.NoError(t, err)
	return existe
}
