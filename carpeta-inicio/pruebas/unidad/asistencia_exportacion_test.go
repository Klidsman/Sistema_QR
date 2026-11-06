package unidad

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"
	"time"

	"asistencia/internal/asistencia"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerarCSV_IncluyeColumnasYRespetaComas(t *testing.T) {
	registros := []asistencia.RegistroExportacion{{
		Fecha:         time.Date(2026, time.November, 6, 0, 0, 0, 0, time.UTC),
		Curso:         "Matemática",
		Seccion:       "3.° A (2026)",
		Bloque:        1,
		Identificador: "80001111",
		Alumno:        "Ríos, Ana",
		Estado:        asistencia.EstadoTardanza,
		Observacion:   "Llegó tarde, con permiso",
		HoraQR:        "07:42",
		Docente:       "Docente Uno",
	}}

	contenido, err := asistencia.GenerarCSV(registros)
	require.NoError(t, err)
	assert.True(t, bytes.HasPrefix(contenido, []byte{0xEF, 0xBB, 0xBF}), "debe incluir BOM")

	lector := csv.NewReader(bytes.NewReader(contenido[3:]))
	filas, err := lector.ReadAll()
	require.NoError(t, err)
	require.Len(t, filas, 2)
	assert.Equal(t, []string{
		"fecha", "curso", "seccion", "bloque", "identificador",
		"alumno", "estado", "observacion", "hora_qr", "docente",
	}, filas[0])
	assert.Equal(t, "Ríos, Ana", filas[1][5])
	assert.Equal(t, "Llegó tarde, con permiso", filas[1][7])
	assert.True(t, strings.Contains(string(contenido), "Matemática"))
}
