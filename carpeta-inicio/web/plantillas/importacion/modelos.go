// Package vista contiene los modelos que reciben las plantillas de importación.
package vista

// Importacion es el resumen de una carga masiva ya procesada.
type Importacion struct {
	ID              int64
	ArchivoNombre   string
	FilasLeidas     int
	FilasCreadas    int
	FilasRechazadas int
	CreadaEn        string
}

// FilaRechazada es una fila que no se pudo importar, con su motivo legible.
type FilaRechazada struct {
	NumeroFila    int
	Identificador string
	Motivo        string
}

// DetalleImportacion incluye el resumen y las filas rechazadas.
type DetalleImportacion struct {
	Importacion Importacion
	Rechazadas  []FilaRechazada
}

// TotalRechazadas devuelve cuántas filas rechazadas se listan.
func (d DetalleImportacion) TotalRechazadas() int {
	return len(d.Rechazadas)
}
