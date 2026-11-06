package alumnos

import "strings"

// SeleccionCarnets indica el grado o la sección elegidos para generar un lote.
// Si hay sección, prevalece sobre el grado.
type SeleccionCarnets struct {
	SeccionID int64
	GradoID   int64
}

// Valida indica si la selección permite buscar alumnos.
func (s SeleccionCarnets) Valida() bool {
	return s.SeccionID > 0 || s.GradoID > 0
}

// AlumnoCarnet reúne los datos que se imprimen en un carné.
type AlumnoCarnet struct {
	ID            int64
	Identificador string
	Nombres       string
	Apellidos     string
	SeccionID     int64
	GradoID       int64
	Grado         string
	Seccion       string
	AnioEscolar   string
	FotoClave     string
	TokenQR       string
}

// NombreCompleto devuelve nombres y apellidos en el orden en que se muestran.
func (a AlumnoCarnet) NombreCompleto() string {
	return a.Apellidos + ", " + a.Nombres
}

// TieneFoto indica si el alumno tiene una foto registrada.
func (a AlumnoCarnet) TieneFoto() bool {
	return strings.TrimSpace(a.FotoClave) != ""
}

// TieneQR indica si el alumno tiene una credencial vigente.
func (a AlumnoCarnet) TieneQR() bool {
	return strings.TrimSpace(a.TokenQR) != ""
}

// AvisosCarnets son los alumnos que requieren atención antes de imprimir.
type AvisosCarnets struct {
	SinFoto []AlumnoCarnet
	SinQR   []AlumnoCarnet
}

// HayAvisos indica si existe algún alumno que deba revisarse antes de imprimir.
func (a AvisosCarnets) HayAvisos() bool {
	return len(a.SinFoto) > 0 || len(a.SinQR) > 0
}

// LoteCarnets es el resultado de preparar una selección para impresión.
type LoteCarnets struct {
	Carnets []AlumnoCarnet
	Avisos  AvisosCarnets
}

// PrepararLote separa los carnés imprimibles de los avisos previos.
// Un alumno sin QR vigente no se imprime para no generar un carné con el QR vacío.
func PrepararLote(alumnos []AlumnoCarnet) LoteCarnets {
	lote := LoteCarnets{
		Carnets: make([]AlumnoCarnet, 0, len(alumnos)),
	}
	for _, alumno := range alumnos {
		if !alumno.TieneQR() {
			lote.Avisos.SinQR = append(lote.Avisos.SinQR, alumno)
			continue
		}
		if !alumno.TieneFoto() {
			lote.Avisos.SinFoto = append(lote.Avisos.SinFoto, alumno)
		}
		lote.Carnets = append(lote.Carnets, alumno)
	}
	return lote
}
