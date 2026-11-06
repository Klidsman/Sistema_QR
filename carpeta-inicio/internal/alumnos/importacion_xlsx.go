package alumnos

import (
	"bytes"
	"fmt"

	"github.com/xuri/excelize/v2"
)

// LeerXLSX lee la primera hoja de un archivo Excel.
func LeerXLSX(contenido []byte, maxFilas int) ([]FilaImportada, error) {
	archivo, err := excelize.OpenReader(bytes.NewReader(contenido))
	if err != nil {
		return nil, fmt.Errorf("abriendo el XLSX: %w", err)
	}
	defer func() { _ = archivo.Close() }()

	hojas := archivo.GetSheetList()
	if len(hojas) == 0 {
		return nil, ErrArchivoVacio
	}
	registros, err := archivo.GetRows(hojas[0])
	if err != nil {
		return nil, fmt.Errorf("leyendo la primera hoja del XLSX: %w", err)
	}
	return filasDesdeRegistros(registros, maxFilas)
}
