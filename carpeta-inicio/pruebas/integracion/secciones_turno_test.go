package integracion

import (
	"context"
	"testing"
	"time"

	"asistencia/internal/academico"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecciones_TurnoVigenteSeDerivaDelHistorial(t *testing.T) {
	pool := AbrirBaseDePruebas(t)
	seccionID := SembrarSeccion(t, pool, "3.°", "A")
	servicio := academico.NuevoServicio(pool)
	repositorio := academico.NuevoRepositorioConfiguracion(pool)
	ctx := context.Background()

	seccion, err := servicio.BuscarSeccionPorID(ctx, seccionID)
	require.NoError(t, err)
	assert.Equal(t, academico.TurnoManana, seccion.Turno)

	hoy := fechaDeHoy()
	manana := hoy.AddDate(0, 0, 1)

	_, err = repositorio.CrearTurnoSeccion(ctx, seccionID, academico.TurnoTarde, manana)
	require.NoError(t, err)

	seccion, err = servicio.BuscarSeccionPorID(ctx, seccionID)
	require.NoError(t, err)
	assert.Equal(t, academico.TurnoManana, seccion.Turno,
		"un cambio futuro no debe aplicarse todavía")

	_, err = repositorio.CrearTurnoSeccion(ctx, seccionID, academico.TurnoTarde, hoy)
	require.NoError(t, err)

	seccion, err = servicio.BuscarSeccionPorID(ctx, seccionID)
	require.NoError(t, err)
	assert.Equal(t, academico.TurnoTarde, seccion.Turno)

	secciones, err := servicio.ListarSeccionesVigentes(ctx)
	require.NoError(t, err)
	require.Len(t, secciones, 1)
	assert.Equal(t, academico.TurnoTarde, secciones[0].Turno)

	turnoAyer, err := repositorio.TurnoVigenteSeccion(ctx, seccionID, hoy.AddDate(0, 0, -1))
	require.NoError(t, err)
	assert.Equal(t, academico.TurnoManana, turnoAyer)

	turnoHoy, err := repositorio.TurnoVigenteSeccion(ctx, seccionID, hoy)
	require.NoError(t, err)
	assert.Equal(t, academico.TurnoTarde, turnoHoy)
}

func fechaDeHoy() time.Time {
	ahora := time.Now().UTC()
	return time.Date(ahora.Year(), ahora.Month(), ahora.Day(), 0, 0, 0, 0, time.UTC)
}
