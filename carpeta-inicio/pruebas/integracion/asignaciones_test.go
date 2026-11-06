package integracion

import (
	"context"
	"testing"
	"time"

	"asistencia/internal/academico"
	"asistencia/internal/alumnos"
	"asistencia/internal/asistencia"
	"asistencia/internal/autenticacion"
	"asistencia/internal/plataforma/fechas"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsignaciones_CreaCierraYControlaAccesoPorFecha(t *testing.T) {
	pool := AbrirBaseDePruebas(t)
	usuarioID := CrearUsuarioPrueba(t, pool)
	configuracion := academico.NuevoServicioConfiguracion(pool)
	auth := autenticacion.NuevoServicio(pool)
	servicio := asistencia.NuevoServicio(asistencia.DependenciasServicio{
		Repositorio: asistencia.NuevoRepositorio(pool),
		Academico:   configuracion,
		Alumnos:     alumnos.NuevoServicio(pool),
		Docentes:    auth,
	})
	ctx := context.Background()
	hoy := fechas.Hoy()

	anio := time.Now().Year() + 1
	anioID, err := configuracion.CrearAnio(ctx, usuarioID, academico.DatosAnio{
		Nombre:      time.Date(anio, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006"),
		FechaInicio: time.Date(anio, time.March, 1, 0, 0, 0, 0, time.UTC),
		FechaFin:    time.Date(anio, time.December, 15, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)
	gradoID, err := configuracion.CrearGrado(ctx, usuarioID, academico.DatosGrado{AnioEscolarID: anioID, Nombre: "3.°"})
	require.NoError(t, err)
	grado2ID, err := configuracion.CrearGrado(ctx, usuarioID, academico.DatosGrado{AnioEscolarID: anioID, Nombre: "4.°"})
	require.NoError(t, err)
	seccionID, err := configuracion.CrearSeccion(ctx, usuarioID, academico.DatosSeccion{
		GradoID: gradoID, Nombre: "A", Turno: academico.TurnoManana, VigenteDesde: hoy,
	})
	require.NoError(t, err)
	cursoID, err := configuracion.CrearCurso(ctx, usuarioID, academico.DatosCurso{
		AnioEscolarID: anioID, Nombre: "Matemática", Grados: []int64{gradoID},
	})
	require.NoError(t, err)
	cursoAjenoID, err := configuracion.CrearCurso(ctx, usuarioID, academico.DatosCurso{
		AnioEscolarID: anioID, Nombre: "Historia", Grados: []int64{grado2ID},
	})
	require.NoError(t, err)

	require.NoError(t, configuracion.ConfigurarHorario(ctx, usuarioID, academico.HorarioTurno{
		Turno: academico.TurnoManana, HoraIngreso: horaDePrueba(t, "07:30"),
		ToleranciaMinutos: 10, HoraLimiteTardanza: horaDePrueba(t, "08:00"),
		HoraCierre: horaDePrueba(t, "12:30"), VigenteDesde: hoy,
	}))
	_, err = configuracion.CrearBloque(ctx, usuarioID, academico.BloqueClase{
		Turno: academico.TurnoManana, Numero: 1,
		HoraInicio: horaDePrueba(t, "07:45"), HoraFin: horaDePrueba(t, "08:30"),
		VigenteDesde: hoy,
	})
	require.NoError(t, err)
	require.NoError(t, configuracion.GuardarCeldas(ctx, usuarioID, seccionID, hoy,
		[]academico.CambioCelda{{DiaSemana: int(hoy.Weekday()), BloqueNumero: 1, CursoID: &cursoID}}, false))

	titularID := CrearDocentePrueba(t, pool, "titular@colegio.pe", "Docente Titular")
	reemplazoID := CrearDocentePrueba(t, pool, "reemplazo@colegio.pe", "Docente Reemplazo")
	ajenoID := CrearDocentePrueba(t, pool, "ajeno@colegio.pe", "Docente Ajeno")

	titularAsignacion, err := servicio.CrearAsignacion(ctx, usuarioID, asistencia.DatosAsignacion{
		DocenteID: titularID, CursoID: cursoID, SeccionID: seccionID,
		FechaInicio: hoy, Titular: true,
	})
	require.NoError(t, err)

	manana := hoy.AddDate(0, 0, 1)
	reemplazoAsignacion, err := servicio.CrearAsignacion(ctx, usuarioID, asistencia.DatosAsignacion{
		DocenteID: reemplazoID, CursoID: cursoID, SeccionID: seccionID,
		FechaInicio: manana, FechaFin: &manana, Titular: false,
	})
	require.NoError(t, err)

	asignado, err := asistencia.NuevoRepositorio(pool).DocenteAsignadoAClase(ctx, titularID, cursoID, seccionID, hoy)
	require.NoError(t, err)
	assert.True(t, asignado)

	asignado, err = asistencia.NuevoRepositorio(pool).DocenteAsignadoAClase(ctx, reemplazoID, cursoID, seccionID, hoy)
	require.NoError(t, err)
	assert.False(t, asignado, "el reemplazo empieza mañana")

	asignado, err = asistencia.NuevoRepositorio(pool).DocenteAsignadoAClase(ctx, reemplazoID, cursoID, seccionID, manana)
	require.NoError(t, err)
	assert.True(t, asignado)

	clases, err := servicio.ClasesAsignadas(ctx, titularID, hoy)
	require.NoError(t, err)
	require.Len(t, clases, 1)
	assert.Equal(t, "Matemática", clases[0].Curso)
	assert.Equal(t, 1, clases[0].BloqueNumero)

	clases, err = servicio.ClasesAsignadas(ctx, ajenoID, hoy)
	require.NoError(t, err)
	assert.Empty(t, clases, "un docente sin asignación no debe ver la clase")

	_, err = servicio.CrearAsignacion(ctx, usuarioID, asistencia.DatosAsignacion{
		DocenteID: titularID, CursoID: cursoAjenoID, SeccionID: seccionID,
		FechaInicio: hoy, Titular: true,
	})
	assert.ErrorIs(t, err, asistencia.ErrCursoNoVinculado)

	_, err = servicio.CrearAsignacion(ctx, usuarioID, asistencia.DatosAsignacion{
		DocenteID: usuarioID, CursoID: cursoID, SeccionID: seccionID,
		FechaInicio: hoy, Titular: true,
	})
	assert.ErrorIs(t, err, asistencia.ErrDocenteNoEncontrado)

	finAnterior := hoy.AddDate(0, 0, -1)
	_, err = servicio.CrearAsignacion(ctx, usuarioID, asistencia.DatosAsignacion{
		DocenteID: titularID, CursoID: cursoID, SeccionID: seccionID,
		FechaInicio: hoy, FechaFin: &finAnterior, Titular: true,
	})
	assert.ErrorIs(t, err, asistencia.ErrAsignacionFueraDeRango)

	require.NoError(t, servicio.CerrarAsignacion(ctx, usuarioID, titularAsignacion, hoy))
	asignado, err = asistencia.NuevoRepositorio(pool).DocenteAsignadoAClase(ctx, titularID, cursoID, seccionID, manana)
	require.NoError(t, err)
	assert.False(t, asignado, "al cerrar hoy ya no tiene acceso mañana")

	vistas, err := servicio.ListarAsignaciones(ctx)
	require.NoError(t, err)
	assert.Len(t, vistas, 2)
	assert.NotEqual(t, int64(0), reemplazoAsignacion)

	futuro := hoy.AddDate(0, 0, 5)
	_, err = servicio.ClasesAsignadas(ctx, titularID, futuro)
	assert.ErrorIs(t, err, asistencia.ErrFechaFutura)

	var creadas, cerradas int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM auditoria WHERE entidad = $1 AND accion = $2`,
		asistencia.EntidadAsignacionDocente, asistencia.AccionCrearAsignacion).Scan(&creadas))
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM auditoria WHERE entidad = $1 AND accion = $2`,
		asistencia.EntidadAsignacionDocente, asistencia.AccionCerrarAsignacion).Scan(&cerradas))
	assert.Equal(t, 2, creadas)
	assert.Equal(t, 1, cerradas)
}
