package integracion

import (
	"context"
	"testing"
	"time"

	"asistencia/internal/alertas"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcesarTardanzas_GeneraUnaAlertaParaElPortero(t *testing.T) {
	pool, _ := AbrirBaseDePruebasConURL(t)
	ctx := context.Background()
	datos := sembrarDatosAlertas(t, pool)

	var porteroID int64
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO usuarios (correo, contrasenia_hash, nombre_completo, rol)
		 VALUES ('portero@prueba.pe', 'x', 'Portero', 'porteria') RETURNING id`).Scan(&porteroID))

	fecha := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	var marcacionID int64
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO marcaciones (alumno_id, fecha, hora, estado, registrada_por)
		 VALUES ($1, $2, $3, 'tardanza', $4) RETURNING id`,
		datos.alumnoID, fecha, fecha.Add(8*time.Hour), porteroID).Scan(&marcacionID))

	fuentes := nuevasFuentesFalsas(datos.seccionID, datos.alumnoID, datos.administradorID)
	fuentes.asistencia.marcaciones = []alertas.EventoMarcacion{{
		MarcacionID:   marcacionID,
		AlumnoID:      datos.alumnoID,
		Fecha:         fecha,
		Hora:          fecha.Add(8 * time.Hour),
		Estado:        "tardanza",
		RegistradaPor: &porteroID,
	}}
	// Solo existe el umbral de faltas seguidas: no se genera reiteración.
	fuentes.academico.umbrales = map[string]int{"faltas_seguidas": 3}

	servicio := alertas.NuevoServicio(alertas.Dependencias{
		Asistencia:  fuentes.asistencia,
		Alumnos:     fuentes.alumnos,
		Academico:   fuentes.academico,
		Usuarios:    fuentes.usuarios,
		Repositorio: alertas.NuevoRepositorio(pool),
	})

	require.NoError(t, servicio.ProcesarTardanzas(ctx))
	require.NoError(t, servicio.ProcesarTardanzas(ctx))

	var total int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM alertas WHERE tipo = 'tardanza'`).Scan(&total))
	assert.Equal(t, 1, total, "la ejecución repetida no debe duplicar la alerta")

	var destinatario int64
	var tipo string
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT destinatario_id, tipo FROM alertas WHERE tipo = 'tardanza'`).Scan(&destinatario, &tipo))
	assert.Equal(t, porteroID, destinatario, "la alerta debe llegar al portero que registró el ingreso")
	assert.Equal(t, alertas.TipoTardanza, tipo)
}

func TestProcesarCierre_SinDocenteLlegaAlAdministrador(t *testing.T) {
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
	require.NoError(t, servicio.ProcesarCierre(ctx, fecha, "manana"))

	var destinatario int64
	var sinResponsable bool
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT destinatario_id, sin_responsable FROM alertas WHERE tipo = 'reiteracion_faltas'`).
		Scan(&destinatario, &sinResponsable))
	assert.Equal(t, datos.administradorID, destinatario)
	assert.True(t, sinResponsable, "la alerta sin docente debe marcarse como sin responsable")
}
