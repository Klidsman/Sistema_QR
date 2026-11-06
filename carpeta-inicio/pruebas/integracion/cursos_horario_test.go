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

func TestCursosYHorarioSemanal_FlujoCompleto(t *testing.T) {
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

	gradoID, err := servicio.CrearGrado(ctx, usuarioID, academico.DatosGrado{
		AnioEscolarID: anioID,
		Nombre:        "3.°",
	})
	require.NoError(t, err)
	grado2ID, err := servicio.CrearGrado(ctx, usuarioID, academico.DatosGrado{
		AnioEscolarID: anioID,
		Nombre:        "4.°",
	})
	require.NoError(t, err)
	seccionID, err := servicio.CrearSeccion(ctx, usuarioID, academico.DatosSeccion{
		GradoID:      gradoID,
		Nombre:       "A",
		Turno:        academico.TurnoManana,
		VigenteDesde: hoy,
	})
	require.NoError(t, err)

	require.NoError(t, servicio.ConfigurarHorario(ctx, usuarioID, academico.HorarioTurno{
		Turno:              academico.TurnoManana,
		HoraIngreso:        horaDePrueba(t, "07:30"),
		ToleranciaMinutos:  10,
		HoraLimiteTardanza: horaDePrueba(t, "08:00"),
		HoraCierre:         horaDePrueba(t, "12:30"),
		VigenteDesde:       hoy,
	}))
	_, err = servicio.CrearBloque(ctx, usuarioID, academico.BloqueClase{
		Turno:        academico.TurnoManana,
		Numero:       1,
		HoraInicio:   horaDePrueba(t, "07:45"),
		HoraFin:      horaDePrueba(t, "08:30"),
		VigenteDesde: hoy,
	})
	require.NoError(t, err)
	_, err = servicio.CrearBloque(ctx, usuarioID, academico.BloqueClase{
		Turno:        academico.TurnoManana,
		Numero:       2,
		HoraInicio:   horaDePrueba(t, "08:30"),
		HoraFin:      horaDePrueba(t, "09:15"),
		VigenteDesde: hoy,
	})
	require.NoError(t, err)

	cursoID, err := servicio.CrearCurso(ctx, usuarioID, academico.DatosCurso{
		AnioEscolarID: anioID,
		Nombre:        "  Matemática  ",
		Grados:        []int64{gradoID},
	})
	require.NoError(t, err)

	cursos, err := servicio.ListarCursos(ctx, anioID)
	require.NoError(t, err)
	require.Len(t, cursos, 1)
	assert.Equal(t, "Matemática", cursos[0].Nombre)
	assert.True(t, cursos[0].Activo)
	assert.Equal(t, []int64{gradoID}, cursos[0].Grados)

	_, err = servicio.CrearCurso(ctx, usuarioID, academico.DatosCurso{
		AnioEscolarID: anioID,
		Nombre:        "matemática",
		Grados:        []int64{gradoID},
	})
	assert.ErrorIs(t, err, academico.ErrCursoExistente)

	_, err = servicio.CrearCurso(ctx, usuarioID, academico.DatosCurso{
		AnioEscolarID: anioID,
		Nombre:        "Sin grados",
	})
	assert.ErrorIs(t, err, academico.ErrCursoSinGrados)

	_, err = servicio.CrearCurso(ctx, usuarioID, academico.DatosCurso{
		AnioEscolarID: anioID,
		Nombre:        "Grado ajeno",
		Grados:        []int64{9999},
	})
	assert.ErrorIs(t, err, academico.ErrGradoNoEncontrado)

	grilla, err := servicio.ObtenerGrilla(ctx, seccionID, hoy)
	require.NoError(t, err)
	assert.Equal(t, academico.TurnoManana, grilla.Turno)
	require.Len(t, grilla.Bloques, 2)
	assert.Empty(t, grilla.Celdas)
	require.Len(t, grilla.Cursos, 1)

	require.NoError(t, servicio.GuardarCeldas(ctx, usuarioID, seccionID, hoy,
		[]academico.CambioCelda{{DiaSemana: 1, BloqueNumero: 1, CursoID: &cursoID}}, false))

	cursoCelda, err := servicio.CursoDeBloque(ctx, seccionID, 1, 1, hoy)
	require.NoError(t, err)
	require.NotNil(t, cursoCelda)
	assert.Equal(t, cursoID, *cursoCelda)

	require.NoError(t, servicio.GuardarCeldas(ctx, usuarioID, seccionID, hoy,
		[]academico.CambioCelda{{DiaSemana: 1, BloqueNumero: 1, CursoID: &cursoID}}, false),
		"guardar el mismo curso no debe pedir confirmación")

	cursoOtroGrado, err := servicio.CrearCurso(ctx, usuarioID, academico.DatosCurso{
		AnioEscolarID: anioID,
		Nombre:        "Educación Física",
		Grados:        []int64{grado2ID},
	})
	require.NoError(t, err)
	err = servicio.GuardarCeldas(ctx, usuarioID, seccionID, hoy,
		[]academico.CambioCelda{{DiaSemana: 2, BloqueNumero: 1, CursoID: &cursoOtroGrado}}, false)
	assert.ErrorIs(t, err, academico.ErrCursoNoVinculado)

	err = servicio.GuardarCeldas(ctx, usuarioID, seccionID, hoy,
		[]academico.CambioCelda{{DiaSemana: 1, BloqueNumero: 1, CursoID: nil}}, false)
	assert.ErrorIs(t, err, academico.ErrRequiereConfirmacion)

	require.NoError(t, servicio.GuardarCeldas(ctx, usuarioID, seccionID, hoy,
		[]academico.CambioCelda{{DiaSemana: 1, BloqueNumero: 1, CursoID: nil}}, true))

	cursoCelda, err = servicio.CursoDeBloque(ctx, seccionID, 1, 1, hoy)
	require.NoError(t, err)
	assert.Nil(t, cursoCelda)

	err = servicio.GuardarCeldas(ctx, usuarioID, seccionID, hoy,
		[]academico.CambioCelda{{DiaSemana: 1, BloqueNumero: 99, CursoID: &cursoID}}, false)
	assert.ErrorIs(t, err, academico.ErrBloqueNoEncontrado)

	futuro := hoy.AddDate(0, 1, 0)
	require.NoError(t, servicio.GuardarCeldas(ctx, usuarioID, seccionID, futuro,
		[]academico.CambioCelda{{DiaSemana: 3, BloqueNumero: 2, CursoID: &cursoID}}, false))

	cursoCelda, err = servicio.CursoDeBloque(ctx, seccionID, 3, 2, futuro)
	require.NoError(t, err)
	require.NotNil(t, cursoCelda)
	assert.Equal(t, cursoID, *cursoCelda)
	cursoCelda, err = servicio.CursoDeBloque(ctx, seccionID, 3, 2, hoy)
	require.NoError(t, err)
	assert.Nil(t, cursoCelda, "el cambio futuro no debe aplicarse antes de su fecha")

	require.NoError(t, servicio.DesactivarCurso(ctx, usuarioID, cursoID))
	cursos, err = servicio.ListarCursos(ctx, anioID)
	require.NoError(t, err)
	var desactivado *academico.Curso
	for i := range cursos {
		if cursos[i].ID == cursoID {
			desactivado = &cursos[i]
		}
	}
	require.NotNil(t, desactivado)
	assert.False(t, desactivado.Activo)
	require.NoError(t, servicio.DesactivarCurso(ctx, usuarioID, cursoID), "desactivar dos veces no debe fallar")

	err = servicio.GuardarCeldas(ctx, usuarioID, seccionID, hoy,
		[]academico.CambioCelda{{DiaSemana: 5, BloqueNumero: 1, CursoID: &cursoID}}, false)
	assert.ErrorIs(t, err, academico.ErrCursoNoVinculado)

	var creados, desactivados, asignaciones int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM auditoria WHERE entidad = $1 AND accion = $2`,
		academico.EntidadCurso, academico.AccionCrearCurso).Scan(&creados))
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM auditoria WHERE entidad = $1 AND accion = $2`,
		academico.EntidadCurso, academico.AccionDesactivarCurso).Scan(&desactivados))
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM auditoria WHERE entidad = $1 AND accion = $2`,
		academico.EntidadHorarioSeccion, academico.AccionAsignarHorario).Scan(&asignaciones))
	assert.Equal(t, 2, creados)
	assert.Equal(t, 1, desactivados)
	assert.GreaterOrEqual(t, asignaciones, 3)
}
