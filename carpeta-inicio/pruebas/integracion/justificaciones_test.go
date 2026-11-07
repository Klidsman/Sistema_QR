package integracion

import (
	"context"
	"testing"
	"time"

	"asistencia/internal/academico"
	"asistencia/internal/alumnos"
	"asistencia/internal/asistencia"
	"asistencia/internal/justificaciones"
	"asistencia/internal/plataforma/almacenamiento"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJustificaciones_FlujoCompleto(t *testing.T) {
	pool, _ := AbrirBaseDePruebasConURL(t)
	ctx := context.Background()
	datos := sembrarDatosAlertas(t, pool)
	directorID := crearUsuarioPrueba(t, pool, "director@justifica.pe", "director")
	adminID := datos.administradorID

	servicio := nuevoServicioJustificaciones(t, pool)

	// Falta del 19 de marzo, dentro del plazo de 5 días.
	fechaFalta := time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC)
	resultado, err := servicio.Registrar(ctx, adminID, "administrador", datos.alumnoID,
		"Enfermedad con certificado", fechaFalta, fechaFalta, nil)
	require.NoError(t, err)
	require.NotZero(t, resultado.JustificacionID)
	require.Len(t, resultado.Cubiertas, 1)
	assert.Equal(t, justificaciones.TipoFalta, resultado.Cubiertas[0].Tipo)

	justificacion, err := servicio.ObtenerConAcceso(ctx, directorID, "director", resultado.JustificacionID)
	require.NoError(t, err)
	assert.Equal(t, justificaciones.EstadoPendiente, justificacion.Estado)

	require.NoError(t, servicio.Resolver(ctx, adminID, "administrador", resultado.JustificacionID, true, "Aprobado por administración"))
	justificado, err := servicio.EsFaltaJustificada(ctx, datos.alumnoID, fechaFalta)
	require.NoError(t, err)
	assert.True(t, justificado)

	// El director tampoco puede resolver una justificación que registró él mismo.
	otraFecha := fechaFalta.AddDate(0, 0, 1)
	resultadoDirector, err := servicio.Registrar(ctx, directorID, "director", datos.alumnoID,
		"Trámite presentado por dirección", otraFecha, otraFecha, nil)
	require.NoError(t, err)
	assert.ErrorIs(t, servicio.Resolver(ctx, directorID, "director", resultadoDirector.JustificacionID, true, ""),
		justificaciones.ErrAutoresolucion)
	require.NoError(t, servicio.Resolver(ctx, adminID, "administrador", resultadoDirector.JustificacionID, true, ""))
}

func TestJustificaciones_TardanzaAprobadaYAnulada(t *testing.T) {
	pool, _ := AbrirBaseDePruebasConURL(t)
	ctx := context.Background()
	datos := sembrarDatosAlertas(t, pool)
	directorID := crearUsuarioPrueba(t, pool, "director2@justifica.pe", "director")
	adminAnulador := crearUsuarioPrueba(t, pool, "admin2@justifica.pe", "administrador")
	servicio := nuevoServicioJustificaciones(t, pool)

	fecha := time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC)
	_, err := pool.Exec(ctx,
		`INSERT INTO marcaciones (alumno_id, fecha, hora, estado)
		 VALUES ($1, $2, $3, 'tardanza')`, datos.alumnoID, fecha, fecha.Add(8*time.Hour))
	require.NoError(t, err)

	resultado, err := servicio.Registrar(ctx, datos.administradorID, "administrador", datos.alumnoID,
		"Trámite familiar", fecha, fecha, nil)
	require.NoError(t, err)
	require.NoError(t, servicio.Resolver(ctx, directorID, "director", resultado.JustificacionID, true, ""))

	var justificada bool
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT justificada FROM marcaciones WHERE alumno_id = $1 AND fecha = $2`,
		datos.alumnoID, fecha).Scan(&justificada))
	assert.True(t, justificada)

	require.NoError(t, servicio.Anular(ctx, adminAnulador, "administrador", resultado.JustificacionID, "Error de digitación"))
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT justificada FROM marcaciones WHERE alumno_id = $1 AND fecha = $2`,
		datos.alumnoID, fecha).Scan(&justificada))
	assert.False(t, justificada)
}

