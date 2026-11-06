package integracion

import (
	"context"
	"testing"
	"time"

	"asistencia/internal/academico"
	"asistencia/internal/reportes"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReportes_DiarioPanelResumenYReiteracion(t *testing.T) {
	pool, _ := AbrirBaseDePruebasConURL(t)
	ctx := context.Background()
	datos := sembrarReportes(t, pool)

	servicio := reportes.NuevoServicio(reportes.Dependencias{
		Repo:      reportes.NuevoRepositorio(pool),
		Academico: academico.NuevoServicioConfiguracion(pool),
		Reloj:     func() time.Time { return time.Date(2026, 3, 10, 20, 0, 0, 0, time.UTC) },
	})
	fecha := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)

	diario, err := servicio.Diario(ctx, reportes.Filtros{Desde: fecha, Hasta: fecha, Rol: "administrador"})
	require.NoError(t, err)
	require.Len(t, diario, 3)
	estados := map[string]string{}
	for _, fila := range diario {
		estados[fila.Identificador] = fila.Estado
	}
	assert.Equal(t, reportes.EstadoPresente, estados["A001"])
	assert.Equal(t, reportes.EstadoTardanza, estados["A002"])
	assert.Equal(t, reportes.EstadoFaltaJustificada, estados["A003"])

	panel, err := servicio.Panel(ctx, reportes.Filtros{Desde: fecha, Hasta: fecha, Rol: "administrador"})
	require.NoError(t, err)
	assert.Equal(t, 2, panel.Presentes)
	assert.Equal(t, 1, panel.Tardanzas)
	assert.Equal(t, 0, panel.Pendientes+panel.Ausentes, "la falta justificada no cuenta como ausente")
	assert.Equal(t, 3, panel.Total)

	inicioMes := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	resumen, err := servicio.Resumen(ctx, reportes.Filtros{
		Desde: inicioMes, Hasta: fecha, Rol: "administrador",
	})
	require.NoError(t, err)
	require.NotEmpty(t, resumen)
	var alumno1 reportes.ResumenAlumno
	for _, fila := range resumen {
		if fila.Identificador == "A001" {
			alumno1 = fila
		}
	}
	assert.Equal(t, 7, alumno1.DiasLaborables)
	assert.Equal(t, 1, alumno1.Asistencias)
	assert.Equal(t, 6, alumno1.Faltas)
	assert.InDelta(t, 14.3, alumno1.Porcentaje, 0.1)

	reiteracion, err := servicio.Reiteracion(ctx, reportes.Filtros{
		Desde: inicioMes, Hasta: fecha, Rol: "administrador",
	})
	require.NoError(t, err)
	require.Len(t, reiteracion, 1)
	assert.Contains(t, reiteracion[0].Alumno, "Tardanza")
	assert.Equal(t, 1, reiteracion[0].Tardanzas)
	assert.Equal(t, datos.umbral, reiteracion[0].UmbralTardanzas)
}

func TestReportes_DocenteSoloVeSuSeccionYFiltroEstado(t *testing.T) {
	pool, _ := AbrirBaseDePruebasConURL(t)
	ctx := context.Background()
	datos := sembrarReportes(t, pool)
	docenteID := crearUsuarioPrueba(t, pool, "docente-reportes@prueba.pe", "docente")
	_, err := pool.Exec(ctx,
		`INSERT INTO asignaciones_docente (docente_id, curso_id, seccion_id, fecha_inicio, creado_por)
		 VALUES ($1, $2, $3, $4, $5)`,
		docenteID, datos.cursoID, datos.seccionID, time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), datos.administradorID)
	require.NoError(t, err)

	servicio := reportes.NuevoServicio(reportes.Dependencias{
		Repo:      reportes.NuevoRepositorio(pool),
		Academico: academico.NuevoServicioConfiguracion(pool),
		Reloj:     func() time.Time { return time.Date(2026, 3, 10, 20, 0, 0, 0, time.UTC) },
	})
	fecha := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	diario, err := servicio.Diario(ctx, reportes.Filtros{
		Desde: fecha, Hasta: fecha, UsuarioID: docenteID, Rol: "docente", Estado: reportes.EstadoTardanza,
	})
	require.NoError(t, err)
	require.Len(t, diario, 1)
	assert.Equal(t, "A002", diario[0].Identificador)
}

type datosReportes struct {
	administradorID int64
	seccionID       int64
	alumnoID        int64
	cursoID         int64
	umbral          int
}

