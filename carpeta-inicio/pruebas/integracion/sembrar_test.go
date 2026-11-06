package integracion

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// SembrarSeccion crea un año escolar, un grado y una sección de prueba y devuelve la sección.
func SembrarSeccion(t *testing.T, pool *pgxpool.Pool, grado, seccion string) int64 {
	t.Helper()
	ctx := context.Background()

	var anioID int64
	err := pool.QueryRow(ctx,
		`INSERT INTO anios_escolares (nombre, fecha_inicio, fecha_fin)
		 VALUES ('2026', '2026-03-01', '2026-12-31')
		 ON CONFLICT (nombre) DO UPDATE SET nombre = EXCLUDED.nombre
		 RETURNING id`,
	).Scan(&anioID)
	require.NoError(t, err)

	var gradoID int64
	err = pool.QueryRow(ctx,
		`INSERT INTO grados (anio_escolar_id, nombre)
		 VALUES ($1, $2)
		 ON CONFLICT (anio_escolar_id, nombre) DO UPDATE SET nombre = EXCLUDED.nombre
		 RETURNING id`,
		anioID, grado,
	).Scan(&gradoID)
	require.NoError(t, err)

	var seccionID int64
	err = pool.QueryRow(ctx,
		`INSERT INTO secciones (grado_id, nombre)
		 VALUES ($1, $2)
		 ON CONFLICT (grado_id, nombre) DO UPDATE SET nombre = EXCLUDED.nombre
		 RETURNING id`,
		gradoID, seccion,
	).Scan(&seccionID)
	require.NoError(t, err)

	_, err = pool.Exec(ctx,
		`INSERT INTO turnos_seccion (seccion_id, turno, vigente_desde)
		 VALUES ($1, 'manana', '2026-03-01')
		 ON CONFLICT (seccion_id, vigente_desde) DO NOTHING`,
		seccionID,
	)
	require.NoError(t, err)

	return seccionID
}

// CrearDocentePrueba inserta una cuenta con rol docente y devuelve su identificador.
func CrearDocentePrueba(t *testing.T, pool *pgxpool.Pool, correo, nombre string) int64 {
	t.Helper()

	var id int64
	err := pool.QueryRow(context.Background(),
		`INSERT INTO usuarios (correo, contrasenia_hash, nombre_completo, rol)
		 VALUES ($1, 'hash-de-prueba', $2, 'docente')
		 RETURNING id`,
		correo, nombre,
	).Scan(&id)
	require.NoError(t, err)
	return id
}

// CrearUsuarioPrueba inserta un administrador y devuelve su identificador.
func CrearUsuarioPrueba(t *testing.T, pool *pgxpool.Pool) int64 {
	t.Helper()

	var id int64
	err := pool.QueryRow(context.Background(),
		`INSERT INTO usuarios (correo, contrasenia_hash, nombre_completo, rol)
		 VALUES ('pruebas@colegio.pe', 'hash-de-prueba', 'Usuario de pruebas', 'administrador')
		 RETURNING id`,
	).Scan(&id)
	require.NoError(t, err)
	return id
}
