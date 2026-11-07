package manejadores

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"asistencia/internal/academico"
	"asistencia/internal/asistencia"
	"asistencia/internal/autenticacion"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type docentesFalso struct {
	docentes      []autenticacion.Cuenta
	creado        *autenticacion.DatosCuenta
	rolCreado     string
	desactivado   int64
	claveCambiada int64
}

func (d *docentesFalso) ListarCuentasPorRol(_ context.Context, _ string) ([]autenticacion.Cuenta, error) {
	return d.docentes, nil
}

func (d *docentesFalso) CrearCuenta(_ context.Context, _ int64, rol string, datos autenticacion.DatosCuenta) (int64, error) {
	d.creado = &datos
	d.rolCreado = rol
	return 7, nil
}

func (d *docentesFalso) DesactivarCuenta(_ context.Context, _, cuentaID int64, _ string) error {
	d.desactivado = cuentaID
	return nil
}

func (d *docentesFalso) RestablecerClaveCuenta(_ context.Context, _, cuentaID int64, _, _ string) error {
	d.claveCambiada = cuentaID
	return nil
}

func nuevoManejadorDocentes(t *testing.T, docentes *docentesFalso) (*autenticacion.Manejador, *scs.SessionManager) {
	t.Helper()
	sesiones := scs.New()
	manejador := autenticacion.NuevoManejadorConCuentas(nil, docentes, sesiones, nil,
		slog.New(slog.NewTextHandler(io.Discard, nil)))
	return manejador, sesiones
}

func ejecutarDocentes(manejador *autenticacion.Manejador, peticion *http.Request) *httptest.ResponseRecorder {
	enrutador := chi.NewRouter()
	manejador.Montar(enrutador)
	respuesta := httptest.NewRecorder()
	enrutador.ServeHTTP(respuesta, peticion)
	return respuesta
}

func peticionPostFormulario(t *testing.T, sesiones *scs.SessionManager, ruta string, valores url.Values, rol string) *http.Request {
	t.Helper()
	peticion := peticionConSesion(t, sesiones, http.MethodPost, ruta,
		urlValuesReader(valores), "application/x-www-form-urlencoded", rol)
	return peticion
}

func urlValuesReader(valores url.Values) io.Reader {
	return stringReader(valores.Encode())
}

func stringReader(texto string) io.Reader {
	return &textoReader{texto: texto}
}

type textoReader struct {
	texto string
}

func (r *textoReader) Read(destino []byte) (int, error) {
	if r.texto == "" {
		return 0, io.EOF
	}
	n := copy(destino, r.texto)
	r.texto = r.texto[n:]
	return n, nil
}

func TestDocentes_ListadoVisibleParaAdministradorYDirector(t *testing.T) {
	falso := &docentesFalso{docentes: []autenticacion.Cuenta{{
		ID: 3, Nombre: "Docente Uno", Correo: "docente@colegio.pe", Activo: true, Rol: "docente",
	}}}
	manejador, sesiones := nuevoManejadorDocentes(t, falso)

	for _, rol := range []string{"administrador", "director"} {
		peticion := peticionConSesion(t, sesiones, http.MethodGet, "/docentes", nil, "", rol)
		respuesta := ejecutarDocentes(manejador, peticion)
		assert.Equal(t, http.StatusOK, respuesta.Code, rol)
		assert.Contains(t, respuesta.Body.String(), "Docente Uno")
	}
}

func TestDocentes_AdministradorCreaYDirectorNoPuede(t *testing.T) {
	falso := &docentesFalso{}
	manejador, sesiones := nuevoManejadorDocentes(t, falso)

	peticion := peticionPostFormulario(t, sesiones, "/docentes", url.Values{
		"nombre": {"Docente Nuevo"}, "correo": {"nuevo@colegio.pe"}, "contrasenia": {"clave-segura"},
	}, "administrador")
	respuesta := ejecutarDocentes(manejador, peticion)
	require.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Contains(t, respuesta.Header().Get("Location"), "ok=creado")
	require.NotNil(t, falso.creado)
	assert.Equal(t, "nuevo@colegio.pe", falso.creado.Correo)

	peticion = peticionPostFormulario(t, sesiones, "/docentes", url.Values{
		"nombre": {"Otro"}, "correo": {"otro@colegio.pe"}, "contrasenia": {"clave-segura"},
	}, "director")
	respuesta = ejecutarDocentes(manejador, peticion)
	assert.Equal(t, http.StatusForbidden, respuesta.Code)
}

