package alumnos

import (
	"net/http"

	"asistencia/internal/academico"
	"asistencia/internal/plataforma/intermediarios"
	vista "asistencia/web/plantillas/carnets"
)

// seleccionCarnets muestra el formulario para elegir un grado o una sección.
func (m *Manejador) seleccionCarnets(respuesta http.ResponseWriter, peticion *http.Request) {
	secciones, err := m.secciones.ListarSecciones(peticion.Context())
	if err != nil {
		m.errorInterno(respuesta, "listando secciones para carnés", err)
		return
	}

	m.renderizar(respuesta, peticion, http.StatusOK, vista.Seleccion(
		m.nombreUsuario(peticion.Context()),
		intermediarios.RolUsuario(peticion.Context()),
		m.tokenCSRF(peticion.Context()),
		secciones,
		gradosDeSecciones(secciones),
		mensajeCarnetsError(peticion.URL.Query().Get("error")),
	))
}

// loteCarnets genera la vista imprimible de los alumnos activos de la selección.
func (m *Manejador) loteCarnets(respuesta http.ResponseWriter, peticion *http.Request) {
	seleccion := SeleccionCarnets{
		SeccionID: enteroDeTexto(peticion.URL.Query().Get("seccion_id")),
		GradoID:   enteroDeTexto(peticion.URL.Query().Get("grado_id")),
	}
	if !seleccion.Valida() {
		http.Redirect(respuesta, peticion, "/carnets?error=seleccion", http.StatusSeeOther)
		return
	}

	secciones, err := m.secciones.ListarSecciones(peticion.Context())
	if err != nil {
		m.errorInterno(respuesta, "listando secciones para carnés", err)
		return
	}

	titulo := ""
	if seleccion.SeccionID > 0 {
		seleccion.GradoID = 0
		seccion, err := m.secciones.BuscarSeccionPorID(peticion.Context(), seleccion.SeccionID)
		if err != nil {
			http.Redirect(respuesta, peticion, "/carnets?error=seccion_invalida", http.StatusSeeOther)
			return
		}
		titulo = seccion.Etiqueta()
	} else {
		grado, encontrado := buscarGrado(gradosDeSecciones(secciones), seleccion.GradoID)
		if !encontrado {
			http.Redirect(respuesta, peticion, "/carnets?error=grado_invalido", http.StatusSeeOther)
			return
		}
		titulo = grado.Etiqueta()
	}

	lote, err := m.alumnos.ListarCarnets(peticion.Context(), seleccion)
	if err != nil {
		m.errorInterno(respuesta, "generando los carnés", err)
		return
	}

	m.renderizar(respuesta, peticion, http.StatusOK, vista.Lote(
		m.nombreUsuario(peticion.Context()),
		intermediarios.RolUsuario(peticion.Context()),
		m.tokenCSRF(peticion.Context()),
		titulo,
		aLoteVista(lote),
		m.carnetAnchoMM,
		m.carnetAltoMM,
	))
}

func gradosDeSecciones(secciones []academico.Seccion) []vista.Grado {
	grados := make([]vista.Grado, 0, len(secciones))
	vistos := make(map[int64]bool)
	for _, seccion := range secciones {
		if vistos[seccion.GradoID] {
			continue
		}
		vistos[seccion.GradoID] = true
		grados = append(grados, vista.Grado{
			ID:          seccion.GradoID,
			Nombre:      seccion.Grado,
			AnioEscolar: seccion.AnioEscolar,
		})
	}
	return grados
}

func buscarGrado(grados []vista.Grado, id int64) (vista.Grado, bool) {
	for _, grado := range grados {
		if grado.ID == id {
			return grado, true
		}
	}
	return vista.Grado{}, false
}

func aLoteVista(lote LoteCarnets) vista.ContenidoLote {
	return vista.ContenidoLote{
		Carnets: aCarnetsVista(lote.Carnets),
		Avisos: vista.Avisos{
			SinFoto: aCarnetsVista(lote.Avisos.SinFoto),
			SinQR:   aCarnetsVista(lote.Avisos.SinQR),
		},
	}
}

func aCarnetsVista(carnets []AlumnoCarnet) []vista.Alumno {
	filas := make([]vista.Alumno, 0, len(carnets))
	for _, carnet := range carnets {
		filas = append(filas, vista.Alumno{
			ID:            carnet.ID,
			Identificador: carnet.Identificador,
			Nombres:       carnet.Nombres,
			Apellidos:     carnet.Apellidos,
			Grado:         carnet.Grado,
			Seccion:       carnet.Seccion,
			AnioEscolar:   carnet.AnioEscolar,
			TieneFoto:     carnet.TieneFoto(),
			TokenQR:       carnet.TokenQR,
		})
	}
	return filas
}

func mensajeCarnetsError(codigo string) string {
	switch codigo {
	case "seleccion":
		return "Elija un grado o una sección para generar los carnés."
	case "seccion_invalida":
		return "La sección elegida no existe."
	case "grado_invalido":
		return "El grado elegido no existe."
	default:
		return ""
	}
}
