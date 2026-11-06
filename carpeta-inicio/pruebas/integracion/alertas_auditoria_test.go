package integracion

import (
	"context"
	"testing"
	"time"

	"asistencia/internal/alertas"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAlertas_AuditanTodasLasAcciones(t *testing.T) {
	pool, _ := AbrirBaseDePruebasConURL(t)
	ctx := context.Background()
	datos := sembrarDatosAlertas(t, pool)
	docenteID := crearUsuarioPrueba(t, pool, "docente-auditoria@prueba.pe", "docente")

	primeraFecha := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	segundaFecha := primeraFecha.AddDate(0, 0, 1)
	var primeraID, segundaID int64
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO marcaciones (alumno_id, fecha, hora, estado)
		 VALUES ($1, $2, $3, 'tardanza') RETURNING id`,
		datos.alumnoID, primeraFecha, primeraFecha.Add(8*time.Hour)).Scan(&primeraID))
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO marcaciones (alumno_id, fecha, hora, estado)
		 VALUES ($1, $2, $3, 'tardanza') RETURNING id`,
		datos.alumnoID, segundaFecha, segundaFecha.Add(8*time.Hour)).Scan(&segundaID))

	fuentes := nuevasFuentesFalsas(datos.seccionID, datos.alumnoID, datos.administradorID)
	fuentes.asistencia.marcaciones = []alertas.EventoMarcacion{
		{MarcacionID: primeraID, AlumnoID: datos.alumnoID, Fecha: primeraFecha,
			Hora: primeraFecha.Add(8 * time.Hour), Estado: "tardanza"},
		{MarcacionID: segundaID, AlumnoID: datos.alumnoID, Fecha: segundaFecha,
			Hora: segundaFecha.Add(8 * time.Hour), Estado: "tardanza"},
	}
	fuentes.academico.umbrales = map[string]int{}

	servicio := alertas.NuevoServicio(alertas.Dependencias{
		Asistencia:  fuentes.asistencia,
		Alumnos:     fuentes.alumnos,
		Academico:   fuentes.academico,
		Usuarios:    fuentes.usuarios,
		Repositorio: alertas.NuevoRepositorio(pool),
	})

	require.NoError(t, servicio.ProcesarTardanzas(ctx))
	ids := idsDeAlertas(t, pool)
	require.Len(t, ids, 2)

	require.NoError(t, servicio.MarcarVista(ctx, ids[0], datos.administradorID, true))
	_, err := servicio.MarcarVistas(ctx, []int64{ids[1]}, datos.administradorID, true)
	require.NoError(t, err)
	require.NoError(t, servicio.AsignarResponsable(ctx, ids[1], datos.administradorID, docenteID))
	require.NoError(t, servicio.MarcarSituacionCambiada(ctx, datos.alumnoID, primeraFecha, "Corregida"))
	require.NoError(t, servicio.ProcesarCierre(ctx, primeraFecha, "manana"))

	filas, err := pool.Query(ctx,
		`SELECT DISTINCT accion FROM auditoria WHERE entidad IN ('alerta', 'cierre_alerta')`)
	require.NoError(t, err)
	defer filas.Close()

	acciones := map[string]bool{}
	for filas.Next() {
		var accion string
		require.NoError(t, filas.Scan(&accion))
		acciones[accion] = true
	}
	require.NoError(t, filas.Err())

	for _, esperada := range []string{
		alertas.AccionCrearAlerta,
		alertas.AccionMarcarVista,
		alertas.AccionMarcarVistas,
		alertas.AccionAsignarResponsable,
		alertas.AccionMarcarSituacionCambiada,
		alertas.AccionEjecutarCierre,
	} {
		assert.True(t, acciones[esperada], "falta la auditoría %s", esperada)
	}
}

func idsDeAlertas(t *testing.T, pool *pgxpool.Pool) []int64 {
	t.Helper()
	filas, err := pool.Query(context.Background(), `SELECT id FROM alertas ORDER BY id`)
	require.NoError(t, err)
	defer filas.Close()
	var ids []int64
	for filas.Next() {
		var id int64
		require.NoError(t, filas.Scan(&id))
		ids = append(ids, id)
	}
	require.NoError(t, filas.Err())
	return ids
}
