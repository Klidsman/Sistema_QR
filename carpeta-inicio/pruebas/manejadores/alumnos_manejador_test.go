package manejadores

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"asistencia/internal/academico"
	"asistencia/internal/alumnos"
	"asistencia/internal/plataforma/almacenamiento"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type seccionesFalso struct {
	secciones []academico.Seccion
}

func (s seccionesFalso) BuscarSeccion(ctx context.Context, anioID int64, grado, seccion string) (int64, error) {
	return s.secciones[0].ID, nil
}

func (s seccionesFalso) ListarSecciones(ctx context.Context) ([]academico.Seccion, error) {
	return s.secciones, nil
}

func (s seccionesFalso) BuscarSeccionPorID(ctx context.Context, id int64) (academico.Seccion, error) {
	return s.secciones[0], nil
}

type alumnosFalso struct {
	listado             []alumnos.AlumnoListado
	alumno              alumnos.Alumno
	obtenerErr          error
	crearID             int64
	crearErr            error
	datosRecibidos      alumnos.DatosAlumno
	desactivado         bool
	fotoActualizada     string
	historialSecciones  []alumnos.HistorialSeccion
	cambiarSeccionID    int64
	cambiarSeccionFecha time.Time
	cambiarSeccionErr   error
	salidaTipo          string
	salidaFecha         time.Time
	salidaErr           error
	carnets             []alumnos.AlumnoCarnet
	carnetSeleccion     alumnos.SeleccionCarnets
}

func (a *alumnosFalso) Listar(ctx context.Context, filtros alumnos.FiltrosListado) ([]alumnos.AlumnoListado, error) {
	return a.listado, nil
}

func (a *alumnosFalso) Obtener(ctx context.Context, id int64) (alumnos.Alumno, error) {
	if a.obtenerErr != nil {
		return alumnos.Alumno{}, a.obtenerErr
	}
	return a.alumno, nil
}

func (a *alumnosFalso) Crear(ctx context.Context, usuarioID int64, datos alumnos.DatosAlumno) (int64, error) {
	a.datosRecibidos = datos
	if a.crearErr != nil {
		return 0, a.crearErr
	}
	return a.crearID, nil
}

func (a *alumnosFalso) Actualizar(ctx context.Context, usuarioID, alumnoID int64, datos alumnos.DatosAlumno) error {
	a.datosRecibidos = datos
	return nil
}

func (a *alumnosFalso) ActualizarFoto(ctx context.Context, usuarioID, alumnoID int64, clave string) error {
	a.fotoActualizada = clave
	return nil
}

func (a *alumnosFalso) Desactivar(ctx context.Context, usuarioID, alumnoID int64) error {
	a.desactivado = true
	return nil
}

func (a *alumnosFalso) HistorialSecciones(ctx context.Context, alumnoID int64) ([]alumnos.HistorialSeccion, error) {
	return a.historialSecciones, nil
}

func (a *alumnosFalso) CambiarSeccion(ctx context.Context, usuarioID, alumnoID, seccionID int64, fecha time.Time) error {
	a.cambiarSeccionID = seccionID
	a.cambiarSeccionFecha = fecha
	return a.cambiarSeccionErr
}

func (a *alumnosFalso) RegistrarSalida(ctx context.Context, usuarioID, alumnoID int64, tipo string, fecha time.Time) error {
	a.salidaTipo = tipo
	a.salidaFecha = fecha
	return a.salidaErr
}

func (a *alumnosFalso) ListarCarnets(ctx context.Context, seleccion alumnos.SeleccionCarnets) (alumnos.LoteCarnets, error) {
	a.carnetSeleccion = seleccion
	return alumnos.PrepararLote(a.carnets), nil
}

type qrFalso struct {
	credencial     *alumnos.CredencialQR
	historial      []alumnos.CredencialQR
	emitirErr      error
	reemplazarErr  error
	motivoRecibido string
}

func (q *qrFalso) CredencialActiva(ctx context.Context, alumnoID int64) (alumnos.CredencialQR, error) {
	if q.credencial == nil {
		return alumnos.CredencialQR{}, alumnos.ErrQRSinActivo
	}
	return *q.credencial, nil
}

