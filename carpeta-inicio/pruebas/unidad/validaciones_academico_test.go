package unidad

import (
	"testing"
	"time"

	"asistencia/internal/academico"
	"asistencia/internal/plataforma/fechas"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func hora(t *testing.T, texto string) time.Time {
	t.Helper()
	valor, err := fechas.ParsearHora(texto)
	require.NoError(t, err)
	return valor
}

func horarioValido(t *testing.T) academico.HorarioTurno {
	t.Helper()
	return academico.HorarioTurno{
		Turno:              academico.TurnoManana,
		HoraIngreso:        hora(t, "07:30"),
		ToleranciaMinutos:  10,
		HoraLimiteTardanza: hora(t, "08:00"),
		HoraCierre:         hora(t, "08:30"),
	}
}

func TestValidarHorarioTurno_AceptaHorarioValido(t *testing.T) {
	assert.NoError(t, academico.ValidarHorarioTurno(horarioValido(t)))
}

func TestValidarHorarioTurno_RechazaCombinacionesInvalidas(t *testing.T) {
	casos := []struct {
		nombre    string
		modificar func(*academico.HorarioTurno)
		esperado  error
	}{
		{"cierre no posterior al limite", func(h *academico.HorarioTurno) {
			h.HoraCierre = hora(t, "08:00")
		}, academico.ErrCierreNoPosterior},
		{"tolerancia excedida", func(h *academico.HorarioTurno) {
			h.ToleranciaMinutos = 40
		}, academico.ErrToleranciaExcedida},
		{"tolerancia negativa", func(h *academico.HorarioTurno) {
			h.ToleranciaMinutos = -1
		}, academico.ErrToleranciaInvalida},
		{"limite antes del ingreso", func(h *academico.HorarioTurno) {
			h.HoraLimiteTardanza = hora(t, "07:20")
			h.HoraCierre = hora(t, "07:40")
		}, academico.ErrHorarioFueraDeOrden},
		{"cruce de medianoche", func(h *academico.HorarioTurno) {
			h.HoraIngreso = hora(t, "23:40")
			h.HoraLimiteTardanza = hora(t, "23:50")
			h.HoraCierre = hora(t, "23:59")
			h.ToleranciaMinutos = 30
		}, academico.ErrCruzaMedianoche},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			horario := horarioValido(t)
			caso.modificar(&horario)
			assert.ErrorIs(t, academico.ValidarHorarioTurno(horario), caso.esperado)
		})
	}
}

func TestValidarSuperposicionHorarios_DetectaCruceEntreTurnos(t *testing.T) {
	manana := horarioValido(t)
	manana.Turno = academico.TurnoManana
	manana.HoraCierre = hora(t, "12:30")

	tarde := horarioValido(t)
	tarde.Turno = academico.TurnoTarde
	tarde.HoraIngreso = hora(t, "12:00")
	tarde.HoraLimiteTardanza = hora(t, "12:30")
	tarde.HoraCierre = hora(t, "13:00")

	assert.ErrorIs(t, academico.ValidarSuperposicionHorarios(manana, tarde),
		academico.ErrHorariosSuperpuestos)

	tarde.HoraIngreso = hora(t, "12:45")
	tarde.HoraLimiteTardanza = hora(t, "13:15")
	tarde.HoraCierre = hora(t, "13:45")
	assert.NoError(t, academico.ValidarSuperposicionHorarios(manana, tarde))
}

func TestValidarVigencia_RechazaFechasPasadas(t *testing.T) {
	hoy := time.Date(2026, time.November, 6, 0, 0, 0, 0, time.UTC)

	assert.NoError(t, academico.ValidarVigencia(hoy, hoy))
	assert.NoError(t, academico.ValidarVigencia(hoy.AddDate(0, 0, 1), hoy))
	assert.ErrorIs(t, academico.ValidarVigencia(hoy.AddDate(0, 0, -1), hoy),
		academico.ErrVigenciaPasada)
}

