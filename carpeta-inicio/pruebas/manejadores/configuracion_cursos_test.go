package manejadores

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"asistencia/internal/academico"
	"asistencia/internal/plataforma/fechas"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func grillaDePrueba(t *testing.T) academico.GrillaHorario {
	t.Helper()
	return academico.GrillaHorario{
		Seccion: academico.Seccion{
			ID:            15,
			Nombre:        "A",
			Grado:         "3.°",
			AnioEscolar:   "2026",
			AnioEscolarID: 1,
			GradoID:       3,
			Turno:         academico.TurnoManana,
		},
		Turno: academico.TurnoManana,
		Bloques: []academico.BloqueClase{{
			ID:           4,
			Turno:        academico.TurnoManana,
			Numero:       1,
			HoraInicio:   mustHora(t, "07:45"),
			HoraFin:      mustHora(t, "08:30"),
			VigenteDesde: time.Date(2026, time.November, 6, 0, 0, 0, 0, time.UTC),
		}},
		Cursos: []academico.Curso{{
			ID:            2,
			AnioEscolarID: 1,
			Nombre:        "Matemática",
			Activo:        true,
			Grados:        []int64{3},
		}},
	}
}

func mustHora(t *testing.T, texto string) time.Time {
	t.Helper()
	valor, err := fechas.ParsearHora(texto)
	require.NoError(t, err)
	return valor
}

func TestConfiguracionCursos_MuestraElListado(t *testing.T) {
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen: academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
		grados:  []academico.Grado{{ID: 3, Nombre: "3.°"}},
		cursos: []academico.Curso{{
			ID:            2,
			AnioEscolarID: 1,
			Nombre:        "Matemática",
			Activo:        true,
			Grados:        []int64{3},
		}},
	})
	peticion := peticionConSesion(t, sesiones, http.MethodGet, "/configuracion/cursos", nil, "", "director")

	respuesta := ejecutarConfiguracion(manejador, peticion)

	assert.Equal(t, http.StatusOK, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "Matemática")
	assert.Contains(t, respuesta.Body.String(), "Activo")
}

func TestConfiguracionCursos_AdministradorCreaYDesactiva(t *testing.T) {
	var creado academico.DatosCurso
	var desactivado int64
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen: academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
		grados:  []academico.Grado{{ID: 3, Nombre: "3.°"}},
		crearCurso: func(datos academico.DatosCurso) (int64, error) {
			creado = datos
			return 9, nil
		},
		desactivarCurso: func(cursoID int64) error {
			desactivado = cursoID
			return nil
		},
	})

	peticion := peticionFormulario(t, sesiones, http.MethodPost, "/configuracion/cursos", url.Values{
		"nombre": {"Matemática"},
		"grados": {"3"},
	}, "administrador")
	respuesta := ejecutarConfiguracion(manejador, peticion)

	require.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Equal(t, "/configuracion/cursos?ok=curso", respuesta.Header().Get("Location"))
	assert.Equal(t, "Matemática", creado.Nombre)
	assert.Equal(t, []int64{3}, creado.Grados)

	peticion = peticionFormulario(t, sesiones, http.MethodPost,
		"/configuracion/cursos/9/desactivacion", url.Values{}, "administrador")
	respuesta = ejecutarConfiguracion(manejador, peticion)

	require.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Equal(t, int64(9), desactivado)
}

func TestConfiguracionCursos_DirectorNoPuedeEscribir(t *testing.T) {
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen: academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
	})

	for _, ruta := range []string{"/configuracion/cursos", "/configuracion/cursos/9/desactivacion"} {
		peticion := peticionFormulario(t, sesiones, http.MethodPost, ruta,
			url.Values{"nombre": {"Matemática"}, "grados": {"3"}}, "director")
		respuesta := ejecutarConfiguracion(manejador, peticion)
		assert.Equal(t, http.StatusForbidden, respuesta.Code, ruta)
	}
}

func TestConfiguracionHorario_MuestraLaGrilla(t *testing.T) {
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen: academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
		secciones: []academico.Seccion{{
			ID: 15, Nombre: "A", Grado: "3.°", AnioEscolar: "2026", AnioEscolarID: 1,
		}},
		grilla: grillaDePrueba(t),
	})
	peticion := peticionConSesion(t, sesiones, http.MethodGet,
		"/configuracion/horarios?seccion_id=15&fecha=2026-11-06", nil, "", "administrador")

	respuesta := ejecutarConfiguracion(manejador, peticion)

	assert.Equal(t, http.StatusOK, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "Horario semanal")
	assert.Contains(t, respuesta.Body.String(), "Matemática")
}

func TestConfiguracionHorario_GuardaCeldasYConfirmaReemplazo(t *testing.T) {
	var recibidos []academico.CambioCelda
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen: academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
		secciones: []academico.Seccion{{
			ID: 15, Nombre: "A", Grado: "3.°", AnioEscolar: "2026", AnioEscolarID: 1,
		}},
		grilla: grillaDePrueba(t),
		guardarCeldas: func(_ int64, _ time.Time, cambios []academico.CambioCelda, confirmar bool) error {
			recibidos = cambios
			if !confirmar {
				return academico.ErrRequiereConfirmacion
			}
			return nil
		},
	})

	peticion := peticionFormulario(t, sesiones, http.MethodPost, "/configuracion/horarios", url.Values{
		"seccion_id": {"15"},
		"fecha":      {"2026-11-06"},
		"celda_1_1":  {"2"},
	}, "administrador")
	respuesta := ejecutarConfiguracion(manejador, peticion)

	require.Equal(t, http.StatusOK, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "Confirmar reemplazo")
	require.Len(t, recibidos, 1)
	assert.Equal(t, 1, recibidos[0].DiaSemana)
	assert.Equal(t, 1, recibidos[0].BloqueNumero)
	require.NotNil(t, recibidos[0].CursoID)
	assert.Equal(t, int64(2), *recibidos[0].CursoID)

	peticion = peticionFormulario(t, sesiones, http.MethodPost, "/configuracion/horarios", url.Values{
		"seccion_id": {"15"},
		"fecha":      {"2026-11-06"},
		"celda_1_1":  {"2"},
		"confirmar":  {"si"},
	}, "administrador")
	respuesta = ejecutarConfiguracion(manejador, peticion)

	assert.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Contains(t, respuesta.Header().Get("Location"), "ok=horario")
}

func TestConfiguracionHorario_DirectorNoPuedeGuardar(t *testing.T) {
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen: academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
	})
	peticion := peticionFormulario(t, sesiones, http.MethodPost, "/configuracion/horarios", url.Values{
		"seccion_id": {"15"},
		"fecha":      {"2026-11-06"},
		"celda_1_1":  {"2"},
	}, "director")

	respuesta := ejecutarConfiguracion(manejador, peticion)

	assert.Equal(t, http.StatusForbidden, respuesta.Code)
}
