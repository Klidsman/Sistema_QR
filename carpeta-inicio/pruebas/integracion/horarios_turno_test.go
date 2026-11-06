package integracion

import (
	"context"
	"testing"
	"time"

	"asistencia/internal/academico"
	"asistencia/internal/plataforma/fechas"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHorariosTurno_VersionaYConsultaPorFecha(t *testing.T) {
	pool := AbrirBaseDePruebas(t)
	usuarioID := CrearUsuarioPrueba(t, pool)
	servicio := academico.NuevoServicioConfiguracion(pool)
	ctx := context.Background()
	hoy := fechas.Hoy()

	anio := time.Now().Year() + 1
	anioID, err := servicio.CrearAnio(ctx, usuarioID, academico.DatosAnio{
		Nombre:      time.Date(anio, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006"),
		FechaInicio: time.Date(anio, time.March, 1, 0, 0, 0, 0, time.UTC),
		FechaFin:    time.Date(anio, time.December, 15, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)
	_ = anioID

	manana := academico.HorarioTurno{
		Turno:              academico.TurnoManana,
		HoraIngreso:        horaDePrueba(t, "07:30"),
		ToleranciaMinutos:  10,
		HoraLimiteTardanza: horaDePrueba(t, "08:00"),
		HoraCierre:         horaDePrueba(t, "08:30"),
		VigenteDesde:       hoy,
	}
	require.NoError(t, servicio.ConfigurarHorario(ctx, usuarioID, manana))

	vigente, err := servicio.HorarioVigente(ctx, academico.TurnoManana, hoy)
	require.NoError(t, err)
	assert.Equal(t, 10, vigente.ToleranciaMinutos)
	assert.Equal(t, "07:30", fechas.FormatearHora(vigente.HoraIngreso))

	futuro := hoy.AddDate(0, 1, 0)
	mananaFuturo := manana
	mananaFuturo.ToleranciaMinutos = 5
	mananaFuturo.VigenteDesde = futuro
	require.NoError(t, servicio.ConfigurarHorario(ctx, usuarioID, mananaFuturo))

	vigente, err = servicio.HorarioVigente(ctx, academico.TurnoManana, hoy)
	require.NoError(t, err)
	assert.Equal(t, 10, vigente.ToleranciaMinutos, "el cambio futuro no debe aplicarse antes de su fecha")

	vigente, err = servicio.HorarioVigente(ctx, academico.TurnoManana, futuro)
	require.NoError(t, err)
	assert.Equal(t, 5, vigente.ToleranciaMinutos)

	historial, err := servicio.ListarHorarios(ctx, academico.TurnoManana)
	require.NoError(t, err)
	assert.Len(t, historial, 2)

	tardeSolapada := academico.HorarioTurno{
		Turno:              academico.TurnoTarde,
		HoraIngreso:        horaDePrueba(t, "08:00"),
		ToleranciaMinutos:  10,
		HoraLimiteTardanza: horaDePrueba(t, "08:15"),
		HoraCierre:         horaDePrueba(t, "08:30"),
		VigenteDesde:       hoy,
	}
	assert.ErrorIs(t, servicio.ConfigurarHorario(ctx, usuarioID, tardeSolapada),
		academico.ErrHorariosSuperpuestos)

	tarde := tardeSolapada
	tarde.HoraIngreso = horaDePrueba(t, "12:45")
	tarde.HoraLimiteTardanza = horaDePrueba(t, "13:15")
	tarde.HoraCierre = horaDePrueba(t, "13:45")
	require.NoError(t, servicio.ConfigurarHorario(ctx, usuarioID, tarde))

	mananaCorregido := manana
	mananaCorregido.ToleranciaMinutos = 15
	require.NoError(t, servicio.ConfigurarHorario(ctx, usuarioID, mananaCorregido),
		"un cambio con la misma fecha debe reemplazar la versión")

	vigente, err = servicio.HorarioVigente(ctx, academico.TurnoManana, hoy)
	require.NoError(t, err)
	assert.Equal(t, 15, vigente.ToleranciaMinutos)
	historial, err = servicio.ListarHorarios(ctx, academico.TurnoManana)
	require.NoError(t, err)
	assert.Len(t, historial, 2, "el reemplazo no debe duplicar la versión del mismo día")

	pasado := manana
	pasado.VigenteDesde = hoy.AddDate(0, 0, -1)
	assert.ErrorIs(t, servicio.ConfigurarHorario(ctx, usuarioID, pasado), academico.ErrVigenciaPasada)

	var auditorias int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM auditoria WHERE entidad = $1 AND accion = $2`,
		academico.EntidadTurno, academico.AccionConfigurarHorario,
	).Scan(&auditorias))
	assert.GreaterOrEqual(t, auditorias, 4)
}

func horaDePrueba(t *testing.T, texto string) time.Time {
	t.Helper()
	valor, err := fechas.ParsearHora(texto)
	require.NoError(t, err)
	return valor
}
