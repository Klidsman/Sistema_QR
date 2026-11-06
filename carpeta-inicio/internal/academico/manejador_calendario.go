package academico

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"asistencia/internal/plataforma/fechas"
	"asistencia/internal/plataforma/intermediarios"
	vista "asistencia/web/plantillas/configuracion"
)

func (m *Manejador) calendario(respuesta http.ResponseWriter, peticion *http.Request) {
	formulario := vista.FormularioExcepcion{}
	m.renderizarCalendario(respuesta, peticion, http.StatusOK, "", peticion.URL.Query().Get("ok"), formulario, false)
}

func (m *Manejador) registrarExcepcion(respuesta http.ResponseWriter, peticion *http.Request) {
	ctx := peticion.Context()
	activo, err := m.configuracion.AnioActivo(ctx)
	if err != nil || activo == nil {
		m.renderizarCalendario(respuesta, peticion, http.StatusUnprocessableEntity,
			mensajeDeError(ErrAnioNoEncontrado), "", vista.FormularioExcepcion{}, false)
		return
	}

	datos, formulario, err := datosExcepcionDeFormulario(activo.ID, peticion)
	if err != nil {
		m.renderizarCalendario(respuesta, peticion, http.StatusUnprocessableEntity,
			mensajeDeError(err), "", formulario, false)
		return
	}
	confirmar := peticion.PostFormValue("confirmar") == "si"

	if _, err := m.configuracion.RegistrarExcepcion(ctx, m.usuarioID(ctx), datos, confirmar); err != nil {
		estado := http.StatusUnprocessableEntity
		nuevoConfirmar := false
		if errors.Is(err, ErrExcepcionSolapada) {
			estado = http.StatusOK
			nuevoConfirmar = true
		}
		m.renderizarCalendario(respuesta, peticion, estado, mensajeDeError(err), "", formulario, nuevoConfirmar)
		return
	}
	http.Redirect(respuesta, peticion, "/configuracion/calendario?ok=excepcion", http.StatusSeeOther)
}

func (m *Manejador) renderizarCalendario(respuesta http.ResponseWriter, peticion *http.Request, estado int, mensajeError, codigoCorrecto string, formulario vista.FormularioExcepcion, confirmar bool) {
	ctx := peticion.Context()
	activo, err := m.configuracion.AnioActivo(ctx)
	if err != nil {
		m.registro.Error("no se pudo buscar el año activo", "error", err)
		http.Error(respuesta, "No se pudo cargar la configuración", http.StatusInternalServerError)
		return
	}

	var grados []Grado
	var excepciones []ExcepcionCalendario
	if activo != nil {
		grados, err = m.configuracion.ListarGrados(ctx, activo.ID)
		if err != nil {
			m.registro.Error("no se pudieron listar los grados", "error", err)
			http.Error(respuesta, "No se pudo cargar la configuración", http.StatusInternalServerError)
			return
		}
		excepciones, err = m.configuracion.ListarExcepciones(ctx, activo.ID)
		if err != nil {
			m.registro.Error("no se pudieron listar las excepciones", "error", err)
			http.Error(respuesta, "No se pudo cargar la configuración", http.StatusInternalServerError)
			return
		}
		excepciones = filtrarExcepciones(excepciones, peticion)
	}

	m.renderizar(respuesta, peticion, estado, vista.Calendario(
		m.nombreUsuario(ctx),
		intermediarios.RolUsuario(ctx),
		m.tokenCSRF(ctx),
		aAnioVista(activo),
		aGradosOpciones(grados),
		aExcepcionesVista(excepciones, grados),
		formulario,
		confirmar,
		mensajeCorrectoExcepcion(codigoCorrecto),
		mensajeError,
	))
}

