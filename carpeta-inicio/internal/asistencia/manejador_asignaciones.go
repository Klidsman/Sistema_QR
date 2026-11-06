package asistencia

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"asistencia/internal/academico"
	"asistencia/internal/autenticacion"
	"asistencia/internal/plataforma/fechas"
	"asistencia/internal/plataforma/intermediarios"
	vista "asistencia/web/plantillas/asistencia"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
)

// ServicioAsistencia son las operaciones que usa el manejador.
type ServicioAsistencia interface {
	ListarAsignaciones(ctx context.Context) ([]AsignacionVista, error)
	CrearAsignacion(ctx context.Context, usuarioID int64, datos DatosAsignacion) (int64, error)
	CerrarAsignacion(ctx context.Context, usuarioID, asignacionID int64, fecha time.Time) error
	DocentesDisponibles(ctx context.Context) ([]autenticacion.Docente, error)
	SeccionesDisponibles(ctx context.Context) ([]academico.Seccion, error)
	CursosDisponibles(ctx context.Context) ([]academico.Curso, error)
	ClasesAsignadas(ctx context.Context, docenteID int64, fecha time.Time) ([]ClaseAsignada, error)
	ListarClase(ctx context.Context, usuarioID int64, administrador bool, fecha time.Time, seccionID, cursoID int64, bloque int) (ClaseDetalle, error)
	GuardarRegistros(ctx context.Context, usuarioID int64, administrador bool, fecha time.Time, seccionID, cursoID int64, bloque int, cambios []CambioRegistro) error
	MarcarTodosPresentes(ctx context.Context, usuarioID int64, administrador bool, fecha time.Time, seccionID, cursoID int64, bloque int) (int, error)
	ConsultaRegistros(ctx context.Context, fecha time.Time, seccionID int64) ([]ConsultaRegistro, error)
	CorregirRegistro(ctx context.Context, usuarioID, registroID int64, estado, observacion, motivo string) error
	PlazoEdicion(ctx context.Context) (ParametrosAsistencia, error)
	ConfigurarPlazo(ctx context.Context, usuarioID int64, horas int) error
	ConfigurarPlazoJustificacion(ctx context.Context, usuarioID int64, dias int) error
	ExportarDocente(ctx context.Context, docenteID int64, desde, hasta time.Time) ([]RegistroExportacion, error)
	ExportarSeccion(ctx context.Context, seccionID, cursoID int64, desde, hasta time.Time) ([]RegistroExportacion, error)
}

// Dependencias reúne lo que el manejador de asistencia necesita.
type Dependencias struct {
	Servicio ServicioAsistencia
	Sesiones *scs.SessionManager
	Registro *slog.Logger
}

// Manejador atiende las pantallas de asistencia del docente.
type Manejador struct {
	servicio ServicioAsistencia
	sesiones *scs.SessionManager
	registro *slog.Logger
}

// NuevoManejador crea el manejador de asistencia.
func NuevoManejador(dependencias Dependencias) *Manejador {
	return &Manejador{
		servicio: dependencias.Servicio,
		sesiones: dependencias.Sesiones,
		registro: dependencias.Registro,
	}
}

// Montar registra las rutas de asistencia agrupadas por rol.
func (m *Manejador) Montar(enrutador chi.Router) {
	enrutador.Group(func(lectura chi.Router) {
		lectura.Use(intermediarios.RequerirRol(m.sesiones,
			intermediarios.RolAdministrador, intermediarios.RolDirector))
		lectura.Get("/asistencia/asignaciones", m.listarAsignaciones)
		lectura.Get("/asistencia/consulta", m.consultarAsistencia)
		lectura.Get("/configuracion/asistencia", m.verParametros)
	})

	enrutador.Group(func(escritura chi.Router) {
		escritura.Use(intermediarios.RequerirRol(m.sesiones, intermediarios.RolAdministrador))
		escritura.Post("/asistencia/asignaciones", m.crearAsignacion)
		escritura.Post("/asistencia/asignaciones/{id}/cierre", m.cerrarAsignacion)
		escritura.Post("/asistencia/consulta/correccion", m.corregirAsistencia)
		escritura.Post("/configuracion/asistencia", m.configurarParametros)
		escritura.Post("/configuracion/asistencia/justificaciones", m.configurarPlazoJustificacion)
	})

	enrutador.Group(func(exportacion chi.Router) {
		exportacion.Use(intermediarios.RequerirRol(m.sesiones,
			intermediarios.RolAdministrador, intermediarios.RolDocente, intermediarios.RolDirector))
		exportacion.Get("/asistencia/exportar", m.formularioExportacion)
		exportacion.Get("/asistencia/exportar/archivo", m.descargarExportacion)
	})

	enrutador.Group(func(clase chi.Router) {
		clase.Use(intermediarios.RequerirRol(m.sesiones,
			intermediarios.RolAdministrador, intermediarios.RolDocente))
		clase.Get("/asistencia", m.seleccionClase)
		clase.Get("/asistencia/clase", m.verClase)
		clase.Post("/asistencia/clase", m.guardarClase)
		clase.Post("/asistencia/clase/todos", m.marcarTodos)
	})
}

