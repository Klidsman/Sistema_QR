package alertas

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"asistencia/internal/plataforma/fechas"
	"asistencia/internal/plataforma/intermediarios"
	vista "asistencia/web/plantillas/alertas"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
)

// ServicioAlertas son las operaciones que usa el manejador.
type ServicioAlertas interface {
	ListarAlertas(ctx context.Context, filtros FiltrosAlerta) ([]Alerta, error)
	Detallar(ctx context.Context, alertas []Alerta) ([]AlertaDetalle, error)
	ContarNuevas(ctx context.Context, destinatarioID int64, todos bool) (int, error)
	MarcarVista(ctx context.Context, id, usuarioID int64, todos bool) error
	MarcarVistas(ctx context.Context, ids []int64, usuarioID int64, todos bool) (int, error)
	AsignarResponsable(ctx context.Context, id, usuarioID, destinatarioID int64) error
}

// DependenciasManejador reúne lo que necesita el manejador de alertas.
type DependenciasManejador struct {
	Servicio ServicioAlertas
	Sesiones *scs.SessionManager
	Registro *slog.Logger
}

// Manejador atiende el panel y el indicador de alertas.
type Manejador struct {
	servicio ServicioAlertas
	sesiones *scs.SessionManager
	registro *slog.Logger
}

// NuevoManejador crea el manejador de alertas.
func NuevoManejador(dependencias DependenciasManejador) *Manejador {
	return &Manejador{
		servicio: dependencias.Servicio,
		sesiones: dependencias.Sesiones,
		registro: dependencias.Registro,
	}
}

// Montar registra las rutas de alertas para los roles autorizados.
func (m *Manejador) Montar(enrutador chi.Router) {
	enrutador.Group(func(autorizados chi.Router) {
		autorizados.Use(intermediarios.RequerirRol(m.sesiones,
			intermediarios.RolAdministrador, intermediarios.RolDocente, intermediarios.RolPorteria))
		autorizados.Get("/alertas", m.panel)
		autorizados.Get("/alertas/indicador", m.indicador)
		autorizados.Post("/alertas/vistas", m.marcarVistas)
		autorizados.Post("/alertas/{id}/vista", m.marcarVista)
	})

	enrutador.Group(func(administracion chi.Router) {
		administracion.Use(intermediarios.RequerirRol(m.sesiones, intermediarios.RolAdministrador))
		administracion.Post("/alertas/{id}/responsable", m.asignarResponsable)
	})
}

