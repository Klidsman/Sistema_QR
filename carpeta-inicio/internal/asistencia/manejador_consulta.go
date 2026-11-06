package asistencia

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"asistencia/internal/plataforma/fechas"
	"asistencia/internal/plataforma/intermediarios"
	vista "asistencia/web/plantillas/asistencia"
)

func (m *Manejador) consultarAsistencia(respuesta http.ResponseWriter, peticion *http.Request) {
	seccionID, _ := strconv.ParseInt(peticion.URL.Query().Get("seccion_id"), 10, 64)
	m.renderizarConsulta(respuesta, peticion, http.StatusOK, fechaDeConsulta(peticion), seccionID,
		mensajeCorrectoConsulta(peticion.URL.Query().Get("ok")), "")
}

func (m *Manejador) corregirAsistencia(respuesta http.ResponseWriter, peticion *http.Request) {
	if err := peticion.ParseForm(); err != nil {
		http.Error(respuesta, "Solicitud no válida", http.StatusBadRequest)
		return
	}
	fecha := fechaDeConsulta(peticion)
	seccionID, _ := strconv.ParseInt(peticion.PostFormValue("seccion_id"), 10, 64)
	registroID, err := strconv.ParseInt(peticion.PostFormValue("registro_id"), 10, 64)
	if err != nil || registroID <= 0 {
		m.renderizarConsulta(respuesta, peticion, http.StatusUnprocessableEntity, fecha, seccionID, "",
			mensajeErrorConsulta(ErrRegistroNoEncontrado))
		return
	}
	err = m.servicio.CorregirRegistro(peticion.Context(), m.usuarioID(peticion.Context()), registroID,
		peticion.PostFormValue("estado"), strings.TrimSpace(peticion.PostFormValue("observacion")),
		peticion.PostFormValue("motivo"))
	if err != nil {
		m.renderizarConsulta(respuesta, peticion, http.StatusUnprocessableEntity, fecha, seccionID, "",
			mensajeErrorConsulta(err))
		return
	}
	http.Redirect(respuesta, peticion, "/asistencia/consulta?fecha="+fechas.FormatearISO(fecha)+
		"&seccion_id="+strconv.FormatInt(seccionID, 10)+"&ok=corregido", http.StatusSeeOther)
}

func (m *Manejador) renderizarConsulta(respuesta http.ResponseWriter, peticion *http.Request, estado int, fecha time.Time, seccionID int64, mensaje, mensajeError string) {
	ctx := peticion.Context()
	secciones, err := m.servicio.SeccionesDisponibles(ctx)
	if err != nil {
		m.registro.Error("no se pudieron listar las secciones", "error", err)
		http.Error(respuesta, "No se pudo cargar la consulta", http.StatusInternalServerError)
		return
	}
	var registros []ConsultaRegistro
	if seccionID > 0 {
		registros, err = m.servicio.ConsultaRegistros(ctx, fecha, seccionID)
		if err != nil {
			m.registro.Error("no se pudieron cargar los registros", "error", err)
			http.Error(respuesta, "No se pudo cargar la consulta", http.StatusInternalServerError)
			return
		}
	}

	m.renderizar(respuesta, peticion, estado, vista.Consulta(
		m.nombreUsuario(ctx),
		intermediarios.RolUsuario(ctx),
		m.tokenCSRF(ctx),
		fechas.FormatearISO(fecha),
		aSeccionesVista(secciones),
		seccionID,
		aConsultaVista(registros),
		mensaje,
		mensajeError,
	))
}

func (m *Manejador) verParametros(respuesta http.ResponseWriter, peticion *http.Request) {
	m.renderizarParametros(respuesta, peticion, http.StatusOK,
		mensajeCorrectoParametros(peticion.URL.Query().Get("ok")), "")
}

func (m *Manejador) configurarParametros(respuesta http.ResponseWriter, peticion *http.Request) {
	horas, err := strconv.Atoi(strings.TrimSpace(peticion.PostFormValue("plazo_edicion_horas")))
	if err != nil {
		m.renderizarParametros(respuesta, peticion, http.StatusUnprocessableEntity, "",
			mensajeErrorConsulta(ErrPlazoFueraDeRango))
		return
	}
	if err := m.servicio.ConfigurarPlazo(peticion.Context(), m.usuarioID(peticion.Context()), horas); err != nil {
		m.renderizarParametros(respuesta, peticion, http.StatusUnprocessableEntity, "",
			mensajeErrorConsulta(err))
		return
	}
	http.Redirect(respuesta, peticion, "/configuracion/asistencia?ok=plazo", http.StatusSeeOther)
}

func (m *Manejador) renderizarParametros(respuesta http.ResponseWriter, peticion *http.Request, estado int, mensaje, mensajeError string) {
	parametros, err := m.servicio.PlazoEdicion(peticion.Context())
	if err != nil {
		m.registro.Error("no se pudo leer el plazo de edición", "error", err)
		http.Error(respuesta, "No se pudo cargar la configuración", http.StatusInternalServerError)
		return
	}
	m.renderizar(respuesta, peticion, estado, vista.Parametros(
		m.nombreUsuario(peticion.Context()),
		intermediarios.RolUsuario(peticion.Context()),
		m.tokenCSRF(peticion.Context()),
		vista.ParametrosVista{
			Plazo:       parametros.PlazoEdicionHoras,
			Actualizado: fechas.FormatearDia(parametros.ActualizadoEn),
		},
		mensaje,
		mensajeError,
	))
}

func aConsultaVista(registros []ConsultaRegistro) []vista.ConsultaFilaVista {
	filas := make([]vista.ConsultaFilaVista, 0, len(registros))
	for _, registro := range registros {
		textoQR := "Sin ingreso por QR"
		if registro.Marcacion != nil {
			textoQR = "QR a las " + fechas.Hora(registro.Marcacion.Hora)
		}
		filas = append(filas, vista.ConsultaFilaVista{
			ID:          registro.Registro.ID,
			Alumno:      registro.Alumno,
			Curso:       registro.Curso,
			Seccion:     registro.Seccion,
			Bloque:      registro.Registro.BloqueNumero,
			Estado:      registro.Registro.Estado,
			Observacion: registro.Registro.Observacion,
			QRTexto:     textoQR,
			Docente:     registro.Docente,
			Diferencia:  registro.Diferencia,
		})
	}
	return filas
}

func mensajeCorrectoConsulta(codigo string) string {
	if codigo == "corregido" {
		return "El registro se corrigió correctamente."
	}
	return ""
}

func mensajeCorrectoParametros(codigo string) string {
	if codigo == "plazo" {
		return "El plazo de edición se guardó correctamente."
	}
	return ""
}

func mensajeErrorConsulta(err error) string {
	if errors.Is(err, ErrEstadoInvalido) {
		return "El estado de asistencia no es válido."
	}
	if errors.Is(err, ErrObservacionLarga) {
		return "La observación no puede superar los 300 caracteres."
	}
	if errors.Is(err, ErrMotivoRequerido) {
		return "Indique el motivo de la corrección."
	}
	if errors.Is(err, ErrRegistroNoEncontrado) {
		return "El registro no existe."
	}
	if errors.Is(err, ErrPlazoFueraDeRango) {
		return "El plazo de edición debe estar entre 1 y 168 horas."
	}
	return "No se pudo completar la operación. Intente de nuevo."
}
