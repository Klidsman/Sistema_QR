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

func TestBloquesClase_VersionaValidaYAudita(t *testing.T) {
	pool := AbrirBaseDePruebas(t)
	usuarioID := CrearUsuarioPrueba(t, pool)
	servicio := academico.NuevoServicioConfiguracion(pool)
	ctx := context.Background()
	hoy := fechas.Hoy()

	anio := time.Now().Year() + 1
	_, err := servicio.CrearAnio(ctx, usuarioID, academico.DatosAnio{
		Nombre:      time.Date(anio, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006"),
		FechaInicio: time.Date(anio, time.March, 1, 0, 0, 0, 0, time.UTC),
		FechaFin:    time.Date(anio, time.December, 15, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)

	manana := academico.HorarioTurno{
		Turno:              academico.TurnoManana,
		HoraIngreso:        horaDePrueba(t, "07:30"),
		ToleranciaMinutos:  10,
		HoraLimiteTardanza: horaDePrueba(t, "08:00"),
		HoraCierre:         horaDePrueba(t, "12:30"),
		VigenteDesde:       hoy,
	}
	require.NoError(t, servicio.ConfigurarHorario(ctx, usuarioID, manana))

	bloque1 := academico.BloqueClase{
		Turno:        academico.TurnoManana,
		Numero:       1,
		HoraInicio:   horaDePrueba(t, "07:45"),
		HoraFin:      horaDePrueba(t, "08:30"),
		VigenteDesde: hoy,
	}
	bloque1ID, err := servicio.CrearBloque(ctx, usuarioID, bloque1)
	require.NoError(t, err)

	vigentes, err := servicio.BloquesVigentes(ctx, academico.TurnoManana, hoy)
	require.NoError(t, err)
	require.Len(t, vigentes, 1)
	assert.Equal(t, 1, vigentes[0].Numero)

	solapado := bloque1
	solapado.Numero = 2
	solapado.HoraInicio = horaDePrueba(t, "08:20")
	solapado.HoraFin = horaDePrueba(t, "09:05")
	_, err = servicio.CrearBloque(ctx, usuarioID, solapado)
	assert.ErrorIs(t, err, academico.ErrBloqueSolapado)

	fuera := bloque1
	fuera.Numero = 2
	fuera.HoraInicio = horaDePrueba(t, "12:35")
	fuera.HoraFin = horaDePrueba(t, "13:00")
	_, err = servicio.CrearBloque(ctx, usuarioID, fuera)
	assert.ErrorIs(t, err, academico.ErrBloqueFueraDeHorario)

	bloque2 := bloque1
	bloque2.Numero = 2
	bloque2.HoraInicio = horaDePrueba(t, "08:30")
	bloque2.HoraFin = horaDePrueba(t, "09:15")
	_, err = servicio.CrearBloque(ctx, usuarioID, bloque2)
	require.NoError(t, err)

	futuro := hoy.AddDate(0, 1, 0)
	require.NoError(t, servicio.AjustarHorasBloque(ctx, usuarioID, bloque1ID,
		horaDePrueba(t, "07:45"), horaDePrueba(t, "08:20"), futuro))

	vigentes, err = servicio.BloquesVigentes(ctx, academico.TurnoManana, hoy)
	require.NoError(t, err)
	require.Len(t, vigentes, 2)
	assert.Equal(t, "08:30", fechas.FormatearHora(vigentes[0].HoraFin))

	vigentes, err = servicio.BloquesVigentes(ctx, academico.TurnoManana, futuro)
	require.NoError(t, err)
	require.Len(t, vigentes, 2)
	assert.Equal(t, "08:20", fechas.FormatearHora(vigentes[0].HoraFin))

	sinHorario := bloque1
	sinHorario.Turno = academico.TurnoTarde
	sinHorario.Numero = 1
	_, err = servicio.CrearBloque(ctx, usuarioID, sinHorario)
	assert.ErrorIs(t, err, academico.ErrHorarioNoEncontrado)

	pasado := bloque1
	pasado.VigenteDesde = hoy.AddDate(0, 0, -1)
	_, err = servicio.CrearBloque(ctx, usuarioID, pasado)
	assert.ErrorIs(t, err, academico.ErrVigenciaPasada)

	var creados, ajustados int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM auditoria WHERE entidad = $1 AND accion = $2`,
		academico.EntidadBloqueClase, academico.AccionCrearBloque).Scan(&creados))
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM auditoria WHERE entidad = $1 AND accion = $2`,
		academico.EntidadBloqueClase, academico.AccionAjustarBloque).Scan(&ajustados))
	assert.GreaterOrEqual(t, creados, 2)
	assert.Equal(t, 1, ajustados)
}
