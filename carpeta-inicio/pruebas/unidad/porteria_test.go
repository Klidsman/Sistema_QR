package unidad

import (
	"testing"

	"asistencia/internal/academico"
	"asistencia/internal/asistencia"
	"asistencia/internal/plataforma/fechas"
	"asistencia/internal/porteria"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func horarioDePrueba() academico.HorarioTurno {
	ingreso, _ := fechas.ParsearHora("07:30")
	limite, _ := fechas.ParsearHora("08:00")
	cierre, _ := fechas.ParsearHora("08:30")
	return academico.HorarioTurno{
		HoraIngreso:        ingreso,
		ToleranciaMinutos:  10,
		HoraLimiteTardanza: limite,
		HoraCierre:         cierre,
	}
}

func TestClasificarEstado_AplicaLosLimitesDelTurno(t *testing.T) {
	casos := []struct {
		hora          string
		estado        string
		errorEsperado error
	}{
		{"07:20", "", porteria.ErrAntesDeIngreso},
		{"07:30", asistencia.MarcacionPuntual, nil},
		{"07:40", asistencia.MarcacionPuntual, nil},
		{"07:41", asistencia.MarcacionTardanza, nil},
		{"08:00", asistencia.MarcacionTardanza, nil},
		{"08:01", asistencia.MarcacionTardanzaGrave, nil},
		{"08:30", asistencia.MarcacionTardanzaGrave, nil},
		{"08:31", "", porteria.ErrRegistroCerrado},
	}

	for _, caso := range casos {
		t.Run(caso.hora, func(t *testing.T) {
			hora, err := fechas.ParsearHora(caso.hora)
			require.NoError(t, err)
			estado, err := porteria.ClasificarEstado(hora, horarioDePrueba())
			if caso.errorEsperado != nil {
				assert.ErrorIs(t, err, caso.errorEsperado)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, caso.estado, estado)
		})
	}
}

func TestEtiquetaEstado_TextosVisibles(t *testing.T) {
	assert.Equal(t, "Puntual", porteria.EtiquetaEstado(asistencia.MarcacionPuntual))
	assert.Equal(t, "Tardanza", porteria.EtiquetaEstado(asistencia.MarcacionTardanza))
	assert.Equal(t, "Tardanza grave", porteria.EtiquetaEstado(asistencia.MarcacionTardanzaGrave))
}
