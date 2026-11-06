package academico

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"asistencia/internal/plataforma/fechas"
	"asistencia/internal/plataforma/intermediarios"
	vista "asistencia/web/plantillas/configuracion"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
)

// ServicioAcademico son las operaciones de configuración que usa el manejador.
type ServicioAcademico interface {
	Resumen(ctx context.Context) (ResumenConfiguracion, error)
	ListarAnios(ctx context.Context) ([]AnioEscolar, error)
	AnioActivo(ctx context.Context) (*AnioEscolar, error)
	ListarGrados(ctx context.Context, anioID int64) ([]Grado, error)
	ListarSecciones(ctx context.Context, anioID int64) ([]Seccion, error)
	ListarTurnosSeccion(ctx context.Context, seccionID int64) ([]TurnoSeccion, error)
	CrearAnio(ctx context.Context, usuarioID int64, datos DatosAnio) (int64, error)
	CerrarAnio(ctx context.Context, usuarioID, anioID int64) error
	CrearGrado(ctx context.Context, usuarioID int64, datos DatosGrado) (int64, error)
	CrearSeccion(ctx context.Context, usuarioID int64, datos DatosSeccion) (int64, error)
	CambiarTurno(ctx context.Context, usuarioID, seccionID int64, turno string, vigenteDesde time.Time) error
}

// DependenciasConfiguracion reúne lo que el manejador de configuración necesita.
type DependenciasConfiguracion struct {
	Configuracion ServicioAcademico
	Sesiones      *scs.SessionManager
	Registro      *slog.Logger
}

// Manejador atiende las pantallas de configuración académica.
type Manejador struct {
	configuracion ServicioAcademico
	sesiones      *scs.SessionManager
	registro      *slog.Logger
}

// NuevoManejador crea el manejador de configuración académica.
func NuevoManejador(dependencias DependenciasConfiguracion) *Manejador {
	return &Manejador{
		configuracion: dependencias.Configuracion,
		sesiones:      dependencias.Sesiones,
		registro:      dependencias.Registro,
	}
}

// Montar registra las rutas de configuración agrupadas por rol.
func (m *Manejador) Montar(enrutador chi.Router) {
	enrutador.Group(func(lectura chi.Router) {
		lectura.Use(intermediarios.RequerirRol(m.sesiones,
			intermediarios.RolAdministrador, intermediarios.RolDirector))
		lectura.Get("/configuracion", m.panel)
		lectura.Get("/configuracion/anio", m.anio)
		lectura.Get("/configuracion/grados", m.grados)
		lectura.Get("/configuracion/secciones", m.secciones)
	})

	enrutador.Group(func(escritura chi.Router) {
		escritura.Use(intermediarios.RequerirRol(m.sesiones, intermediarios.RolAdministrador))
		escritura.Post("/configuracion/anio", m.crearAnio)
		escritura.Post("/configuracion/anio/cierre", m.cerrarAnio)
		escritura.Post("/configuracion/grados", m.crearGrado)
		escritura.Post("/configuracion/secciones", m.crearSeccion)
		escritura.Post("/configuracion/secciones/{id}/turno", m.cambiarTurno)
	})
}

func (m *Manejador) panel(respuesta http.ResponseWriter, peticion *http.Request) {
	resumen, err := m.configuracion.Resumen(peticion.Context())
	if err != nil {
		m.registro.Error("no se pudo cargar el resumen de configuración", "error", err)
		http.Error(respuesta, "No se pudo cargar la configuración", http.StatusInternalServerError)
		return
	}

	anio := ""
	estado := ""
	if resumen.Anio != nil {
		anio = resumen.Anio.Nombre
		estado = resumen.Anio.Estado
	}

	m.renderizar(respuesta, peticion, http.StatusOK, vista.Panel(
		m.nombreUsuario(peticion.Context()),
		intermediarios.RolUsuario(peticion.Context()),
		m.tokenCSRF(peticion.Context()),
		anio,
		estado,
		resumen.Grados,
		resumen.Secciones,
	))
}

func (m *Manejador) anio(respuesta http.ResponseWriter, peticion *http.Request) {
	m.renderizarAnio(respuesta, peticion, http.StatusOK, "", peticion.URL.Query().Get("ok"), "")
}

func (m *Manejador) crearAnio(respuesta http.ResponseWriter, peticion *http.Request) {
	datos, err := datosAnioDeFormulario(peticion)
	if err != nil {
		m.renderizarAnio(respuesta, peticion, http.StatusUnprocessableEntity, mensajeDeError(err), "", "")
		return
	}
	if _, err := m.configuracion.CrearAnio(peticion.Context(), m.usuarioID(peticion.Context()), datos); err != nil {
		m.renderizarAnio(respuesta, peticion, http.StatusUnprocessableEntity, mensajeDeError(err), "", "")
		return
	}
	http.Redirect(respuesta, peticion, "/configuracion/anio?ok=creado", http.StatusSeeOther)
}

