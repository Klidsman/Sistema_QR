package integracion

import (
	"context"
	"testing"

	"asistencia/internal/alumnos"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCredencialesQR_UnSoloVigenteYReemplazoTransaccional(t *testing.T) {
	pool := AbrirBaseDePruebas(t)
	seccionID := SembrarSeccion(t, pool, "3.°", "A")
	usuarioID := CrearUsuarioPrueba(t, pool)
	servicio := alumnos.NuevoServicio(pool)
	ctx := context.Background()

	alumnoID, err := servicio.Crear(ctx, usuarioID, alumnos.DatosAlumno{
		Identificador: "11223344",
		Nombres:       "Luis",
		Apellidos:     "Ramos",
		SeccionID:     seccionID,
	})
	require.NoError(t, err)

	primera, err := servicio.EmitirQR(ctx, usuarioID, alumnoID)
	require.NoError(t, err)
	assert.Len(t, primera.Token, 43)

	_, err = servicio.EmitirQR(ctx, usuarioID, alumnoID)
	assert.ErrorIs(t, err, alumnos.ErrQRYaActivo)

	segunda, err := servicio.ReemplazarQR(ctx, usuarioID, alumnoID, alumnos.MotivoRobo)
	require.NoError(t, err)
	assert.NotEqual(t, primera.Token, segunda.Token)

	activa, err := servicio.CredencialActiva(ctx, alumnoID)
	require.NoError(t, err)
	assert.Equal(t, segunda.Token, activa.Token)

	historial, err := servicio.HistorialQR(ctx, alumnoID)
	require.NoError(t, err)
	require.Len(t, historial, 2)
	assert.Equal(t, alumnos.CredencialActiva, historial[0].Estado)
	assert.Equal(t, alumnos.CredencialRevocada, historial[1].Estado)
	assert.Equal(t, alumnos.MotivoRobo, historial[1].Motivo)
}

func TestCredencialesQR_RechazaMotivoInvalido(t *testing.T) {
	pool := AbrirBaseDePruebas(t)
	seccionID := SembrarSeccion(t, pool, "3.°", "A")
	usuarioID := CrearUsuarioPrueba(t, pool)
	servicio := alumnos.NuevoServicio(pool)
	ctx := context.Background()

	alumnoID, err := servicio.Crear(ctx, usuarioID, alumnos.DatosAlumno{
		Identificador: "22334455",
		Nombres:       "Elena",
		Apellidos:     "Paredes",
		SeccionID:     seccionID,
	})
	require.NoError(t, err)
	_, err = servicio.EmitirQR(ctx, usuarioID, alumnoID)
	require.NoError(t, err)

	_, err = servicio.ReemplazarQR(ctx, usuarioID, alumnoID, "porque si")
	assert.ErrorIs(t, err, alumnos.ErrMotivoRequerido)
}

func TestCredencialesQR_ElIndiceUnicoImpideDosActivas(t *testing.T) {
	pool := AbrirBaseDePruebas(t)
	seccionID := SembrarSeccion(t, pool, "3.°", "A")
	usuarioID := CrearUsuarioPrueba(t, pool)
	servicio := alumnos.NuevoServicio(pool)
	ctx := context.Background()

	alumnoID, err := servicio.Crear(ctx, usuarioID, alumnos.DatosAlumno{
		Identificador: "33445566",
		Nombres:       "Mario",
		Apellidos:     "Salas",
		SeccionID:     seccionID,
	})
	require.NoError(t, err)
	_, err = servicio.EmitirQR(ctx, usuarioID, alumnoID)
	require.NoError(t, err)

	_, err = pool.Exec(ctx,
		`INSERT INTO credenciales_qr (alumno_id, token) VALUES ($1, $2)`,
		alumnoID, "token_manual_con_longitud_invalida",
	)
	assert.Error(t, err, "la segunda credencial activa debe ser rechazada")
}

func TestCredencialesQR_DesactivarRevocaElQRVigente(t *testing.T) {
	pool := AbrirBaseDePruebas(t)
	seccionID := SembrarSeccion(t, pool, "3.°", "A")
	usuarioID := CrearUsuarioPrueba(t, pool)
	servicio := alumnos.NuevoServicio(pool)
	ctx := context.Background()

	alumnoID, err := servicio.Crear(ctx, usuarioID, alumnos.DatosAlumno{
		Identificador: "44556677",
		Nombres:       "Rosa",
		Apellidos:     "Huamán",
		SeccionID:     seccionID,
	})
	require.NoError(t, err)
	_, err = servicio.EmitirQR(ctx, usuarioID, alumnoID)
	require.NoError(t, err)

	require.NoError(t, servicio.Desactivar(ctx, usuarioID, alumnoID))

	_, err = servicio.CredencialActiva(ctx, alumnoID)
	assert.ErrorIs(t, err, alumnos.ErrQRSinActivo)

	historial, err := servicio.HistorialQR(ctx, alumnoID)
	require.NoError(t, err)
	require.Len(t, historial, 1)
	assert.Equal(t, alumnos.CredencialRevocada, historial[0].Estado)
	assert.Equal(t, alumnos.MotivoDesactivacion, historial[0].Motivo)
}