func (q *qrFalso) HistorialQR(ctx context.Context, alumnoID int64) ([]alumnos.CredencialQR, error) {
	return q.historial, nil
}

func (q *qrFalso) EmitirQR(ctx context.Context, usuarioID, alumnoID int64) (alumnos.CredencialQR, error) {
	if q.emitirErr != nil {
		return alumnos.CredencialQR{}, q.emitirErr
	}
	return alumnos.CredencialQR{ID: 1, Token: "token-de-prueba", Estado: alumnos.CredencialActiva, EmitidaEn: time.Now()}, nil
}

func (q *qrFalso) ReemplazarQR(ctx context.Context, usuarioID, alumnoID int64, motivo string) (alumnos.CredencialQR, error) {
	q.motivoRecibido = motivo
	if q.reemplazarErr != nil {
		return alumnos.CredencialQR{}, q.reemplazarErr
	}
	return alumnos.CredencialQR{ID: 2, Token: "token-nuevo", Estado: alumnos.CredencialActiva, EmitidaEn: time.Now()}, nil
}

func nuevoManejadorDePrueba(t *testing.T, servicio *alumnosFalso, qr *qrFalso) (*alumnos.Manejador, *scs.SessionManager) {
	t.Helper()
	return nuevoManejadorDePruebaConImportacion(t, servicio, qr, &importacionFalso{})
}

func nuevoManejadorDePruebaConImportacion(t *testing.T, servicio *alumnosFalso, qr *qrFalso, importacion *importacionFalso) (*alumnos.Manejador, *scs.SessionManager) {
	t.Helper()
	sesiones := scs.New()
	manejador := alumnos.NuevoManejador(alumnos.Dependencias{
		Alumnos: servicio,
		QR:      qr,
		Secciones: seccionesFalso{secciones: []academico.Seccion{
			{ID: 1, Nombre: "A", Grado: "3.°", AnioEscolar: "2026", GradoID: 1, AnioEscolarID: 1},
		}},
		Almacen:      almacenamiento.NuevoAlmacenDisco(t.TempDir()),
		Sesiones:     sesiones,
		Registro:     slog.New(slog.NewTextHandler(io.Discard, nil)),
		FotoMaximaMB: 2,
		Importacion:  importacion,
	})
	return manejador, sesiones
}

func peticionConSesion(t *testing.T, sesiones *scs.SessionManager, metodo, ruta string, cuerpo io.Reader, tipoContenido string, rol string) *http.Request {
	t.Helper()
	peticion := httptest.NewRequest(metodo, ruta, cuerpo)
	if tipoContenido != "" {
		peticion.Header.Set("Content-Type", tipoContenido)
	}
	ctx, err := sesiones.Load(peticion.Context(), "token-de-prueba")
	require.NoError(t, err)
	sesiones.Put(ctx, "usuario_id", int64(9))
	sesiones.Put(ctx, "rol", rol)
	sesiones.Put(ctx, "nombre_completo", "Admin de prueba")
	return peticion.WithContext(ctx)
}

func cuerpoMultipart(campos map[string]string) (io.Reader, string) {
	var buffer bytes.Buffer
	escritor := multipart.NewWriter(&buffer)
	for clave, valor := range campos {
		_ = escritor.WriteField(clave, valor)
	}
	_ = escritor.Close()
	return &buffer, escritor.FormDataContentType()
}

func ejecutar(manejador *alumnos.Manejador, peticion *http.Request) *httptest.ResponseRecorder {
	enrutador := chi.NewRouter()
	manejador.Montar(enrutador)
	respuesta := httptest.NewRecorder()
	enrutador.ServeHTTP(respuesta, peticion)
	return respuesta
}

func alumnoDePrueba() alumnos.Alumno {
	return alumnos.Alumno{
		ID:            7,
		Identificador: "00123456",
		Nombres:       "José",
		Apellidos:     "Núñez",
		Estado:        alumnos.EstadoActivo,
		SeccionID:     1,
		Grado:         "3.°",
		Seccion:       "A",
		Turno:         "manana",
		AnioEscolar:   "2026",
		FotoClave:     "alumnos/7/foto.jpg",
	}
}

