package integracion

import (
	"context"
	"sync"
	"testing"
	"time"

	"asistencia/internal/alertas"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcesarTardanzas_ReiteracionMensualUnaSolaVez(t *testing.T) {
	pool, _ := AbrirBaseDePruebasConURL(t)
	ctx := context.Background()
	datos := sembrarDatosAlertas(t, pool)
	docenteID := crearUsuarioPrueba(t, pool, "docente@prueba.pe", "docente")

	fecha := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	var primeraID int64
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO marcaciones (alumno_id, fecha, hora, estado)
		 VALUES ($1, $2, $3, 'tardanza') RETURNING id`,
		datos.alumnoID, fecha, fecha.Add(8*time.Hour)).Scan(&primeraID))

	fuentes := nuevasFuentesFalsas(datos.seccionID, datos.alumnoID, datos.administradorID)
	fuentes.asistencia.marcaciones = []alertas.EventoMarcacion{{
		MarcacionID: primeraID,
		AlumnoID:    datos.alumnoID,
		Fecha:       fecha,
		Hora:        fecha.Add(8 * time.Hour),
		Estado:      "tardanza",
	}}
	fuentes.asistencia.contarTardanzas = 3
	fuentes.asistencia.docentes = []int64{docenteID}
	fuentes.academico.umbrales = map[string]int{"tardanzas": 3}

	servicio := alertas.NuevoServicio(alertas.Dependencias{
		Asistencia:  fuentes.asistencia,
		Alumnos:     fuentes.alumnos,
		Academico:   fuentes.academico,
		Usuarios:    fuentes.usuarios,
		Repositorio: alertas.NuevoRepositorio(pool),
	})

	require.NoError(t, servicio.ProcesarTardanzas(ctx))
	require.NoError(t, servicio.ProcesarTardanzas(ctx))
	assert.Equal(t, 1, contarAlertas(t, pool, alertas.TipoReiteracionTardanza))

	var destinatario int64
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT destinatario_id FROM alertas WHERE tipo = $1`, alertas.TipoReiteracionTardanza).
		Scan(&destinatario))
	assert.Equal(t, docenteID, destinatario)

	// Una cuarta tardanza del mismo mes no duplica la reiteración.
	segundaFecha := fecha.AddDate(0, 0, 1)
	var segundaID int64
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO marcaciones (alumno_id, fecha, hora, estado)
		 VALUES ($1, $2, $3, 'tardanza') RETURNING id`,
		datos.alumnoID, segundaFecha, segundaFecha.Add(8*time.Hour)).Scan(&segundaID))
	fuentes.asistencia.marcaciones = []alertas.EventoMarcacion{{
		MarcacionID: segundaID,
		AlumnoID:    datos.alumnoID,
		Fecha:       segundaFecha,
		Hora:        segundaFecha.Add(8 * time.Hour),
		Estado:      "tardanza",
	}}
	fuentes.asistencia.contarTardanzas = 4

	require.NoError(t, servicio.ProcesarTardanzas(ctx))
	assert.Equal(t, 1, contarAlertas(t, pool, alertas.TipoReiteracionTardanza))
}

func TestProcesarCierre_FaltaJustificadaNoCuenta(t *testing.T) {
	pool, _ := AbrirBaseDePruebasConURL(t)
	ctx := context.Background()
	datos := sembrarDatosAlertas(t, pool)

	fuentes := nuevasFuentesFalsas(datos.seccionID, datos.alumnoID, datos.administradorID)
	fuentes.asistencia.marcacionEn = time.Date(2026, 3, 8, 0, 0, 0, 0, time.UTC)
	fuentes.academico.umbrales = map[string]int{"faltas_seguidas": 2}

	justificaciones := &justificacionesFalsas{
		fechas: map[string]bool{"2026-03-09": true},
	}
	servicio := alertas.NuevoServicio(alertas.Dependencias{
		Asistencia:  fuentes.asistencia,
		Alumnos:     fuentes.alumnos,
		Academico:   fuentes.academico,
		Usuarios:    fuentes.usuarios,
		Repositorio: alertas.NuevoRepositorio(pool),
	}).ConJustificaciones(justificaciones)

	fecha := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	require.NoError(t, servicio.ProcesarCierre(ctx, fecha, "manana"))
	assert.Equal(t, 0, contarAlertas(t, pool, alertas.TipoReiteracionFalta),
		"la falta justificada no debe contar para la racha")

	// Al día siguiente la racha llega a dos días laborables: 11 y 10 de marzo.
	require.NoError(t, servicio.ProcesarCierre(ctx, fecha.AddDate(0, 0, 1), "manana"))
	assert.Equal(t, 1, contarAlertas(t, pool, alertas.TipoReiteracionFalta))
}

func TestMarcarSituacionCambiada_ConservaLaAlerta(t *testing.T) {
	pool, _ := AbrirBaseDePruebasConURL(t)
	ctx := context.Background()
	datos := sembrarDatosAlertas(t, pool)

	fecha := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	var marcacionID int64
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO marcaciones (alumno_id, fecha, hora, estado)
		 VALUES ($1, $2, $3, 'tardanza') RETURNING id`,
		datos.alumnoID, fecha, fecha.Add(8*time.Hour)).Scan(&marcacionID))

	fuentes := nuevasFuentesFalsas(datos.seccionID, datos.alumnoID, datos.administradorID)
	fuentes.asistencia.marcaciones = []alertas.EventoMarcacion{{
		MarcacionID: marcacionID,
		AlumnoID:    datos.alumnoID,
		Fecha:       fecha,
		Hora:        fecha.Add(8 * time.Hour),
		Estado:      "tardanza",
	}}
	fuentes.academico.umbrales = map[string]int{}

	servicio := alertas.NuevoServicio(alertas.Dependencias{
		Asistencia:  fuentes.asistencia,
		Alumnos:     fuentes.alumnos,
		Academico:   fuentes.academico,
		Usuarios:    fuentes.usuarios,
		Repositorio: alertas.NuevoRepositorio(pool),
	})
	require.NoError(t, servicio.ProcesarTardanzas(ctx))
	require.NoError(t, servicio.MarcarSituacionCambiada(ctx, datos.alumnoID, fecha, "Justificada por la familia"))

	var cambiada bool
	var motivo string
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT situacion_cambiada, COALESCE(motivo_cambio, '') FROM alertas WHERE tipo = 'tardanza'`).
		Scan(&cambiada, &motivo))
	assert.True(t, cambiada)
	assert.Equal(t, "Justificada por la familia", motivo)
	assert.Equal(t, 1, contarAlertas(t, pool, alertas.TipoTardanza), "la alerta no se elimina")

	var auditorias int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM auditoria WHERE entidad = 'alerta' AND accion = 'marcar_situacion_cambiada'`).
		Scan(&auditorias))
	assert.Equal(t, 1, auditorias)
}

