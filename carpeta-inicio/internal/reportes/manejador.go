package reportes

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"asistencia/internal/plataforma/fechas"
	"asistencia/internal/plataforma/intermediarios"
	vista "asistencia/web/plantillas/reportes"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
)

// ServicioReportes son las operaciones que usa el manejador.
type ServicioReportes interface {
	Diario(ctx context.Context, filtros Filtros) ([]FilaDiaria, error)
	Panel(ctx context.Context, filtros Filtros) (PanelDelDia, error)
	Resumen(ctx context.Context, filtros Filtros) ([]ResumenAlumno, error)
	Reiteracion(ctx context.Context, filtros Filtros) ([]Reiteracion, error)
}

// DependenciasManejador reúne lo necesario para atender reportes.
type DependenciasManejador struct {
	Servicio ServicioReportes
	Sesiones *scs.SessionManager
	Registro *slog.Logger
}

// Manejador atiende los reportes.
type Manejador struct {
	servicio ServicioReportes
	sesiones *scs.SessionManager
	registro *slog.Logger
}

// NuevoManejador crea el manejador.
func NuevoManejador(dependencias DependenciasManejador) *Manejador {
	return &Manejador{servicio: dependencias.Servicio, sesiones: dependencias.Sesiones, registro: dependencias.Registro}
}

// Montar registra las rutas de reportes.
func (m *Manejador) Montar(enrutador chi.Router) {
	enrutador.Group(func(autorizados chi.Router) {
		autorizados.Use(intermediarios.RequerirRol(m.sesiones,
			intermediarios.RolAdministrador, intermediarios.RolDirector, intermediarios.RolDocente))
		autorizados.Get("/reportes", func(respuesta http.ResponseWriter, peticion *http.Request) {
			http.Redirect(respuesta, peticion, "/reportes/diario", http.StatusSeeOther)
		})
		autorizados.Get("/reportes/diario", m.diario)
		autorizados.Get("/reportes/panel", m.panelPagina)
		autorizados.Get("/reportes/panel/indicadores", m.panelIndicadores)
		autorizados.Get("/reportes/resumen", m.resumen)
		autorizados.Get("/reportes/reiteracion", m.reiteracion)
		autorizados.Get("/reportes/diario.xlsx", m.diarioXLSX)
		autorizados.Get("/reportes/diario.pdf", m.diarioPDF)
		autorizados.Get("/reportes/resumen.xlsx", m.resumenXLSX)
		autorizados.Get("/reportes/resumen.pdf", m.resumenPDF)
	})
}

func (m *Manejador) diario(respuesta http.ResponseWriter, peticion *http.Request) {
	ctx := peticion.Context()
	filtros := m.filtros(peticion, true)
	filas, err := m.servicio.Diario(ctx, filtros)
	if err != nil {
		m.errorInterno(respuesta, "calculando el reporte diario", err)
		return
	}
	panel, err := m.servicio.Panel(ctx, filtros)
	if err != nil {
		m.errorInterno(respuesta, "calculando el panel del día", err)
		return
	}
	m.renderizar(respuesta, peticion, vista.DiarioReporte(vista.Diario{
		Nombre:    m.nombreUsuario(ctx),
		Rol:       intermediarios.RolUsuario(ctx),
		TokenCSRF: m.tokenCSRF(ctx),
		Filtros:   aFiltrosVista(filtros),
		Filas:     aFilasVista(filas),
		Panel:     aPanelVista(panel),
		Fecha:     fechas.FormatearISO(filtros.Desde),
	}))
}

func (m *Manejador) panelPagina(respuesta http.ResponseWriter, peticion *http.Request) {
	ctx := peticion.Context()
	filtros := m.filtros(peticion, true)
	panel, err := m.servicio.Panel(ctx, filtros)
	if err != nil {
		m.errorInterno(respuesta, "calculando el panel", err)
		return
	}
	m.renderizar(respuesta, peticion, vista.PanelReporte(vista.Diario{
		Nombre:    m.nombreUsuario(ctx),
		Rol:       intermediarios.RolUsuario(ctx),
		TokenCSRF: m.tokenCSRF(ctx),
		Filtros:   aFiltrosVista(filtros),
		Panel:     aPanelVista(panel),
	}))
}

func (m *Manejador) panelIndicadores(respuesta http.ResponseWriter, peticion *http.Request) {
	filtros := m.filtros(peticion, true)
	panel, err := m.servicio.Panel(peticion.Context(), filtros)
	if err != nil {
		m.errorInterno(respuesta, "calculando los indicadores", err)
		return
	}
	m.renderizar(respuesta, peticion, vista.PanelIndicadores(aPanelVista(panel), fechas.FormatearISO(filtros.Desde)))
}

