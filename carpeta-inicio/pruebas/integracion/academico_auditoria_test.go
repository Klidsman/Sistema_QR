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

func TestAuditoria_RegistraTodasLasAccionesDeConfiguracion(t *testing.T) {
	pool := AbrirBaseDePruebas(t)
	usuarioID := CrearUsuarioPrueba(t, pool)
	servicio := academico.NuevoServicioConfiguracion(pool)
	ctx := context.Background()
	hoy := fechas.Hoy()

	anio := time.Now().Year() + 1
	inicio := time.Date(anio, time.March, 1, 0, 0, 0, 0, time.UTC)
	anioID, err := servicio.CrearAnio(ctx, usuarioID, academico.DatosAnio{
		Nombre:      time.Date(anio, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006"),
		FechaInicio: inicio,
		FechaFin:    time.Date(anio, time.December, 15, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)
	gradoID, err := servicio.CrearGrado(ctx, usuarioID, academico.DatosGrado{
		AnioEscolarID: anioID, Nombre: "3.°",
	})
	require.NoError(t, err)
	seccionID, err := servicio.CrearSeccion(ctx, usuarioID, academico.DatosSeccion{
		GradoID: gradoID, Nombre: "A", Turno: academico.TurnoManana, VigenteDesde: hoy,
	})
	require.NoError(t, err)
	require.NoError(t, servicio.CambiarTurno(ctx, usuarioID, seccionID, academico.TurnoTarde, hoy))
	require.NoError(t, servicio.ConfigurarHorario(ctx, usuarioID, academico.HorarioTurno{
		Turno:              academico.TurnoTarde,
		HoraIngreso:        horaDePrueba(t, "12:45"),
		ToleranciaMinutos:  10,
		HoraLimiteTardanza: horaDePrueba(t, "13:15"),
		HoraCierre:         horaDePrueba(t, "17:30"),
		VigenteDesde:       hoy,
	}))
	bloqueID, err := servicio.CrearBloque(ctx, usuarioID, academico.BloqueClase{
		Turno:        academico.TurnoTarde,
		Numero:       1,
		HoraInicio:   horaDePrueba(t, "13:00"),
		HoraFin:      horaDePrueba(t, "13:45"),
		VigenteDesde: hoy,
	})
	require.NoError(t, err)
	require.NoError(t, servicio.AjustarHorasBloque(ctx, usuarioID, bloqueID,
		horaDePrueba(t, "13:00"), horaDePrueba(t, "13:40"), hoy))
	cursoID, err := servicio.CrearCurso(ctx, usuarioID, academico.DatosCurso{
		AnioEscolarID: anioID, Nombre: "Matemática", Grados: []int64{gradoID},
	})
	require.NoError(t, err)
	require.NoError(t, servicio.GuardarCeldas(ctx, usuarioID, seccionID, hoy,
		[]academico.CambioCelda{{DiaSemana: 1, BloqueNumero: 1, CursoID: &cursoID}}, true))
	require.NoError(t, servicio.DesactivarCurso(ctx, usuarioID, cursoID))
	_, err = servicio.RegistrarExcepcion(ctx, usuarioID, academico.DatosExcepcion{
		AnioEscolarID: anioID,
		Tipo:          academico.TipoFeriado,
		FechaInicio:   hoy,
		FechaFin:      hoy,
		Motivo:        "Feriado de prueba",
	}, false)
	require.NoError(t, err)
	require.NoError(t, servicio.ConfigurarUmbral(ctx, usuarioID, academico.UmbralAlerta{
		AnioEscolarID: anioID,
		Tipo:          academico.TipoUmbralTardanzas,
		Valor:         3,
		VigenteDesde:  hoy,
	}))
	require.NoError(t, servicio.CerrarAnio(ctx, usuarioID, anioID))

	esperadas := [][2]string{
		{academico.EntidadAnioEscolar, academico.AccionCrear},
		{academico.EntidadAnioEscolar, academico.AccionCerrar},
		{academico.EntidadGrado, academico.AccionCrear},
		{academico.EntidadSeccion, academico.AccionCrear},
		{academico.EntidadTurno, academico.AccionCambiarTurno},
		{academico.EntidadTurno, academico.AccionConfigurarHorario},
		{academico.EntidadBloqueClase, academico.AccionCrearBloque},
		{academico.EntidadBloqueClase, academico.AccionAjustarBloque},
		{academico.EntidadCurso, academico.AccionCrearCurso},
		{academico.EntidadCurso, academico.AccionDesactivarCurso},
		{academico.EntidadHorarioSeccion, academico.AccionAsignarHorario},
		{academico.EntidadCalendario, academico.AccionRegistrarExcepcion},
		{academico.EntidadUmbral, academico.AccionConfigurarUmbral},
	}

	for _, esperada := range esperadas {
		var cantidad int
		require.NoError(t, pool.QueryRow(ctx,
			`SELECT count(*) FROM auditoria WHERE entidad = $1 AND accion = $2 AND usuario_id = $3`,
			esperada[0], esperada[1], usuarioID,
		).Scan(&cantidad))
		assert.GreaterOrEqual(t, cantidad, 1, "falta auditoría de %s %s", esperada[0], esperada[1])
	}
}