func datosExcepcionDeFormulario(anioID int64, peticion *http.Request) (DatosExcepcion, vista.FormularioExcepcion, error) {
	if err := peticion.ParseForm(); err != nil {
		return DatosExcepcion{}, vista.FormularioExcepcion{}, err
	}
	formulario := vista.FormularioExcepcion{
		Tipo:        peticion.PostFormValue("tipo"),
		FechaInicio: peticion.PostFormValue("fecha_inicio"),
		FechaFin:    peticion.PostFormValue("fecha_fin"),
		Motivo:      peticion.PostFormValue("motivo"),
	}
	inicio, err := fechas.ParsearDia(formulario.FechaInicio)
	if err != nil {
		return DatosExcepcion{}, formulario, err
	}
	fin, err := fechas.ParsearDia(formulario.FechaFin)
	if err != nil {
		return DatosExcepcion{}, formulario, err
	}
	grados := make([]int64, 0)
	for _, texto := range peticion.PostForm["grados"] {
		gradoID, err := strconv.ParseInt(strings.TrimSpace(texto), 10, 64)
		if err != nil || gradoID <= 0 {
			return DatosExcepcion{}, formulario, ErrGradoNoEncontrado
		}
		grados = append(grados, gradoID)
	}
	formulario.Grados = grados
	return DatosExcepcion{
		AnioEscolarID: anioID,
		Tipo:          formulario.Tipo,
		FechaInicio:   inicio,
		FechaFin:      fin,
		Motivo:        formulario.Motivo,
		Grados:        grados,
	}, formulario, nil
}

func filtrarExcepciones(excepciones []ExcepcionCalendario, peticion *http.Request) []ExcepcionCalendario {
	mes := strings.TrimSpace(peticion.URL.Query().Get("mes"))
	gradoID, _ := strconv.ParseInt(peticion.URL.Query().Get("grado_id"), 10, 64)

	var inicioMes, finMes time.Time
	if mes != "" {
		if fecha, err := fechas.ParsearDia(mes + "-01"); err == nil {
			inicioMes = fecha
			finMes = fecha.AddDate(0, 1, -1)
		}
	}

	filtradas := make([]ExcepcionCalendario, 0, len(excepciones))
	for _, excepcion := range excepciones {
		if !inicioMes.IsZero() {
			if fechas.CompararDia(excepcion.FechaInicio, finMes) > 0 ||
				fechas.CompararDia(excepcion.FechaFin, inicioMes) < 0 {
				continue
			}
		}
		if gradoID > 0 && !alcanzaGrado(excepcion, gradoID) {
			continue
		}
		filtradas = append(filtradas, excepcion)
	}
	return filtradas
}

func alcanzaGrado(excepcion ExcepcionCalendario, gradoID int64) bool {
	if len(excepcion.Grados) == 0 {
		return true
	}
	for _, alcanzado := range excepcion.Grados {
		if alcanzado == gradoID {
			return true
		}
	}
	return false
}

func mensajeCorrectoExcepcion(codigo string) string {
	if codigo == "excepcion" {
		return "La excepción de calendario se registró correctamente."
	}
	return ""
}

func aExcepcionesVista(excepciones []ExcepcionCalendario, grados []Grado) []vista.ExcepcionVista {
	nombres := make(map[int64]string, len(grados))
	for _, grado := range grados {
		nombres[grado.ID] = grado.Nombre
	}
	resultado := make([]vista.ExcepcionVista, 0, len(excepciones))
	for _, excepcion := range excepciones {
		alcance := "Todo el colegio"
		if len(excepcion.Grados) > 0 {
			etiquetas := make([]string, 0, len(excepcion.Grados))
			for _, gradoID := range excepcion.Grados {
				if nombre, existe := nombres[gradoID]; existe {
					etiquetas = append(etiquetas, nombre)
				}
			}
			alcance = strings.Join(etiquetas, ", ")
		}
		resultado = append(resultado, vista.ExcepcionVista{
			ID:     excepcion.ID,
			Tipo:   etiquetaTipoExcepcion(excepcion.Tipo),
			Inicio: fechas.FormatearDia(excepcion.FechaInicio),
			Fin:    fechas.FormatearDia(excepcion.FechaFin),
			Motivo: excepcion.Motivo,
			Grados: alcance,
		})
	}
	return resultado
}

func etiquetaTipoExcepcion(tipo string) string {
	switch tipo {
	case TipoFeriado:
		return "Feriado"
	case TipoNoLaborable:
		return "No laborable"
	case TipoSuspension:
		return "Suspensión"
	default:
		return tipo
	}
}
