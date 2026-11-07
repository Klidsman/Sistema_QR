package justificaciones

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"asistencia/internal/alumnos"
	"asistencia/internal/plataforma/fechas"
	"asistencia/internal/plataforma/intermediarios"
	vista "asistencia/web/plantillas/justificaciones"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
)

// ServicioJustificaciones son las operaciones que usa el manejador.
type ServicioJustificaciones interface {
	Registrar(ctx context.Context, usuarioID int64, rol string, alumnoID int64, motivo string, desde, hasta time.Time, adjuntos []EntradaAdjunto) (ResultadoRegistro, error)
	Resolver(ctx context.Context, usuarioID int64, rol string, id int64, aprobar bool, comentario string) error
	Anular(ctx context.Context, usuarioID int64, rol string, id int64, motivo string) error
	Listar(ctx context.Context, usuarioID int64, rol string) ([]Justificacion, error)
	HistorialDeAlumno(ctx context.Context, usuarioID int64, rol string, alumnoID int64) ([]Justificacion, error)
	ObtenerConAcceso(ctx context.Context, usuarioID int64, rol string, id int64) (Justificacion, error)
	AbrirAdjunto(ctx context.Context, usuarioID int64, rol string, adjuntoID int64) (Adjunto, []byte, error)
	Describir(ctx context.Context, lista []Justificacion) ([]JustificacionDetalle, error)
	Alumno(ctx context.Context, alumnoID int64) (alumnos.Alumno, error)
}

// DependenciasManejador reúne lo necesario para atender justificaciones.
type DependenciasManejador struct {
	Servicio ServicioJustificaciones
	Sesiones *scs.SessionManager
	Registro *slog.Logger
}

// Manejador atiende las pantallas de justificaciones.
type Manejador struct {
	servicio ServicioJustificaciones
	sesiones *scs.SessionManager
	registro *slog.Logger
}

// NuevoManejador crea el manejador.
func NuevoManejador(dependencias DependenciasManejador) *Manejador {
	return &Manejador{servicio: dependencias.Servicio, sesiones: dependencias.Sesiones, registro: dependencias.Registro}
}

// Montar registra las rutas de justificaciones.
func (m *Manejador) Montar(enrutador chi.Router) {
	enrutador.Group(func(autorizados chi.Router) {
		autorizados.Use(intermediarios.RequerirRol(m.sesiones,
			intermediarios.RolAdministrador, intermediarios.RolDirector, intermediarios.RolDocente))
		autorizados.Get("/justificaciones", m.panel)
		autorizados.Get("/justificaciones/nueva", m.formulario)
		autorizados.Post("/justificaciones", m.registrar)
		autorizados.Get("/justificaciones/adjuntos/{adjuntoID}", m.adjunto)
		autorizados.Get("/justificaciones/{id}", m.detalle)
		autorizados.Get("/alumnos/{id}/justificaciones", m.historialAlumno)
	})

	enrutador.Group(func(resolutores chi.Router) {
		resolutores.Use(intermediarios.RequerirRolConEscritura(m.sesiones,
			intermediarios.RolAdministrador, intermediarios.RolDirector))
		resolutores.Post("/justificaciones/{id}/aprobar", m.aprobar)
		resolutores.Post("/justificaciones/{id}/rechazar", m.rechazar)
	})

	enrutador.Group(func(administracion chi.Router) {
		administracion.Use(intermediarios.RequerirRol(m.sesiones, intermediarios.RolAdministrador))
		administracion.Post("/justificaciones/{id}/anular", m.anular)
	})
}

func (m *Manejador) panel(respuesta http.ResponseWriter, peticion *http.Request) {
	ctx := peticion.Context()
	usuarioID := intermediarios.UsuarioID(ctx)
	rol := intermediarios.RolUsuario(ctx)
	lista, err := m.servicio.Listar(ctx, usuarioID, rol)
	if err != nil {
		m.errorInterno(respuesta, "listando justificaciones", err)
		return
	}
	titulo := "Justificaciones pendientes"
	if rol == "docente" {
		titulo = "Mis justificaciones"
	}
	m.renderizarPanel(respuesta, peticion, titulo, lista,
		mensajeCorrecto(peticion.URL.Query().Get("ok")), "")
}