func TestDocentes_AdministradorRestableceYDesactiva(t *testing.T) {
	falso := &docentesFalso{}
	manejador, sesiones := nuevoManejadorDocentes(t, falso)

	peticion := peticionPostFormulario(t, sesiones, "/docentes/9/clave",
		url.Values{"contrasenia": {"nueva-clave"}}, "administrador")
	respuesta := ejecutarDocentes(manejador, peticion)
	require.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Equal(t, int64(9), falso.claveCambiada)

	peticion = peticionPostFormulario(t, sesiones, "/docentes/9/desactivacion",
		url.Values{}, "administrador")
	respuesta = ejecutarDocentes(manejador, peticion)
	require.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Equal(t, int64(9), falso.desactivado)
}

type asistenciaFalsa struct {
	asignaciones      []asistencia.AsignacionVista
	creada            *asistencia.DatosAsignacion
	cerrada           int64
	clases            []asistencia.ClaseAsignada
	detalle           asistencia.ClaseDetalle
	guardados         []asistencia.CambioRegistro
	marcarCreados     int
	resumen           []asistencia.ResumenAsignacion
	alumnoQR          asistencia.AlumnoClase
	qrErr             error
	consulta          []asistencia.ConsultaRegistro
	corregido         *asistencia.ConsultaRegistro
	correccionErr     error
	plazo             asistencia.ParametrosAsistencia
	plazoNuevo        int
	bloqueRecibido    int
	exportados        []asistencia.RegistroExportacion
	exportarDocenteID int64
	exportarSeccionID int64
}

func (a *asistenciaFalsa) ListarAsignaciones(context.Context) ([]asistencia.AsignacionVista, error) {
	return a.asignaciones, nil
}

func (a *asistenciaFalsa) CrearAsignacion(_ context.Context, _ int64, datos asistencia.DatosAsignacion) (int64, error) {
	a.creada = &datos
	return 4, nil
}

func (a *asistenciaFalsa) CerrarAsignacion(_ context.Context, _, asignacionID int64, _ time.Time) error {
	a.cerrada = asignacionID
	return nil
}

func (a *asistenciaFalsa) DocentesDisponibles(context.Context) ([]autenticacion.Docente, error) {
	return []autenticacion.Docente{{ID: 3, Nombre: "Docente Uno", Activo: true, Rol: "docente"}}, nil
}

func (a *asistenciaFalsa) SeccionesDisponibles(context.Context) ([]academico.Seccion, error) {
	return []academico.Seccion{{ID: 5, Nombre: "A", Grado: "3.°", AnioEscolar: "2026"}}, nil
}

func (a *asistenciaFalsa) CursosDisponibles(context.Context) ([]academico.Curso, error) {
	return []academico.Curso{{ID: 2, Nombre: "Matemática", Activo: true}}, nil
}

func (a *asistenciaFalsa) ClasesAsignadas(_ context.Context, _ int64, _ time.Time) ([]asistencia.ClaseAsignada, error) {
	return a.clases, nil
}

func (a *asistenciaFalsa) AsignacionesDelDia(_ context.Context, _ int64, _ time.Time) ([]asistencia.ResumenAsignacion, error) {
	return a.resumen, nil
}

func (a *asistenciaFalsa) AlumnoPorTokenEnClase(_ context.Context, _ int64, _ bool, _ time.Time, _, _ int64, _ int, _ string) (asistencia.AlumnoClase, error) {
	return a.alumnoQR, a.qrErr
}

func (a *asistenciaFalsa) ListarClase(_ context.Context, _ int64, _ bool, _ time.Time, _, _ int64, _ int) (asistencia.ClaseDetalle, error) {
	return a.detalle, nil
}

func (a *asistenciaFalsa) GuardarRegistros(_ context.Context, _ int64, _ bool, _ time.Time, _, _ int64, bloque int, cambios []asistencia.CambioRegistro) error {
	a.guardados = cambios
	a.bloqueRecibido = bloque
	return nil
}