func TestAlumnosManejador_ListadoVisibleParaDirector(t *testing.T) {
	servicio := &alumnosFalso{listado: []alumnos.AlumnoListado{{
		ID: 7, Identificador: "00123456", Nombres: "José", Apellidos: "Núñez",
		Estado: alumnos.EstadoActivo, Grado: "3.°", Seccion: "A", AnioEscolar: "2026",
	}}}
	manejador, sesiones := nuevoManejadorDePrueba(t, servicio, &qrFalso{})

	respuesta := ejecutar(manejador, peticionConSesion(t, sesiones, http.MethodGet, "/alumnos", nil, "", "director"))

	require.Equal(t, http.StatusOK, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "Núñez, José")
	assert.NotContains(t, respuesta.Body.String(), "Nuevo alumno")
}

func TestAlumnosManejador_CrearAlumno(t *testing.T) {
	servicio := &alumnosFalso{crearID: 7}
	manejador, sesiones := nuevoManejadorDePrueba(t, servicio, &qrFalso{})
	cuerpo, tipo := cuerpoMultipart(map[string]string{
		"identificador": "00123456",
		"nombres":       "José",
		"apellidos":     "Núñez",
		"seccion_id":    "1",
	})

	respuesta := ejecutar(manejador, peticionConSesion(t, sesiones, http.MethodPost, "/alumnos", cuerpo, tipo, "administrador"))

	require.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Equal(t, "/alumnos/7?ok=creado", respuesta.Header().Get("Location"))
	assert.Equal(t, "00123456", servicio.datosRecibidos.Identificador)
	assert.Equal(t, int64(1), servicio.datosRecibidos.SeccionID)
}

func TestAlumnosManejador_IdentificadorRepetido(t *testing.T) {
	servicio := &alumnosFalso{crearErr: alumnos.ErrIdentificadorExistente}
	manejador, sesiones := nuevoManejadorDePrueba(t, servicio, &qrFalso{})
	cuerpo, tipo := cuerpoMultipart(map[string]string{
		"identificador": "00123456",
		"nombres":       "José",
		"apellidos":     "Núñez",
		"seccion_id":    "1",
	})

	respuesta := ejecutar(manejador, peticionConSesion(t, sesiones, http.MethodPost, "/alumnos", cuerpo, tipo, "administrador"))

	require.Equal(t, http.StatusUnprocessableEntity, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "Ya existe un alumno con este DNI o código")
}

func TestAlumnosManejador_DirectorNoPuedeCrear(t *testing.T) {
	servicio := &alumnosFalso{crearID: 7}
	manejador, sesiones := nuevoManejadorDePrueba(t, servicio, &qrFalso{})
	cuerpo, tipo := cuerpoMultipart(map[string]string{"identificador": "00123456"})

	respuesta := ejecutar(manejador, peticionConSesion(t, sesiones, http.MethodPost, "/alumnos", cuerpo, tipo, "director"))

	assert.Equal(t, http.StatusForbidden, respuesta.Code)
}

func TestAlumnosManejador_CrearAlumnoConFotoInvalidaLoGuardaSinFoto(t *testing.T) {
	servicio := &alumnosFalso{crearID: 7}
	manejador, sesiones := nuevoManejadorDePrueba(t, servicio, &qrFalso{})

	var buffer bytes.Buffer
	escritor := multipart.NewWriter(&buffer)
	_ = escritor.WriteField("identificador", "00123456")
	_ = escritor.WriteField("nombres", "José")
	_ = escritor.WriteField("apellidos", "Núñez")
	_ = escritor.WriteField("seccion_id", "1")
	archivo, err := escritor.CreateFormFile("foto", "foto.txt")
	require.NoError(t, err)
	_, err = archivo.Write([]byte("esto no es una imagen"))
	require.NoError(t, err)
	require.NoError(t, escritor.Close())

	respuesta := ejecutar(manejador, peticionConSesion(t, sesiones, http.MethodPost, "/alumnos", &buffer, escritor.FormDataContentType(), "administrador"))

	require.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Equal(t, "/alumnos/7?ok=creado&error=foto_invalida", respuesta.Header().Get("Location"))
}

