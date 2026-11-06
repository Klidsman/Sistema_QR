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
	HorarioVigente(ctx context.Context, turno string, fecha time.Time) (HorarioTurno, error)
	ListarHorarios(ctx context.Context, turno string) ([]HorarioTurno, error)
	ConfigurarHorario(ctx context.Context, usuarioID int64, horario HorarioTurno) error
	BloquesVigentes(ctx context.Context, turno string, fecha time.Time) ([]BloqueClase, error)
	CrearBloque(ctx context.Context, usuarioID int64, bloque BloqueClase) (int64, error)
	AjustarHorasBloque(ctx context.Context, usuarioID, bloqueID int64, inicio, fin, vigencia time.Time) error
	ListarCursos(ctx context.Context, anioID int64) ([]Curso, error)
	CrearCurso(ctx context.Context, usuarioID int64, datos DatosCurso) (int64, error)
	DesactivarCurso(ctx context.Context, usuarioID, cursoID int64) error
	ObtenerGrilla(ctx context.Context, seccionID int64, fecha time.Time) (GrillaHorario, error)
	GuardarCeldas(ctx context.Context, usuarioID, seccionID int64, fecha time.Time, cambios []CambioCelda, confirmar bool) error
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
		lectura.Get("/configuracion/turnos", m.turnos)
		lectura.Get("/configuracion/bloques", m.bloques)
		lectura.Get("/configuracion/cursos", m.cursos)
		lectura.Get("/configuracion/horarios", m.horarios)
	})

	enrutador.Group(func(edicionCompartida chi.Router) {
		edicionCompartida.Use(intermediarios.RequerirRolConEscritura(m.sesiones,
			intermediarios.RolAdministrador, intermediarios.RolDirector))
		edicionCompartida.Post("/configuracion/turnos/{turno}", m.configurarHorario)
		edicionCompartida.Post("/configuracion/bloques/{id}/horas", m.ajustarHorasBloque)
	})

	enrutador.Group(func(escritura chi.Router) {
		escritura.Use(intermediarios.RequerirRol(m.sesiones, intermediarios.RolAdministrador))
		escritura.Post("/configuracion/anio", m.crearAnio)
		escritura.Post("/configuracion/anio/cierre", m.cerrarAnio)
		escritura.Post("/configuracion/grados", m.crearGrado)
		escritura.Post("/configuracion/secciones", m.crearSeccion)
		escritura.Post("/configuracion/secciones/{id}/turno", m.cambiarTurno)
		escritura.Post("/configuracion/bloques", m.crearBloque)
		escritura.Post("/configuracion/cursos", m.crearCurso)
		escritura.Post("/configuracion/cursos/{id}/desactivacion", m.desactivarCurso)
		escritura.Post("/configuracion/horarios", m.guardarHorario)
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

func (m *Manejador) turnos(respuesta http.ResponseWriter, peticion *http.Request) {
	m.renderizarTurnos(respuesta, peticion, http.StatusOK, "", peticion.URL.Query().Get("ok"))
}

func (m *Manejador) configurarHorario(respuesta http.ResponseWriter, peticion *http.Request) {
	turno := chi.URLParam(peticion, "turno")
	horario, err := horarioDeFormulario(peticion, turno)
	if err != nil {
		m.renderizarTurnos(respuesta, peticion, http.StatusUnprocessableEntity, mensajeDeError(err), "")
		return
	}
	if err := m.configuracion.ConfigurarHorario(peticion.Context(), m.usuarioID(peticion.Context()), horario); err != nil {
		m.renderizarTurnos(respuesta, peticion, http.StatusUnprocessableEntity, mensajeDeError(err), "")
		return
	}
	http.Redirect(respuesta, peticion, "/configuracion/turnos?ok=horario", http.StatusSeeOther)
}

func (m *Manejador) renderizarTurnos(respuesta http.ResponseWriter, peticion *http.Request, estado int, mensajeError, codigoCorrecto string) {
	ctx := peticion.Context()
	activo, err := m.configuracion.AnioActivo(ctx)
	if err != nil {
		m.registro.Error("no se pudo buscar el año activo", "error", err)
		http.Error(respuesta, "No se pudo cargar la configuración", http.StatusInternalServerError)
		return
	}

	hoy := fechas.Hoy()
	var manana, tarde *HorarioTurno
	var historial []HorarioTurno
	if activo != nil {
		manana, err = m.horarioVigente(ctx, TurnoManana, hoy)
		if err != nil {
			m.registro.Error("no se pudo cargar el horario de mañana", "error", err)
			http.Error(respuesta, "No se pudo cargar la configuración", http.StatusInternalServerError)
			return
		}
		tarde, err = m.horarioVigente(ctx, TurnoTarde, hoy)
		if err != nil {
			m.registro.Error("no se pudo cargar el horario de tarde", "error", err)
			http.Error(respuesta, "No se pudo cargar la configuración", http.StatusInternalServerError)
			return
		}
		for _, turno := range Turnos() {
			horarios, err := m.configuracion.ListarHorarios(ctx, turno)
			if err != nil {
				m.registro.Error("no se pudo listar el historial de horarios", "error", err)
				http.Error(respuesta, "No se pudo cargar la configuración", http.StatusInternalServerError)
				return
			}
			historial = append(historial, horarios...)
		}
	}

	m.renderizar(respuesta, peticion, estado, vista.Turnos(
		m.nombreUsuario(ctx),
		intermediarios.RolUsuario(ctx),
		m.tokenCSRF(ctx),
		aAnioVista(activo),
		aHorarioVista(manana),
		aHorarioVista(tarde),
		aHorariosVista(historial),
		mensajeCorrectoHorario(codigoCorrecto),
		mensajeError,
	))
}

func (m *Manejador) horarioVigente(ctx context.Context, turno string, fecha time.Time) (*HorarioTurno, error) {
	horario, err := m.configuracion.HorarioVigente(ctx, turno, fecha)
	if errors.Is(err, ErrHorarioNoEncontrado) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &horario, nil
}

func horarioDeFormulario(peticion *http.Request, turno string) (HorarioTurno, error) {
	if err := ValidarTurno(turno); err != nil {
		return HorarioTurno{}, err
	}
	ingreso, err := fechas.ParsearHora(peticion.PostFormValue("hora_ingreso"))
	if err != nil {
		return HorarioTurno{}, err
	}
	tolerancia, err := strconv.Atoi(strings.TrimSpace(peticion.PostFormValue("tolerancia_minutos")))
	if err != nil {
		return HorarioTurno{}, ErrToleranciaInvalida
	}
	limite, err := fechas.ParsearHora(peticion.PostFormValue("hora_limite_tardanza"))
	if err != nil {
		return HorarioTurno{}, err
	}
	cierre, err := fechas.ParsearHora(peticion.PostFormValue("hora_cierre"))
	if err != nil {
		return HorarioTurno{}, err
	}
	vigenteDesde, err := fechas.ParsearDia(peticion.PostFormValue("vigente_desde"))
	if err != nil {
		return HorarioTurno{}, err
	}
	return HorarioTurno{
		Turno:              turno,
		HoraIngreso:        ingreso,
		ToleranciaMinutos:  tolerancia,
		HoraLimiteTardanza: limite,
		HoraCierre:         cierre,
		VigenteDesde:       vigenteDesde,
	}, nil
}

func mensajeCorrectoHorario(codigo string) string {
	if codigo == "horario" {
		return "El horario del turno se guardó correctamente."
	}
	return ""
}

func aHorarioVista(horario *HorarioTurno) *vista.HorarioVista {
	if horario == nil {
		return nil
	}
	return &vista.HorarioVista{
		Turno:              horario.Turno,
		HoraIngreso:        fechas.FormatearHora(horario.HoraIngreso),
		ToleranciaMinutos:  horario.ToleranciaMinutos,
		HoraLimiteTardanza: fechas.FormatearHora(horario.HoraLimiteTardanza),
		HoraCierre:         fechas.FormatearHora(horario.HoraCierre),
		VigenteDesde:       horario.VigenteDesde,
	}
}

func aHorariosVista(horarios []HorarioTurno) []vista.HorarioVista {
	resultado := make([]vista.HorarioVista, 0, len(horarios))
	for i := range horarios {
		resultado = append(resultado, *aHorarioVista(&horarios[i]))
	}
	return resultado
}

func (m *Manejador) bloques(respuesta http.ResponseWriter, peticion *http.Request) {
	m.renderizarBloques(respuesta, peticion, http.StatusOK, turnoDeConsulta(peticion),
		"", peticion.URL.Query().Get("ok"))
}

func (m *Manejador) crearBloque(respuesta http.ResponseWriter, peticion *http.Request) {
	bloque, err := bloqueDeFormulario(peticion)
	if err != nil {
		m.renderizarBloques(respuesta, peticion, http.StatusUnprocessableEntity, turnoDeConsulta(peticion),
			mensajeDeError(err), "")
		return
	}
	if _, err := m.configuracion.CrearBloque(peticion.Context(), m.usuarioID(peticion.Context()), bloque); err != nil {
		m.renderizarBloques(respuesta, peticion, http.StatusUnprocessableEntity, bloque.Turno,
			mensajeDeError(err), "")
		return
	}
	http.Redirect(respuesta, peticion,
		"/configuracion/bloques?turno="+bloque.Turno+"&ok=creado", http.StatusSeeOther)
}

func (m *Manejador) ajustarHorasBloque(respuesta http.ResponseWriter, peticion *http.Request) {
	turno := turnoDeConsulta(peticion)
	bloqueID, err := strconv.ParseInt(chi.URLParam(peticion, "id"), 10, 64)
	if err != nil || bloqueID <= 0 {
		m.renderizarBloques(respuesta, peticion, http.StatusUnprocessableEntity, turno,
			mensajeDeError(ErrBloqueNoEncontrado), "")
		return
	}
	inicio, err := fechas.ParsearHora(peticion.PostFormValue("hora_inicio"))
	if err != nil {
		m.renderizarBloques(respuesta, peticion, http.StatusUnprocessableEntity, turno,
			mensajeDeError(err), "")
		return
	}
	fin, err := fechas.ParsearHora(peticion.PostFormValue("hora_fin"))
	if err != nil {
		m.renderizarBloques(respuesta, peticion, http.StatusUnprocessableEntity, turno,
			mensajeDeError(err), "")
		return
	}
	vigencia, err := fechas.ParsearDia(peticion.PostFormValue("vigente_desde"))
	if err != nil {
		m.renderizarBloques(respuesta, peticion, http.StatusUnprocessableEntity, turno,
			mensajeDeError(err), "")
		return
	}
	if err := m.configuracion.AjustarHorasBloque(peticion.Context(), m.usuarioID(peticion.Context()),
		bloqueID, inicio, fin, vigencia); err != nil {
		m.renderizarBloques(respuesta, peticion, http.StatusUnprocessableEntity, turno,
			mensajeDeError(err), "")
		return
	}
	http.Redirect(respuesta, peticion,
		"/configuracion/bloques?turno="+turno+"&ok=ajustado", http.StatusSeeOther)
}

func (m *Manejador) renderizarBloques(respuesta http.ResponseWriter, peticion *http.Request, estado int, turno, mensajeError, codigoCorrecto string) {
	ctx := peticion.Context()
	activo, err := m.configuracion.AnioActivo(ctx)
	if err != nil {
		m.registro.Error("no se pudo buscar el año activo", "error", err)
		http.Error(respuesta, "No se pudo cargar la configuración", http.StatusInternalServerError)
		return
	}
	var bloques []BloqueClase
	if activo != nil {
		bloques, err = m.configuracion.BloquesVigentes(ctx, turno, fechas.Hoy())
		if err != nil {
			m.registro.Error("no se pudieron cargar los bloques", "error", err)
			http.Error(respuesta, "No se pudo cargar la configuración", http.StatusInternalServerError)
			return
		}
	}

	m.renderizar(respuesta, peticion, estado, vista.Bloques(
		m.nombreUsuario(ctx),
		intermediarios.RolUsuario(ctx),
		m.tokenCSRF(ctx),
		aAnioVista(activo),
		turno,
		aBloquesVista(bloques),
		mensajeCorrectoBloque(codigoCorrecto),
		mensajeError,
	))
}

func turnoDeConsulta(peticion *http.Request) string {
	turno := peticion.URL.Query().Get("turno")
	if !TurnoValido(turno) {
		return TurnoManana
	}
	return turno
}

func bloqueDeFormulario(peticion *http.Request) (BloqueClase, error) {
	turno := peticion.PostFormValue("turno")
	if err := ValidarTurno(turno); err != nil {
		return BloqueClase{}, err
	}
	numero, err := strconv.Atoi(strings.TrimSpace(peticion.PostFormValue("numero")))
	if err != nil {
		return BloqueClase{}, ErrBloqueNumeroInvalido
	}
	inicio, err := fechas.ParsearHora(peticion.PostFormValue("hora_inicio"))
	if err != nil {
		return BloqueClase{}, err
	}
	fin, err := fechas.ParsearHora(peticion.PostFormValue("hora_fin"))
	if err != nil {
		return BloqueClase{}, err
	}
	vigencia, err := fechas.ParsearDia(peticion.PostFormValue("vigente_desde"))
	if err != nil {
		return BloqueClase{}, err
	}
	return BloqueClase{
		Turno:        turno,
		Numero:       numero,
		HoraInicio:   inicio,
		HoraFin:      fin,
		VigenteDesde: vigencia,
	}, nil
}

func mensajeCorrectoBloque(codigo string) string {
	switch codigo {
	case "creado":
		return "El bloque de clase se creó correctamente."
	case "ajustado":
		return "Las horas del bloque se actualizaron correctamente."
	default:
		return ""
	}
}

func aBloquesVista(bloques []BloqueClase) []vista.BloqueVista {
	resultado := make([]vista.BloqueVista, 0, len(bloques))
	for _, bloque := range bloques {
		resultado = append(resultado, vista.BloqueVista{
			ID:           bloque.ID,
			Turno:        bloque.Turno,
			Numero:       bloque.Numero,
			HoraInicio:   fechas.FormatearHora(bloque.HoraInicio),
			HoraFin:      fechas.FormatearHora(bloque.HoraFin),
			VigenteDesde: bloque.VigenteDesde,
		})
	}
	return resultado
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
	case errors.Is(err, ErrHorarioNoEncontrado):
		return "No hay un horario vigente para ese turno en la fecha."
	case errors.Is(err, ErrToleranciaInvalida):
		return "La tolerancia debe estar entre 0 y 120 minutos."
	case errors.Is(err, ErrCierreNoPosterior):
		return "La hora de cierre debe ser posterior a la hora límite de tardanza."
	case errors.Is(err, ErrToleranciaExcedida):
		return "La hora de ingreso más la tolerancia no puede superar la hora límite de tardanza."
	case errors.Is(err, ErrHorarioFueraDeOrden):
		return "Las horas del turno deben ir de ingreso a límite de tardanza y a cierre."
	case errors.Is(err, ErrCruzaMedianoche):
		return "Los horarios deben estar dentro del mismo día."
	case errors.Is(err, ErrHorariosSuperpuestos):
		return "Los horarios de mañana y tarde se superponen."
	case errors.Is(err, ErrBloqueNoEncontrado):
		return "El bloque de clase no existe."
	case errors.Is(err, ErrBloqueNumeroInvalido):
		return "El número de bloque debe estar entre 1 y 20."
	case errors.Is(err, ErrBloqueHorasInvalidas):
		return "La hora de fin del bloque debe ser posterior a la de inicio."
	case errors.Is(err, ErrBloqueFueraDeHorario):
		return "El bloque debe estar dentro del horario del turno."
	case errors.Is(err, ErrBloqueSolapado):
		return "El bloque se superpone con otro bloque del mismo turno."
	case errors.Is(err, ErrCursoSinGrados):
		return "El curso debe estar vinculado al menos a un grado."
	case errors.Is(err, ErrCursoExistente):
		return "Ya existe un curso con ese nombre en el año."
	case errors.Is(err, ErrCursoNoEncontrado):
		return "El curso indicado no existe."
	case errors.Is(err, ErrCursoNoVinculado):
		return "El curso no está vinculado al grado de la sección."
	case errors.Is(err, ErrSeccionSinTurno):
		return "La sección no tiene un turno vigente en la fecha indicada."
	case errors.Is(err, ErrSeccionNoEncontrada):
		return "La sección no existe en el año escolar."
	case errors.Is(err, ErrRequiereConfirmacion):
		return "La celda ya tiene un curso asignado. Confirme el reemplazo para continuar."
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