func TestJustificaciones_RechazoYDuplicados(t *testing.T) {
	pool, _ := AbrirBaseDePruebasConURL(t)
	ctx := context.Background()
	datos := sembrarDatosAlertas(t, pool)
	directorID := crearUsuarioPrueba(t, pool, "director3@justifica.pe", "director")
	servicio := nuevoServicioJustificaciones(t, pool)

	fecha := time.Date(2026, 3, 17, 0, 0, 0, 0, time.UTC)
	resultado, err := servicio.Registrar(ctx, datos.administradorID, "administrador", datos.alumnoID,
		"Motivo válido", fecha, fecha, nil)
	require.NoError(t, err)

	err = servicio.Resolver(ctx, directorID, "director", resultado.JustificacionID, false, "")
	assert.ErrorIs(t, err, justificaciones.ErrComentarioRequerido)
	require.NoError(t, servicio.Resolver(ctx, directorID, "director", resultado.JustificacionID, false, "Sin sustento"))

	justificado, err := servicio.EsFaltaJustificada(ctx, datos.alumnoID, fecha)
	require.NoError(t, err)
	assert.False(t, justificado)

	// Una justificación rechazada permite presentar otra.
	_, err = servicio.Registrar(ctx, datos.administradorID, "administrador", datos.alumnoID,
		"Segundo intento", fecha, fecha, nil)
	require.NoError(t, err)
}

func TestJustificaciones_FueraDePlazoYDuplicada(t *testing.T) {
	pool, _ := AbrirBaseDePruebasConURL(t)
	ctx := context.Background()
	datos := sembrarDatosAlertas(t, pool)
	servicio := nuevoServicioJustificaciones(t, pool)

	antigua := time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC)
	_, err := servicio.Registrar(ctx, datos.administradorID, "administrador", datos.alumnoID,
		"Fuera de plazo", antigua, antigua, nil)
	assert.ErrorIs(t, err, justificaciones.ErrFueraDePlazo)

	fecha := time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC)
	_, err = servicio.Registrar(ctx, datos.administradorID, "administrador", datos.alumnoID,
		"Pendiente", fecha, fecha, nil)
	require.NoError(t, err)

	_, err = servicio.Registrar(ctx, datos.administradorID, "administrador", datos.alumnoID,
		"Duplicada", fecha, fecha, nil)
	assert.ErrorIs(t, err, justificaciones.ErrSinFechasAplicables)
}

func nuevoServicioJustificaciones(t *testing.T, pgxPool *pgxpool.Pool) *justificaciones.Servicio {
	t.Helper()
	servicioAlumnos := alumnos.NuevoServicio(pgxPool)
	servicioConfiguracion := academico.NuevoServicioConfiguracion(pgxPool)
	servicioAsistencia := asistencia.NuevoServicio(asistencia.DependenciasServicio{
		Repositorio: asistencia.NuevoRepositorio(pgxPool),
		Academico:   servicioConfiguracion,
		Alumnos:     servicioAlumnos,
	})
	return justificaciones.NuevoServicio(justificaciones.Dependencias{
		Alumnos:    servicioAlumnos,
		Academico:  servicioConfiguracion,
		Asistencia: servicioAsistencia,
		Alertas:    &alertasJustificacionesFalsas{},
		Almacen:    almacenamiento.NuevoAlmacenDisco(t.TempDir()),
		Repo:       justificaciones.NuevoRepositorio(pgxPool),
		Reloj:      func() time.Time { return time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC) },
	})
}

type alertasJustificacionesFalsas struct{}

func (a *alertasJustificacionesFalsas) MarcarSituacionCambiada(context.Context, int64, time.Time, string) error {
	return nil
}
