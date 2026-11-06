package academico

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"asistencia/internal/plataforma/intermediarios"
	vista "asistencia/web/plantillas/configuracion"

	"github.com/go-chi/chi/v5"
)

func (m *Manejador) cursos(respuesta http.ResponseWriter, peticion *http.Request) {
	m.renderizarCursos(respuesta, peticion, http.StatusOK, "", peticion.URL.Query().Get("ok"))
}

func (m *Manejador) crearCurso(respuesta http.ResponseWriter, peticion *http.Request) {
	ctx := peticion.Context()
	activo, err := m.configuracion.AnioActivo(ctx)
	if err != nil || activo == nil {
		m.renderizarCursos(respuesta, peticion, http.StatusUnprocessableEntity,
			mensajeDeError(ErrAnioNoEncontrado), "")
		return
	}
	datos, err := datosCursoDeFormulario(activo.ID, peticion)
	if err != nil {
		m.renderizarCursos(respuesta, peticion, http.StatusUnprocessableEntity, mensajeDeError(err), "")
		return
	}
	if _, err := m.configuracion.CrearCurso(ctx, m.usuarioID(ctx), datos); err != nil {
		m.renderizarCursos(respuesta, peticion, http.StatusUnprocessableEntity, mensajeDeError(err), "")
		return
	}
	http.Redirect(respuesta, peticion, "/configuracion/cursos?ok=curso", http.StatusSeeOther)
}

func (m *Manejador) desactivarCurso(respuesta http.ResponseWriter, peticion *http.Request) {
	cursoID, err := strconv.ParseInt(chi.URLParam(peticion, "id"), 10, 64)
	if err != nil || cursoID <= 0 {
		m.renderizarCursos(respuesta, peticion, http.StatusUnprocessableEntity,
			mensajeDeError(ErrCursoNoEncontrado), "")
		return
	}
	if err := m.configuracion.DesactivarCurso(peticion.Context(), m.usuarioID(peticion.Context()), cursoID); err != nil {
		m.renderizarCursos(respuesta, peticion, http.StatusUnprocessableEntity, mensajeDeError(err), "")
		return
	}
	http.Redirect(respuesta, peticion, "/configuracion/cursos?ok=desactivado", http.StatusSeeOther)
}

func (m *Manejador) renderizarCursos(respuesta http.ResponseWriter, peticion *http.Request, estado int, mensajeError, codigoCorrecto string) {
	ctx := peticion.Context()
	activo, err := m.configuracion.AnioActivo(ctx)
	if err != nil {
		m.registro.Error("no se pudo buscar el año activo", "error", err)
		http.Error(respuesta, "No se pudo cargar la configuración", http.StatusInternalServerError)
		return
	}
	var cursos []Curso
	var grados []Grado
	if activo != nil {
		cursos, err = m.configuracion.ListarCursos(ctx, activo.ID)
		if err != nil {
			m.registro.Error("no se pudieron listar los cursos", "error", err)
			http.Error(respuesta, "No se pudo cargar la configuración", http.StatusInternalServerError)
			return
		}
		grados, err = m.configuracion.ListarGrados(ctx, activo.ID)
		if err != nil {
			m.registro.Error("no se pudieron listar los grados", "error", err)
			http.Error(respuesta, "No se pudo cargar la configuración", http.StatusInternalServerError)
			return
		}
	}

	m.renderizar(respuesta, peticion, estado, vista.Cursos(
		m.nombreUsuario(ctx),
		intermediarios.RolUsuario(ctx),
		m.tokenCSRF(ctx),
		aAnioVista(activo),
		aGradosOpciones(grados),
		aCursosVista(cursos, grados),
		mensajeCorrectoCurso(codigoCorrecto),
		mensajeError,
	))
}

func datosCursoDeFormulario(anioID int64, peticion *http.Request) (DatosCurso, error) {
	if err := peticion.ParseForm(); err != nil {
		return DatosCurso{}, err
	}
	grados := make([]int64, 0)
	for _, texto := range peticion.PostForm["grados"] {
		gradoID, err := strconv.ParseInt(strings.TrimSpace(texto), 10, 64)
		if err != nil || gradoID <= 0 {
			return DatosCurso{}, ErrGradoNoEncontrado
		}
		grados = append(grados, gradoID)
	}
	return DatosCurso{
		AnioEscolarID: anioID,
		Nombre:        peticion.PostFormValue("nombre"),
		Grados:        grados,
	}, nil
}

func mensajeCorrectoCurso(codigo string) string {
	switch codigo {
	case "curso":
		return "El curso se creó correctamente."
	case "desactivado":
		return "El curso se desactivó correctamente."
	default:
		return ""
	}
}

func aCursosVista(cursos []Curso, grados []Grado) []vista.CursoVista {
	nombres := make(map[int64]string, len(grados))
	for _, grado := range grados {
		nombres[grado.ID] = grado.Nombre
	}
	resultado := make([]vista.CursoVista, 0, len(cursos))
	for _, curso := range cursos {
		etiquetas := make([]string, 0, len(curso.Grados))
		for _, gradoID := range curso.Grados {
			if nombre, existe := nombres[gradoID]; existe {
				etiquetas = append(etiquetas, nombre)
			}
		}
		resultado = append(resultado, vista.CursoVista{
			ID:     curso.ID,
			Nombre: curso.Nombre,
			Grados: strings.Join(etiquetas, ", "),
			Activo: curso.Activo,
		})
	}
	return resultado
}

func aGradosOpciones(grados []Grado) []vista.GradoOpcion {
	resultado := make([]vista.GradoOpcion, 0, len(grados))
	for _, grado := range grados {
		resultado = append(resultado, vista.GradoOpcion{ID: grado.ID, Nombre: grado.Nombre})
	}
	return resultado
}

func aCursoOpciones(cursos []Curso) []vista.CursoOpcion {
	resultado := make([]vista.CursoOpcion, 0)
	for _, curso := range cursos {
		if curso.Activo {
			resultado = append(resultado, vista.CursoOpcion{ID: curso.ID, Nombre: curso.Nombre})
		}
	}
	return resultado
}

func etiquetaSeccion(seccion Seccion) string {
	return fmt.Sprintf("%s %s (%s)", seccion.Grado, seccion.Nombre, seccion.AnioEscolar)
}
