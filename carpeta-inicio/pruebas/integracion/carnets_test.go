package integracion

import (
	"context"
	"testing"

	"asistencia/internal/alumnos"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCarnets_SeleccionPorSeccionYGradoConAvisos(t *testing.T) {
	pool := AbrirBaseDePruebas(t)
	seccionA := SembrarSeccion(t, pool, "3.°", "A")
	seccionB := SembrarSeccion(t, pool, "3.°", "B")
	usuarioID := CrearUsuarioPrueba(t, pool)
	servicio := alumnos.NuevoServicio(pool)
	ctx := context.Background()

	crear := func(identificador, nombres, apellidos string, seccionID int64) int64 {
		id, err := servicio.Crear(ctx, usuarioID, alumnos.DatosAlumno{
			Identificador: identificador,
			Nombres:       nombres,
			Apellidos:     apellidos,
			SeccionID:     seccionID,
		})
		require.NoError(t, err)
		return id
	}

	conFotoYQR := crear("30000001", "Ana", "Ríos", seccionA)
	require.NoError(t, servicio.ActualizarFoto(ctx, usuarioID, conFotoYQR, "alumnos/ana.jpg"))
	_, err := servicio.EmitirQR(ctx, usuarioID, conFotoYQR)
	require.NoError(t, err)

	sinFotoConQR := crear("30000002", "Luis", "Pérez", seccionA)
	_, err = servicio.EmitirQR(ctx, usuarioID, sinFotoConQR)
	require.NoError(t, err)

	conFotoSinQR := crear("30000003", "Elena", "Ñuño", seccionA)
	require.NoError(t, servicio.ActualizarFoto(ctx, usuarioID, conFotoSinQR, "alumnos/elena.jpg"))

	deOtraSeccion := crear("30000004", "Mario", "Quispe", seccionB)
	require.NoError(t, servicio.ActualizarFoto(ctx, usuarioID, deOtraSeccion, "alumnos/mario.jpg"))
	_, err = servicio.EmitirQR(ctx, usuarioID, deOtraSeccion)
	require.NoError(t, err)

	loteA, err := servicio.ListarCarnets(ctx, alumnos.SeleccionCarnets{SeccionID: seccionA})
	require.NoError(t, err)
	require.Len(t, loteA.Carnets, 2)
	assert.Len(t, loteA.Avisos.SinFoto, 1)
	assert.Len(t, loteA.Avisos.SinQR, 1)
	assert.Equal(t, int64(sinFotoConQR), loteA.Avisos.SinFoto[0].ID)
	assert.Equal(t, int64(conFotoSinQR), loteA.Avisos.SinQR[0].ID)

	loteGrado, err := servicio.ListarCarnets(ctx, alumnos.SeleccionCarnets{GradoID: loteA.Carnets[0].GradoID})
	require.NoError(t, err)
	assert.Len(t, loteGrado.Carnets, 3)

	_, err = servicio.ListarCarnets(ctx, alumnos.SeleccionCarnets{})
	assert.ErrorIs(t, err, alumnos.ErrSeleccionCarnetsInvalida)
}
