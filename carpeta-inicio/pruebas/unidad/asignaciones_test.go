package unidad

import (
	"testing"
	"time"

	"asistencia/internal/asistencia"

	"github.com/stretchr/testify/assert"
)

func TestValidarFechasAsignacion_RechazaFinAnterior(t *testing.T) {
	inicio := time.Date(2026, time.November, 10, 0, 0, 0, 0, time.UTC)
	finIgual := inicio
	finAnterior := inicio.AddDate(0, 0, -1)
	finPosterior := inicio.AddDate(0, 0, 5)

	assert.NoError(t, asistencia.ValidarFechasAsignacion(inicio, nil))
	assert.NoError(t, asistencia.ValidarFechasAsignacion(inicio, &finIgual))
	assert.NoError(t, asistencia.ValidarFechasAsignacion(inicio, &finPosterior))
	assert.ErrorIs(t, asistencia.ValidarFechasAsignacion(inicio, &finAnterior),
		asistencia.ErrAsignacionFueraDeRango)
}

func TestAsignacionVigenteEnFecha_IncluyeLosBordes(t *testing.T) {
	inicio := time.Date(2026, time.November, 10, 0, 0, 0, 0, time.UTC)
	fin := time.Date(2026, time.November, 12, 0, 0, 0, 0, time.UTC)
	asignacion := asistencia.AsignacionDocente{FechaInicio: inicio, FechaFin: &fin}

	assert.False(t, asistencia.AsignacionVigenteEnFecha(asignacion, inicio.AddDate(0, 0, -1)))
	assert.True(t, asistencia.AsignacionVigenteEnFecha(asignacion, inicio))
	assert.True(t, asistencia.AsignacionVigenteEnFecha(asignacion, fin))
	assert.False(t, asistencia.AsignacionVigenteEnFecha(asignacion, fin.AddDate(0, 0, 1)))

	asignacion.FechaFin = nil
	assert.True(t, asistencia.AsignacionVigenteEnFecha(asignacion, fin.AddDate(0, 0, 100)))
}