func (m *Manejador) cerrarAnio(respuesta http.ResponseWriter, peticion *http.Request) {
	anio, err := m.configuracion.AnioActivo(peticion.Context())
	if err != nil || anio == nil {
		m.renderizarAnio(respuesta, peticion, http.StatusUnprocessableEntity,
			mensajeDeError(ErrAnioNoEncontrado), "", "")
		return
	}
	if err := m.configuracion.CerrarAnio(peticion.Context(), m.usuarioID(peticion.Context()), anio.ID); err != nil {
		m.renderizarAnio(respuesta, peticion, http.StatusUnprocessableEntity, mensajeDeError(err), "", "")
		return
	}
	http.Redirect(respuesta, peticion, "/configuracion/anio?ok=cerrado", http.StatusSeeOther)
}

func (m *Manejador) grados(respuesta http.ResponseWriter, peticion *http.Request) {
	m.renderizarGrados(respuesta, peticion, http.StatusOK, "", peticion.URL.Query().Get("ok"), "")
}

func (m *Manejador) crearGrado(respuesta http.ResponseWriter, peticion *http.Request) {
	anio, err := m.configuracion.AnioActivo(peticion.Context())
	if err != nil || anio == nil {
		m.renderizarGrados(respuesta, peticion, http.StatusUnprocessableEntity,
			mensajeDeError(ErrAnioNoEncontrado), "", "")
		return
	}
	nombre := peticion.PostFormValue("nombre")
	if _, err := m.configuracion.CrearGrado(peticion.Context(), m.usuarioID(peticion.Context()), DatosGrado{
		AnioEscolarID: anio.ID,
		Nombre:        nombre,
	}); err != nil {
		m.renderizarGrados(respuesta, peticion, http.StatusUnprocessableEntity, mensajeDeError(err), "", nombre)
		return
	}
	http.Redirect(respuesta, peticion, "/configuracion/grados?ok=creado", http.StatusSeeOther)
}

func (m *Manejador) secciones(respuesta http.ResponseWriter, peticion *http.Request) {
	m.renderizarSecciones(respuesta, peticion, http.StatusOK, "", peticion.URL.Query().Get("ok"), "")
}

func (m *Manejador) crearSeccion(respuesta http.ResponseWriter, peticion *http.Request) {
	anio, err := m.configuracion.AnioActivo(peticion.Context())
	if err != nil || anio == nil {
		m.renderizarSecciones(respuesta, peticion, http.StatusUnprocessableEntity,
			mensajeDeError(ErrAnioNoEncontrado), "", "")
		return
	}
	datos, err := datosSeccionDeFormulario(peticion)
	if err != nil {
		m.renderizarSecciones(respuesta, peticion, http.StatusUnprocessableEntity, mensajeDeError(err), "", "")
		return
	}
	if _, err := m.configuracion.CrearSeccion(peticion.Context(), m.usuarioID(peticion.Context()), datos); err != nil {
		m.renderizarSecciones(respuesta, peticion, http.StatusUnprocessableEntity, mensajeDeError(err), "", datos.Nombre)
		return
	}
	http.Redirect(respuesta, peticion, "/configuracion/secciones?ok=creado", http.StatusSeeOther)
}

func (m *Manejador) cambiarTurno(respuesta http.ResponseWriter, peticion *http.Request) {
	seccionID, err := strconv.ParseInt(chi.URLParam(peticion, "id"), 10, 64)
	if err != nil || seccionID <= 0 {
		m.renderizarSecciones(respuesta, peticion, http.StatusUnprocessableEntity,
			mensajeDeError(ErrGradoNoEncontrado), "", "")
		return
	}
	vigenteDesde, err := fechas.ParsearDia(peticion.PostFormValue("vigente_desde"))
	if err != nil {
		m.renderizarSecciones(respuesta, peticion, http.StatusUnprocessableEntity, mensajeDeError(err), "", "")
		return
	}
	if err := m.configuracion.CambiarTurno(peticion.Context(), m.usuarioID(peticion.Context()),
		seccionID, peticion.PostFormValue("turno"), vigenteDesde); err != nil {
		m.renderizarSecciones(respuesta, peticion, http.StatusUnprocessableEntity, mensajeDeError(err), "", "")
		return
	}
	http.Redirect(respuesta, peticion, "/configuracion/secciones?ok=turno", http.StatusSeeOther)
}

