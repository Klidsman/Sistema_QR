package academico

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"asistencia/internal/plataforma/fechas"
	"asistencia/internal/plataforma/intermediarios"
	vista "asistencia/web/plantillas/configuracion"
)

func (m *Manejador) horarios(respuesta http.ResponseWriter, peticion *http.Request) {
	seccionID, _ := strconv.ParseInt(peticion.URL.Query().Get("seccion_id"), 10, 64)
	m.renderizarHorario(respuesta, peticion, http.StatusOK, seccionID, fechaDeConsulta(peticion),
		"", peticion.URL.Query().Get("ok"), nil, false)
}

func (m *Manejador) guardarHorario(respuesta http.ResponseWriter, peticion *http.Request) {
	ctx := peticion.Context()
	seccionID, err := strconv.ParseInt(strings.TrimSpace(peticion.PostFormValue("seccion_id")), 10, 64)
	if err != nil || seccionID <= 0 {
		m.renderizarHorario(respuesta, peticion, http.StatusUnprocessableEntity, 0, fechas.Hoy(),
			mensajeDeError(ErrSeccionNoEncontrada), "", nil, false)
		return
	}
	fecha, err := fechas.ParsearDia(peticion.PostFormValue("fecha"))
	if err != nil {
		m.renderizarHorario(respuesta, peticion, http.StatusUnprocessableEntity, seccionID, fechas.Hoy(),
			mensajeDeError(err), "", nil, false)
		return
	}
	confirmar := peticion.PostFormValue("confirmar") == "si"

	grilla, err := m.configuracion.ObtenerGrilla(ctx, seccionID, fecha)
	if err != nil {
		m.renderizarHorario(respuesta, peticion, http.StatusUnprocessableEntity, seccionID, fecha,
			mensajeDeError(err), "", nil, false)
		return
	}

	seleccion := seleccionDeFormulario(peticion, grilla)
	cambios, err := cambiosDeHorario(grilla, seleccion)
	if err != nil {
		m.renderizarHorario(respuesta, peticion, http.StatusUnprocessableEntity, seccionID, fecha,
			mensajeDeError(err), "", seleccion, confirmar)
		return
	}

	if err := m.configuracion.GuardarCeldas(ctx, m.usuarioID(ctx), seccionID, fecha, cambios, confirmar); err != nil {
		estado := http.StatusUnprocessableEntity
		if err == ErrRequiereConfirmacion {
			estado = http.StatusOK
			confirmar = true
		}
		m.renderizarHorario(respuesta, peticion, estado, seccionID, fecha,
			mensajeDeError(err), "", seleccion, confirmar)
		return
	}

	http.Redirect(respuesta, peticion, fmt.Sprintf(
		"/configuracion/horarios?seccion_id=%d&fecha=%s&ok=horario",
		seccionID, fechas.FormatearISO(fecha)), http.StatusSeeOther)
}

func (m *Manejador) renderizarHorario(respuesta http.ResponseWriter, peticion *http.Request, estado int, seccionID int64, fecha time.Time, mensajeError, codigoCorrecto string, seleccion map[string]string, confirmar bool) {
	ctx := peticion.Context()
	activo, err := m.configuracion.AnioActivo(ctx)
	if err != nil {
		m.registro.Error("no se pudo buscar el año activo", "error", err)
		http.Error(respuesta, "No se pudo cargar la configuración", http.StatusInternalServerError)
		return
	}

	var secciones []Seccion
	var grillaVista *vista.GrillaVista
	if activo != nil {
		secciones, err = m.configuracion.ListarSecciones(ctx, activo.ID)
		if err != nil {
			m.registro.Error("no se pudieron listar las secciones", "error", err)
			http.Error(respuesta, "No se pudo cargar la configuración", http.StatusInternalServerError)
			return
		}
		if seccionID > 0 {
			grilla, err := m.configuracion.ObtenerGrilla(ctx, seccionID, fecha)
			if err == nil {
				if seleccion == nil {
					seleccion = seleccionDeGrilla(grilla)
				}
				grillaVista = aGrillaVista(grilla, seleccion)
			} else if err != ErrSeccionSinTurno && err != ErrSeccionNoEncontrada {
				m.registro.Error("no se pudo cargar el horario de la sección", "error", err)
				http.Error(respuesta, "No se pudo cargar la configuración", http.StatusInternalServerError)
				return
			}
		}
	}

	m.renderizar(respuesta, peticion, estado, vista.Horarios(
		m.nombreUsuario(ctx),
		intermediarios.RolUsuario(ctx),
		m.tokenCSRF(ctx),
		aAnioVista(activo),
		aSeccionesVista(secciones),
		seccionID,
		fechas.FormatearISO(fecha),
		grillaVista,
		confirmar,
		mensajeCorrectoHorarioSemanal(codigoCorrecto),
		mensajeError,
	))
}

