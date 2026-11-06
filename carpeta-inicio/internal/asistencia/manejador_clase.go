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

type parametrosClase struct {
	Fecha     time.Time
	SeccionID int64
	CursoID   int64
	Bloque    int
}

func (m *Manejador) seleccionClase(respuesta http.ResponseWriter, peticion *http.Request) {
	fecha := fechaDeConsulta(peticion)
	var clases []ClaseAsignada
	if intermediarios.RolUsuario(peticion.Context()) == intermediarios.RolDocente {
		var err error
		clases, err = m.servicio.ClasesAsignadas(peticion.Context(), m.usuarioID(peticion.Context()), fecha)
		if err != nil {
			m.renderizarSeleccion(respuesta, peticion, http.StatusUnprocessableEntity,
				fecha, nil, mensajeErrorClase(err))
			return
		}
	}
	m.renderizarSeleccion(respuesta, peticion, http.StatusOK, fecha, clases, "")
}

func (m *Manejador) verClase(respuesta http.ResponseWriter, peticion *http.Request) {
	parametros, err := leerParametrosClase(peticion)
	if err != nil {
		m.renderizarClase(respuesta, peticion, http.StatusUnprocessableEntity, nil, mensajeErrorClase(err))
		return
	}
	detalle, err := m.servicio.ListarClase(peticion.Context(), m.usuarioID(peticion.Context()),
		m.esAdministrador(peticion), parametros.Fecha, parametros.SeccionID, parametros.CursoID, parametros.Bloque)
	if err != nil {
		m.renderizarClase(respuesta, peticion, http.StatusUnprocessableEntity, nil, mensajeErrorClase(err))
		return
	}
	m.renderizarClase(respuesta, peticion, http.StatusOK, &detalle, "")
}

func (m *Manejador) guardarClase(respuesta http.ResponseWriter, peticion *http.Request) {
	if err := peticion.ParseForm(); err != nil {
		m.renderizarClase(respuesta, peticion, http.StatusBadRequest, nil, "Solicitud no válida")
		return
	}
	parametros, err := leerParametrosClase(peticion)
	if err != nil {
		m.renderizarClase(respuesta, peticion, http.StatusUnprocessableEntity, nil, mensajeErrorClase(err))
		return
	}
	cambios := cambiosDeFormulario(peticion)
	if err != nil {
		m.renderizarClase(respuesta, peticion, http.StatusUnprocessableEntity, nil, mensajeErrorClase(err))
		return
	}

	detalle, err := m.servicio.ListarClase(peticion.Context(), m.usuarioID(peticion.Context()),
		m.esAdministrador(peticion), parametros.Fecha, parametros.SeccionID, parametros.CursoID, parametros.Bloque)
	if err != nil {
		m.renderizarClase(respuesta, peticion, http.StatusUnprocessableEntity, nil, mensajeErrorClase(err))
		return
	}
	if err := m.servicio.GuardarRegistros(peticion.Context(), m.usuarioID(peticion.Context()),
		m.esAdministrador(peticion), parametros.Fecha, parametros.SeccionID, parametros.CursoID, parametros.Bloque, cambios); err != nil {
		m.renderizarClase(respuesta, peticion, http.StatusUnprocessableEntity, &detalle, mensajeErrorClase(err))
		return
	}
	http.Redirect(respuesta, peticion, urlClase(parametros)+"&ok=guardado", http.StatusSeeOther)
}

func (m *Manejador) marcarTodos(respuesta http.ResponseWriter, peticion *http.Request) {
	parametros, err := leerParametrosClase(peticion)
	if err != nil {
		m.renderizarClase(respuesta, peticion, http.StatusUnprocessableEntity, nil, mensajeErrorClase(err))
		return
	}
	if _, err := m.servicio.MarcarTodosPresentes(peticion.Context(), m.usuarioID(peticion.Context()),
		m.esAdministrador(peticion), parametros.Fecha, parametros.SeccionID, parametros.CursoID, parametros.Bloque); err != nil {
		detalle, errDetalle := m.servicio.ListarClase(peticion.Context(), m.usuarioID(peticion.Context()),
			m.esAdministrador(peticion), parametros.Fecha, parametros.SeccionID, parametros.CursoID, parametros.Bloque)
		if errDetalle != nil {
			detalle = ClaseDetalle{}
		}
		m.renderizarClase(respuesta, peticion, http.StatusUnprocessableEntity, &detalle, mensajeErrorClase(err))
		return
	}
	http.Redirect(respuesta, peticion, urlClase(parametros)+"&ok=todos", http.StatusSeeOther)
}