func (m *Manejador) renderizarAnio(respuesta http.ResponseWriter, peticion *http.Request, estado int, mensajeError, codigoCorrecto, nombre string) {
	anios, err := m.configuracion.ListarAnios(peticion.Context())
	if err != nil {
		m.registro.Error("no se pudieron listar los años escolares", "error", err)
		http.Error(respuesta, "No se pudo cargar la configuración", http.StatusInternalServerError)
		return
	}
	activo, err := m.configuracion.AnioActivo(peticion.Context())
	if err != nil {
		m.registro.Error("no se pudo buscar el año activo", "error", err)
		http.Error(respuesta, "No se pudo cargar la configuración", http.StatusInternalServerError)
		return
	}
	m.renderizar(respuesta, peticion, estado, vista.Anio(
		m.nombreUsuario(peticion.Context()),
		intermediarios.RolUsuario(peticion.Context()),
		m.tokenCSRF(peticion.Context()),
		aAnioVista(activo),
		aAniosVista(anios),
		nombre,
		mensajeCorrectoAnio(codigoCorrecto),
		mensajeError,
	))
}

func (m *Manejador) renderizarGrados(respuesta http.ResponseWriter, peticion *http.Request, estado int, mensajeError, codigoCorrecto, nombre string) {
	activo, grados, err := m.datosGrados(peticion.Context())
	if err != nil {
		m.registro.Error("no se pudieron cargar los grados", "error", err)
		http.Error(respuesta, "No se pudo cargar la configuración", http.StatusInternalServerError)
		return
	}
	m.renderizar(respuesta, peticion, estado, vista.Grados(
		m.nombreUsuario(peticion.Context()),
		intermediarios.RolUsuario(peticion.Context()),
		m.tokenCSRF(peticion.Context()),
		aAnioVista(activo),
		aGradosVista(grados),
		nombre,
		mensajeCorrectoGrado(codigoCorrecto),
		mensajeError,
	))
}

func (m *Manejador) renderizarSecciones(respuesta http.ResponseWriter, peticion *http.Request, estado int, mensajeError, codigoCorrecto, nombre string) {
	activo, err := m.configuracion.AnioActivo(peticion.Context())
	if err != nil {
		m.registro.Error("no se pudo buscar el año activo", "error", err)
		http.Error(respuesta, "No se pudo cargar la configuración", http.StatusInternalServerError)
		return
	}
	var grados []Grado
	var secciones []Seccion
	if activo != nil {
		grados, err = m.configuracion.ListarGrados(peticion.Context(), activo.ID)
		if err != nil {
			m.registro.Error("no se pudieron listar los grados", "error", err)
			http.Error(respuesta, "No se pudo cargar la configuración", http.StatusInternalServerError)
			return
		}
		secciones, err = m.configuracion.ListarSecciones(peticion.Context(), activo.ID)
		if err != nil {
			m.registro.Error("no se pudieron listar las secciones", "error", err)
			http.Error(respuesta, "No se pudo cargar la configuración", http.StatusInternalServerError)
			return
		}
	}
	m.renderizar(respuesta, peticion, estado, vista.Secciones(
		m.nombreUsuario(peticion.Context()),
		intermediarios.RolUsuario(peticion.Context()),
		m.tokenCSRF(peticion.Context()),
		aAnioVista(activo),
		aGradosVista(grados),
		aSeccionesVista(secciones),
		nombre,
		mensajeCorrectoSeccion(codigoCorrecto),
		mensajeError,
	))
}

func (m *Manejador) datosGrados(ctx context.Context) (*AnioEscolar, []Grado, error) {
	activo, err := m.configuracion.AnioActivo(ctx)
	if err != nil || activo == nil {
		return activo, nil, err
	}
	grados, err := m.configuracion.ListarGrados(ctx, activo.ID)
	if err != nil {
		return nil, nil, err
	}
	return activo, grados, nil
}

