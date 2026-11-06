package unidad

import (
	"testing"
	"time"

	"asistencia/internal/asistencia"
)

func TestClaveClase_DistingueElBloqueYLosDemasCampos(t *testing.T) {
	fecha := time.Date(2026, time.November, 6, 0, 0, 0, 0, time.UTC)

	bloque1 := asistencia.ClaveClase(fecha, 5, 2, 1)
	bloque2 := asistencia.ClaveClase(fecha, 5, 2, 2)
	if bloque1 == bloque2 {
		t.Fatal("dos bloques distintos no pueden compartir la clave")
	}

	if bloque1 != asistencia.ClaveClase(fecha, 5, 2, 1) {
		t.Fatal("la misma clase debe producir la misma clave")
	}
	if bloque1 == asistencia.ClaveClase(fecha.AddDate(0, 0, 1), 5, 2, 1) {
		t.Fatal("dos fechas distintas no pueden compartir la clave")
	}
	if bloque1 == asistencia.ClaveClase(fecha, 6, 2, 1) {
		t.Fatal("dos secciones distintas no pueden compartir la clave")
	}
	if bloque1 == asistencia.ClaveClase(fecha, 5, 3, 1) {
		t.Fatal("dos cursos distintos no pueden compartir la clave")
	}
}
