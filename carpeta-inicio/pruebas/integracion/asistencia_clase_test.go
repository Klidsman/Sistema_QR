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

func TestAsistenciaClase_ListaGuardaMarcasYDiferencias(t *testing.T) {
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

	docenteID := CrearDocentePrueba(t, pool, "clase@colegio.pe", "Docente de Clase")
	_, err = servicio.CrearAsignacion(ctx, usuarioID, asistencia.DatosAsignacion{
		DocenteID: docenteID, CursoID: cursoID, SeccionID: seccionID,
		FechaInicio: hoy, Titular: true,
	})
	require.NoError(t, err)

	alumnoConQR, err := servicioAlumnos.Crear(ctx, usuarioID, alumnos.DatosAlumno{
		Identificador: "80001111", Nombres: "Ana", Apellidos: "Ríos", SeccionID: seccionID,
	})
	require.NoError(t, err)
	alumnoSinQR, err := servicioAlumnos.Crear(ctx, usuarioID, alumnos.DatosAlumno{
		Identificador: "80002222", Nombres: "Luis", Apellidos: "Pérez", SeccionID: seccionID,
	})
	require.NoError(t, err)

	_, err = pool.Exec(ctx,
		`INSERT INTO marcaciones (alumno_id, fecha, hora, estado)
		 VALUES ($1, $2, $3, 'puntual')`,
		alumnoConQR, fechaSQL(hoy), hoy.Add(7*time.Hour+42*time.Minute))
	require.NoError(t, err)

	detalle, err := servicio.ListarClase(ctx, docenteID, false, hoy, seccionID, cursoID, 1)
	require.NoError(t, err)
	require.Len(t, detalle.Alumnos, 2)
	assert.Equal(t, "Matemática", detalle.Clase.Curso)
	for _, alumno := range detalle.Alumnos {
		if alumno.Alumno.ID == alumnoConQR {
			require.NotNil(t, alumno.Marcacion)
			assert.Equal(t, "07:42", fechas.Hora(alumno.Marcacion.Hora))
		} else {
			assert.Nil(t, alumno.Marcacion)
		}
	}

	require.NoError(t, servicio.GuardarRegistros(ctx, docenteID, false, hoy, seccionID, cursoID, 1,
		[]asistencia.CambioRegistro{
			{AlumnoID: alumnoConQR, Estado: asistencia.EstadoFalta, Observacion: "Llegó tarde y se retiró"},
		}))

	creados, err := servicio.MarcarTodosPresentes(ctx, docenteID, false, hoy, seccionID, cursoID, 1)
	require.NoError(t, err)
	assert.Equal(t, 1, creados, "solo faltaba el alumno sin registro")

	detalle, err = servicio.ListarClase(ctx, docenteID, false, hoy, seccionID, cursoID, 1)
	require.NoError(t, err)
	for _, alumno := range detalle.Alumnos {
		require.NotNil(t, alumno.Registro)
		if alumno.Alumno.ID == alumnoSinQR {
			assert.Equal(t, asistencia.EstadoPresente, alumno.Registro.Estado)
			assert.True(t, alumno.Diferencia, "sin QR y presente es una diferencia informativa")
		}
		if alumno.Alumno.ID == alumnoConQR {
			assert.Equal(t, asistencia.EstadoFalta, alumno.Registro.Estado)
			assert.True(t, alumno.Diferencia, "con QR y falta es una diferencia informativa")
			assert.Equal(t, "Llegó tarde y se retiró", alumno.Registro.Observacion)
		}
	}

	creados, err = servicio.MarcarTodosPresentes(ctx, docenteID, false, hoy, seccionID, cursoID, 1)
	require.NoError(t, err)
	assert.Equal(t, 0, creados, "no debe duplicar ni tocar los registros existentes")

	err = servicio.GuardarRegistros(ctx, docenteID, false, hoy, seccionID, cursoID, 1,
		[]asistencia.CambioRegistro{{AlumnoID: 999999, Estado: asistencia.EstadoPresente}})
	assert.ErrorIs(t, err, asistencia.ErrAlumnoFueraDeSeccion)

	err = servicio.GuardarRegistros(ctx, docenteID, false, hoy.AddDate(0, 0, 1), seccionID, cursoID, 1,
		[]asistencia.CambioRegistro{{AlumnoID: alumnoConQR, Estado: asistencia.EstadoPresente}})
	assert.ErrorIs(t, err, asistencia.ErrFechaFutura)

	exportados, err := servicio.ExportarDocente(ctx, docenteID, hoy, hoy)
	require.NoError(t, err)
	require.Len(t, exportados, 2)
	for _, fila := range exportados {
		if fila.Identificador == "80001111" {
			assert.Equal(t, "Ríos, Ana", fila.Alumno)
			assert.Equal(t, "07:42", fila.HoraQR)
			assert.Equal(t, asistencia.EstadoFalta, fila.Estado)
			assert.Equal(t, "Matemática", fila.Curso)
		}
	}

	csvExportado, err := asistencia.GenerarCSV(exportados)
	require.NoError(t, err)
	assert.Contains(t, string(csvExportado), "identificador")
	assert.Contains(t, string(csvExportado), "Ríos, Ana")

	porSeccion, err := servicio.ExportarSeccion(ctx, seccionID, 0, hoy, hoy)
	require.NoError(t, err)
	assert.Len(t, porSeccion, 2)

	var creadosAuditados, marcasMasivas int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM auditoria WHERE entidad = $1 AND accion = $2`,
		asistencia.EntidadRegistroDocente, asistencia.AccionCrearRegistro).Scan(&creadosAuditados))
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM auditoria WHERE entidad = $1 AND accion = $2`,
		asistencia.EntidadRegistroDocente, asistencia.AccionMarcarTodos).Scan(&marcasMasivas))
	assert.GreaterOrEqual(t, creadosAuditados, 1)
	assert.Equal(t, 1, marcasMasivas)
}

func fechaSQL(fecha time.Time) time.Time {
	return time.Date(fecha.Year(), fecha.Month(), fecha.Day(), 0, 0, 0, 0, time.UTC)
}