func (a *asistenciaFalsa) MarcarTodosPresentes(_ context.Context, _ int64, _ bool, _ time.Time, _, _ int64, _ int) (int, error) {
	return a.marcarCreados, nil
}

func (a *asistenciaFalsa) ConsultaRegistros(_ context.Context, _ time.Time, _ int64) ([]asistencia.ConsultaRegistro, error) {
	return a.consulta, nil
}

func (a *asistenciaFalsa) CorregirRegistro(_ context.Context, _, _ int64, estado, observacion, motivo string) error {
	if a.correccionErr != nil {
		return a.correccionErr
	}
	a.corregido = &asistencia.ConsultaRegistro{Registro: asistencia.RegistroDocente{Estado: estado, Observacion: observacion}}
	return nil
}

func (a *asistenciaFalsa) PlazoEdicion(context.Context) (asistencia.ParametrosAsistencia, error) {
	return a.plazo, nil
}

func (a *asistenciaFalsa) ConfigurarPlazoJustificacion(_ context.Context, _ int64, _ int) error {
	return nil
}

func (a *asistenciaFalsa) ConfigurarPlazo(_ context.Context, _ int64, horas int) error {
	a.plazoNuevo = horas
	return nil
}

func (a *asistenciaFalsa) ExportarDocente(_ context.Context, docenteID int64, _, _ time.Time) ([]asistencia.RegistroExportacion, error) {
	a.exportarDocenteID = docenteID
	return a.exportados, nil
}

func (a *asistenciaFalsa) ExportarSeccion(_ context.Context, seccionID, _ int64, _, _ time.Time) ([]asistencia.RegistroExportacion, error) {
	a.exportarSeccionID = seccionID
	return a.exportados, nil
}

func nuevoManejadorAsistencia(t *testing.T, falso *asistenciaFalsa) (*asistencia.Manejador, *scs.SessionManager) {
	t.Helper()
	sesiones := scs.New()
	manejador := asistencia.NuevoManejador(asistencia.Dependencias{
		Servicio: falso,
		Sesiones: sesiones,
		Registro: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	return manejador, sesiones
}

func ejecutarAsistencia(manejador *asistencia.Manejador, peticion *http.Request) *httptest.ResponseRecorder {
	enrutador := chi.NewRouter()
	manejador.Montar(enrutador)
	respuesta := httptest.NewRecorder()
	enrutador.ServeHTTP(respuesta, peticion)
	return respuesta
}

func TestAsignaciones_ListaYCrea(t *testing.T) {
	falso := &asistenciaFalsa{asignaciones: []asistencia.AsignacionVista{{
		AsignacionDocente: asistencia.AsignacionDocente{ID: 4},
		Docente:           "Docente Uno",
		Curso:             "Matemática",
		Seccion:           "3.° A (2026)",
		Vigente:           true,
	}}}
	manejador, sesiones := nuevoManejadorAsistencia(t, falso)

	peticion := peticionConSesion(t, sesiones, http.MethodGet, "/asistencia/asignaciones", nil, "", "director")
	respuesta := ejecutarAsistencia(manejador, peticion)
	assert.Equal(t, http.StatusOK, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "Docente Uno")

	peticion = peticionPostFormulario(t, sesiones, "/asistencia/asignaciones", url.Values{
		"docente_id":   {"3"},
		"curso_id":     {"2"},
		"seccion_id":   {"5"},
		"fecha_inicio": {"2026-11-06"},
		"titular":      {"si"},
	}, "administrador")
	respuesta = ejecutarAsistencia(manejador, peticion)
	require.Equal(t, http.StatusSeeOther, respuesta.Code)
	require.NotNil(t, falso.creada)
	assert.Equal(t, int64(3), falso.creada.DocenteID)
	assert.True(t, falso.creada.Titular)
}

func TestAsignaciones_DirectorNoPuedeEscribir(t *testing.T) {
	manejador, sesiones := nuevoManejadorAsistencia(t, &asistenciaFalsa{})
	peticion := peticionPostFormulario(t, sesiones, "/asistencia/asignaciones",
		url.Values{}, "director")

	respuesta := ejecutarAsistencia(manejador, peticion)

	assert.Equal(t, http.StatusForbidden, respuesta.Code)
}