func (m *Manejador) resumen(respuesta http.ResponseWriter, peticion *http.Request) {
	ctx := peticion.Context()
	filtros := m.filtros(peticion, false)
	filas, err := m.servicio.Resumen(ctx, filtros)
	if err != nil {
		m.errorInterno(respuesta, "calculando el resumen", err)
		return
	}
	m.renderizar(respuesta, peticion, vista.ResumenReporte(vista.Resumen{
		Nombre:    m.nombreUsuario(ctx),
		Rol:       intermediarios.RolUsuario(ctx),
		TokenCSRF: m.tokenCSRF(ctx),
		Filtros:   aFiltrosVista(filtros),
		Filas:     aResumenVista(filas),
	}))
}

func (m *Manejador) reiteracion(respuesta http.ResponseWriter, peticion *http.Request) {
	ctx := peticion.Context()
	filtros := m.filtros(peticion, false)
	filas, err := m.servicio.Reiteracion(ctx, filtros)
	if err != nil {
		m.errorInterno(respuesta, "calculando la reiteración", err)
		return
	}
	vistas := make([]vista.ReiteracionFila, 0, len(filas))
	for _, fila := range filas {
		vistas = append(vistas, vista.ReiteracionFila{
			Alumno:          fila.Alumno,
			Seccion:         fila.Seccion,
			Tardanzas:       fila.Tardanzas,
			FaltasSeguidas:  fila.FaltasSeguidas,
			UmbralTardanzas: fila.UmbralTardanzas,
			UmbralFaltas:    fila.UmbralFaltas,
		})
	}
	m.renderizar(respuesta, peticion, vista.ReiteracionReporte(vista.Reiteracion{
		Nombre:    m.nombreUsuario(ctx),
		Rol:       intermediarios.RolUsuario(ctx),
		TokenCSRF: m.tokenCSRF(ctx),
		Filtros:   aFiltrosVista(filtros),
		Filas:     vistas,
	}))
}