func (m *Manejador) renderizar(respuesta http.ResponseWriter, peticion *http.Request, estado int, componente interface {
	Render(context.Context, io.Writer) error
}) {
	respuesta.Header().Set("Content-Type", "text/html; charset=utf-8")
	respuesta.WriteHeader(estado)
	if err := componente.Render(peticion.Context(), respuesta); err != nil {
		m.registro.Error("no se pudo renderizar la vista de configuración", "error", err)
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

func datosAnioDeFormulario(peticion *http.Request) (DatosAnio, error) {
	inicio, err := fechas.ParsearDia(peticion.PostFormValue("fecha_inicio"))
	if err != nil {
		return DatosAnio{}, err
	}
	fin, err := fechas.ParsearDia(peticion.PostFormValue("fecha_fin"))
	if err != nil {
		return DatosAnio{}, err
	}
	return DatosAnio{
		Nombre:      peticion.PostFormValue("nombre"),
		FechaInicio: inicio,
		FechaFin:    fin,
	}, nil
}

func datosSeccionDeFormulario(peticion *http.Request) (DatosSeccion, error) {
	gradoID, err := strconv.ParseInt(strings.TrimSpace(peticion.PostFormValue("grado_id")), 10, 64)
	if err != nil || gradoID <= 0 {
		return DatosSeccion{}, ErrGradoNoEncontrado
	}
	vigenteDesde, err := fechas.ParsearDia(peticion.PostFormValue("vigente_desde"))
	if err != nil {
		return DatosSeccion{}, err
	}
	return DatosSeccion{
		GradoID:      gradoID,
		Nombre:       peticion.PostFormValue("nombre"),
		Turno:        peticion.PostFormValue("turno"),
		VigenteDesde: vigenteDesde,
	}, nil
}

func mensajeDeError(err error) string {
	switch {
	case errors.Is(err, ErrAnioActivoExistente):
		return "Ya existe un año escolar activo. Cierre el año actual antes de crear otro."
	case errors.Is(err, ErrAnioNombreExistente):
		return "Ya existe un año escolar con ese nombre."
	case errors.Is(err, ErrAnioNoEncontrado):
		return "No hay un año escolar activo. Cree el año antes de continuar."
	case errors.Is(err, ErrAnioCerrado):
		return "El año escolar está cerrado y no admite cambios."
	case errors.Is(err, ErrGradoNoEncontrado):
		return "El grado indicado no existe en el año activo."
	case errors.Is(err, ErrGradoExistente):
		return "Ya existe un grado con ese nombre en el año."
	case errors.Is(err, ErrSeccionExistente):
		return "Ya existe una sección con ese nombre en el grado."
	case errors.Is(err, ErrFechasAnioInvalidas):
		return "La fecha de fin debe ser posterior a la fecha de inicio."
	case errors.Is(err, ErrVigenciaPasada):
		return "La fecha de vigencia debe ser hoy o una fecha futura."
	case errors.Is(err, ErrNombreInvalido):
		return "El nombre debe tener entre 2 y 80 caracteres."
	case errors.Is(err, ErrNombreSeccionInvalido):
		return "El nombre de la sección debe tener entre 1 y 10 caracteres."
	case errors.Is(err, ErrTurnoInvalido):
		return "El turno debe ser mañana o tarde."
	default:
		return "No se pudo guardar el cambio. Revise los datos e intente de nuevo."
	}
}

func mensajeCorrectoAnio(codigo string) string {
	switch codigo {
	case "creado":
		return "El año escolar se creó correctamente."
	case "cerrado":
		return "El año escolar se cerró correctamente."
	default:
		return ""
	}
}

func mensajeCorrectoGrado(codigo string) string {
	if codigo == "creado" {
		return "El grado se creó correctamente."
	}
	return ""
}

func aAnioVista(anio *AnioEscolar) *vista.AnioVista {
	if anio == nil {
		return nil
	}
	return &vista.AnioVista{
		ID:          anio.ID,
		Nombre:      anio.Nombre,
		FechaInicio: anio.FechaInicio,
		FechaFin:    anio.FechaFin,
		Estado:      anio.Estado,
	}
}

func aAniosVista(anios []AnioEscolar) []vista.AnioVista {
	resultado := make([]vista.AnioVista, 0, len(anios))
	for _, anio := range anios {
		resultado = append(resultado, vista.AnioVista{
			ID:          anio.ID,
			Nombre:      anio.Nombre,
			FechaInicio: anio.FechaInicio,
			FechaFin:    anio.FechaFin,
			Estado:      anio.Estado,
		})
	}
	return resultado
}

func aGradosVista(grados []Grado) []vista.GradoVista {
	resultado := make([]vista.GradoVista, 0, len(grados))
	for _, grado := range grados {
		resultado = append(resultado, vista.GradoVista{ID: grado.ID, Nombre: grado.Nombre})
	}
	return resultado
}

func aSeccionesVista(secciones []Seccion) []vista.SeccionVista {
	resultado := make([]vista.SeccionVista, 0, len(secciones))
	for _, seccion := range secciones {
		resultado = append(resultado, vista.SeccionVista{
			ID:     seccion.ID,
			Grado:  seccion.Grado,
			Nombre: seccion.Nombre,
			Turno:  seccion.Turno,
		})
	}
	return resultado
}

func mensajeCorrectoSeccion(codigo string) string {
	switch codigo {
	case "creado":
		return "La sección se creó correctamente."
	case "turno":
		return "El turno de la sección se actualizó correctamente."
	default:
		return ""
	}
}
