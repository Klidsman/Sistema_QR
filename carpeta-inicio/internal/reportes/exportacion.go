package reportes

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/go-pdf/fpdf"
	"github.com/xuri/excelize/v2"
	"golang.org/x/image/font/gofont/goregular"
)

// ExportarDiarioXLSX genera el reporte diario en Excel.
func ExportarDiarioXLSX(filas []FilaDiaria, fecha string) ([]byte, error) {
	archivo := excelize.NewFile()
	hoja := "Diario"
	indice, _ := archivo.NewSheet(hoja)
	archivo.SetActiveSheet(indice)
	encabezados := []string{"Fecha", "Turno", "Grado", "Sección", "Identificador", "Alumno", "Estado", "Hora"}
	for columna, encabezado := range encabezados {
		celda, _ := excelize.CoordinatesToCellName(columna+1, 1)
		_ = archivo.SetCellValue(hoja, celda, encabezado)
	}
	for fila, registro := range filas {
		valores := []string{fecha, registro.Turno, registro.Grado, registro.Seccion,
			registro.Identificador, registro.Alumno, EtiquetaEstado(registro.Estado), registro.Hora}
		for columna, valor := range valores {
			celda, _ := excelize.CoordinatesToCellName(columna+1, fila+2)
			_ = archivo.SetCellValue(hoja, celda, valor)
		}
	}
	var buffer bytes.Buffer
	if err := archivo.Write(&buffer); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// ExportarResumenXLSX genera el resumen acumulado en Excel.
func ExportarResumenXLSX(resumen []ResumenAlumno, desde, hasta string) ([]byte, error) {
	archivo := excelize.NewFile()
	hoja := "Resumen"
	indice, _ := archivo.NewSheet(hoja)
	archivo.SetActiveSheet(indice)
	encabezados := []string{"Desde", "Hasta", "Grado", "Sección", "Identificador", "Alumno",
		"Asistencias", "Tardanzas", "Faltas", "Faltas justificadas", "Tardanzas justificadas", "Porcentaje"}
	for columna, encabezado := range encabezados {
		celda, _ := excelize.CoordinatesToCellName(columna+1, 1)
		_ = archivo.SetCellValue(hoja, celda, encabezado)
	}
	for fila, registro := range resumen {
		valores := []any{desde, hasta, registro.Grado, registro.Seccion, registro.Identificador,
			registro.Alumno, registro.Asistencias, registro.Tardanzas, registro.Faltas,
			registro.FaltasJustificadas, registro.TardanzasJustificadas,
			fmt.Sprintf("%.1f", registro.Porcentaje)}
		for columna, valor := range valores {
			celda, _ := excelize.CoordinatesToCellName(columna+1, fila+2)
			_ = archivo.SetCellValue(hoja, celda, valor)
		}
	}
	var buffer bytes.Buffer
	if err := archivo.Write(&buffer); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// ExportarDiarioPDF genera el reporte diario en PDF.
func ExportarDiarioPDF(filas []FilaDiaria, fecha string) ([]byte, error) {
	pdf := nuevoPDF("Reporte diario " + fecha)
	filaPDF(pdf, []string{"Turno", "Grado", "Sección", "Identificador", "Alumno", "Estado", "Hora"}, true)
	for _, registro := range filas {
		if pdf.GetY() > 180 {
			pdf.AddPage()
		}
		filaPDF(pdf, []string{registro.Turno, registro.Grado, registro.Seccion,
			registro.Identificador, registro.Alumno, EtiquetaEstado(registro.Estado), registro.Hora}, false)
	}
	return salidaPDF(pdf)
}

// ExportarResumenPDF genera el resumen acumulado en PDF.
func ExportarResumenPDF(resumen []ResumenAlumno, desde, hasta string) ([]byte, error) {
	pdf := nuevoPDF("Reporte del " + desde + " al " + hasta)
	filaPDF(pdf, []string{"Grado", "Sección", "Alumno", "Asist.", "Tard.", "Faltas", "Justif.", "%"}, true)
	for _, registro := range resumen {
		if pdf.GetY() > 180 {
			pdf.AddPage()
		}
		filaPDF(pdf, []string{
			registro.Grado,
			registro.Seccion,
			registro.Alumno,
			strconv.Itoa(registro.Asistencias),
			strconv.Itoa(registro.Tardanzas),
			strconv.Itoa(registro.Faltas),
			strconv.Itoa(registro.FaltasJustificadas + registro.TardanzasJustificadas),
			fmt.Sprintf("%.1f", registro.Porcentaje),
		}, false)
	}
	return salidaPDF(pdf)
}

func nuevoPDF(titulo string) *fpdf.Fpdf {
	pdf := fpdf.New("L", "mm", "A4", "")
	pdf.AddUTF8FontFromBytes("gofont", "", goregular.TTF)
	pdf.AddPage()
	pdf.SetFont("gofont", "", 14)
	pdf.Cell(0, 10, titulo)
	pdf.Ln(12)
	pdf.SetFont("gofont", "", 9)
	return pdf
}

func filaPDF(pdf *fpdf.Fpdf, valores []string, encabezado bool) {
	anchos := []float64{22, 24, 24, 32, 70, 34, 20}
	if encabezado {
		pdf.SetFont("gofont", "", 9)
	}
	for indice, valor := range valores {
		if indice >= len(anchos) {
			break
		}
		pdf.CellFormat(anchos[indice], 7, valor, "1", 0, "", false, 0, "")
	}
	pdf.Ln(-1)
}

func salidaPDF(pdf *fpdf.Fpdf) ([]byte, error) {
	var buffer bytes.Buffer
	if err := pdf.Output(&buffer); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