func sembrarReportes(t *testing.T, pool *pgxpool.Pool) datosReportes {
	t.Helper()
	ctx := context.Background()
	base := sembrarDatosAlertas(t, pool)
	datos := datosReportes{administradorID: base.administradorID, seccionID: base.seccionID, alumnoID: base.alumnoID}
	var anioID, gradoID int64
	require.NoError(t, pool.QueryRow(ctx, `SELECT anio_escolar_id FROM grados WHERE id = (SELECT grado_id FROM secciones WHERE id = $1)`, datos.seccionID).Scan(&anioID))
	require.NoError(t, pool.QueryRow(ctx, `SELECT grado_id FROM secciones WHERE id = $1`, datos.seccionID).Scan(&gradoID))

	for _, alumno := range []struct{ id int64 }{{datos.alumnoID}} {
		_, err := pool.Exec(ctx,
			`INSERT INTO historial_secciones (alumno_id, seccion_id, fecha_inicio)
			 VALUES ($1, $2, '2026-03-01')`, alumno.id, datos.seccionID)
		require.NoError(t, err)
	}
	otros := []struct {
		identificador string
		apellidos     string
	}{
		{"A002", "Tardanza"},
		{"A003", "Justificada"},
	}
	for _, alumno := range otros {
		var alumnoID int64
		require.NoError(t, pool.QueryRow(ctx,
			`INSERT INTO alumnos (identificador, nombres, apellidos, nombres_busqueda, apellidos_busqueda, seccion_id)
			 VALUES ($1, 'Alumno', $2, 'alumno', lower($2), $3) RETURNING id`,
			alumno.identificador, alumno.apellidos, datos.seccionID).Scan(&alumnoID))
		_, err := pool.Exec(ctx,
			`INSERT INTO historial_secciones (alumno_id, seccion_id, fecha_inicio)
			 VALUES ($1, $2, '2026-03-01')`, alumnoID, datos.seccionID)
		require.NoError(t, err)
	}

	_, err := pool.Exec(ctx,
		`INSERT INTO horarios_turno (turno, hora_ingreso, tolerancia_minutos, hora_limite_tardanza, hora_cierre, vigente_desde)
		 VALUES ('manana', '07:30', 10, '08:00', '08:30', '2026-03-01')`)
	require.NoError(t, err)

	var cursoID int64
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO cursos (anio_escolar_id, nombre) VALUES ($1, 'Curso') RETURNING id`,
		anioID).Scan(&cursoID))
	datos.cursoID = cursoID

	datos.umbral = 1
	_, err = pool.Exec(ctx,
		`INSERT INTO umbrales_alerta (anio_escolar_id, tipo, valor, vigente_desde)
		 VALUES ($1, 'tardanzas', $2, '2026-03-01')`, anioID, datos.umbral)
	require.NoError(t, err)

	fecha := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	var segundoID int64
	require.NoError(t, pool.QueryRow(ctx, `SELECT id FROM alumnos WHERE identificador = 'A002'`).Scan(&segundoID))
	_, err = pool.Exec(ctx,
		`INSERT INTO marcaciones (alumno_id, fecha, hora, estado) VALUES ($1, $2, $3, 'puntual')`,
		datos.alumnoID, fecha, fecha.Add(7*time.Hour+30*time.Minute))
	require.NoError(t, err)
	_, err = pool.Exec(ctx,
		`INSERT INTO marcaciones (alumno_id, fecha, hora, estado) VALUES ($1, $2, $3, 'tardanza')`,
		segundoID, fecha, fecha.Add(7*time.Hour+45*time.Minute))
	require.NoError(t, err)

	var terceroID int64
	require.NoError(t, pool.QueryRow(ctx, `SELECT id FROM alumnos WHERE identificador = 'A003'`).Scan(&terceroID))
	var justificacionID int64
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO justificaciones (alumno_id, seccion_id, motivo, estado, registrante_id)
		 VALUES ($1, $2, 'Enfermedad', 'aprobada', $3) RETURNING id`,
		terceroID, datos.seccionID, datos.administradorID).Scan(&justificacionID))
	_, err = pool.Exec(ctx,
		`INSERT INTO justificacion_fechas (justificacion_id, alumno_id, fecha_evento, tipo, estado)
		 VALUES ($1, $2, $3, 'falta', 'aprobada')`, justificacionID, terceroID, fecha)
	require.NoError(t, err)

	_ = gradoID
	return datos
}
