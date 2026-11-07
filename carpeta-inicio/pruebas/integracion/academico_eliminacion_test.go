package integracion

import (
	"context"
	"testing"
	"time"

	"asistencia/internal/academico"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEliminarGradoYSeccion_RespetaDependencias(t *testing.T) {
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
		AnioEscolarID: anioID, Nombre: "3.°",
	})
	require.NoError(t, err)
	seccionID, err := servicio.CrearSeccion(ctx, usuarioID, academico.DatosSeccion{
		GradoID: gradoID, Nombre: "A", Turno: academico.TurnoManana, VigenteDesde: inicio,
	})
	require.NoError(t, err)

	assert.ErrorIs(t, servicio.EliminarGrado(ctx, usuarioID, gradoID), academico.ErrGradoConDependencias,
		"un grado con secciones no se puede eliminar")

	require.NoError(t, servicio.EliminarSeccion(ctx, usuarioID, seccionID))
	grados, err := servicio.ListarGrados(ctx, anioID)
	require.NoError(t, err)
	assert.Len(t, grados, 1, "eliminar una sección no debe eliminar el grado")

	require.NoError(t, servicio.EliminarGrado(ctx, usuarioID, gradoID))
	grados, err = servicio.ListarGrados(ctx, anioID)
	require.NoError(t, err)
	assert.Empty(t, grados)

	grado2ID, err := servicio.CrearGrado(ctx, usuarioID, academico.DatosGrado{
		AnioEscolarID: anioID, Nombre: "4.°",
	})
	require.NoError(t, err)
	seccion2ID, err := servicio.CrearSeccion(ctx, usuarioID, academico.DatosSeccion{
		GradoID: grado2ID, Nombre: "B", Turno: academico.TurnoManana, VigenteDesde: inicio,
	})
	require.NoError(t, err)
	_, err = pool.Exec(ctx,
		`INSERT INTO alumnos (identificador, nombres, apellidos, nombres_busqueda, apellidos_busqueda, seccion_id)
		 VALUES ('90000001', 'Ana', 'Ríos', 'ana', 'rios', $1)`, seccion2ID)
	require.NoError(t, err)

	assert.ErrorIs(t, servicio.EliminarSeccion(ctx, usuarioID, seccion2ID), academico.ErrSeccionConDependencias,
		"una sección con alumnos no se puede eliminar")

	var eliminados int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM auditoria WHERE accion = $1`, academico.AccionEliminar).Scan(&eliminados))
	assert.GreaterOrEqual(t, eliminados, 2)
}
