package integracion

import (
	"context"
	"testing"
	"time"

	"asistencia/internal/academico"
	"asistencia/internal/alumnos"
	"asistencia/internal/asistencia"
	"asistencia/internal/plataforma/fechas"
	"asistencia/internal/porteria"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPorteria_RegistraIngresoYDisparaAlerta(t *testing.T) {
	fuentes := nuevasFuentesPorteria()
	servicio := porteria.NuevoServicio(porteria.Dependencias{
		Alumnos:     fuentes.alumnos,
		Academico:   fuentes.academico,
		Marcaciones: fuentes.marcaciones,
		Alertas:     fuentes.alertas,
		Reloj:       func() time.Time { return fechaHoraLima(7) },
	})

	resultado, err := servicio.Registrar(context.Background(), 9, "token-valido")
	require.NoError(t, err)
	assert.True(t, resultado.Registrado)
	assert.Equal(t, "Tardanza", resultado.Estado)
	assert.Equal(t, "Perez, Ana", resultado.Alumno)
	assert.Equal(t, 1, fuentes.marcaciones.llamadas)
	assert.Equal(t, int64(9), fuentes.marcaciones.usuarioID)
	assert.Equal(t, asistencia.MarcacionTardanza, fuentes.marcaciones.estado)
	assert.Equal(t, 1, fuentes.alertas.llamadas)
}

func TestPorteria_RechazaTokenInvalidoYDuplicado(t *testing.T) {
	fuentes := nuevasFuentesPorteria()
	fuentes.alumnos.error = alumnos.ErrAlumnoNoEncontrado
	servicio := porteria.NuevoServicio(porteria.Dependencias{
		Alumnos:     fuentes.alumnos,
		Academico:   fuentes.academico,
		Marcaciones: fuentes.marcaciones,
		Alertas:     fuentes.alertas,
		Reloj:       func() time.Time { return fechaHoraLima(7) },
	})
	_, err := servicio.Registrar(context.Background(), 9, "no-existe")
	assert.ErrorIs(t, err, porteria.ErrTokenInvalido)

	fuentes = nuevasFuentesPorteria()
	fuentes.marcaciones.creado = false
	servicio = porteria.NuevoServicio(porteria.Dependencias{
		Alumnos:     fuentes.alumnos,
		Academico:   fuentes.academico,
		Marcaciones: fuentes.marcaciones,
		Alertas:     fuentes.alertas,
		Reloj:       func() time.Time { return fechaHoraLima(7) },
	})
	_, err = servicio.Registrar(context.Background(), 9, "token-valido")
	assert.ErrorIs(t, err, porteria.ErrIngresoDuplicado)
	assert.Equal(t, 0, fuentes.alertas.llamadas)
}

func TestPorteria_RechazaFueraDeHorario(t *testing.T) {
	fuentes := nuevasFuentesPorteria()
	servicio := porteria.NuevoServicio(porteria.Dependencias{
		Alumnos:     fuentes.alumnos,
		Academico:   fuentes.academico,
		Marcaciones: fuentes.marcaciones,
		Alertas:     fuentes.alertas,
		Reloj:       func() time.Time { return fechaHoraLima(8) },
	})
	_, err := servicio.Registrar(context.Background(), 9, "token-valido")
	assert.ErrorIs(t, err, porteria.ErrRegistroCerrado)
	assert.Equal(t, 0, fuentes.marcaciones.llamadas)
}

func TestPorteria_RechazaDiaNoLaborable(t *testing.T) {
	fuentes := nuevasFuentesPorteria()
	fuentes.academico.noLaborable = true
	servicio := porteria.NuevoServicio(porteria.Dependencias{
		Alumnos:     fuentes.alumnos,
		Academico:   fuentes.academico,
		Marcaciones: fuentes.marcaciones,
		Alertas:     fuentes.alertas,
		Reloj:       func() time.Time { return fechaHoraLima(7) },
	})
	_, err := servicio.Registrar(context.Background(), 9, "token-valido")
	assert.ErrorIs(t, err, porteria.ErrDiaNoLaborable)
}

func fechaHoraLima(hora int) time.Time {
	return time.Date(2026, time.March, 10, hora+5, 45, 0, 0, time.UTC)
}

type fuentesPorteria struct {
	alumnos     *alumnosPorteriaFalsos
	academico   *academicoPorteriaFalso
	marcaciones *marcacionesPorteriaFalsas
	alertas     *alertasPorteriaFalsas
}

func nuevasFuentesPorteria() fuentesPorteria {
	return fuentesPorteria{
		alumnos:     &alumnosPorteriaFalsos{},
		academico:   &academicoPorteriaFalso{},
		marcaciones: &marcacionesPorteriaFalsas{creado: true},
		alertas:     &alertasPorteriaFalsas{},
	}
}

type alumnosPorteriaFalsos struct {
	error error
}

func (a *alumnosPorteriaFalsos) AlumnoPorToken(context.Context, string, time.Time) (alumnos.Alumno, error) {
	if a.error != nil {
		return alumnos.Alumno{}, a.error
	}
	return alumnos.Alumno{ID: 5, Identificador: "A001", Nombres: "Ana", Apellidos: "Perez", SeccionID: 2}, nil
}

type academicoPorteriaFalso struct {
	noLaborable bool
}

func (a *academicoPorteriaFalso) TurnoDeSeccion(context.Context, int64, time.Time) (string, error) {
	return academico.TurnoManana, nil
}

func (a *academicoPorteriaFalso) HorarioVigente(context.Context, string, time.Time) (academico.HorarioTurno, error) {
	ingreso, _ := fechas.ParsearHora("07:30")
	limite, _ := fechas.ParsearHora("08:00")
	cierre, _ := fechas.ParsearHora("08:30")
	return academico.HorarioTurno{
		HoraIngreso:        ingreso,
		ToleranciaMinutos:  10,
		HoraLimiteTardanza: limite,
		HoraCierre:         cierre,
	}, nil
}

func (a *academicoPorteriaFalso) EsDiaNoLaborable(context.Context, int64, time.Time) (bool, error) {
	return a.noLaborable, nil
}

func (a *academicoPorteriaFalso) EtiquetaSeccion(context.Context, int64) (string, error) {
	return "1.° A", nil
}

type marcacionesPorteriaFalsas struct {
	creado    bool
	llamadas  int
	usuarioID int64
	estado    string
}

func (m *marcacionesPorteriaFalsas) RegistrarMarcacionPorteria(_ context.Context, usuarioID, _ int64, _, _ time.Time, estado string) (int64, bool, error) {
	m.llamadas++
	m.usuarioID = usuarioID
	m.estado = estado
	if !m.creado {
		return 0, false, nil
	}
	return 1, true, nil
}

type alertasPorteriaFalsas struct {
	llamadas int
}

func (a *alertasPorteriaFalsas) ProcesarTardanzas(context.Context) error {
	a.llamadas++
	return nil
}
