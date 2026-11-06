package unidad

import (
	"testing"

	"asistencia/internal/alumnos"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func alumnoConCarnet(id int64, foto, token string) alumnos.AlumnoCarnet {
	return alumnos.AlumnoCarnet{
		ID:            id,
		Identificador: "00123456",
		Nombres:       "María",
		Apellidos:     "Núñez Ríos",
		Grado:         "3.°",
		Seccion:       "A",
		AnioEscolar:   "2026",
		FotoClave:     foto,
		TokenQR:       token,
	}
}

func TestPrepararLoteCarnets_SeparaCarnetsImprimiblesYAvisos(t *testing.T) {
	lote := alumnos.PrepararLote([]alumnos.AlumnoCarnet{
		alumnoConCarnet(1, "alumnos/1/foto.jpg", "token-uno"),
		alumnoConCarnet(2, "alumnos/2/foto.jpg", "token-dos"),
		alumnoConCarnet(3, "", "token-tres"),
		alumnoConCarnet(4, "alumnos/4/foto.jpg", ""),
		alumnoConCarnet(5, "", ""),
	})

	require.Len(t, lote.Carnets, 3)
	assert.Len(t, lote.Avisos.SinFoto, 1)
	assert.Len(t, lote.Avisos.SinQR, 2)

	// Los alumnos sin QR vigente no se imprimen porque el carné quedaría sin código.
	for _, carnet := range lote.Carnets {
		assert.True(t, carnet.TieneQR())
	}
	assert.Equal(t, int64(3), lote.Avisos.SinFoto[0].ID)
	assert.Equal(t, int64(4), lote.Avisos.SinQR[0].ID)
	assert.Equal(t, int64(5), lote.Avisos.SinQR[1].ID)
}

func TestPrepararLoteCarnets_SinAvisosCuandoTodosTienenFotoYQR(t *testing.T) {
	lote := alumnos.PrepararLote([]alumnos.AlumnoCarnet{
		alumnoConCarnet(1, "alumnos/1/foto.jpg", "token-uno"),
	})

	require.Len(t, lote.Carnets, 1)
	assert.Empty(t, lote.Avisos.SinFoto)
	assert.Empty(t, lote.Avisos.SinQR)
	assert.False(t, lote.Avisos.HayAvisos())
}

func TestPrepararLoteCarnets_ListaVaciaNoProduceCarnets(t *testing.T) {
	lote := alumnos.PrepararLote(nil)

	assert.Empty(t, lote.Carnets)
	assert.Empty(t, lote.Avisos.SinFoto)
	assert.Empty(t, lote.Avisos.SinQR)
}

func TestSeleccionCarnets_RequiereSeccionOGrado(t *testing.T) {
	assert.False(t, alumnos.SeleccionCarnets{}.Valida())
	assert.True(t, alumnos.SeleccionCarnets{SeccionID: 4}.Valida())
	assert.True(t, alumnos.SeleccionCarnets{GradoID: 2}.Valida())
}