func (m *Manejador) renderizarSeleccion(respuesta http.ResponseWriter, peticion *http.Request, estado int, fecha time.Time, clases []ClaseAsignada, mensajeError string) {
	vistas := make([]vista.ClaseSeleccionVista, 0, len(clases))
	for _, clase := range clases {
		vistas = append(vistas, vista.ClaseSeleccionVista{
			Curso:   clase.Curso,
			Seccion: clase.Seccion,
			Horario: fechas.FormatearHora(clase.HoraInicio) + " – " + fechas.FormatearHora(clase.HoraFin),
			Bloque:  clase.BloqueNumero,
			URL: "/asistencia/clase?" + parametrosURL(fecha, clase.SeccionID,
				clase.CursoID, clase.BloqueNumero),
		})
	}
	m.renderizar(respuesta, peticion, estado, vista.Seleccion(
		m.nombreUsuario(peticion.Context()),
		intermediarios.RolUsuario(peticion.Context()),
		m.tokenCSRF(peticion.Context()),
		fechas.FormatearISO(fecha),
		vistas,
		mensajeError,
	))
}

func (m *Manejador) renderizarClase(respuesta http.ResponseWriter, peticion *http.Request, estado int, detalle *ClaseDetalle, mensajeError string) {
	var cabecera *vista.ClaseVista
	var filas []vista.AlumnoFilaVista
	if detalle != nil && detalle.Clase.SeccionID != 0 {
		cabecera = &vista.ClaseVista{
			SeccionID:  detalle.Clase.SeccionID,
			CursoID:    detalle.Clase.CursoID,
			Seccion:    detalle.Clase.Seccion,
			Curso:      detalle.Clase.Curso,
			Turno:      detalle.Clase.Turno,
			Fecha:      fechas.FormatearDia(detalle.Fecha),
			FechaISO:   fechas.FormatearISO(detalle.Fecha),
			Bloque:     detalle.Clase.Bloque,
			HoraInicio: fechas.FormatearHora(detalle.Clase.HoraInicio),
			HoraFin:    fechas.FormatearHora(detalle.Clase.HoraFin),
		}
		filas = aFilasVista(detalle.Alumnos)
	}
	m.renderizar(respuesta, peticion, estado, vista.Clase(
		m.nombreUsuario(peticion.Context()),
		intermediarios.RolUsuario(peticion.Context()),
		m.tokenCSRF(peticion.Context()),
		cabecera,
		filas,
		estadosVista(),
		mensajeCorrectoClase(peticion.URL.Query().Get("ok")),
		mensajeError,
	))
}

func leerParametrosClase(peticion *http.Request) (parametrosClase, error) {
	textoFecha := peticion.FormValue("fecha")
	if textoFecha == "" {
		textoFecha = peticion.URL.Query().Get("fecha")
	}
	fecha, err := fechas.ParsearDia(textoFecha)
	if err != nil {
		return parametrosClase{}, err
	}
	seccionID, err := strconv.ParseInt(peticion.FormValue("seccion_id"), 10, 64)
	if err != nil || seccionID <= 0 {
		return parametrosClase{}, ErrClaseNoProgramada
	}
	cursoID, err := strconv.ParseInt(peticion.FormValue("curso_id"), 10, 64)
	if err != nil || cursoID <= 0 {
		return parametrosClase{}, ErrClaseNoProgramada
	}
	bloque, err := strconv.Atoi(peticion.FormValue("bloque"))
	if err != nil || bloque <= 0 {
		return parametrosClase{}, ErrBloqueNoEncontrado
	}
	return parametrosClase{Fecha: fecha, SeccionID: seccionID, CursoID: cursoID, Bloque: bloque}, nil
}

