// Package vista contiene los modelos que reciben las plantillas de carnés.
package vista

import (
	"strconv"

	"github.com/a-h/templ"
)

// Grado es una opción del selector de grado, construida a partir de las secciones.
type Grado struct {
	ID          int64
	Nombre      string
	AnioEscolar string
}

// Etiqueta devuelve el texto que se muestra en los selectores.
func (g Grado) Etiqueta() string {
	return g.Nombre + " (" + g.AnioEscolar + ")"
}

// Alumno es un carné ya preparado para la plantilla.
type Alumno struct {
	ID            int64
	Identificador string
	Nombres       string
	Apellidos     string
	Grado         string
	Seccion       string
	AnioEscolar   string
	TieneFoto     bool
	TokenQR       string
}

// NombreCompleto devuelve apellidos y nombres en el orden en que se muestran.
func (a Alumno) NombreCompleto() string {
	return a.Apellidos + ", " + a.Nombres
}

// Avisos son los alumnos que requieren atención antes de imprimir.
type Avisos struct {
	SinFoto []Alumno
	SinQR   []Alumno
}

// TotalSinFoto devuelve cuántos alumnos del lote no tienen foto.
func (a Avisos) TotalSinFoto() int {
	return len(a.SinFoto)
}

// TotalSinQR devuelve cuántos alumnos del lote no tienen QR vigente.
func (a Avisos) TotalSinQR() int {
	return len(a.SinQR)
}

// HayAvisos indica si debe mostrarse el bloque de advertencias.
func (a Avisos) HayAvisos() bool {
	return len(a.SinFoto) > 0 || len(a.SinQR) > 0
}

// TextoSinFoto describe cuántos alumnos del lote no tienen foto.
func (a Avisos) TextoSinFoto() string {
	return textoAlumnos(len(a.SinFoto)) + " sin foto"
}

// TextoSinQR describe cuántos alumnos del lote no tienen QR vigente.
func (a Avisos) TextoSinQR() string {
	return textoAlumnos(len(a.SinQR)) + " sin QR vigente"
}

func textoAlumnos(cantidad int) string {
	if cantidad == 1 {
		return "1 alumno"
	}
	return strconv.Itoa(cantidad) + " alumnos"
}

// ContenidoLote es el contenido imprimible y sus avisos previos.
type ContenidoLote struct {
	Carnets []Alumno
	Avisos  Avisos
}

const estilosLote = `<style>
	@page { size: A4; margin: 8mm; }
	.carnets { display: flex; flex-wrap: wrap; gap: 5mm; }
	@media print {
		header, .sin-imprimir { display: none !important; }
		body { background: #ffffff !important; }
		main { max-width: none !important; margin: 0 !important; padding: 0 !important; }
		.carnets { gap: 3mm; }
	}
</style>`

// EstilosLote entrega las reglas de impresión del lote.
func EstilosLote() templ.Component {
	return templ.Raw(estilosLote)
}
