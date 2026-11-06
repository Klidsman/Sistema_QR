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

func TestAsistenciaConsulta_PlazoCorreccionYConsulta(t *testing.T) {
	pool := AbrirBaseDePruebas(t)
	usuarioID := CrearUsuarioPrueba(t, pool)
	configuracion := academico.NuevoServicioConfiguracion(pool)
	servicioAlumnos := alumnos.NuevoServicio(pool)
	auth := autenticacion.NuevoServicio(pool)
	servicio := asistencia.NuevoServicio(asistencia.DependenciasServicio{
		Repositorio: asistencia.NuevoRepositorio(pool),
		Academico:   configuracion,
		Alumnos:     servicioAlumnos,
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
	seccionID, err := configuracion.CrearSeccion(ctx, usuarioID, academico.DatosSeccion{
		GradoID: gradoID, Nombre: "A", Turno: academico.TurnoManana, VigenteDesde: hoy,
	})
	require.NoError(t, err)
	cursoID, err := configuracion.CrearCurso(ctx, usuarioID, academico.DatosCurso{
		AnioEscolarID: anioID, Nombre: "Matemática", Grados: []int64{gradoID},
	})
	require.NoError(t, err)
	require.NoError(t, configuracion.ConfigurarHorario(ctx, usuarioID, academico.HorarioTurno{
		Turno: academico.TurnoManana, HoraIngreso: horaDePrueba(t, "07:30"),
		ToleranciaMinutos: 10, HoraLimiteTardanza: horaDePrueba(t, "08:00"),
		HoraCierre: horaDePrueba(t, "12:30"), VigenteDesde: hoy,
	}))
	_, err = configuracion.CrearBloque(ctx, usuarioID, academico.BloqueClase{
		Turno: academico.TurnoManana, Numero: 1,
		HoraInicio: horaDePrueba(t, "07:45"), HoraFin: horaDePrueba(t, "08:30"), VigenteDesde: hoy,
	})
	require.NoError(t, err)
	require.NoError(t, configuracion.GuardarCeldas(ctx, usuarioID, seccionID, hoy,
		[]academico.CambioCelda{{DiaSemana: int(hoy.Weekday()), BloqueNumero: 1, CursoID: &cursoID}}, false))
	docenteID := CrearDocentePrueba(t, pool, "consulta@colegio.pe", "Docente Consulta")
	_, err = servicio.CrearAsignacion(ctx, usuarioID, asistencia.DatosAsignacion{
		DocenteID: docenteID, CursoID: cursoID, SeccionID: seccionID, FechaInicio: hoy, Titular: true,
	})
	require.NoError(t, err)
	alumnoID, err := servicioAlumnos.Crear(ctx, usuarioID, alumnos.DatosAlumno{
		Identificador: "81001111", Nombres: "Ana", Apellidos: "Ríos", SeccionID: seccionID,
	})
	require.NoError(t, err)

	require.NoError(t, servicio.GuardarRegistros(ctx, docenteID, false, hoy, seccionID, cursoID, 1,
		[]asistencia.CambioRegistro{{AlumnoID: alumnoID, Estado: asistencia.EstadoPresente}}))

	require.NoError(t, servicio.ConfigurarPlazo(ctx, usuarioID, 1))
	_, err = pool.Exec(ctx,
		`UPDATE registros_docente SET creado_en = now() - interval '3 hours' WHERE alumno_id = $1`,
		alumnoID)
	require.NoError(t, err)

	err = servicio.GuardarRegistros(ctx, docenteID, false, hoy, seccionID, cursoID, 1,
		[]asistencia.CambioRegistro{{AlumnoID: alumnoID, Estado: asistencia.EstadoFalta}})
	assert.ErrorIs(t, err, asistencia.ErrPlazoVencido)

	err = servicio.CorregirRegistro(ctx, usuarioID, 1, asistencia.EstadoTardanza, "Corregido", "")
	assert.ErrorIs(t, err, asistencia.ErrMotivoRequerido)

	var registroID int64
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT id FROM registros_docente WHERE alumno_id = $1`, alumnoID).Scan(&registroID))
	require.NoError(t, servicio.CorregirRegistro(ctx, usuarioID, registroID,
		asistencia.EstadoTardanza, "Corregido por administrador", "Error del docente"))

	registros, err := servicio.ConsultaRegistros(ctx, hoy, seccionID)
	require.NoError(t, err)
	require.Len(t, registros, 1)
	assert.Equal(t, "Ríos, Ana", registros[0].Alumno)
	assert.Equal(t, "Matemática", registros[0].Curso)
	assert.Equal(t, "Usuario de pruebas", registros[0].Docente)
	assert.Equal(t, asistencia.EstadoTardanza, registros[0].Registro.Estado)
	assert.Equal(t, "Corregido por administrador", registros[0].Registro.Observacion)
	assert.True(t, registros[0].Diferencia, "sin QR y tardanza es una diferencia informativa")

	assert.ErrorIs(t, servicio.ConfigurarPlazo(ctx, usuarioID, 0), asistencia.ErrPlazoFueraDeRango)
	assert.ErrorIs(t, servicio.ConfigurarPlazo(ctx, usuarioID, 169), asistencia.ErrPlazoFueraDeRango)

	var correcciones, plazos int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM auditoria WHERE entidad = $1 AND accion = $2 AND motivo IS NOT NULL`,
		asistencia.EntidadRegistroDocente, asistencia.AccionCorregirRegistro).Scan(&correcciones))
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM auditoria WHERE entidad = $1 AND accion = $2`,
		asistencia.EntidadParametroAsistencia, asistencia.AccionConfigurarPlazo).Scan(&plazos))
	assert.Equal(t, 1, correcciones)
	assert.Equal(t, 1, plazos)

	parametros, err := servicio.PlazoEdicion(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, parametros.PlazoEdicionHoras)
}