func (m *Manejador) historialAlumno(respuesta http.ResponseWriter, peticion *http.Request) {
	ctx := peticion.Context()
	id, err := strconv.ParseInt(chi.URLParam(peticion, "id"), 10, 64)
	if err != nil {
		http.Error(respuesta, "El alumno no existe.", http.StatusNotFound)
		return
	}
	lista, err := m.servicio.HistorialDeAlumno(ctx, intermediarios.UsuarioID(ctx), intermediarios.RolUsuario(ctx), id)
	if err != nil {
		m.errorInterno(respuesta, "listando el historial de justificaciones", err)
		return
	}
	alumno, err := m.servicio.Alumno(ctx, id)
	if err != nil {
		m.errorInterno(respuesta, "buscando el alumno", err)
		return
	}
	m.renderizarPanel(respuesta, peticion, "Justificaciones de "+alumno.NombreCompleto(), lista,
		mensajeCorrecto(peticion.URL.Query().Get("ok")), "")
}

func (m *Manejador) formulario(respuesta http.ResponseWriter, peticion *http.Request) {
	ctx := peticion.Context()
	alumnoID, err := strconv.ParseInt(peticion.URL.Query().Get("alumno_id"), 10, 64)
	if err != nil {
		http.Error(respuesta, "Seleccione un alumno.", http.StatusUnprocessableEntity)
		return
	}
	alumno, err := m.servicio.Alumno(ctx, alumnoID)
	if err != nil {
		m.errorInterno(respuesta, "buscando el alumno", err)
		return
	}
	fechaConsulta := peticion.URL.Query().Get("fecha")
	if fechaConsulta == "" {
		fechaConsulta = fechas.FormatearISO(fechas.Hoy())
	}
	m.renderizarFormulario(respuesta, peticion, vista.Formulario{
		AlumnoID: alumnoID,
		Alumno:   alumno.NombreCompleto(),
		Seccion:  alumno.Grado + " " + alumno.Seccion,
		Desde:    fechaConsulta,
		Hasta:    fechaConsulta,
	}, "")
}

func (m *Manejador) registrar(respuesta http.ResponseWriter, peticion *http.Request) {
	ctx := peticion.Context()
	if err := peticion.ParseMultipartForm(20 << 20); err != nil {
		http.Error(respuesta, "El formulario no es válido.", http.StatusUnprocessableEntity)
		return
	}
	alumnoID, err := strconv.ParseInt(peticion.PostFormValue("alumno_id"), 10, 64)
	if err != nil {
		http.Error(respuesta, "Seleccione un alumno.", http.StatusUnprocessableEntity)
		return
	}
	desde, err := fechas.ParsearDia(peticion.PostFormValue("desde"))
	if err != nil {
		m.errorFormulario(respuesta, peticion, alumnoID, "La fecha inicial no es válida.")
		return
	}
	hasta, err := fechas.ParsearDia(peticion.PostFormValue("hasta"))
	if err != nil {
		m.errorFormulario(respuesta, peticion, alumnoID, "La fecha final no es válida.")
		return
	}
	adjuntos, err := adjuntosDeFormulario(peticion.MultipartForm.File)
	if err != nil {
		m.errorFormulario(respuesta, peticion, alumnoID, mensajeDeError(err))
		return
	}
	_, err = m.servicio.Registrar(ctx, intermediarios.UsuarioID(ctx), intermediarios.RolUsuario(ctx),
		alumnoID, peticion.PostFormValue("motivo"), desde, hasta, adjuntos)
	if err != nil {
		m.errorFormulario(respuesta, peticion, alumnoID, mensajeDeError(err))
		return
	}
	http.Redirect(respuesta, peticion,
		"/alumnos/"+strconv.FormatInt(alumnoID, 10)+"/justificaciones?ok=creada", http.StatusSeeOther)
}

func (m *Manejador) detalle(respuesta http.ResponseWriter, peticion *http.Request) {
	ctx := peticion.Context()
	id, err := strconv.ParseInt(chi.URLParam(peticion, "id"), 10, 64)
	if err != nil {
		http.Error(respuesta, "La justificación no existe.", http.StatusNotFound)
		return
	}
	justificacion, err := m.servicio.ObtenerConAcceso(ctx, intermediarios.UsuarioID(ctx), intermediarios.RolUsuario(ctx), id)
	if errors.Is(err, ErrSinAcceso) {
		http.Error(respuesta, "No tiene acceso a esta justificación.", http.StatusForbidden)
		return
	}
	if err != nil {
		m.errorInterno(respuesta, "buscando la justificación", err)
		return
	}
	detalles, err := m.servicio.Describir(ctx, []Justificacion{justificacion})
	if err != nil {
		m.errorInterno(respuesta, "describiendo la justificación", err)
		return
	}
	rol := intermediarios.RolUsuario(ctx)
	m.renderizarDetalle(respuesta, peticion, detalles[0], rol, mensajeCorrecto(peticion.URL.Query().Get("ok")), "")
}