func fechaDeConsulta(peticion *http.Request) time.Time {
	fecha, err := fechas.ParsearDia(peticion.URL.Query().Get("fecha"))
	if err != nil {
		return fechas.Hoy()
	}
	return fecha
}

func seleccionDeGrilla(grilla GrillaHorario) map[string]string {
	seleccion := make(map[string]string, len(grilla.Celdas))
	for _, celda := range grilla.Celdas {
		valor := ""
		if celda.CursoID != nil {
			valor = strconv.FormatInt(*celda.CursoID, 10)
		}
		seleccion[claveCelda(celda.DiaSemana, celda.BloqueNumero)] = valor
	}
	return seleccion
}

func seleccionDeFormulario(peticion *http.Request, grilla GrillaHorario) map[string]string {
	seleccion := make(map[string]string, len(grilla.Bloques)*5)
	for _, bloque := range grilla.Bloques {
		for dia := 1; dia <= 5; dia++ {
			seleccion[claveCelda(dia, bloque.Numero)] = strings.TrimSpace(
				peticion.PostFormValue(claveCampoCelda(dia, bloque.Numero)))
		}
	}
	return seleccion
}

func cambiosDeHorario(grilla GrillaHorario, seleccion map[string]string) ([]CambioCelda, error) {
	cambios := make([]CambioCelda, 0)
	for _, bloque := range grilla.Bloques {
		for dia := 1; dia <= 5; dia++ {
			clave := claveCelda(dia, bloque.Numero)
			nuevo := seleccion[clave]
			actual, err := cursoIDDeTexto(nuevo)
			if err != nil {
				return nil, err
			}
			if seleccionDeGrillaTiene(grilla, dia, bloque.Numero, nuevo) {
				continue
			}
			cambios = append(cambios, CambioCelda{
				DiaSemana:    dia,
				BloqueNumero: bloque.Numero,
				CursoID:      actual,
			})
		}
	}
	return cambios, nil
}

func seleccionDeGrillaTiene(grilla GrillaHorario, dia, bloque int, valor string) bool {
	for _, celda := range grilla.Celdas {
		if celda.DiaSemana != dia || celda.BloqueNumero != bloque {
			continue
		}
		actual := ""
		if celda.CursoID != nil {
			actual = strconv.FormatInt(*celda.CursoID, 10)
		}
		return actual == valor
	}
	return valor == ""
}

func cursoIDDeTexto(texto string) (*int64, error) {
	if texto == "" {
		return nil, nil
	}
	cursoID, err := strconv.ParseInt(texto, 10, 64)
	if err != nil || cursoID <= 0 {
		return nil, ErrCursoNoEncontrado
	}
	return &cursoID, nil
}

func claveCelda(dia, bloque int) string {
	return fmt.Sprintf("%d-%d", dia, bloque)
}

func claveCampoCelda(dia, bloque int) string {
	return fmt.Sprintf("celda_%d_%d", dia, bloque)
}

func mensajeCorrectoHorarioSemanal(codigo string) string {
	if codigo == "horario" {
		return "El horario semanal se guardó correctamente."
	}
	return ""
}

func aGrillaVista(grilla GrillaHorario, seleccion map[string]string) *vista.GrillaVista {
	return &vista.GrillaVista{
		SeccionID: grilla.Seccion.ID,
		Seccion:   etiquetaSeccion(grilla.Seccion),
		Turno:     grilla.Turno,
		Bloques:   aBloquesVista(grilla.Bloques),
		Cursos:    aCursoOpciones(grilla.Cursos),
		Seleccion: seleccion,
	}
}