func TestProcesarCierre_ConcurrenteNoDuplica(t *testing.T) {
	pool, _ := AbrirBaseDePruebasConURL(t)
	ctx := context.Background()
	datos := sembrarDatosAlertas(t, pool)

	fuentes := nuevasFuentesFalsas(datos.seccionID, datos.alumnoID, datos.administradorID)
	fuentes.asistencia.marcacionEn = time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC)
	fuentes.academico.umbrales = map[string]int{"faltas_seguidas": 1}

	servicio := alertas.NuevoServicio(alertas.Dependencias{
		Asistencia:  fuentes.asistencia,
		Alumnos:     fuentes.alumnos,
		Academico:   fuentes.academico,
		Usuarios:    fuentes.usuarios,
		Repositorio: alertas.NuevoRepositorio(pool),
	})

	fecha := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	var espera sync.WaitGroup
	for i := 0; i < 2; i++ {
		espera.Add(1)
		go func() {
			defer espera.Done()
			assert.NoError(t, servicio.ProcesarCierre(ctx, fecha, "manana"))
		}()
	}
	espera.Wait()

	assert.Equal(t, 1, contarAlertas(t, pool, alertas.TipoReiteracionFalta))
	var cierres int
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM cierres_alerta`).Scan(&cierres))
	assert.Equal(t, 1, cierres)
}

type justificacionesFalsas struct {
	fechas map[string]bool
}

func (j *justificacionesFalsas) EsFaltaJustificada(_ context.Context, _ int64, fecha time.Time) (bool, error) {
	return j.fechas[fecha.Format("2006-01-02")], nil
}

func crearUsuarioPrueba(t *testing.T, pool *pgxpool.Pool, correo, rol string) int64 {
	t.Helper()
	var id int64
	require.NoError(t, pool.QueryRow(context.Background(),
		`INSERT INTO usuarios (correo, contrasenia_hash, nombre_completo, rol)
		 VALUES ($1, 'x', $2, $3) RETURNING id`, correo, correo, rol).Scan(&id))
	return id
}

func contarAlertas(t *testing.T, pool *pgxpool.Pool, tipo string) int {
	t.Helper()
	var total int
	require.NoError(t, pool.QueryRow(context.Background(),
		`SELECT count(*) FROM alertas WHERE tipo = $1`, tipo).Scan(&total))
	return total
}
