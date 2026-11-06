package integracion

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestRendimientoAlertas revisa el plan del panel y de los conteos con 2,000
// alumnos. Es opcional para no alargar la suite normal:
//
//	$env:PRUEBA_RENDIMIENTO = "1"; go test ./pruebas/integracion -run TestRendimientoAlertas -v
func TestRendimientoAlertas(t *testing.T) {
	if os.Getenv("PRUEBA_RENDIMIENTO") == "" {
		t.Skip("defina PRUEBA_RENDIMIENTO=1 para ejecutar la revisión de rendimiento")
	}

	pool, _ := AbrirBaseDePruebasConURL(t)
	ctx := context.Background()
	datos := sembrarDatosAlertas(t, pool)

	_, err := pool.Exec(ctx, `
		INSERT INTO alumnos (identificador, nombres, apellidos, nombres_busqueda, apellidos_busqueda, seccion_id)
		SELECT 'R' || lpad(i::text, 5, '0'), 'Alumno', 'Rendimiento ' || i,
		       'alumno', 'rendimiento ' || i, $1
		FROM generate_series(1, 2000) AS i`, datos.seccionID)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		INSERT INTO marcaciones (alumno_id, fecha, hora, estado)
		SELECT a.id, date '2026-03-10', timestamptz '2026-03-10 08:00:00-05', 'tardanza'
		FROM alumnos a
		WHERE a.identificador LIKE 'R%'`)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		INSERT INTO alertas (tipo, alumno_id, seccion_id, fecha_evento, marcacion_id, destinatario_id)
		SELECT 'tardanza', m.alumno_id, $2, m.fecha, m.id, $1
		FROM marcaciones m
		JOIN alumnos a ON a.id = m.alumno_id
		WHERE a.identificador LIKE 'R%'`, datos.administradorID, datos.seccionID)
	require.NoError(t, err)

	planes := []struct {
		nombre string
		sql    string
		args   []any
	}{
		{"panel", `
			EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
			SELECT id, tipo, alumno_id, seccion_id, fecha_evento, estado
			FROM alertas
			WHERE destinatario_id = $1
			  AND fecha_evento >= date '2026-03-01'
			  AND fecha_evento <= date '2026-03-31'
			  AND tipo = 'tardanza'
			  AND estado = 'nueva'
			ORDER BY creada_en DESC
			LIMIT 200`, []any{datos.administradorID}},
		{"conteo_tardanzas", `
			EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
			SELECT count(*)
			FROM marcaciones
			WHERE alumno_id = (SELECT min(id) FROM alumnos)
			  AND fecha BETWEEN date '2026-03-01' AND date '2026-03-31'
			  AND estado IN ('tardanza', 'tardanza_grave')`, nil},
	}

	inicio := time.Now()
	for _, plan := range planes {
		filas, err := pool.Query(ctx, plan.sql, plan.args...)
		require.NoError(t, err)
		var lineas []string
		for filas.Next() {
			var linea string
			require.NoError(t, filas.Scan(&linea))
			lineas = append(lineas, linea)
		}
		require.NoError(t, filas.Err())
		filas.Close()
		t.Logf("plan %s:\n%s", plan.nombre, joinLineas(lineas))
	}
	t.Logf("consulta de planes completada en %s con 2,000 alumnos", time.Since(inicio))
}

func joinLineas(lineas []string) string {
	resultado := ""
	for _, linea := range lineas {
		resultado += linea + "\n"
	}
	return resultado
}
