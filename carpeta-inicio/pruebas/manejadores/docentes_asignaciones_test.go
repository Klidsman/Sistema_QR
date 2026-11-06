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
	docentes      []autenticacion.Docente
	creado        *autenticacion.DatosDocente
	desactivado   int64
	claveCambiada int64
}

func (d *docentesFalso) ListarDocentes(context.Context) ([]autenticacion.Docente, error) {
	return d.docentes, nil
}

func (d *docentesFalso) CrearDocente(_ context.Context, _ int64, datos autenticacion.DatosDocente) (int64, error) {
	d.creado = &datos
	return 7, nil
}

func (d *docentesFalso) DesactivarDocente(_ context.Context, _, docenteID int64) error {
	d.desactivado = docenteID
	return nil
}

func (d *docentesFalso) RestablecerClave(_ context.Context, _, docenteID int64, _ string) error {
	d.claveCambiada = docenteID
	return nil
}

func nuevoManejadorDocentes(t *testing.T, docentes *docentesFalso) (*autenticacion.Manejador, *scs.SessionManager) {
	t.Helper()
	sesiones := scs.New()
	manejador := autenticacion.NuevoManejadorConDocentes(nil, docentes, sesiones, nil,
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
	falso := &docentesFalso{docentes: []autenticacion.Docente{{
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
	asignaciones []asistencia.AsignacionVista
	creada       *asistencia.DatosAsignacion
	cerrada      int64
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
