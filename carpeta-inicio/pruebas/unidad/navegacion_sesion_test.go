package unidad

import (
	"testing"

	"asistencia/internal/plataforma/intermediarios"

	"github.com/stretchr/testify/assert"
)

func TestRutaInicioPorRol(t *testing.T) {
	assert.Equal(t, "/alumnos", intermediarios.RutaInicioPorRol("administrador"))
	assert.Equal(t, "/alumnos", intermediarios.RutaInicioPorRol("director"))
	assert.Equal(t, "/asistencia", intermediarios.RutaInicioPorRol("docente"))
	assert.Equal(t, "/porteria", intermediarios.RutaInicioPorRol("porteria"))
}

func TestRutaPermitidaParaRol(t *testing.T) {
	casos := []struct {
		rol       string
		ruta      string
		permitida bool
	}{
		{"docente", "/asistencia", true},
		{"docente", "/reportes/diario", true},
		{"docente", "/alumnos", false},
		{"docente", "/porteria", false},
		{"porteria", "/porteria", true},
		{"porteria", "/alertas", true},
		{"porteria", "/alumnos", false},
		{"director", "/docentes", true},
		{"director", "/porteria", false},
		{"administrador", "/cualquier/ruta", true},
		{"docente", "https://ejemplo.com/alumnos", false},
		{"docente", "//ejemplo.com/alumnos", false},
		{"docente", "/ingreso", false},
		{"", "/alumnos", false},
	}
	for _, caso := range casos {
		assert.Equal(t, caso.permitida, intermediarios.RutaPermitidaParaRol(caso.rol, caso.ruta),
			"rol %s ruta %s", caso.rol, caso.ruta)
	}
}
