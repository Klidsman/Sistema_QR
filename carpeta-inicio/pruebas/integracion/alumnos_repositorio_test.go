package integracion

import (
	"context"
	"testing"

	"asistencia/internal/alumnos"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAlumnosRepositorio_CrearYBuscar(t *testing.T) {
	pool := AbrirBaseDePruebas(t)
	seccionID := SembrarSeccion(t, pool, "3.°", "A")
	usuarioID := CrearUsuarioPrueba(t, pool)
	servicio := alumnos.NuevoServicio(pool)
	ctx := context.Background()

	id, err := servicio.Crear(ctx, usuarioID, alumnos.DatosAlumno{
		Identificador:   "00123456",
		Nombres:         "José",
		Apellidos:       "Núñez",
		SeccionID:       seccionID,
		ApoderadoNombre: "María Núñez",
	})
	require.NoError(t, err)

	alumno, err := servicio.Obtener(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, "José", alumno.Nombres)
	assert.Equal(t, "Núñez", alumno.Apellidos)
	assert.Equal(t, "3.°", alumno.Grado)
	assert.Equal(t, "A", alumno.Seccion)
	assert.Equal(t, alumnos.EstadoActivo, alumno.Estado)
}

func TestAlumnosRepositorio_RechazaIdentificadorRepetido(t *testing.T) {
	pool := AbrirBaseDePruebas(t)
	seccionID := SembrarSeccion(t, pool, "3.°", "A")
	usuarioID := CrearUsuarioPrueba(t, pool)
	servicio := alumnos.NuevoServicio(pool)
	ctx := context.Background()

	datos := alumnos.DatosAlumno{
		Identificador: "00123456",
		Nombres:       "José",
		Apellidos:     "Núñez",
		SeccionID:     seccionID,
	}
	_, err := servicio.Crear(ctx, usuarioID, datos)
	require.NoError(t, err)

	datos.Nombres = "Otro"
	_, err = servicio.Crear(ctx, usuarioID, datos)
	assert.ErrorIs(t, err, alumnos.ErrIdentificadorExistente)
}

func TestAlumnosRepositorio_BusquedaSinTildesNiMayusculas(t *testing.T) {
	pool := AbrirBaseDePruebas(t)
	seccionID := SembrarSeccion(t, pool, "3.°", "A")
	usuarioID := CrearUsuarioPrueba(t, pool)
	servicio := alumnos.NuevoServicio(pool)
	ctx := context.Background()

	_, err := servicio.Crear(ctx, usuarioID, alumnos.DatosAlumno{
		Identificador: "87654321",
		Nombres:       "José",
		Apellidos:     "Núñez",
		SeccionID:     seccionID,
	})
	require.NoError(t, err)

	casos := []string{"jose", "JOSE", "nunez", "NÚÑEZ", "8765"}
	for _, texto := range casos {
		listado, err := servicio.Listar(ctx, alumnos.FiltrosListado{Texto: texto})
		require.NoError(t, err)
		assert.Len(t, listado, 1, "búsqueda %q", texto)
	}
}

func TestAlumnosRepositorio_DesactivarConservaElRegistro(t *testing.T) {
	pool := AbrirBaseDePruebas(t)
	seccionID := SembrarSeccion(t, pool, "3.°", "A")
	usuarioID := CrearUsuarioPrueba(t, pool)
	servicio := alumnos.NuevoServicio(pool)
	ctx := context.Background()

	id, err := servicio.Crear(ctx, usuarioID, alumnos.DatosAlumno{
		Identificador: "55667788",
		Nombres:       "Ana",
		Apellidos:     "Quispe",
		SeccionID:     seccionID,
	})
	require.NoError(t, err)
	require.NoError(t, servicio.Desactivar(ctx, usuarioID, id))

	alumno, err := servicio.Obtener(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, alumnos.EstadoDesactivado, alumno.Estado)

	listado, err := servicio.Listar(ctx, alumnos.FiltrosListado{Estado: alumnos.EstadoDesactivado})
	require.NoError(t, err)
	assert.Len(t, listado, 1)
}
