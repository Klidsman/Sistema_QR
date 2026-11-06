package integracion

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"asistencia/internal/academico"
	"asistencia/internal/alertas"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcesarCierre_EsIdempotenteYRegistraCierre(t *testing.T) {
	pool, _ := AbrirBaseDePruebasConURL(t)
	ctx := context.Background()

	datos := sembrarDatosAlertas(t, pool)
	fuentes := nuevasFuentesFalsas(datos.seccionID, datos.alumnoID, datos.administradorID)
	fuentes.asistencia.marcacionEn = time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC)
	servicio := alertas.NuevoServicio(alertas.Dependencias{
		Asistencia:  fuentes.asistencia,
		Alumnos:     fuentes.alumnos,
		Academico:   fuentes.academico,
		Usuarios:    fuentes.usuarios,
		Repositorio: alertas.NuevoRepositorio(pool),
	})

	fecha := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	require.NoError(t, servicio.ProcesarCierre(ctx, fecha, "manana"))
	require.NoError(t, servicio.ProcesarCierre(ctx, fecha, "manana"))

	var total int
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM alertas`).Scan(&total))
	assert.Equal(t, 1, total, "el cierre repetido no debe duplicar la alerta")
	assert.Equal(t, 1, fuentes.alumnos.llamadas, "el segundo cierre no debe recalcular la racha")

	var cierres int
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM cierres_alerta`).Scan(&cierres))
	assert.Equal(t, 1, cierres)
}

func TestProgramador_IniciaYSeDetiene(t *testing.T) {
	procesador := &procesadorFalso{}
	registro := slog.New(slog.NewTextHandler(io.Discard, nil))
	programador := alertas.NuevoProgramador(procesador, registro, alertas.OpcionesProgramador{
		HoraCierreManana: "13:00",
		HoraCierreTarde:  "19:00",
	})

	ctx, cancelar := context.WithCancel(context.Background())
	require.NoError(t, programador.Iniciar(ctx))
	cancelar()
}

type datosAlertas struct {
	administradorID int64
	seccionID       int64
	alumnoID        int64
}

func sembrarDatosAlertas(t *testing.T, pool *pgxpool.Pool) datosAlertas {
	t.Helper()
	ctx := context.Background()
	var datos datosAlertas
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO usuarios (correo, contrasenia_hash, nombre_completo, rol)
		 VALUES ('admin@prueba.pe', 'x', 'Administrador', 'administrador') RETURNING id`).Scan(&datos.administradorID))
	var anioID, gradoID int64
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO anios_escolares (nombre, fecha_inicio, fecha_fin)
		 VALUES ('2026', '2026-03-01', '2026-12-31') RETURNING id`).Scan(&anioID))
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO grados (anio_escolar_id, nombre) VALUES ($1, '1') RETURNING id`, anioID).Scan(&gradoID))
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO secciones (grado_id, nombre) VALUES ($1, 'A') RETURNING id`, gradoID).Scan(&datos.seccionID))
	_, err := pool.Exec(ctx,
		`INSERT INTO turnos_seccion (seccion_id, turno, vigente_desde) VALUES ($1, 'manana', '2026-03-01')`, datos.seccionID)
	require.NoError(t, err)
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO alumnos (identificador, nombres, apellidos, nombres_busqueda, apellidos_busqueda, seccion_id)
		 VALUES ('A001', 'Ana', 'Perez', 'ana', 'perez', $1) RETURNING id`, datos.seccionID).Scan(&datos.alumnoID))
	return datos
}

type fuentesFalsas struct {
	asistencia *asistenciaFalsa
	alumnos    *alumnosFalsos
	academico  *academicoFalso
	usuarios   *usuariosFalsos
}

func nuevasFuentesFalsas(seccionID, alumnoID, administradorID int64) fuentesFalsas {
	return fuentesFalsas{
		asistencia: &asistenciaFalsa{},
		alumnos:    &alumnosFalsos{seccionID: seccionID, alumnoID: alumnoID},
		academico:  &academicoFalso{seccionID: seccionID},
		usuarios:   &usuariosFalsos{administradorID: administradorID},
	}
}

type asistenciaFalsa struct {
	marcacionEn     time.Time
	marcaciones     []alertas.EventoMarcacion
	contarTardanzas int
	docentes        []int64
}

func (a *asistenciaFalsa) MarcacionesDelRango(context.Context, time.Time, time.Time) ([]alertas.EventoMarcacion, error) {
	return a.marcaciones, nil
}

func (a *asistenciaFalsa) ContarTardanzas(context.Context, int64, time.Time, time.Time) (int, error) {
	return a.contarTardanzas, nil
}

func (a *asistenciaFalsa) TieneMarcacion(_ context.Context, _ int64, fecha time.Time) (bool, error) {
	return fecha.Equal(a.marcacionEn), nil
}

func (a *asistenciaFalsa) DocentesDeSeccionEnFecha(context.Context, int64, time.Time) ([]int64, error) {
	return a.docentes, nil
}

type alumnosFalsos struct {
	seccionID int64
	alumnoID  int64
	llamadas  int
}

func (a *alumnosFalsos) AlumnosActivosDeSeccionEnFecha(context.Context, int64, time.Time) ([]alertas.AlumnoEvento, error) {
	a.llamadas++
	return []alertas.AlumnoEvento{{ID: a.alumnoID, Nombres: "Ana", Apellidos: "Perez"}}, nil
}

func (a *alumnosFalsos) SeccionDeAlumnoEnFecha(context.Context, int64, time.Time) (int64, error) {
	return a.seccionID, nil
}

func (a *alumnosFalsos) AlumnoPorID(context.Context, int64) (alertas.AlumnoEvento, error) {
	return alertas.AlumnoEvento{ID: a.alumnoID, Nombres: "Ana", Apellidos: "Perez"}, nil
}

type academicoFalso struct {
	seccionID int64
	umbrales  map[string]int
}

func (a *academicoFalso) SeccionesActivas(context.Context) ([]academico.Seccion, error) {
	return []academico.Seccion{{ID: a.seccionID, Turno: "manana"}}, nil
}

func (a *academicoFalso) EsDiaNoLaborable(context.Context, int64, time.Time) (bool, error) {
	return false, nil
}

func (a *academicoFalso) UmbralVigente(_ context.Context, tipo string, _ time.Time) (int, bool, error) {
	if a.umbrales != nil {
		valor, existe := a.umbrales[tipo]
		return valor, existe, nil
	}
	return 1, true, nil
}

func (a *academicoFalso) EtiquetaSeccion(context.Context, int64) (string, error) {
	return "1.° A", nil
}

type usuariosFalsos struct {
	administradorID int64
}

func (u *usuariosFalsos) AdministradorActivo(context.Context) (int64, error) {
	return u.administradorID, nil
}

type procesadorFalso struct{}

func (p *procesadorFalso) ProcesarTardanzas(context.Context) error { return nil }

func (p *procesadorFalso) ProcesarCierre(context.Context, time.Time, string) error { return nil }