func (m *Manejador) aprobar(respuesta http.ResponseWriter, peticion *http.Request) {
	m.resolver(respuesta, peticion, true)
}

func (m *Manejador) rechazar(respuesta http.ResponseWriter, peticion *http.Request) {
	m.resolver(respuesta, peticion, false)
}

func (m *Manejador) resolver(respuesta http.ResponseWriter, peticion *http.Request, aprobar bool) {
	ctx := peticion.Context()
	id, err := strconv.ParseInt(chi.URLParam(peticion, "id"), 10, 64)
	if err != nil {
		http.Error(respuesta, "La justificación no existe.", http.StatusNotFound)
		return
	}
	err = m.servicio.Resolver(ctx, intermediarios.UsuarioID(ctx), intermediarios.RolUsuario(ctx), id,
		aprobar, peticion.PostFormValue("comentario"))
	if err != nil {
		m.mostrarErrorDetalle(respuesta, peticion, id, mensajeDeError(err))
		return
	}
	codigo := "rechazada"
	if aprobar {
		codigo = "aprobada"
	}
	http.Redirect(respuesta, peticion, "/justificaciones/"+strconv.FormatInt(id, 10)+"?ok="+codigo, http.StatusSeeOther)
}

func (m *Manejador) anular(respuesta http.ResponseWriter, peticion *http.Request) {
	ctx := peticion.Context()
	id, err := strconv.ParseInt(chi.URLParam(peticion, "id"), 10, 64)
	if err != nil {
		http.Error(respuesta, "La justificación no existe.", http.StatusNotFound)
		return
	}
	err = m.servicio.Anular(ctx, intermediarios.UsuarioID(ctx), intermediarios.RolUsuario(ctx), id,
		peticion.PostFormValue("motivo"))
	if err != nil {
		m.mostrarErrorDetalle(respuesta, peticion, id, mensajeDeError(err))
		return
	}
	http.Redirect(respuesta, peticion, "/justificaciones/"+strconv.FormatInt(id, 10)+"?ok=anulada", http.StatusSeeOther)
}

func (m *Manejador) adjunto(respuesta http.ResponseWriter, peticion *http.Request) {
	ctx := peticion.Context()
	id, err := strconv.ParseInt(chi.URLParam(peticion, "adjuntoID"), 10, 64)
	if err != nil {
		http.Error(respuesta, "El documento no existe.", http.StatusNotFound)
		return
	}
	adjunto, contenido, err := m.servicio.AbrirAdjunto(ctx, intermediarios.UsuarioID(ctx), intermediarios.RolUsuario(ctx), id)
	if errors.Is(err, ErrSinAcceso) {
		http.Error(respuesta, "No tiene acceso a este documento.", http.StatusForbidden)
		return
	}
	if err != nil {
		m.errorInterno(respuesta, "abriendo el adjunto", err)
		return
	}
	respuesta.Header().Set("Content-Type", adjunto.Tipo)
	respuesta.Header().Set("Content-Disposition", "attachment; filename=\""+adjunto.Nombre+"\"")
	_, _ = respuesta.Write(contenido)
}

func (m *Manejador) errorFormulario(respuesta http.ResponseWriter, peticion *http.Request, alumnoID int64, mensaje string) {
	alumno, err := m.servicio.Alumno(peticion.Context(), alumnoID)
	if err != nil {
		m.errorInterno(respuesta, "buscando el alumno", err)
		return
	}
	m.renderizarFormulario(respuesta, peticion, vista.Formulario{
		AlumnoID: alumnoID,
		Alumno:   alumno.NombreCompleto(),
		Seccion:  alumno.Grado + " " + alumno.Seccion,
		Desde:    peticion.PostFormValue("desde"),
		Hasta:    peticion.PostFormValue("hasta"),
		Motivo:   peticion.PostFormValue("motivo"),
	}, mensaje)
}

func (m *Manejador) mostrarErrorDetalle(respuesta http.ResponseWriter, peticion *http.Request, id int64, mensaje string) {
	justificacion, err := m.servicio.ObtenerConAcceso(peticion.Context(),
		intermediarios.UsuarioID(peticion.Context()), intermediarios.RolUsuario(peticion.Context()), id)
	if err != nil {
		m.errorInterno(respuesta, "buscando la justificación", err)
		return
	}
	detalles, err := m.servicio.Describir(peticion.Context(), []Justificacion{justificacion})
	if err != nil {
		m.errorInterno(respuesta, "describiendo la justificación", err)
		return
	}
	m.renderizarDetalle(respuesta, peticion, detalles[0], intermediarios.RolUsuario(peticion.Context()), "", mensaje)
}

