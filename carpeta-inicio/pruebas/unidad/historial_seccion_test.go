package unidad

import (
	"testing"
	"time"

	"asistencia/internal/alumnos"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fechaDePrueba(dia int) time.Time {
	return time.Date(2026, time.May, dia, 0, 0, 0, 0, time.UTC)
}

func historialAbierto() *alumnos.HistorialSeccion {
	return &alumnos.HistorialSeccion{
		ID:          1,
		SeccionID:   10,
		Grado:       "3.°",
		Seccion:     "A",
		FechaInicio: fechaDePrueba(1),
	}
}

func TestPrepararCambioSeccion_CierraElTramoActualYAbreUnoNuevo(t *testing.T) {
	actual := historialAbierto()
	fecha := fechaDePrueba(15)

	cambio, err := alumnos.PrepararCambioSeccion(actual, 20, fecha)
	require.NoError(t, err)

	assert.Equal(t, int64(10), cambio.SeccionAnterior)
	assert.Equal(t, int64(20), cambio.SeccionNueva)
	assert.Equal(t, fecha, cambio.FechaCierre)
	assert.Equal(t, fecha, cambio.FechaApertura)
}

func TestPrepararCambioSeccion_SinHistorialPrevioAbreElPrimerTramo(t *testing.T) {
	fecha := fechaDePrueba(10)

	cambio, err := alumnos.PrepararCambioSeccion(nil, 20, fecha)
	require.NoError(t, err)

	assert.Equal(t, int64(0), cambio.SeccionAnterior)
	assert.Equal(t, int64(20), cambio.SeccionNueva)
	assert.Equal(t, fecha, cambio.FechaCierre)
	assert.Equal(t, fecha, cambio.FechaApertura)
}

func TestPrepararCambioSeccion_RechazaLaMismaSeccion(t *testing.T) {
	_, err := alumnos.PrepararCambioSeccion(historialAbierto(), 10, fechaDePrueba(15))

	assert.ErrorIs(t, err, alumnos.ErrMismaSeccion)
}

func TestPrepararCambioSeccion_RechazaFechaAnteriorAlTramoActual(t *testing.T) {
	_, err := alumnos.PrepararCambioSeccion(historialAbierto(), 20, fechaDePrueba(1).AddDate(0, 0, -1))

	assert.ErrorIs(t, err, alumnos.ErrFechaInvalida)
}

func TestPrepararCambioSeccion_RechazaTramoYaCerrado(t *testing.T) {
	actual := historialAbierto()
	cierre := fechaDePrueba(5)
	actual.FechaFin = &cierre

	_, err := alumnos.PrepararCambioSeccion(actual, 20, fechaDePrueba(15))

	assert.ErrorIs(t, err, alumnos.ErrHistorialYaCerrado)
}

func TestPrepararCambioSeccion_RechazaSeccionYFechaInvalidas(t *testing.T) {
	_, err := alumnos.PrepararCambioSeccion(historialAbierto(), 0, fechaDePrueba(15))
	assert.ErrorIs(t, err, alumnos.ErrSeccionInvalida)

	_, err = alumnos.PrepararCambioSeccion(historialAbierto(), 20, time.Time{})
	assert.ErrorIs(t, err, alumnos.ErrFechaInvalida)
}

func TestTipoSalidaValido_SoloRetiradoYTrasladado(t *testing.T) {
	assert.True(t, alumnos.TipoSalidaValido(alumnos.EstadoRetirado))
	assert.True(t, alumnos.TipoSalidaValido(alumnos.EstadoTrasladado))

	assert.False(t, alumnos.TipoSalidaValido(alumnos.EstadoActivo))
	assert.False(t, alumnos.TipoSalidaValido(alumnos.EstadoDesactivado))
	assert.False(t, alumnos.TipoSalidaValido(""))
	assert.False(t, alumnos.TipoSalidaValido("RETIRADO"))
}
