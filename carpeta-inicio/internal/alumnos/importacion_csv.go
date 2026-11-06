package alumnos

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"
)

// LeerCSV lee un archivo CSV con BOM opcional y separador coma o punto y coma.
func LeerCSV(contenido []byte, maxFilas int) ([]FilaImportada, error) {
	contenido = bytes.TrimPrefix(contenido, []byte{0xEF, 0xBB, 0xBF})
	if len(bytes.TrimSpace(contenido)) == 0 {
		return nil, ErrArchivoVacio
	}

	lector := csv.NewReader(bytes.NewReader(contenido))
	lector.Comma = separadorCSV(contenido)
	lector.FieldsPerRecord = -1
	lector.LazyQuotes = true
	lector.TrimLeadingSpace = true

	registros, err := lector.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("leyendo el CSV: %w", err)
	}
	return filasDesdeRegistros(registros, maxFilas)
}

func separadorCSV(contenido []byte) rune {
	primera := primeraLinea(contenido)
	camposComa := contarCamposCSV(primera, ',')
	camposPuntoYComa := contarCamposCSV(primera, ';')
	if camposPuntoYComa > camposComa {
		return ';'
	}
	return ','
}

func primeraLinea(contenido []byte) string {
	if posicion := bytes.IndexByte(contenido, '\n'); posicion >= 0 {
		return string(contenido[:posicion])
	}
	return string(contenido)
}

func contarCamposCSV(linea string, separador rune) int {
	lector := csv.NewReader(strings.NewReader(linea))
	lector.Comma = separador
	lector.FieldsPerRecord = -1
	lector.LazyQuotes = true
	lector.TrimLeadingSpace = true
	registro, err := lector.Read()
	if err != nil {
		return 0
	}
	return len(registro)
}