func (m *Manejador) renderizarPanel(respuesta http.ResponseWriter, peticion *http.Request, titulo string, lista []Justificacion, mensaje, mensajeError string) {
	detalles, err := m.servicio.Describir(peticion.Context(), lista)
	if err != nil {
		m.errorInterno(respuesta, "describiendo justificaciones", err)
		return
	}
	rol := intermediarios.RolUsuario(peticion.Context())
	vistas := make([]vista.JustificacionVista, 0, len(detalles))
	for _, detalle := range detalles {
		vistas = append(vistas, m.aVista(detalle, rol, 0))
	}
	m.renderizar(respuesta, peticion, vista.PanelJustificaciones(vista.Panel{
		Nombre:          m.nombreUsuario(peticion.Context()),
		Rol:             rol,
		TokenCSRF:       m.tokenCSRF(peticion.Context()),
		Justificaciones: vistas,
		Mensaje:         mensaje,
		Error:           mensajeError,
		Titulo:          titulo,
	}))
}

func (m *Manejador) renderizarFormulario(respuesta http.ResponseWriter, peticion *http.Request, formulario vista.Formulario, mensajeError string) {
	formulario.Nombre = m.nombreUsuario(peticion.Context())
	formulario.Rol = intermediarios.RolUsuario(peticion.Context())
	formulario.TokenCSRF = m.tokenCSRF(peticion.Context())
	formulario.Error = mensajeError
	m.renderizar(respuesta, peticion, vista.FormularioJustificacion(formulario))
}

func (m *Manejador) renderizarDetalle(respuesta http.ResponseWriter, peticion *http.Request, detalle JustificacionDetalle, rol, mensaje, mensajeError string) {
	m.renderizar(respuesta, peticion, vista.DetalleJustificacion(vista.Detalle{
		Nombre:        m.nombreUsuario(peticion.Context()),
		Rol:           rol,
		TokenCSRF:     m.tokenCSRF(peticion.Context()),
		Justificacion: m.aVista(detalle, rol, intermediarios.UsuarioID(peticion.Context())),
		Mensaje:       mensaje,
		Error:         mensajeError,
	}))
}

func (m *Manejador) aVista(detalle JustificacionDetalle, rol string, usuarioID int64) vista.JustificacionVista {
	fechasVista := make([]vista.FechaVista, 0, len(detalle.Fechas))
	for _, fecha := range detalle.Fechas {
		fechasVista = append(fechasVista, vista.FechaVista{
			Fecha:  fechas.FormatearDia(fecha.Fecha),
			Tipo:   etiquetaTipo(fecha.Tipo),
			Estado: fecha.Estado,
		})
	}
	adjuntos := make([]vista.AdjuntoVista, 0, len(detalle.Adjuntos))
	for _, adjunto := range detalle.Adjuntos {
		adjuntos = append(adjuntos, vista.AdjuntoVista{
			ID:     adjunto.ID,
			Nombre: adjunto.Nombre,
			Tamano: tamanoLegible(adjunto.Tamano),
		})
	}
	puedeResolver := detalle.Estado == EstadoPendiente &&
		(rol == "administrador" || rol == "director") &&
		(detalle.RegistranteID != usuarioID || rol == "administrador")
	return vista.JustificacionVista{
		ID:              detalle.ID,
		Alumno:          detalle.Alumno,
		Seccion:         detalle.Seccion,
		Motivo:          detalle.Motivo,
		Estado:          etiquetaEstado(detalle.Estado),
		Registrante:     strconv.FormatInt(detalle.RegistranteID, 10),
		CreadaEn:        fechas.Formatear(detalle.CreadaEn),
		Comentario:      detalle.Comentario,
		MotivoAnulacion: detalle.MotivoAnulacion,
		Fechas:          fechasVista,
		Adjuntos:        adjuntos,
		Pendiente:       detalle.Estado == EstadoPendiente,
		Aprobada:        detalle.Estado == EstadoAprobada,
		PuedeResolver:   puedeResolver,
		PuedeAnular:     detalle.Estado == EstadoAprobada && rol == "administrador",
	}
}

