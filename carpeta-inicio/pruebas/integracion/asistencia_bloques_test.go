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

func TestAsistenciaBloques_CadaHoraTieneSuPropioRegistro(t *testing.T) {
	pool := AbrirBaseDePruebas(t)
	usuarioID := CrearUsuarioPrueba(t, pool)
	configuracion := academico.NuevoServicioConfiguracion(pool)
	servicioAlumnos := alumnos.NuevoServicio(pool)
	servicio := asistencia.NuevoServicio(asistencia.DependenciasServicio{
		Repositorio: asistencia.NuevoRepositorio(pool),
		Academico:   configuracion,
		Alumnos:     servicioAlumnos,
		Docentes:    autenticacion.NuevoServicio(pool),
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
	_, err = configuracion.CrearBloque(ctx, usuarioID, academico.BloqueClase{
		Turno: academico.TurnoManana, Numero: 2,
		HoraInicio: horaDePrueba(t, "08:30"), HoraFin: horaDePrueba(t, "09:15"), VigenteDesde: hoy,
	})
	require.NoError(t, err)
	require.NoError(t, configuracion.GuardarCeldas(ctx, usuarioID, seccionID, hoy,
		[]academico.CambioCelda{
			{DiaSemana: int(hoy.Weekday()), BloqueNumero: 1, CursoID: &cursoID},
			{DiaSemana: int(hoy.Weekday()), BloqueNumero: 2, CursoID: &cursoID},
		}, false))

	docenteID := CrearDocentePrueba(t, pool, "bloques@colegio.pe", "Docente Bloques")
	_, err = servicio.CrearAsignacion(ctx, usuarioID, asistencia.DatosAsignacion{
		DocenteID: docenteID, CursoID: cursoID, SeccionID: seccionID, FechaInicio: hoy, Titular: true,
	})
	require.NoError(t, err)
	alumnoID, err := servicioAlumnos.Crear(ctx, usuarioID, alumnos.DatosAlumno{
		Identificador: "82001111", Nombres: "Ana", Apellidos: "Ríos", SeccionID: seccionID,
	})
	require.NoError(t, err)

	require.NoError(t, servicio.GuardarRegistros(ctx, docenteID, false, hoy, seccionID, cursoID, 1,
		[]asistencia.CambioRegistro{{AlumnoID: alumnoID, Estado: asistencia.EstadoPresente}}))
	require.NoError(t, servicio.GuardarRegistros(ctx, docenteID, false, hoy, seccionID, cursoID, 2,
		[]asistencia.CambioRegistro{{AlumnoID: alumnoID, Estado: asistencia.EstadoTardanza}}))

	bloque1, err := servicio.ListarClase(ctx, docenteID, false, hoy, seccionID, cursoID, 1)
	require.NoError(t, err)
	require.Len(t, bloque1.Alumnos, 1)
	require.NotNil(t, bloque1.Alumnos[0].Registro)
	assert.Equal(t, asistencia.EstadoPresente, bloque1.Alumnos[0].Registro.Estado)

	bloque2, err := servicio.ListarClase(ctx, docenteID, false, hoy, seccionID, cursoID, 2)
	require.NoError(t, err)
	require.Len(t, bloque2.Alumnos, 1)
	require.NotNil(t, bloque2.Alumnos[0].Registro)
	assert.Equal(t, asistencia.EstadoTardanza, bloque2.Alumnos[0].Registro.Estado)

	var registros int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM registros_docente WHERE alumno_id = $1 AND fecha = $2`,
		alumnoID, fechaSQL(hoy)).Scan(&registros))
	assert.Equal(t, 2, registros, "el mismo alumno tiene un registro por bloque")

	creados, err := servicio.MarcarTodosPresentes(ctx, docenteID, false, hoy, seccionID, cursoID, 1)
	require.NoError(t, err)
	assert.Equal(t, 0, creados, "el bloque 1 ya tenía registro")

	var auditorias int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM auditoria WHERE entidad = $1 AND accion = $2`,
		asistencia.EntidadRegistroDocente, asistencia.AccionCrearRegistro).Scan(&auditorias))
	assert.Equal(t, 2, auditorias)
}