func TestAlumnosManejador_FichaMuestraQR(t *testing.T) {
	credencial := &alumnos.CredencialQR{ID: 1, Token: "token", Estado: alumnos.CredencialActiva, EmitidaEn: time.Now()}
	servicio := &alumnosFalso{alumno: alumnoDePrueba()}
	manejador, sesiones := nuevoManejadorDePrueba(t, servicio, &qrFalso{credencial: credencial})

	respuesta := ejecutar(manejador, peticionConSesion(t, sesiones, http.MethodGet, "/alumnos/7", nil, "", "administrador"))

	require.Equal(t, http.StatusOK, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "Código QR")
	assert.Contains(t, respuesta.Body.String(), "/alumnos/7/qr/imagen")
	assert.Contains(t, respuesta.Body.String(), "Editar")
}

func TestAlumnosManejador_DirectorNoVeAcciones(t *testing.T) {
	servicio := &alumnosFalso{alumno: alumnoDePrueba()}
	manejador, sesiones := nuevoManejadorDePrueba(t, servicio, &qrFalso{})

	respuesta := ejecutar(manejador, peticionConSesion(t, sesiones, http.MethodGet, "/alumnos/7", nil, "", "director"))

	require.Equal(t, http.StatusOK, respuesta.Code)
	assert.NotContains(t, respuesta.Body.String(), ">Editar<")
}

func TestAlumnosManejador_EmitirYReemplazarQR(t *testing.T) {
	servicio := &alumnosFalso{alumno: alumnoDePrueba()}
	qr := &qrFalso{}
	manejador, sesiones := nuevoManejadorDePrueba(t, servicio, qr)

	emitir := ejecutar(manejador, peticionConSesion(t, sesiones, http.MethodPost, "/alumnos/7/qr", strings.NewReader(""), "", "administrador"))
	require.Equal(t, http.StatusSeeOther, emitir.Code)
	assert.Equal(t, "/alumnos/7?ok=qr_emitido", emitir.Header().Get("Location"))

	cuerpo, tipo := cuerpoMultipart(map[string]string{"motivo": "robo"})
	reemplazar := ejecutar(manejador, peticionConSesion(t, sesiones, http.MethodPost, "/alumnos/7/qr/reemplazo", cuerpo, tipo, "administrador"))
	require.Equal(t, http.StatusSeeOther, reemplazar.Code)
	assert.Equal(t, "/alumnos/7?ok=qr_reemplazado", reemplazar.Header().Get("Location"))
	assert.Equal(t, "robo", qr.motivoRecibido)
}

func TestAlumnosManejador_ImagenQRSoloContieneElToken(t *testing.T) {
	credencial := &alumnos.CredencialQR{ID: 1, Token: "token", Estado: alumnos.CredencialActiva, EmitidaEn: time.Now()}
	servicio := &alumnosFalso{alumno: alumnoDePrueba()}
	manejador, sesiones := nuevoManejadorDePrueba(t, servicio, &qrFalso{credencial: credencial})

	respuesta := ejecutar(manejador, peticionConSesion(t, sesiones, http.MethodGet, "/alumnos/7/qr/imagen", nil, "", "director"))

	require.Equal(t, http.StatusOK, respuesta.Code)
	assert.Equal(t, "image/png", respuesta.Header().Get("Content-Type"))
	assert.True(t, bytes.HasPrefix(respuesta.Body.Bytes(), []byte("\x89PNG")))
}

func TestAlumnosManejador_DirectorNoPuedeEmitirQR(t *testing.T) {
	servicio := &alumnosFalso{alumno: alumnoDePrueba()}
	manejador, sesiones := nuevoManejadorDePrueba(t, servicio, &qrFalso{})

	respuesta := ejecutar(manejador, peticionConSesion(t, sesiones, http.MethodPost, "/alumnos/7/qr", strings.NewReader(""), "", "director"))

	assert.Equal(t, http.StatusForbidden, respuesta.Code)
}
