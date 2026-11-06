package manejadores

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"asistencia/internal/reportes"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReportes_PermisosYExportaciones(t *testing.T) {
	servicio := &servicioReportesFalso{}
	manejador, sesiones := nuevoManejadorReportes(servicio)

	for _, rol := range []string{"administrador", "director", "docente"} {
		peticion := peticionConSesion(t, sesiones, http.MethodGet, "/reportes/diario", nil, "", rol)
		respuesta := ejecutarReportes(manejador, peticion)
		assert.Equal(t, http.StatusOK, respuesta.Code, rol)
		assert.Contains(t, respuesta.Body.String(), "Reporte diario")
	}

	peticion := peticionConSesion(t, sesiones, http.MethodGet, "/reportes/diario", nil, "", "porteria")
	respuesta := ejecutarReportes(manejador, peticion)
	assert.Equal(t, http.StatusForbidden, respuesta.Code)

	peticion = peticionConSesion(t, sesiones, http.MethodGet, "/reportes/diario.xlsx", nil, "", "administrador")
	respuesta = ejecutarReportes(manejador, peticion)
	require.Equal(t, http.StatusOK, respuesta.Code)
	assert.Contains(t, respuesta.Header().Get("Content-Type"), "spreadsheetml")
	assert.NotEmpty(t, respuesta.Body.Bytes())

	peticion = peticionConSesion(t, sesiones, http.MethodGet, "/reportes/diario.pdf", nil, "", "administrador")
	respuesta = ejecutarReportes(manejador, peticion)
	require.Equal(t, http.StatusOK, respuesta.Code)
	assert.Equal(t, "application/pdf", respuesta.Header().Get("Content-Type"))
	assert.NotEmpty(t, respuesta.Body.Bytes())

	peticion = peticionConSesion(t, sesiones, http.MethodGet, "/reportes/reiteracion", nil, "", "director")
	respuesta = ejecutarReportes(manejador, peticion)
	require.Equal(t, http.StatusOK, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "Alumnos con reiteración")
}

func nuevoManejadorReportes(servicio reportes.ServicioReportes) (*reportes.Manejador, *scs.SessionManager) {
	sesiones := scs.New()
	return reportes.NuevoManejador(reportes.DependenciasManejador{
		Servicio: servicio,
		Sesiones: sesiones,
		Registro: registroDePrueba(),
	}), sesiones
}

func ejecutarReportes(manejador *reportes.Manejador, peticion *http.Request) *httptest.ResponseRecorder {
	enrutador := chi.NewRouter()
	manejador.Montar(enrutador)
	respuesta := httptest.NewRecorder()
	enrutador.ServeHTTP(respuesta, peticion)
	return respuesta
}

type servicioReportesFalso struct{}

func (s *servicioReportesFalso) Diario(context.Context, reportes.Filtros) ([]reportes.FilaDiaria, error) {
	return []reportes.FilaDiaria{{
		Turno: "manana", Grado: "3.°", Seccion: "A", Identificador: "A001",
		Alumno: "Pérez, Ana", Estado: reportes.EstadoPresente, Hora: "07:35",
	}}, nil
}

func (s *servicioReportesFalso) Panel(context.Context, reportes.Filtros) (reportes.PanelDelDia, error) {
	return reportes.PanelDelDia{Fecha: time.Now(), Presentes: 1, Total: 1}, nil
}

func (s *servicioReportesFalso) Resumen(context.Context, reportes.Filtros) ([]reportes.ResumenAlumno, error) {
	return []reportes.ResumenAlumno{{
		Grado: "3.°", Seccion: "A", Identificador: "A001", Alumno: "Pérez, Ana",
		Asistencias: 18, Faltas: 2, DiasLaborables: 20, Porcentaje: 90,
	}}, nil
}

func (s *servicioReportesFalso) Reiteracion(context.Context, reportes.Filtros) ([]reportes.Reiteracion, error) {
	return []reportes.Reiteracion{{
		Alumno: "Pérez, Ana", Seccion: "3.° A", Tardanzas: 3,
		UmbralTardanzas: 3, FaltasSeguidas: 2, UmbralFaltas: 2,
	}}, nil
}
