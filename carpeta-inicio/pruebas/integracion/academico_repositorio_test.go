package integracion

import (
	"context"
	"testing"
	"time"

	"asistencia/internal/academico"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfiguracionAcademica_CreaAnioGradosSeccionesYTurnos(t *testing.T) {
	pool := AbrirBaseDePruebas(t)
	usuarioID := CrearUsuarioPrueba(t, pool)
	servicio := academico.NuevoServicioConfiguracion(pool)
	ctx := context.Background()

	anio := time.Now().Year() + 1
	inicio := time.Date(anio, time.March, 1, 0, 0, 0, 0, time.UTC)
	fin := time.Date(anio, time.December, 15, 0, 0, 0, 0, time.UTC)

	anioID, err := servicio.CrearAnio(ctx, usuarioID, academico.DatosAnio{
		Nombre:      "  " + time.Date(anio, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006") + "  ",
		FechaInicio: inicio,
		FechaFin:    fin,
	})
	require.NoError(t, err)

	activo, err := servicio.AnioActivo(ctx)
	require.NoError(t, err)
	require.NotNil(t, activo)
	assert.Equal(t, anioID, activo.ID)
	assert.Equal(t, academico.EstadoAnioActivo, activo.Estado)

	_, err = servicio.CrearAnio(ctx, usuarioID, academico.DatosAnio{
		Nombre:      "Otro año",
		FechaInicio: inicio.AddDate(1, 0, 0),
		FechaFin:    fin.AddDate(1, 0, 0),
	})
	assert.ErrorIs(t, err, academico.ErrAnioActivoExistente)

	_, err = servicio.CrearAnio(ctx, usuarioID, academico.DatosAnio{
		Nombre:      "Año inválido",
		FechaInicio: fin,
		FechaFin:    inicio,
	})
	assert.ErrorIs(t, err, academico.ErrFechasAnioInvalidas)

	gradoID, err := servicio.CrearGrado(ctx, usuarioID, academico.DatosGrado{
		AnioEscolarID: anioID,
		Nombre:        "  3.°  ",
	})
	require.NoError(t, err)

	grados, err := servicio.ListarGrados(ctx, anioID)
	require.NoError(t, err)
	require.Len(t, grados, 1)
	assert.Equal(t, "3.°", grados[0].Nombre, "el nombre debe normalizarse")

	_, err = servicio.CrearGrado(ctx, usuarioID, academico.DatosGrado{
		AnioEscolarID: anioID,
		Nombre:        "3.°",
	})
	assert.ErrorIs(t, err, academico.ErrGradoExistente)

	seccionID, err := servicio.CrearSeccion(ctx, usuarioID, academico.DatosSeccion{
		GradoID:      gradoID,
		Nombre:       "A",
		Turno:        academico.TurnoManana,
		VigenteDesde: inicio,
	})
	require.NoError(t, err)

	secciones, err := servicio.ListarSecciones(ctx, anioID)
	require.NoError(t, err)
	require.Len(t, secciones, 1)
	assert.Equal(t, academico.TurnoManana, secciones[0].Turno)
	assert.Equal(t, "3.°", secciones[0].Grado)

	_, err = servicio.CrearSeccion(ctx, usuarioID, academico.DatosSeccion{
		GradoID:      gradoID,
		Nombre:       "A",
		Turno:        academico.TurnoTarde,
		VigenteDesde: inicio,
	})
	assert.ErrorIs(t, err, academico.ErrSeccionExistente)

	cambio := inicio.AddDate(0, 1, 0)
	require.NoError(t, servicio.CambiarTurno(ctx, usuarioID, seccionID, academico.TurnoTarde, cambio))

	turnos, err := servicio.ListarTurnosSeccion(ctx, seccionID)
	require.NoError(t, err)
	require.Len(t, turnos, 2)
	assert.Equal(t, academico.TurnoTarde, turnos[0].Turno)
	assert.Equal(t, academico.TurnoManana, turnos[1].Turno)

	require.NoError(t, servicio.CerrarAnio(ctx, usuarioID, anioID))

	activo, err = servicio.AnioActivo(ctx)
	require.NoError(t, err)
	assert.Nil(t, activo, "al cerrar no debe quedar un año activo")

	_, err = servicio.CrearGrado(ctx, usuarioID, academico.DatosGrado{
		AnioEscolarID: anioID,
		Nombre:        "4.°",
	})
	assert.ErrorIs(t, err, academico.ErrAnioCerrado)
}

func TestConfiguracionAcademica_RegistraLaAuditoriaDeCadaCambio(t *testing.T) {
	pool := AbrirBaseDePruebas(t)
	usuarioID := CrearUsuarioPrueba(t, pool)
	servicio := academico.NuevoServicioConfiguracion(pool)
	ctx := context.Background()

	anio := time.Now().Year() + 1
	inicio := time.Date(anio, time.March, 1, 0, 0, 0, 0, time.UTC)
	fin := time.Date(anio, time.December, 15, 0, 0, 0, 0, time.UTC)

	anioID, err := servicio.CrearAnio(ctx, usuarioID, academico.DatosAnio{
		Nombre:      time.Date(anio, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006"),
		FechaInicio: inicio,
		FechaFin:    fin,
	})
	require.NoError(t, err)
	gradoID, err := servicio.CrearGrado(ctx, usuarioID, academico.DatosGrado{
		AnioEscolarID: anioID,
		Nombre:        "5.°",
	})
	require.NoError(t, err)
	seccionID, err := servicio.CrearSeccion(ctx, usuarioID, academico.DatosSeccion{
		GradoID:      gradoID,
		Nombre:       "B",
		Turno:        academico.TurnoManana,
		VigenteDesde: inicio,
	})
	require.NoError(t, err)
	require.NoError(t, servicio.CambiarTurno(ctx, usuarioID, seccionID, academico.TurnoTarde, inicio.AddDate(0, 1, 0)))
	require.NoError(t, servicio.CerrarAnio(ctx, usuarioID, anioID))

	casos := map[string]string{
		academico.EntidadAnioEscolar: academico.AccionCrear,
		academico.EntidadGrado:       academico.AccionCrear,
		academico.EntidadSeccion:     academico.AccionCrear,
		academico.EntidadTurno:       academico.AccionCambiarTurno,
	}
	for entidad, accion := range casos {
		var cantidad int
		err := pool.QueryRow(ctx,
			`SELECT count(*) FROM auditoria WHERE entidad = $1 AND accion = $2 AND usuario_id = $3`,
			entidad, accion, usuarioID,
		).Scan(&cantidad)
		require.NoError(t, err)
		assert.Equal(t, 1, cantidad, "auditoría de %s %s", entidad, accion)
	}

	var cierres int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM auditoria
		 WHERE entidad = $1 AND accion = $2 AND valor_nuevo->>'estado' = 'cerrado'`,
		academico.EntidadAnioEscolar, academico.AccionCerrar,
	).Scan(&cierres))
	assert.Equal(t, 1, cierres)
}