func (m *Manejador) panel(respuesta http.ResponseWriter, peticion *http.Request) {
	ctx := peticion.Context()
	usuarioID := intermediarios.UsuarioID(ctx)
	todos := intermediarios.RolUsuario(ctx) == intermediarios.RolAdministrador
	filtros, valores, err := filtrosDePeticion(peticion, usuarioID, todos)
	if err != nil {
		http.Error(respuesta, "Los filtros no son válidos.", http.StatusUnprocessableEntity)
		return
	}

	alertas, err := m.servicio.ListarAlertas(ctx, filtros)
	if err != nil {
		m.registro.Error("listando alertas", "error", err)
		http.Error(respuesta, "No se pudieron cargar las alertas.", http.StatusInternalServerError)
		return
	}
	detalles, err := m.servicio.Detallar(ctx, alertas)
	if err != nil {
		m.registro.Error("detallando alertas", "error", err)
		http.Error(respuesta, "No se pudieron cargar las alertas.", http.StatusInternalServerError)
		return
	}
	nuevas, err := m.servicio.ContarNuevas(ctx, usuarioID, todos)
	if err != nil {
		m.registro.Error("contando alertas nuevas", "error", err)
		http.Error(respuesta, "No se pudieron cargar las alertas.", http.StatusInternalServerError)
		return
	}
	token, err := intermediarios.TokenCSRF(ctx, m.sesiones)
	if err != nil {
		http.Error(respuesta, "No se pudo preparar el formulario.", http.StatusInternalServerError)
		return
	}

	detallesVista := make([]vista.AlertaVista, 0, len(detalles))
	for _, detalle := range detalles {
		detallesVista = append(detallesVista, vista.AlertaVista{
			ID:                detalle.ID,
			Tipo:              detalle.Tipo,
			TipoEtiqueta:      EtiquetaTipo(detalle.Tipo),
			Alumno:            detalle.AlumnoNombre,
			Seccion:           detalle.SeccionEtiqueta,
			Fecha:             detalle.FechaEvento.Format("2006-01-02"),
			Estado:            estadoVisible(detalle.Estado),
			Nueva:             detalle.Estado == EstadoNueva,
			SinResponsable:    detalle.SinResponsable,
			SituacionCambiada: detalle.SituacionCambiada,
			MotivoCambio:      detalle.MotivoCambio,
		})
	}

	m.renderizar(respuesta, peticion, vista.PanelAlertas(vista.Panel{
		Nombre:    m.nombreUsuario(ctx),
		Rol:       intermediarios.RolUsuario(ctx),
		TokenCSRF: token,
		Nuevas:    nuevas,
		Alertas:   detallesVista,
		Desde:     valores.desde,
		Hasta:     valores.hasta,
		Tipo:      valores.tipo,
		Estado:    valores.estado,
		Mensaje:   mensajeCorrecto(peticion.URL.Query().Get("ok")),
		EsAdmin:   todos,
	}))
}

func (m *Manejador) indicador(respuesta http.ResponseWriter, peticion *http.Request) {
	ctx := peticion.Context()
	todos := intermediarios.RolUsuario(ctx) == intermediarios.RolAdministrador
	nuevas, err := m.servicio.ContarNuevas(ctx, intermediarios.UsuarioID(ctx), todos)
	if err != nil {
		m.registro.Error("contando alertas nuevas", "error", err)
		http.Error(respuesta, "No se pudo cargar el indicador.", http.StatusInternalServerError)
		return
	}
	m.renderizar(respuesta, peticion, vista.IndicadorAlertas(vista.Indicador{Nuevas: nuevas}))
}

func (m *Manejador) marcarVista(respuesta http.ResponseWriter, peticion *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(peticion, "id"), 10, 64)
	if err != nil {
		http.Error(respuesta, "La alerta no existe.", http.StatusNotFound)
		return
	}
	todos := intermediarios.RolUsuario(peticion.Context()) == intermediarios.RolAdministrador
	err = m.servicio.MarcarVista(peticion.Context(), id, intermediarios.UsuarioID(peticion.Context()), todos)
	if errors.Is(err, ErrAlertaSinAcceso) {
		http.Error(respuesta, "No tiene acceso a esta alerta.", http.StatusForbidden)
		return
	}
	if err != nil {
		m.registro.Error("marcando la alerta como vista", "error", err)
		http.Error(respuesta, "No se pudo marcar la alerta.", http.StatusInternalServerError)
		return
	}
	http.Redirect(respuesta, peticion, "/alertas?ok=vista", http.StatusSeeOther)
}

func (m *Manejador) marcarVistas(respuesta http.ResponseWriter, peticion *http.Request) {
	if err := peticion.ParseForm(); err != nil {
		http.Error(respuesta, "El formulario no es válido.", http.StatusUnprocessableEntity)
		return
	}
	ids := make([]int64, 0, len(peticion.Form["ids"]))
	for _, valor := range peticion.Form["ids"] {
		id, err := strconv.ParseInt(valor, 10, 64)
		if err == nil {
			ids = append(ids, id)
		}
	}
	todos := intermediarios.RolUsuario(peticion.Context()) == intermediarios.RolAdministrador
	if _, err := m.servicio.MarcarVistas(peticion.Context(), ids, intermediarios.UsuarioID(peticion.Context()), todos); err != nil {
		m.registro.Error("marcando alertas como vistas", "error", err)
		http.Error(respuesta, "No se pudieron marcar las alertas.", http.StatusInternalServerError)
		return
	}
	http.Redirect(respuesta, peticion, "/alertas?ok=vistas", http.StatusSeeOther)
}

