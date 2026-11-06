package alumnos

import (
	"errors"
	"strings"
	"time"
)

// Motivos de rechazo de una fila de importación.
const (
	MotivoCampoObligatorioVacio          = "campo_obligatorio_vacio"
	MotivoIdentificadorRepetidoEnArchivo = "identificador_repetido_en_archivo"
	MotivoIdentificadorExistente         = "identificador_existente"
	MotivoSeccionInexistente             = "seccion_inexistente"
	MotivoFormatoInvalido                = "formato_invalido"
)

// Errores de negocio de la importación.
var (
	ErrArchivoVacio            = errors.New("el archivo no contiene filas para importar")
	ErrArchivoMuyGrande        = errors.New("el archivo supera el tamaño máximo permitido")
	ErrFormatoArchivo          = errors.New("el archivo debe ser CSV o XLSX")
	ErrDemasiadasFilas         = errors.New("el archivo supera el máximo de filas permitido")
	ErrImportacionNoEncontrada = errors.New("la importación no existe")
)

// ErrorColumnasFaltantes indica que el archivo no tiene las columnas obligatorias.
type ErrorColumnasFaltantes struct {
	Columnas []string
}

// Error describe las columnas que faltan.
func (e *ErrorColumnasFaltantes) Error() string {
	return "faltan columnas obligatorias: " + strings.Join(e.Columnas, ", ")
}

// FilaImportada es una fila leída del archivo, antes de validarla contra la base de datos.
type FilaImportada struct {
	Numero            int
	Identificador     string
	Nombres           string
	Apellidos         string
	Grado             string
	Seccion           string
	ApoderadoNombre   string
	ApoderadoTelefono string
}

// FilaRechazada es una fila que no se pudo importar y el motivo del rechazo.
type FilaRechazada struct {
	NumeroFila    int
	Identificador string
	Motivo        string
}

// Importacion es el resumen de una carga masiva ya procesada.
type Importacion struct {
	ID              int64
	UsuarioID       int64
	ArchivoNombre   string
	FilasLeidas     int
	FilasCreadas    int
	FilasRechazadas int
	CreadaEn        time.Time
}

// ImportacionDetalle incluye las filas rechazadas de una importación.
type ImportacionDetalle struct {
	Importacion
	Rechazadas []FilaRechazada
}

// MotivoImportacionLegible devuelve el texto en español del motivo de rechazo.
func MotivoImportacionLegible(motivo string) string {
	switch motivo {
	case MotivoCampoObligatorioVacio:
		return "Campo obligatorio vacío"
	case MotivoIdentificadorRepetidoEnArchivo:
		return "Identificador repetido en el archivo"
	case MotivoIdentificadorExistente:
		return "El identificador ya existe"
	case MotivoSeccionInexistente:
		return "Grado o sección inexistente"
	case MotivoFormatoInvalido:
		return "Formato inválido"
	default:
		return motivo
	}
}
