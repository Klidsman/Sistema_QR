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

func TestCalendario_RegistraExcepcionesPorAlcanceYDetectaSolapamientos(t *testing.T) {
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
	grado3ID, err := servicio.CrearGrado(ctx, usuarioID, academico.DatosGrado{
		AnioEscolarID: anioID, Nombre: "3.°",
	})
	require.NoError(t, err)
	grado4ID, err := servicio.CrearGrado(ctx, usuarioID, academico.DatosGrado{
		AnioEscolarID: anioID, Nombre: "4.°",
	})
	require.NoError(t, err)
	seccion3ID, err := servicio.CrearSeccion(ctx, usuarioID, academico.DatosSeccion{
		GradoID: grado3ID, Nombre: "A", Turno: academico.TurnoManana, VigenteDesde: hoy,
	})
	require.NoError(t, err)
	seccion4ID, err := servicio.CrearSeccion(ctx, usuarioID, academico.DatosSeccion{
		GradoID: grado4ID, Nombre: "A", Turno: academico.TurnoTarde, VigenteDesde: hoy,
	})
	require.NoError(t, err)

	feriadoID, err := servicio.RegistrarExcepcion(ctx, usuarioID, academico.DatosExcepcion{
		AnioEscolarID: anioID,
		Tipo:          academico.TipoFeriado,
		FechaInicio:   hoy,
		FechaFin:      hoy,
		Motivo:        "Fiestas Patrias",
	}, false)
	require.NoError(t, err)
	require.Positive(t, feriadoID)

	vigentes, err := servicio.ExcepcionesVigentes(ctx, anioID, hoy)
	require.NoError(t, err)
	require.Len(t, vigentes, 1)
	assert.Equal(t, academico.TipoFeriado, vigentes[0].Tipo)

	alcance3, err := servicio.ExcepcionesDeSeccion(ctx, seccion3ID, hoy)
	require.NoError(t, err)
	require.Len(t, alcance3, 1)

	diasHastaSabado := (6 - int(hoy.Weekday()) + 7) % 7
	if diasHastaSabado == 0 {
		diasHastaSabado = 7
	}
	sabado := hoy.AddDate(0, 0, diasHastaSabado)
	_, err = servicio.RegistrarExcepcion(ctx, usuarioID, academico.DatosExcepcion{
		AnioEscolarID: anioID,
		Tipo:          academico.TipoNoLaborable,
		FechaInicio:   sabado,
		FechaFin:      sabado,
		Motivo:        "Actividad interna",
	}, false)
	require.NoError(t, err)

	vigentes, err = servicio.ExcepcionesVigentes(ctx, anioID, sabado)
	require.NoError(t, err)
	assert.Len(t, vigentes, 1)

	suspensionFecha := hoy.AddDate(0, 0, 14)
	_, err = servicio.RegistrarExcepcion(ctx, usuarioID, academico.DatosExcepcion{
		AnioEscolarID: anioID,
		Tipo:          academico.TipoSuspension,
		FechaInicio:   suspensionFecha,
		FechaFin:      suspensionFecha.AddDate(0, 0, 2),
		Motivo:        "Suspensión de clases",
		Grados:        []int64{grado3ID},
	}, false)
	require.NoError(t, err)

	alcance3, err = servicio.ExcepcionesDeSeccion(ctx, seccion3ID, suspensionFecha)
	require.NoError(t, err)
	require.Len(t, alcance3, 1)

	alcance4, err := servicio.ExcepcionesDeSeccion(ctx, seccion4ID, suspensionFecha)
	require.NoError(t, err)
	assert.Empty(t, alcance4)

	_, err = servicio.RegistrarExcepcion(ctx, usuarioID, academico.DatosExcepcion{
		AnioEscolarID: anioID,
		Tipo:          academico.TipoSuspension,
		FechaInicio:   hoy,
		FechaFin:      hoy,
		Motivo:        "Otra suspensión global",
	}, false)
	assert.ErrorIs(t, err, academico.ErrExcepcionSolapada)

	_, err = servicio.RegistrarExcepcion(ctx, usuarioID, academico.DatosExcepcion{
		AnioEscolarID: anioID,
		Tipo:          academico.TipoSuspension,
		FechaInicio:   hoy,
		FechaFin:      hoy,
		Motivo:        "Suspensión confirmada",
	}, true)
	require.NoError(t, err)

	_, err = servicio.RegistrarExcepcion(ctx, usuarioID, academico.DatosExcepcion{
		AnioEscolarID: anioID,
		Tipo:          academico.TipoSuspension,
		FechaInicio:   hoy,
		FechaFin:      hoy,
		Motivo:        "Solo cuarto",
		Grados:        []int64{grado4ID},
	}, false)
	require.NoError(t, err, "una excepción por grado no se superpone con una global")

	_, err = servicio.RegistrarExcepcion(ctx, usuarioID, academico.DatosExcepcion{
		AnioEscolarID: anioID,
		Tipo:          academico.TipoFeriado,
		FechaInicio:   hoy.AddDate(0, 0, 14),
		FechaFin:      hoy.AddDate(0, 0, 21),
		Motivo:        "Grado ajeno",
		Grados:        []int64{9999},
	}, false)
	assert.ErrorIs(t, err, academico.ErrGradoNoEncontrado)

	var auditorias int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM auditoria WHERE entidad = $1 AND accion = $2`,
		academico.EntidadCalendario, academico.AccionRegistrarExcepcion,
	).Scan(&auditorias))
	assert.GreaterOrEqual(t, auditorias, 4)
}
