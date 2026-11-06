package manejadores

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"testing"
	"time"

	"asistencia/internal/alumnos"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type importacionFalso struct {
	resultado  alumnos.Importacion
	err        error
	archivo    string
	contenido  []byte
	lista      []alumnos.Importacion
	detalle    alumnos.ImportacionDetalle
	detalleErr error
}

func (i *importacionFalso) Importar(ctx context.Context, usuarioID int64, archivoNombre string, contenido []byte) (alumnos.Importacion, error) {
	i.archivo = archivoNombre
	i.contenido = contenido
	if i.err != nil {
		return alumnos.Importacion{}, i.err
	}
	return i.resultado, nil
}

func (i *importacionFalso) ListarImportaciones(ctx context.Context) ([]alumnos.Importacion, error) {
	return i.lista, nil
}

func (i *importacionFalso) ObtenerImportacion(ctx context.Context, id int64) (alumnos.ImportacionDetalle, error) {
	if i.detalleErr != nil {
		return alumnos.ImportacionDetalle{}, i.detalleErr
	}
	return i.detalle, nil
}

func cuerpoConArchivo(nombreArchivo string, contenido []byte) (io.Reader, string) {
	var buffer bytes.Buffer
	escritor := multipart.NewWriter(&buffer)
	archivo, _ := escritor.CreateFormFile("archivo", nombreArchivo)
	_, _ = archivo.Write(contenido)
	_ = escritor.Close()
	return &buffer, escritor.FormDataContentType()
}

func TestImportacionManejador_MuestraFormularioDeCarga(t *testing.T) {
	manejador, sesiones := nuevoManejadorDePruebaConImportacion(t, &alumnosFalso{}, &qrFalso{}, &importacionFalso{})

	respuesta := ejecutar(manejador, peticionConSesion(t, sesiones, http.MethodGet, "/alumnos/importar", nil, "", "administrador"))

	require.Equal(t, http.StatusOK, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "Importar alumnos")
	assert.Contains(t, respuesta.Body.String(), "identificador")
}

func TestImportacionManejador_SoloAdministradorPuedeCargar(t *testing.T) {
	manejador, sesiones := nuevoManejadorDePruebaConImportacion(t, &alumnosFalso{}, &qrFalso{}, &importacionFalso{})

	respuesta := ejecutar(manejador, peticionConSesion(t, sesiones, http.MethodGet, "/alumnos/importar", nil, "", "director"))

	assert.Equal(t, http.StatusForbidden, respuesta.Code)
}

func TestImportacionManejador_ProcesaArchivoYRedirigeAlReporte(t *testing.T) {
	imp := &importacionFalso{resultado: alumnos.Importacion{ID: 15, ArchivoNombre: "alumnos.csv"}}
	manejador, sesiones := nuevoManejadorDePruebaConImportacion(t, &alumnosFalso{}, &qrFalso{}, imp)
	cuerpo, tipo := cuerpoConArchivo("alumnos.csv", []byte("identificador,nombres\n70000001,Ana\n"))

	respuesta := ejecutar(manejador, peticionConSesion(t, sesiones, http.MethodPost, "/alumnos/importar", cuerpo, tipo, "administrador"))

	require.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Equal(t, "/alumnos/importaciones/15?ok=importado", respuesta.Header().Get("Location"))
	assert.Equal(t, "alumnos.csv", imp.archivo)
	assert.Contains(t, string(imp.contenido), "70000001")
}

func TestImportacionManejador_InformaElLimiteDeFilas(t *testing.T) {
	imp := &importacionFalso{err: alumnos.ErrDemasiadasFilas}
	manejador, sesiones := nuevoManejadorDePruebaConImportacion(t, &alumnosFalso{}, &qrFalso{}, imp)
	cuerpo, tipo := cuerpoConArchivo("alumnos.csv", []byte("contenido"))

	respuesta := ejecutar(manejador, peticionConSesion(t, sesiones, http.MethodPost, "/alumnos/importar", cuerpo, tipo, "administrador"))

	require.Equal(t, http.StatusUnprocessableEntity, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "supera el máximo de filas")
}

func TestImportacionManejador_HistorialYDetalleDisponiblesParaDirector(t *testing.T) {
	imp := &importacionFalso{
		lista: []alumnos.Importacion{{
			ID: 15, ArchivoNombre: "alumnos.csv", FilasLeidas: 4, FilasCreadas: 2, FilasRechazadas: 2, CreadaEn: time.Now(),
		}},
		detalle: alumnos.ImportacionDetalle{
			Importacion: alumnos.Importacion{ID: 15, ArchivoNombre: "alumnos.csv", FilasLeidas: 4, FilasCreadas: 2, FilasRechazadas: 2},
			Rechazadas: []alumnos.FilaRechazada{{
				NumeroFila: 3, Identificador: "70000001", Motivo: alumnos.MotivoCampoObligatorioVacio,
			}},
		},
	}
	manejador, sesiones := nuevoManejadorDePruebaConImportacion(t, &alumnosFalso{}, &qrFalso{}, imp)

	historial := ejecutar(manejador, peticionConSesion(t, sesiones, http.MethodGet, "/alumnos/importaciones", nil, "", "director"))
	require.Equal(t, http.StatusOK, historial.Code)
	assert.Contains(t, historial.Body.String(), "alumnos.csv")
	assert.Contains(t, historial.Body.String(), "/alumnos/importaciones/15")

	detalle := ejecutar(manejador, peticionConSesion(t, sesiones, http.MethodGet, "/alumnos/importaciones/15", nil, "", "director"))
	require.Equal(t, http.StatusOK, detalle.Code)
	assert.Contains(t, detalle.Body.String(), "Campo obligatorio vacío")
	assert.Contains(t, detalle.Body.String(), "70000001")
}

func TestImportacionManejador_DirectorNoPuedeImportar(t *testing.T) {
	manejador, sesiones := nuevoManejadorDePruebaConImportacion(t, &alumnosFalso{}, &qrFalso{}, &importacionFalso{})
	cuerpo, tipo := cuerpoConArchivo("alumnos.csv", []byte("contenido"))

	respuesta := ejecutar(manejador, peticionConSesion(t, sesiones, http.MethodPost, "/alumnos/importar", cuerpo, tipo, "director"))

	assert.Equal(t, http.StatusForbidden, respuesta.Code)
}