func (m *Manejador) listarAsignaciones(respuesta http.ResponseWriter, peticion *http.Request) {
	m.renderizarAsignaciones(respuesta, peticion, http.StatusOK, "", peticion.URL.Query().Get("ok"))
}

func (m *Manejador) crearAsignacion(respuesta http.ResponseWriter, peticion *http.Request) {
	datos, err := datosAsignacionDeFormulario(peticion)
	if err != nil {
		m.renderizarAsignaciones(respuesta, peticion, http.StatusUnprocessableEntity, mensajeErrorAsistencia(err), "")
		return
	}
	if _, err := m.servicio.CrearAsignacion(peticion.Context(), m.usuarioID(peticion.Context()), datos); err != nil {
		m.renderizarAsignaciones(respuesta, peticion, http.StatusUnprocessableEntity, mensajeErrorAsistencia(err), "")
		return
	}
	http.Redirect(respuesta, peticion, "/asistencia/asignaciones?ok=creada", http.StatusSeeOther)
}

func (m *Manejador) cerrarAsignacion(respuesta http.ResponseWriter, peticion *http.Request) {
	asignacionID, err := strconv.ParseInt(chi.URLParam(peticion, "id"), 10, 64)
	if err != nil || asignacionID <= 0 {
		m.renderizarAsignaciones(respuesta, peticion, http.StatusUnprocessableEntity,
			mensajeErrorAsistencia(ErrAsignacionNoEncontrada), "")
		return
	}
	fecha, err := fechas.ParsearDia(peticion.PostFormValue("fecha_fin"))
	if err != nil {
		m.renderizarAsignaciones(respuesta, peticion, http.StatusUnprocessableEntity,
			mensajeErrorAsistencia(err), "")
		return
	}
	if err := m.servicio.CerrarAsignacion(peticion.Context(), m.usuarioID(peticion.Context()), asignacionID, fecha); err != nil {
		m.renderizarAsignaciones(respuesta, peticion, http.StatusUnprocessableEntity, mensajeErrorAsistencia(err), "")
		return
	}
	http.Redirect(respuesta, peticion, "/asistencia/asignaciones?ok=cerrada", http.StatusSeeOther)
}

func (m *Manejador) renderizarAsignaciones(respuesta http.ResponseWriter, peticion *http.Request, estado int, mensajeError, codigoCorrecto string) {
	ctx := peticion.Context()
	asignaciones, err := m.servicio.ListarAsignaciones(ctx)
	if err != nil {
		m.registro.Error("no se pudieron listar las asignaciones", "error", err)
		http.Error(respuesta, "No se pudieron cargar las asignaciones", http.StatusInternalServerError)
		return
	}
	docentes, err := m.servicio.DocentesDisponibles(ctx)
	if err != nil {
		m.registro.Error("no se pudieron listar los docentes", "error", err)
		http.Error(respuesta, "No se pudieron cargar las asignaciones", http.StatusInternalServerError)
		return
	}
	secciones, err := m.servicio.SeccionesDisponibles(ctx)
	if err != nil {
		m.registro.Error("no se pudieron listar las secciones", "error", err)
		http.Error(respuesta, "No se pudieron cargar las asignaciones", http.StatusInternalServerError)
		return
	}
	cursos, err := m.servicio.CursosDisponibles(ctx)
	if err != nil {
		m.registro.Error("no se pudieron listar los cursos", "error", err)
		http.Error(respuesta, "No se pudieron cargar las asignaciones", http.StatusInternalServerError)
		return
	}

	m.renderizar(respuesta, peticion, estado, vista.Asignaciones(
		m.nombreUsuario(ctx),
		intermediarios.RolUsuario(ctx),
		m.tokenCSRF(ctx),
		aDocentesVista(docentes),
		aSeccionesVista(secciones),
		aCursosVista(cursos),
		aAsignacionesVista(asignaciones),
		mensajeCorrectoAsignacion(codigoCorrecto),
		mensajeError,
	))
}

