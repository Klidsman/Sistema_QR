package integracion

import (
	"context"
	"testing"
	"time"

	"asistencia/internal/alumnos"
	"asistencia/internal/plataforma/fechas"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func dia(t *testing.T, texto string) time.Time {
	t.Helper()
	fecha, err := fechas.ParsearDia(texto)
	require.NoError(t, err)
	return fecha
}

func crearAlumnoConQR(t *testing.T, pool *pgxpool.Pool, seccionID, usuarioID int64, identificador string) (int64, *alumnos.Servicio) {
	t.Helper()
	servicio := alumnos.NuevoServicio(pool)
	ctx := context.Background()

	alumnoID, err := servicio.Crear(ctx, usuarioID, alumnos.DatosAlumno{
		Identificador: identificador,
		Nombres:       "Rosa",
		Apellidos:     "Huamán",
		SeccionID:     seccionID,
	})
	require.NoError(t, err)
	_, err = servicio.EmitirQR(ctx, usuarioID, alumnoID)
	require.NoError(t, err)
	return alumnoID, servicio
}

func TestCambioSeccion_CierraElTramoAnteriorYAbreElNuevo(t *testing.T) {
	pool := AbrirBaseDePruebas(t)
	seccionA := SembrarSeccion(t, pool, "3.°", "A")
	seccionB := SembrarSeccion(t, pool, "3.°", "B")
	usuarioID := CrearUsuarioPrueba(t, pool)
	alumnoID, servicio := crearAlumnoConQR(t, pool, seccionA, usuarioID, "50123456")
	ctx := context.Background()

	fecha := dia(t, "2027-05-15")
	require.NoError(t, servicio.CambiarSeccion(ctx, usuarioID, alumnoID, seccionB, fecha))

	alumno, err := servicio.Obtener(ctx, alumnoID)
	require.NoError(t, err)
	assert.Equal(t, seccionB, alumno.SeccionID)
	assert.Equal(t, "B", alumno.Seccion)
	assert.Equal(t, alumnos.EstadoActivo, alumno.Estado)

	historial, err := servicio.HistorialSecciones(ctx, alumnoID)
	require.NoError(t, err)
	require.Len(t, historial, 2)

	assert.True(t, historial[0].Actual())
	assert.Equal(t, seccionB, historial[0].SeccionID)
	assert.Equal(t, fecha.Format("2006-01-02"), historial[0].FechaInicio.Format("2006-01-02"))

	require.NotNil(t, historial[1].FechaFin)
	assert.Equal(t, seccionA, historial[1].SeccionID)
	assert.Equal(t, "A", historial[1].Seccion)
	assert.Equal(t, fecha.Format("2006-01-02"), historial[1].FechaFin.Format("2006-01-02"))

	_, err = servicio.CredencialActiva(ctx, alumnoID)
	assert.NoError(t, err, "el cambio de sección no debe invalidar el QR")
}

func TestCambioSeccion_RechazaLaMismaSeccion(t *testing.T) {
	pool := AbrirBaseDePruebas(t)
	seccionA := SembrarSeccion(t, pool, "3.°", "A")
	usuarioID := CrearUsuarioPrueba(t, pool)
	alumnoID, servicio := crearAlumnoConQR(t, pool, seccionA, usuarioID, "50234567")
	ctx := context.Background()

	err := servicio.CambiarSeccion(ctx, usuarioID, alumnoID, seccionA, time.Now())

	assert.ErrorIs(t, err, alumnos.ErrMismaSeccion)
}

func TestRegistrarSalida_TrasladadoRevocaQYConservaHistorial(t *testing.T) {
	pool := AbrirBaseDePruebas(t)
	seccionA := SembrarSeccion(t, pool, "3.°", "A")
	usuarioID := CrearUsuarioPrueba(t, pool)
	alumnoID, servicio := crearAlumnoConQR(t, pool, seccionA, usuarioID, "50345678")
	ctx := context.Background()

	fecha := dia(t, "2027-06-30")
	require.NoError(t, servicio.RegistrarSalida(ctx, usuarioID, alumnoID, alumnos.EstadoTrasladado, fecha))

	alumno, err := servicio.Obtener(ctx, alumnoID)
	require.NoError(t, err)
	assert.Equal(t, alumnos.EstadoTrasladado, alumno.Estado)
	require.NotNil(t, alumno.FechaSalida)
	assert.Equal(t, fecha.Format("2006-01-02"), alumno.FechaSalida.Format("2006-01-02"))

	_, err = servicio.CredencialActiva(ctx, alumnoID)
	assert.ErrorIs(t, err, alumnos.ErrQRSinActivo)

	credenciales, err := servicio.HistorialQR(ctx, alumnoID)
	require.NoError(t, err)
	require.Len(t, credenciales, 1)
	assert.Equal(t, alumnos.CredencialRevocada, credenciales[0].Estado)
	assert.Equal(t, alumnos.MotivoRetiro, credenciales[0].Motivo)

	historial, err := servicio.HistorialSecciones(ctx, alumnoID)
	require.NoError(t, err)
	require.Len(t, historial, 1)
	assert.False(t, historial[0].Actual())
	require.NotNil(t, historial[0].FechaFin)
	assert.Equal(t, fecha.Format("2006-01-02"), historial[0].FechaFin.Format("2006-01-02"))

	var accion string
	err = pool.QueryRow(ctx,
		`SELECT accion FROM auditoria WHERE entidad = 'alumno' AND entidad_id = $1 AND accion = 'trasladar'`,
		alumnoID,
	).Scan(&accion)
	require.NoError(t, err)
	assert.Equal(t, "trasladar", accion)
}

func TestRegistrarSalida_RetiradoRevocaElQR(t *testing.T) {
	pool := AbrirBaseDePruebas(t)
	seccionA := SembrarSeccion(t, pool, "2.°", "A")
	usuarioID := CrearUsuarioPrueba(t, pool)
	alumnoID, servicio := crearAlumnoConQR(t, pool, seccionA, usuarioID, "50456789")
	ctx := context.Background()

	fecha := dia(t, "2027-07-10")
	require.NoError(t, servicio.RegistrarSalida(ctx, usuarioID, alumnoID, alumnos.EstadoRetirado, fecha))

	alumno, err := servicio.Obtener(ctx, alumnoID)
	require.NoError(t, err)
	assert.Equal(t, alumnos.EstadoRetirado, alumno.Estado)

	_, err = servicio.CredencialActiva(ctx, alumnoID)
	assert.ErrorIs(t, err, alumnos.ErrQRSinActivo)

	var accion string
	err = pool.QueryRow(ctx,
		`SELECT accion FROM auditoria WHERE entidad = 'alumno' AND entidad_id = $1 AND accion = 'retirar'`,
		alumnoID,
	).Scan(&accion)
	require.NoError(t, err)
	assert.Equal(t, "retirar", accion)
}

func TestRegistrarSalida_RechazaAlumnoNoActivo(t *testing.T) {
	pool := AbrirBaseDePruebas(t)
	seccionA := SembrarSeccion(t, pool, "1.°", "A")
	usuarioID := CrearUsuarioPrueba(t, pool)
	alumnoID, servicio := crearAlumnoConQR(t, pool, seccionA, usuarioID, "50567890")
	ctx := context.Background()

	require.NoError(t, servicio.RegistrarSalida(ctx, usuarioID, alumnoID, alumnos.EstadoRetirado, time.Now()))
	err := servicio.RegistrarSalida(ctx, usuarioID, alumnoID, alumnos.EstadoTrasladado, time.Now())

	assert.ErrorIs(t, err, alumnos.ErrAlumnoNoActivo)
}

func TestRegistrarSalida_RechazaTipoInvalido(t *testing.T) {
	pool := AbrirBaseDePruebas(t)
	seccionA := SembrarSeccion(t, pool, "1.°", "A")
	usuarioID := CrearUsuarioPrueba(t, pool)
	alumnoID, servicio := crearAlumnoConQR(t, pool, seccionA, usuarioID, "50678901")
	ctx := context.Background()

	err := servicio.RegistrarSalida(ctx, usuarioID, alumnoID, "activo", time.Now())

	assert.ErrorIs(t, err, alumnos.ErrTipoSalidaInvalido)
}