func (m *Manejador) asignarResponsable(respuesta http.ResponseWriter, peticion *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(peticion, "id"), 10, 64)
	if err != nil {
		http.Error(respuesta, "La alerta no existe.", http.StatusNotFound)
		return
	}
	destinatario, err := strconv.ParseInt(peticion.PostFormValue("destinatario_id"), 10, 64)
	if err != nil || destinatario <= 0 {
		http.Error(respuesta, "Seleccione un docente activo.", http.StatusUnprocessableEntity)
		return
	}
	err = m.servicio.AsignarResponsable(peticion.Context(), id, intermediarios.UsuarioID(peticion.Context()), destinatario)
	if errors.Is(err, ErrAlertaNoEncontrada) {
		http.Error(respuesta, "La alerta no existe.", http.StatusNotFound)
		return
	}
	if err != nil {
		m.registro.Error("asignando responsable", "error", err)
		http.Error(respuesta, "No se pudo asignar el responsable.", http.StatusInternalServerError)
		return
	}
	http.Redirect(respuesta, peticion, "/alertas?ok=responsable", http.StatusSeeOther)
}

func (m *Manejador) renderizar(respuesta http.ResponseWriter, peticion *http.Request, componente interface {
	Render(context.Context, io.Writer) error
}) {
	respuesta.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := componente.Render(peticion.Context(), respuesta); err != nil {
		m.registro.Error("renderizando la vista de alertas", "error", err)
	}
}

func (m *Manejador) nombreUsuario(ctx context.Context) string {
	nombre := intermediarios.NombreUsuario(ctx, m.sesiones)
	if nombre == "" {
		return "Usuario"
	}
	return nombre
}

type valoresFiltro struct {
	desde, hasta, tipo, estado string
}

func filtrosDePeticion(peticion *http.Request, usuarioID int64, todos bool) (FiltrosAlerta, valoresFiltro, error) {
	consulta := peticion.URL.Query()
	valores := valoresFiltro{
		desde:  consulta.Get("desde"),
		hasta:  consulta.Get("hasta"),
		tipo:   consulta.Get("tipo"),
		estado: consulta.Get("estado"),
	}
	filtros := FiltrosAlerta{
		Tipo:           valores.tipo,
		Estado:         valores.estado,
		DestinatarioID: usuarioID,
		Todos:          todos,
	}
	if !todos {
		filtros.DestinatarioID = usuarioID
	}
	if seccion := consulta.Get("seccion"); seccion != "" {
		id, err := strconv.ParseInt(seccion, 10, 64)
		if err != nil {
			return FiltrosAlerta{}, valores, err
		}
		filtros.SeccionID = id
	}
	if valores.desde != "" {
		fecha, err := fechas.ParsearDia(valores.desde)
		if err != nil {
			return FiltrosAlerta{}, valores, err
		}
		filtros.Desde = &fecha
	}
	if valores.hasta != "" {
		fecha, err := fechas.ParsearDia(valores.hasta)
		if err != nil {
			return FiltrosAlerta{}, valores, err
		}
		filtros.Hasta = &fecha
	}
	return filtros, valores, nil
}

func estadoVisible(estado string) string {
	if estado == EstadoNueva {
		return "Nueva"
	}
	return "Vista"
}

func mensajeCorrecto(codigo string) string {
	switch codigo {
	case "vista":
		return "La alerta se marcó como vista."
	case "vistas":
		return "Las alertas seleccionadas se marcaron como vistas."
	case "responsable":
		return "El responsable se asignó correctamente."
	default:
		return ""
	}
}
