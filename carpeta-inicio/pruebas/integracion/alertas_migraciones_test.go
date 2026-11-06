package integracion

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigracionesAlertas_CreanEstructuraYRestricciones(t *testing.T) {
	pool, _ := AbrirBaseDePruebasConURL(t)
	ctx := context.Background()

	var columnas int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM information_schema.columns
		 WHERE table_name = 'marcaciones' AND column_name = 'registrada_por'`).Scan(&columnas))
	assert.Equal(t, 1, columnas, "marcaciones debe tener registrada_por")

	var indices int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM pg_indexes
		 WHERE tablename = 'alertas'
		   AND indexname IN (
		       'alertas_tardanza_unica',
		       'alertas_reiteracion_tardanzas_unica',
		       'alertas_reiteracion_faltas_unica',
		       'alertas_panel_idx')`).Scan(&indices))
	assert.Equal(t, 4, indices, "faltan índices de deduplicación o del panel")

	var definicionUmbral string
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT pg_get_constraintdef(oid) FROM pg_constraint
		 WHERE conname = 'umbrales_alerta_tipo_check'`).Scan(&definicionUmbral))
	assert.Contains(t, definicionUmbral, "faltas_seguidas")

	var definicionAuditoria string
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT pg_get_constraintdef(oid) FROM pg_constraint
		 WHERE conname = 'auditoria_entidad_check'`).Scan(&definicionAuditoria))
	assert.True(t, strings.Contains(definicionAuditoria, "alerta"))
	assert.True(t, strings.Contains(definicionAuditoria, "cierre_alerta"))

	var restricciones int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM pg_constraint
		 WHERE conrelid = 'cierres_alerta'::regclass
		   AND contype = 'u'
		   AND pg_get_constraintdef(oid) LIKE '%fecha%turno%'`).Scan(&restricciones))
	assert.Equal(t, 1, restricciones, "cierres_alerta debe ser único por fecha y turno")
}
