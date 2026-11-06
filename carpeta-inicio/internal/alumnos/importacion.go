package alumnos

import (
	"strings"
)

const maximoFilasImportacionPorDefecto = 5000

type definicionColumna struct {
	clave     string
	requerida bool
	alias     []string
}

var columnasImportacion = []definicionColumna{
	{clave: "identificador", requerida: true, alias: []string{"identificador", "dni", "codigo", "documento"}},
	{clave: "nombres", requerida: true, alias: []string{"nombres", "nombre"}},
	{clave: "apellidos", requerida: true, alias: []string{"apellidos", "apellido"}},
	{clave: "grado", requerida: true, alias: []string{"grado"}},
	{clave: "seccion", requerida: true, alias: []string{"seccion"}},
	{clave: "apoderado_nombre", requerida: false, alias: []string{"apoderado_nombre", "apoderado", "nombre_apoderado"}},
	{clave: "apoderado_telefono", requerida: false, alias: []string{"apoderado_telefono", "telefono_apoderado", "telefono"}},
}

// ClasificarFilas valida los campos obligatorios y detecta identificadores repetidos
// dentro del archivo. Devuelve las filas que pueden importarse y las rechazadas.
func ClasificarFilas(filas []FilaImportada) ([]FilaImportada, []FilaRechazada) {
	limpias := make([]FilaImportada, 0, len(filas))
	conteo := make(map[string]int)
	for _, fila := range filas {
		fila = filaLimpia(fila)
		if fila.Identificador != "" {
			conteo[fila.Identificador]++
		}
		limpias = append(limpias, fila)
	}

	validas := make([]FilaImportada, 0, len(limpias))
	rechazadas := make([]FilaRechazada, 0)
	for _, fila := range limpias {
		motivo := ""
		switch {
		case fila.Identificador == "" || fila.Nombres == "" || fila.Apellidos == "" ||
			fila.Grado == "" || fila.Seccion == "":
			motivo = MotivoCampoObligatorioVacio
		case strings.ContainsAny(fila.Identificador, " \t\r\n"):
			motivo = MotivoFormatoInvalido
		case conteo[fila.Identificador] > 1:
			motivo = MotivoIdentificadorRepetidoEnArchivo
		}

		if motivo == "" {
			validas = append(validas, fila)
			continue
		}
		rechazadas = append(rechazadas, FilaRechazada{
			NumeroFila:    fila.Numero,
			Identificador: fila.Identificador,
			Motivo:        motivo,
		})
	}
	return validas, rechazadas
}

func filaLimpia(fila FilaImportada) FilaImportada {
	fila.Identificador = LimpiarIdentificador(fila.Identificador)
	fila.Nombres = strings.TrimSpace(fila.Nombres)
	fila.Apellidos = strings.TrimSpace(fila.Apellidos)
	fila.Grado = strings.TrimSpace(fila.Grado)
	fila.Seccion = strings.TrimSpace(fila.Seccion)
	fila.ApoderadoNombre = strings.TrimSpace(fila.ApoderadoNombre)
	fila.ApoderadoTelefono = strings.TrimSpace(fila.ApoderadoTelefono)
	return fila
}

func filasDesdeRegistros(registros [][]string, maxFilas int) ([]FilaImportada, error) {
	if maxFilas <= 0 {
		maxFilas = maximoFilasImportacionPorDefecto
	}
	if len(registros) == 0 {
		return nil, ErrArchivoVacio
	}

	indices, err := indicesColumnas(registros[0])
	if err != nil {
		return nil, err
	}

	filas := make([]FilaImportada, 0)
	for posicion, registro := range registros[1:] {
		if registroVacio(registro) {
			continue
		}
		if len(filas) >= maxFilas {
			return nil, ErrDemasiadasFilas
		}
		filas = append(filas, FilaImportada{
			Numero:            posicion + 2,
			Identificador:     valorColumna(registro, indices, "identificador"),
			Nombres:           valorColumna(registro, indices, "nombres"),
			Apellidos:         valorColumna(registro, indices, "apellidos"),
			Grado:             valorColumna(registro, indices, "grado"),
			Seccion:           valorColumna(registro, indices, "seccion"),
			ApoderadoNombre:   valorColumna(registro, indices, "apoderado_nombre"),
			ApoderadoTelefono: valorColumna(registro, indices, "apoderado_telefono"),
		})
	}
	if len(filas) == 0 {
		return nil, ErrArchivoVacio
	}
	return filas, nil
}

func indicesColumnas(encabezados []string) (map[string]int, error) {
	presentes := make(map[string]int)
	for posicion, encabezado := range encabezados {
		clave := claveColumna(encabezado)
		if clave != "" {
			if _, existe := presentes[clave]; !existe {
				presentes[clave] = posicion
			}
		}
	}

	indices := make(map[string]int)
	faltantes := make([]string, 0)
	for _, columna := range columnasImportacion {
		encontrada := -1
		for _, alias := range columna.alias {
			if posicion, existe := presentes[alias]; existe {
				encontrada = posicion
				break
			}
		}
		if encontrada >= 0 {
			indices[columna.clave] = encontrada
			continue
		}
		if columna.requerida {
			faltantes = append(faltantes, columna.clave)
		}
	}
	if len(faltantes) > 0 {
		return nil, &ErrorColumnasFaltantes{Columnas: faltantes}
	}
	return indices, nil
}

func claveColumna(texto string) string {
	clave := Normalizar(strings.ReplaceAll(strings.TrimSpace(texto), "\ufeff", ""))
	clave = strings.ReplaceAll(clave, "-", "_")
	clave = strings.ReplaceAll(clave, " ", "_")
	return strings.Trim(clave, "_")
}

func valorColumna(registro []string, indices map[string]int, clave string) string {
	posicion, existe := indices[clave]
	if !existe || posicion < 0 || posicion >= len(registro) {
		return ""
	}
	return strings.TrimSpace(registro[posicion])
}

func registroVacio(registro []string) bool {
	for _, campo := range registro {
		if strings.TrimSpace(campo) != "" {
			return false
		}
	}
	return true
}
