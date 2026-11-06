package alumnos

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"asistencia/internal/plataforma/fechas"
	"asistencia/internal/plataforma/intermediarios"
	vista "asistencia/web/plantillas/importacion"
)

const maximoArchivoImportacion = 5 * 1024 * 1024

func (m *Manejador) mostrarImportacion(respuesta http.ResponseWriter, peticion *http.Request) {
	m.renderizarCarga(respuesta, peticion, http.StatusOK, "", nil)
}

func (m *Manejador) procesarImportacion(respuesta http.ResponseWriter, peticion *http.Request) {
	limite := int64(maximoArchivoImportacion) + 1024*1024
	peticion.Body = http.MaxBytesReader(respuesta, peticion.Body, limite)
	if err := peticion.ParseMultipartForm(limite); err != nil {
		m.renderizarCarga(respuesta, peticion, http.StatusUnprocessableEntity,
			"No se pudo leer el archivo. Verifique que sea un CSV o XLSX válido.", nil)
		return
	}

	archivo, cabecera, err := peticion.FormFile("archivo")
	if err != nil {
		m.renderizarCarga(respuesta, peticion, http.StatusUnprocessableEntity,
			"Seleccione un archivo CSV o XLSX.", nil)
		return
	}
	defer archivo.Close()

	if cabecera.Size > maximoArchivoImportacion {
		m.renderizarCarga(respuesta, peticion, http.StatusUnprocessableEntity,
			"El archivo supera el tamaño máximo de 5 MB.", nil)
		return
	}
	contenido, err := io.ReadAll(io.LimitReader(archivo, maximoArchivoImportacion+1))
	if err != nil {
		m.errorInterno(respuesta, "leyendo el archivo de importación", err)
		return
	}
	if len(contenido) > maximoArchivoImportacion {
		m.renderizarCarga(respuesta, peticion, http.StatusUnprocessableEntity,
			"El archivo supera el tamaño máximo de 5 MB.", nil)
		return
	}

	resultado, err := m.importacion.Importar(peticion.Context(), intermediarios.UsuarioID(peticion.Context()), cabecera.Filename, contenido)
	if err != nil {
		m.responderErrorImportacion(respuesta, peticion, err)
		return
	}
	http.Redirect(respuesta, peticion,
		fmt.Sprintf("/alumnos/importaciones/%d?ok=importado", resultado.ID), http.StatusSeeOther)
}

func (m *Manejador) listarImportaciones(respuesta http.ResponseWriter, peticion *http.Request) {
	importaciones, err := m.importacion.ListarImportaciones(peticion.Context())
	if err != nil {
		m.errorInterno(respuesta, "listando las importaciones", err)
		return
	}
	m.renderizar(respuesta, peticion, http.StatusOK, vista.Historial(
		m.nombreUsuario(peticion.Context()),
		intermediarios.RolUsuario(peticion.Context()),
		m.tokenCSRF(peticion.Context()),
		aImportacionesVista(importaciones),
		mensajeCorrecto(peticion.URL.Query().Get("ok")),
	))
}

func (m *Manejador) detalleImportacion(respuesta http.ResponseWriter, peticion *http.Request) {
	id, ok := idDeRuta(peticion)
	if !ok {
		http.NotFound(respuesta, peticion)
		return
	}
	detalle, err := m.importacion.ObtenerImportacion(peticion.Context(), id)
	if errors.Is(err, ErrImportacionNoEncontrada) {
		http.NotFound(respuesta, peticion)
		return
	}
	if err != nil {
		m.errorInterno(respuesta, "obteniendo el detalle de la importación", err)
		return
	}
	m.renderizar(respuesta, peticion, http.StatusOK, vista.Detalle(
		m.nombreUsuario(peticion.Context()),
		intermediarios.RolUsuario(peticion.Context()),
		m.tokenCSRF(peticion.Context()),
		aDetalleImportacionVista(detalle),
	))
}

func (m *Manejador) renderizarCarga(respuesta http.ResponseWriter, peticion *http.Request, estado int, mensaje string, faltantes []string) {
	m.renderizar(respuesta, peticion, estado, vista.Carga(
		m.nombreUsuario(peticion.Context()),
		intermediarios.RolUsuario(peticion.Context()),
		m.tokenCSRF(peticion.Context()),
		m.importacionMaxFilas,
		mensaje,
		faltantes,
	))
}

func (m *Manejador) responderErrorImportacion(respuesta http.ResponseWriter, peticion *http.Request, err error) {
	var faltantes *ErrorColumnasFaltantes
	switch {
	case errors.As(err, &faltantes):
		m.renderizarCarga(respuesta, peticion, http.StatusUnprocessableEntity,
			"El archivo no tiene todas las columnas obligatorias.", faltantes.Columnas)
	case errors.Is(err, ErrDemasiadasFilas):
		m.renderizarCarga(respuesta, peticion, http.StatusUnprocessableEntity,
			fmt.Sprintf("El archivo supera el máximo de filas permitido (%d).", m.importacionMaxFilas), nil)
	case errors.Is(err, ErrArchivoVacio):
		m.renderizarCarga(respuesta, peticion, http.StatusUnprocessableEntity,
			"El archivo no contiene filas para importar.", nil)
	case errors.Is(err, ErrFormatoArchivo):
		m.renderizarCarga(respuesta, peticion, http.StatusUnprocessableEntity,
			"El archivo debe ser CSV (.csv) o Excel (.xlsx).", nil)
	case errors.Is(err, ErrArchivoMuyGrande):
		m.renderizarCarga(respuesta, peticion, http.StatusUnprocessableEntity,
			"El archivo supera el tamaño máximo de 5 MB.", nil)
	default:
		m.errorInterno(respuesta, "procesando la importación", err)
	}
}

func aImportacionesVista(importaciones []Importacion) []vista.Importacion {
	filas := make([]vista.Importacion, 0, len(importaciones))
	for _, importacion := range importaciones {
		filas = append(filas, vista.Importacion{
			ID:              importacion.ID,
			ArchivoNombre:   importacion.ArchivoNombre,
			FilasLeidas:     importacion.FilasLeidas,
			FilasCreadas:    importacion.FilasCreadas,
			FilasRechazadas: importacion.FilasRechazadas,
			CreadaEn:        fechas.Formatear(importacion.CreadaEn),
		})
	}
	return filas
}

func aDetalleImportacionVista(detalle ImportacionDetalle) vista.DetalleImportacion {
	rechazadas := make([]vista.FilaRechazada, 0, len(detalle.Rechazadas))
	for _, fila := range detalle.Rechazadas {
		rechazadas = append(rechazadas, vista.FilaRechazada{
			NumeroFila:    fila.NumeroFila,
			Identificador: fila.Identificador,
			Motivo:        MotivoImportacionLegible(fila.Motivo),
		})
	}
	importaciones := aImportacionesVista([]Importacion{detalle.Importacion})
	return vista.DetalleImportacion{
		Importacion: importaciones[0],
		Rechazadas:  rechazadas,
	}
}