func (m *Manejador) diarioXLSX(respuesta http.ResponseWriter, peticion *http.Request) {
	filtros := m.filtros(peticion, true)
	filas, err := m.servicio.Diario(peticion.Context(), filtros)
	if err != nil {
		m.errorInterno(respuesta, "exportando el reporte diario", err)
		return
	}
	contenido, err := ExportarDiarioXLSX(filas, fechas.FormatearISO(filtros.Desde))
	if err != nil {
		m.errorInterno(respuesta, "generando el Excel", err)
		return
	}
	descargar(respuesta, "reporte-diario.xlsx",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", contenido)
}

func (m *Manejador) diarioPDF(respuesta http.ResponseWriter, peticion *http.Request) {
	filtros := m.filtros(peticion, true)
	filas, err := m.servicio.Diario(peticion.Context(), filtros)
	if err != nil {
		m.errorInterno(respuesta, "exportando el reporte diario", err)
		return
	}
	contenido, err := ExportarDiarioPDF(filas, fechas.FormatearISO(filtros.Desde))
	if err != nil {
		m.errorInterno(respuesta, "generando el PDF", err)
		return
	}
	descargar(respuesta, "reporte-diario.pdf", "application/pdf", contenido)
}

func (m *Manejador) resumenXLSX(respuesta http.ResponseWriter, peticion *http.Request) {
	filtros := m.filtros(peticion, false)
	filas, err := m.servicio.Resumen(peticion.Context(), filtros)
	if err != nil {
		m.errorInterno(respuesta, "exportando el resumen", err)
		return
	}
	contenido, err := ExportarResumenXLSX(filas, fechas.FormatearISO(filtros.Desde), fechas.FormatearISO(filtros.Hasta))
	if err != nil {
		m.errorInterno(respuesta, "generando el Excel", err)
		return
	}
	descargar(respuesta, "reporte-resumen.xlsx",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", contenido)
}

func (m *Manejador) resumenPDF(respuesta http.ResponseWriter, peticion *http.Request) {
	filtros := m.filtros(peticion, false)
	filas, err := m.servicio.Resumen(peticion.Context(), filtros)
	if err != nil {
		m.errorInterno(respuesta, "exportando el resumen", err)
		return
	}
	contenido, err := ExportarResumenPDF(filas, fechas.FormatearISO(filtros.Desde), fechas.FormatearISO(filtros.Hasta))
	if err != nil {
		m.errorInterno(respuesta, "generando el PDF", err)
		return
	}
	descargar(respuesta, "reporte-resumen.pdf", "application/pdf", contenido)
}

func (m *Manejador) filtros(peticion *http.Request, diario bool) Filtros {
	consulta := peticion.URL.Query()
	hoy := fechas.Hoy()
	desde := hoy
	hasta := hoy
	if diario {
		if valor := consulta.Get("fecha"); valor != "" {
			if fecha, err := fechas.ParsearDia(valor); err == nil {
				desde, hasta = fecha, fecha
			}
		}
		if valor := consulta.Get("desde"); valor != "" {
			if fecha, err := fechas.ParsearDia(valor); err == nil {
				desde = fecha
			}
		}
	} else {
		desde = time.Date(hoy.Year(), hoy.Month(), 1, 0, 0, 0, 0, hoy.Location())
		hasta = hoy
		if valor := consulta.Get("desde"); valor != "" {
			if fecha, err := fechas.ParsearDia(valor); err == nil {
				desde = fecha
			}
		}
		if valor := consulta.Get("hasta"); valor != "" {
			if fecha, err := fechas.ParsearDia(valor); err == nil {
				hasta = fecha
			}
		}
		if hasta.Before(desde) {
			hasta = desde
		}
	}
	return Filtros{
		Desde:     desde,
		Hasta:     hasta,
		Turno:     consulta.Get("turno"),
		Grado:     consulta.Get("grado"),
		Seccion:   consulta.Get("seccion"),
		Alumno:    consulta.Get("alumno"),
		Estado:    consulta.Get("estado"),
		UsuarioID: intermediarios.UsuarioID(peticion.Context()),
		Rol:       intermediarios.RolUsuario(peticion.Context()),
	}
}

func aFiltrosVista(filtros Filtros) vista.Filtros {
	return vista.Filtros{
		Desde:   fechas.FormatearISO(filtros.Desde),
		Hasta:   fechas.FormatearISO(filtros.Hasta),
		Turno:   filtros.Turno,
		Grado:   filtros.Grado,
		Seccion: filtros.Seccion,
		Alumno:  filtros.Alumno,
		Estado:  filtros.Estado,
	}
}

func aFilasVista(filas []FilaDiaria) []vista.FilaDiaria {
	resultado := make([]vista.FilaDiaria, 0, len(filas))
	for _, fila := range filas {
		resultado = append(resultado, vista.FilaDiaria{
			Turno:         fila.Turno,
			Grado:         fila.Grado,
			Seccion:       fila.Seccion,
			Identificador: fila.Identificador,
			Alumno:        fila.Alumno,
			Estado:        EtiquetaEstado(fila.Estado),
			Hora:          fila.Hora,
		})
	}
	return resultado
}

func aPanelVista(panel PanelDelDia) vista.Panel {
	return vista.Panel{
		Presentes:  panel.Presentes,
		Tardanzas:  panel.Tardanzas,
		Ausentes:   panel.Ausentes,
		Pendientes: panel.Pendientes,
		Total:      panel.Total,
	}
}

func aResumenVista(filas []ResumenAlumno) []vista.ResumenFila {
	resultado := make([]vista.ResumenFila, 0, len(filas))
	for _, fila := range filas {
		resultado = append(resultado, vista.ResumenFila{
			Grado:                 fila.Grado,
			Seccion:               fila.Seccion,
			Identificador:         fila.Identificador,
			Alumno:                fila.Alumno,
			Asistencias:           fila.Asistencias,
			Tardanzas:             fila.Tardanzas,
			Faltas:                fila.Faltas,
			FaltasJustificadas:    fila.FaltasJustificadas,
			TardanzasJustificadas: fila.TardanzasJustificadas,
			Porcentaje:            fmt.Sprintf("%.1f", fila.Porcentaje),
		})
	}
	return resultado
}

func descargar(respuesta http.ResponseWriter, nombre, tipo string, contenido []byte) {
	respuesta.Header().Set("Content-Type", tipo)
	respuesta.Header().Set("Content-Disposition", "attachment; filename=\""+nombre+"\"")
	respuesta.Header().Set("Content-Length", strconv.Itoa(len(contenido)))
	_, _ = respuesta.Write(contenido)
}

func (m *Manejador) renderizar(respuesta http.ResponseWriter, peticion *http.Request, componente interface {
	Render(context.Context, io.Writer) error
}) {
	respuesta.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := componente.Render(peticion.Context(), respuesta); err != nil {
		m.registro.Error("renderizando reportes", "error", err)
	}
}

func (m *Manejador) nombreUsuario(ctx context.Context) string {
	nombre := intermediarios.NombreUsuario(ctx, m.sesiones)
	if nombre == "" {
		return "Usuario"
	}
	return nombre
}

func (m *Manejador) tokenCSRF(ctx context.Context) string {
	token, err := intermediarios.TokenCSRF(ctx, m.sesiones)
	if err != nil {
		return ""
	}
	return token
}

func (m *Manejador) errorInterno(respuesta http.ResponseWriter, mensaje string, err error) {
	m.registro.Error(mensaje, "error", err)
	http.Error(respuesta, "No se pudo completar la operación.", http.StatusInternalServerError)
}