func TestValidarFechasAnio_RechazaFinAnteriorOIgual(t *testing.T) {
	inicio := time.Date(2027, time.March, 1, 0, 0, 0, 0, time.UTC)

	assert.NoError(t, academico.ValidarFechasAnio(inicio, inicio.AddDate(0, 9, 0)))
	assert.ErrorIs(t, academico.ValidarFechasAnio(inicio, inicio), academico.ErrFechasAnioInvalidas)
	assert.ErrorIs(t, academico.ValidarFechasAnio(inicio, inicio.AddDate(0, -1, 0)),
		academico.ErrFechasAnioInvalidas)
}

func bloqueValido(t *testing.T) academico.BloqueClase {
	t.Helper()
	return academico.BloqueClase{
		Turno:      academico.TurnoManana,
		Numero:     1,
		HoraInicio: hora(t, "07:45"),
		HoraFin:    hora(t, "08:30"),
	}
}

func TestValidarBloque_AceptaBloqueDentroDelTurno(t *testing.T) {
	horario := horarioValido(t)
	horario.HoraCierre = hora(t, "12:30")

	assert.NoError(t, academico.ValidarBloque(bloqueValido(t), horario, nil))
}

func TestValidarBloque_RechazaBloquesInvalidos(t *testing.T) {
	horario := horarioValido(t)
	horario.HoraCierre = hora(t, "12:30")

	casos := []struct {
		nombre    string
		modificar func(*academico.BloqueClase)
		otros     []academico.BloqueClase
		esperado  error
	}{
		{"fin anterior al inicio", func(b *academico.BloqueClase) {
			b.HoraFin = hora(t, "07:40")
		}, nil, academico.ErrBloqueHorasInvalidas},
		{"fuera del horario del turno", func(b *academico.BloqueClase) {
			b.HoraInicio = hora(t, "12:35")
			b.HoraFin = hora(t, "13:00")
		}, nil, academico.ErrBloqueFueraDeHorario},
		{"numero fuera de rango", func(b *academico.BloqueClase) {
			b.Numero = 0
		}, nil, academico.ErrBloqueNumeroInvalido},
		{"cruce de medianoche", func(b *academico.BloqueClase) {
			b.HoraInicio = hora(t, "23:50")
			b.HoraFin = fechas.SumarMinutos(hora(t, "23:50"), 20)
		}, nil, academico.ErrCruzaMedianoche},
		{"solapamiento con otro bloque", func(b *academico.BloqueClase) {}, []academico.BloqueClase{{
			Turno:      academico.TurnoManana,
			Numero:     2,
			HoraInicio: hora(t, "08:20"),
			HoraFin:    hora(t, "09:05"),
		}}, academico.ErrBloqueSolapado},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			bloque := bloqueValido(t)
			caso.modificar(&bloque)
			assert.ErrorIs(t, academico.ValidarBloque(bloque, horario, caso.otros), caso.esperado)
		})
	}
}

func TestValidarBloque_IgnoraOtraVersionDelMismoNumero(t *testing.T) {
	horario := horarioValido(t)
	horario.HoraCierre = hora(t, "12:30")

	otros := []academico.BloqueClase{{
		Turno:      academico.TurnoManana,
		Numero:     1,
		HoraInicio: hora(t, "07:45"),
		HoraFin:    hora(t, "08:30"),
	}}

	assert.NoError(t, academico.ValidarBloque(bloqueValido(t), horario, otros))
}

func TestValidarUmbralYExcepcion(t *testing.T) {
	assert.NoError(t, academico.ValidarUmbral(academico.TipoUmbralTardanzas, 3))
	assert.ErrorIs(t, academico.ValidarUmbral(academico.TipoUmbralTardanzas, 0),
		academico.ErrValorUmbralInvalido)
	assert.ErrorIs(t, academico.ValidarUmbral("otro", 3), academico.ErrTipoInvalido)

	inicio := time.Date(2027, time.July, 28, 0, 0, 0, 0, time.UTC)
	assert.NoError(t, academico.ValidarExcepcion(academico.TipoFeriado, "Fiestas Patrias", inicio, inicio))
	assert.ErrorIs(t, academico.ValidarExcepcion(academico.TipoFeriado, "Fiestas Patrias", inicio, inicio.AddDate(0, 0, -1)),
		academico.ErrRangoFechasExcepcion)
	assert.ErrorIs(t, academico.ValidarExcepcion(academico.TipoFeriado, "ok", inicio, inicio),
		academico.ErrMotivoExcepcionCorto)
}
