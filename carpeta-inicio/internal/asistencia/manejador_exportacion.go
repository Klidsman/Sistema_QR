package asistencia

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"asistencia/internal/plataforma/fechas"
	"asistencia/internal/plataforma/intermediarios"
	vista "asistencia/web/plantillas/asistencia"
)

func (m *Manejador) formularioExportacion(respuesta http.ResponseWriter, peticion *http.Request) {
	secciones, err := m.servicio.SeccionesDisponibles(peticion.Context())
	if err != nil {
		m.registro.Error("no se pudieron listar las secciones", "error", err)
		http.Error(respuesta, "No se pudo cargar la exportación", http.StatusInternalServerError)
		return
	}
	m.renderizar(respuesta, peticion, http.StatusOK, vista.Exportacion(
		m.nombreUsuario(peticion.Context()),
		intermediarios.RolUsuario(peticion.Context()),
		m.tokenCSRF(peticion.Context()),
		aSeccionesVista(secciones),
		fechas.FormatearISO(inicioDeMes(fechas.Hoy())),
		fechas.FormatearISO(fechas.Hoy()),
		"",
	))
}

func (m *Manejador) descargarExportacion(respuesta http.ResponseWriter, peticion *http.Request) {
	desde, hasta, err := rangoExportacion(peticion)
	if err != nil {
		m.renderizarErrorExportacion(respuesta, peticion, mensajeExportacion(err))
		return
	}
	seccionID, _ := strconv.ParseInt(peticion.URL.Query().Get("seccion_id"), 10, 64)
	cursoID, _ := strconv.ParseInt(peticion.URL.Query().Get("curso_id"), 10, 64)

	var registros []RegistroExportacion
	rol := intermediarios.RolUsuario(peticion.Context())
	if rol == intermediarios.RolDocente {
		registros, err = m.servicio.ExportarDocente(peticion.Context(), m.usuarioID(peticion.Context()), desde, hasta)
	} else {
		if seccionID <= 0 {
			m.renderizarErrorExportacion(respuesta, peticion,
				"Seleccione una sección para exportar.")
			return
		}
		registros, err = m.servicio.ExportarSeccion(peticion.Context(), seccionID, cursoID, desde, hasta)
	}
	if err != nil {
		m.renderizarErrorExportacion(respuesta, peticion, mensajeExportacion(err))
		return
	}

	contenido, err := GenerarCSV(registros)
	if err != nil {
		m.registro.Error("no se pudo generar el CSV", "error", err)
		http.Error(respuesta, "No se pudo generar el archivo", http.StatusInternalServerError)
		return
	}
	nombre := fmt.Sprintf("asistencia_%s_%s.csv",
		fechas.FormatearISO(desde), fechas.FormatearISO(hasta))
	respuesta.Header().Set("Content-Type", "text/csv; charset=utf-8")
	respuesta.Header().Set("Content-Disposition", `attachment; filename="`+nombre+`"`)
	respuesta.WriteHeader(http.StatusOK)
	_, _ = respuesta.Write(contenido)
}

func (m *Manejador) renderizarErrorExportacion(respuesta http.ResponseWriter, peticion *http.Request, mensaje string) {
	secciones, err := m.servicio.SeccionesDisponibles(peticion.Context())
	if err != nil {
		http.Error(respuesta, "No se pudo cargar la exportación", http.StatusInternalServerError)
		return
	}
	m.renderizar(respuesta, peticion, http.StatusUnprocessableEntity, vista.Exportacion(
		m.nombreUsuario(peticion.Context()),
		intermediarios.RolUsuario(peticion.Context()),
		m.tokenCSRF(peticion.Context()),
		aSeccionesVista(secciones),
		fechas.FormatearISO(inicioDeMes(fechas.Hoy())),
		fechas.FormatearISO(fechas.Hoy()),
		mensaje,
	))
}

func rangoExportacion(peticion *http.Request) (time.Time, time.Time, error) {
	textoDesde := strings.TrimSpace(peticion.URL.Query().Get("desde"))
	textoHasta := strings.TrimSpace(peticion.URL.Query().Get("hasta"))
	if textoDesde == "" {
		textoDesde = fechas.FormatearISO(inicioDeMes(fechas.Hoy()))
	}
	if textoHasta == "" {
		textoHasta = fechas.FormatearISO(fechas.Hoy())
	}
	desde, err := fechas.ParsearDia(textoDesde)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	hasta, err := fechas.ParsearDia(textoHasta)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if err := ValidarRangoExportacion(desde, hasta); err != nil {
		return time.Time{}, time.Time{}, err
	}
	return desde, hasta, nil
}

func inicioDeMes(fecha time.Time) time.Time {
	return time.Date(fecha.Year(), fecha.Month(), 1, 0, 0, 0, 0, fecha.Location())
}

func mensajeExportacion(err error) string {
	if errors.Is(err, ErrRangoFechasInvalido) {
		return "El rango de fechas no es válido."
	}
	return "No se pudo generar la exportación. Revise las fechas y la sección."
}