func datosAsignacionDeFormulario(peticion *http.Request) (DatosAsignacion, error) {
	docenteID, err := strconv.ParseInt(strings.TrimSpace(peticion.PostFormValue("docente_id")), 10, 64)
	if err != nil || docenteID <= 0 {
		return DatosAsignacion{}, ErrDocenteNoEncontrado
	}
	cursoID, err := strconv.ParseInt(strings.TrimSpace(peticion.PostFormValue("curso_id")), 10, 64)
	if err != nil || cursoID <= 0 {
		return DatosAsignacion{}, ErrCursoNoVinculado
	}
	seccionID, err := strconv.ParseInt(strings.TrimSpace(peticion.PostFormValue("seccion_id")), 10, 64)
	if err != nil || seccionID <= 0 {
		return DatosAsignacion{}, ErrAsignacionNoEncontrada
	}
	inicio, err := fechas.ParsearDia(peticion.PostFormValue("fecha_inicio"))
	if err != nil {
		return DatosAsignacion{}, err
	}
	var fin *time.Time
	if texto := strings.TrimSpace(peticion.PostFormValue("fecha_fin")); texto != "" {
		fechaFin, err := fechas.ParsearDia(texto)
		if err != nil {
			return DatosAsignacion{}, err
		}
		fin = &fechaFin
	}
	return DatosAsignacion{
		DocenteID:   docenteID,
		CursoID:     cursoID,
		SeccionID:   seccionID,
		FechaInicio: inicio,
		FechaFin:    fin,
		Titular:     peticion.PostFormValue("titular") == "si",
	}, nil
}

func mensajeCorrectoAsignacion(codigo string) string {
	switch codigo {
	case "creada":
		return "La asignación se creó correctamente."
	case "cerrada":
		return "La asignación se cerró correctamente."
	default:
		return ""
	}
}

func mensajeErrorAsistencia(err error) string {
	if errors.Is(err, ErrDocenteNoEncontrado) {
		return "El docente no existe o no está activo."
	}
	if errors.Is(err, ErrCursoNoVinculado) {
		return "El curso no está vinculado al grado de la sección."
	}
	if errors.Is(err, ErrAsignacionNoEncontrada) {
		return "La asignación no existe."
	}
	if errors.Is(err, ErrAsignacionFueraDeRango) || errors.Is(err, ErrRangoFechasInvalido) {
		return "Las fechas de la asignación no son válidas."
	}
	return "No se pudo guardar el cambio. Revise los datos e intente de nuevo."
}

func (m *Manejador) renderizar(respuesta http.ResponseWriter, peticion *http.Request, estado int, componente interface {
	Render(context.Context, io.Writer) error
}) {
	respuesta.Header().Set("Content-Type", "text/html; charset=utf-8")
	respuesta.WriteHeader(estado)
	if err := componente.Render(peticion.Context(), respuesta); err != nil {
		m.registro.Error("no se pudo renderizar la vista de asistencia", "error", err)
	}
}

func (m *Manejador) tokenCSRF(ctx context.Context) string {
	token, err := intermediarios.TokenCSRF(ctx, m.sesiones)
	if err != nil {
		m.registro.Error("no se pudo preparar el token CSRF", "error", err)
	}
	return token
}

func (m *Manejador) nombreUsuario(ctx context.Context) string {
	nombre := intermediarios.NombreUsuario(ctx, m.sesiones)
	if nombre == "" {
		return "Usuario"
	}
	return nombre
}

func (m *Manejador) usuarioID(ctx context.Context) int64 {
	return intermediarios.UsuarioID(ctx)
}

func aDocentesVista(docentes []autenticacion.Docente) []vista.DocenteOpcion {
	resultado := make([]vista.DocenteOpcion, 0, len(docentes))
	for _, docente := range docentes {
		resultado = append(resultado, vista.DocenteOpcion{ID: docente.ID, Nombre: docente.Nombre})
	}
	return resultado
}

func aSeccionesVista(secciones []academico.Seccion) []vista.SeccionOpcion {
	resultado := make([]vista.SeccionOpcion, 0, len(secciones))
	for _, seccion := range secciones {
		resultado = append(resultado, vista.SeccionOpcion{ID: seccion.ID, Etiqueta: seccion.Etiqueta()})
	}
	return resultado
}

func aCursosVista(cursos []academico.Curso) []vista.CursoOpcion {
	resultado := make([]vista.CursoOpcion, 0, len(cursos))
	for _, curso := range cursos {
		resultado = append(resultado, vista.CursoOpcion{ID: curso.ID, Nombre: curso.Nombre})
	}
	return resultado
}

func aAsignacionesVista(asignaciones []AsignacionVista) []vista.AsignacionVista {
	resultado := make([]vista.AsignacionVista, 0, len(asignaciones))
	for _, asignacion := range asignaciones {
		fin := ""
		if asignacion.FechaFin != nil {
			fin = fechas.FormatearDia(*asignacion.FechaFin)
		}
		tipo := "Reemplazo"
		if asignacion.Titular {
			tipo = "Titular"
		}
		estado := "Cerrada"
		if asignacion.Vigente {
			estado = "Vigente"
		}
		resultado = append(resultado, vista.AsignacionVista{
			ID:      asignacion.ID,
			Docente: asignacion.Docente,
			Curso:   asignacion.Curso,
			Seccion: asignacion.Seccion,
			Inicio:  fechas.FormatearDia(asignacion.FechaInicio),
			Fin:     fin,
			Tipo:    tipo,
			Estado:  estado,
		})
	}
	return resultado
}