func (m *Manejador) renderizar(respuesta http.ResponseWriter, peticion *http.Request, componente interface {
	Render(context.Context, io.Writer) error
}) {
	respuesta.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := componente.Render(peticion.Context(), respuesta); err != nil {
		m.registro.Error("renderizando justificaciones", "error", err)
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

func adjuntosDeFormulario(archivos map[string][]*multipart.FileHeader) ([]EntradaAdjunto, error) {
	var adjuntos []EntradaAdjunto
	for _, clave := range []string{"adjunto1", "adjunto2", "adjunto3"} {
		for _, archivo := range archivos[clave] {
			if archivo.Size <= 0 {
				continue
			}
			if archivo.Size > 5*1024*1024 {
				return nil, ErrAdjuntoInvalido
			}
			extension := strings.ToLower(filepath.Ext(archivo.Filename))
			tipo := ""
			switch extension {
			case ".pdf":
				tipo = "application/pdf"
			case ".jpg", ".jpeg":
				tipo = "image/jpeg"
			case ".png":
				tipo = "image/png"
			default:
				return nil, ErrAdjuntoInvalido
			}
			archivoAbierto, err := archivo.Open()
			if err != nil {
				return nil, err
			}
			contenido, err := io.ReadAll(io.LimitReader(archivoAbierto, 5*1024*1024+1))
			_ = archivoAbierto.Close()
			if err != nil {
				return nil, err
			}
			adjuntos = append(adjuntos, EntradaAdjunto{
				Nombre:    archivo.Filename,
				Tipo:      tipo,
				Tamano:    archivo.Size,
				Contenido: contenido,
			})
		}
	}
	return adjuntos, nil
}

func mensajeCorrecto(codigo string) string {
	switch codigo {
	case "creada":
		return "La justificación quedó pendiente de aprobación."
	case "aprobada":
		return "La justificación se aprobó correctamente."
	case "rechazada":
		return "La justificación se rechazó."
	case "anulada":
		return "La justificación se anuló y el registro volvió a su estado original."
	default:
		return ""
	}
}

func mensajeDeError(err error) string {
	switch {
	case errors.Is(err, ErrMotivoRequerido):
		return "Indique el motivo de la justificación."
	case errors.Is(err, ErrMotivoLargo):
		return "El motivo no puede superar los 500 caracteres."
	case errors.Is(err, ErrRangoInvalido):
		return "La fecha final debe ser igual o posterior a la inicial."
	case errors.Is(err, ErrFueraDePlazo):
		return "La justificación está fuera del plazo permitido."
	case errors.Is(err, ErrSinFechasAplicables):
		return "No hay fechas con falta o tardanza que justificar."
	case errors.Is(err, ErrSinAcceso):
		return "No tiene acceso a esta justificación."
	case errors.Is(err, ErrYaResuelta):
		return "La justificación ya fue resuelta."
	case errors.Is(err, ErrAutoresolucion):
		return "No puede resolver una justificación que usted registró."
	case errors.Is(err, ErrComentarioRequerido):
		return "Indique el comentario del rechazo."
	case errors.Is(err, ErrSoloAdministrador):
		return "Solo el administrador puede anular justificaciones."
	case errors.Is(err, ErrNoAnulable):
		return "Solo se puede anular una justificación aprobada."
	case errors.Is(err, ErrMotivoAnulacion):
		return "Indique el motivo de la anulación."
	case errors.Is(err, ErrAdjuntoInvalido), errors.Is(err, ErrAdjuntosExcedidos):
		return "Solo se aceptan hasta tres documentos PDF, JPG o PNG de 5 MB."
	case errors.Is(err, ErrAlumnoInactivo):
		return "El alumno no está activo."
	default:
		return "No se pudo completar la operación."
	}
}

func etiquetaEstado(estado string) string {
	switch estado {
	case EstadoPendiente:
		return "Pendiente"
	case EstadoAprobada:
		return "Aprobada"
	case EstadoRechazada:
		return "Rechazada"
	case EstadoAnulada:
		return "Anulada"
	case EstadoRegistroNoAplica:
		return "Registro ya no aplica"
	default:
		return estado
	}
}

func etiquetaTipo(tipo string) string {
	if tipo == TipoTardanza {
		return "Tardanza"
	}
	return "Falta"
}

func tamanoLegible(tamano int64) string {
	if tamano >= 1024*1024 {
		return strconv.FormatInt(tamano/(1024*1024), 10) + " MB"
	}
	return strconv.FormatInt(tamano/1024, 10) + " KB"
}