func cambiosDeFormulario(peticion *http.Request) []CambioRegistro {
	cambios := make([]CambioRegistro, 0)
	for clave := range peticion.PostForm {
		if !strings.HasPrefix(clave, "estado_") {
			continue
		}
		alumnoID, err := strconv.ParseInt(strings.TrimPrefix(clave, "estado_"), 10, 64)
		if err != nil || alumnoID <= 0 {
			continue
		}
		cambios = append(cambios, CambioRegistro{
			AlumnoID:    alumnoID,
			Estado:      peticion.PostFormValue(clave),
			Observacion: strings.TrimSpace(peticion.PostFormValue("observacion_" + strconv.FormatInt(alumnoID, 10))),
		})
	}
	return cambios
}

func aFilasVista(alumnos []AlumnoClase) []vista.AlumnoFilaVista {
	filas := make([]vista.AlumnoFilaVista, 0, len(alumnos))
	for _, alumno := range alumnos {
		estado := EstadoPresente
		observacion := ""
		if alumno.Registro != nil {
			estado = alumno.Registro.Estado
			observacion = alumno.Registro.Observacion
		}
		textoQR := "Sin ingreso por QR"
		if alumno.Marcacion != nil {
			textoQR = "QR a las " + fechas.Hora(alumno.Marcacion.Hora)
		}
		filas = append(filas, vista.AlumnoFilaVista{
			ID:            alumno.Alumno.ID,
			Identificador: alumno.Alumno.Identificador,
			Nombre:        alumno.Alumno.NombreCompleto(),
			Estado:        estado,
			Observacion:   observacion,
			QRTexto:       textoQR,
			TieneQR:       alumno.Marcacion != nil,
			Diferencia:    alumno.Diferencia,
		})
	}
	return filas
}

func estadosVista() []vista.EstadoOpcion {
	return []vista.EstadoOpcion{
		{Valor: EstadoPresente, Etiqueta: "Presente"},
		{Valor: EstadoTardanza, Etiqueta: "Tardanza"},
		{Valor: EstadoFalta, Etiqueta: "Falta"},
		{Valor: EstadoFaltaJustificada, Etiqueta: "Falta justificada"},
	}
}

func mensajeCorrectoClase(codigo string) string {
	switch codigo {
	case "guardado":
		return "La asistencia de la clase se guardó correctamente."
	case "todos":
		return "Se marcó como presentes a los alumnos sin registro."
	default:
		return ""
	}
}

func mensajeErrorClase(err error) string {
	if errors.Is(err, ErrDocenteSinAsignacion) {
		return "No tiene acceso a esta clase en la fecha indicada."
	}
	if errors.Is(err, ErrFechaFutura) {
		return "No se puede registrar asistencia de una fecha futura."
	}
	if errors.Is(err, ErrDiaNoLaborable) {
		return "La fecha no es laborable para esta sección."
	}
	if errors.Is(err, ErrClaseNoProgramada) {
		return "La clase no corresponde al horario de la sección."
	}
	if errors.Is(err, ErrBloqueNoEncontrado) {
		return "El bloque indicado no existe en el turno de la sección."
	}
	if errors.Is(err, ErrEstadoInvalido) {
		return "El estado de asistencia no es válido."
	}
	if errors.Is(err, ErrObservacionLarga) {
		return "La observación no puede superar los 300 caracteres."
	}
	if errors.Is(err, ErrAlumnoFueraDeSeccion) {
		return "El alumno no pertenece a la sección en la fecha indicada."
	}
	if errors.Is(err, ErrPlazoVencido) {
		return "El plazo de edición terminó. Solo un administrador puede corregir este registro."
	}
	return "No se pudo completar la operación. Intente de nuevo."
}

func (m *Manejador) esAdministrador(peticion *http.Request) bool {
	return intermediarios.RolUsuario(peticion.Context()) == intermediarios.RolAdministrador
}

func fechaDeConsulta(peticion *http.Request) time.Time {
	fecha, err := fechas.ParsearDia(peticion.URL.Query().Get("fecha"))
	if err != nil {
		return fechas.Hoy()
	}
	return fecha
}

func parametrosURL(fecha time.Time, seccionID, cursoID int64, bloque int) string {
	return fmt.Sprintf("fecha=%s&seccion_id=%d&curso_id=%d&bloque=%d",
		fechas.FormatearISO(fecha), seccionID, cursoID, bloque)
}

func urlClase(parametros parametrosClase) string {
	return "/asistencia/clase?" + parametrosURL(parametros.Fecha, parametros.SeccionID, parametros.CursoID, parametros.Bloque)
}
