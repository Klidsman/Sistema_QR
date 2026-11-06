// Package alumnos gestiona el registro, los estados y las credenciales QR de los alumnos.
package alumnos

import (
	"errors"
	"strings"
	"time"
	"unicode"
)

// AlumnoEnClase son los datos mínimos de un alumno para la lista del docente.
type AlumnoEnClase struct {
	ID            int64
	Identificador string
	Nombres       string
	Apellidos     string
}

// NombreCompleto devuelve apellidos y nombres en el orden visible.
func (a AlumnoEnClase) NombreCompleto() string {
	return a.Apellidos + ", " + a.Nombres
}

// Estados de un alumno.
const (
	EstadoActivo      = "activo"
	EstadoDesactivado = "desactivado"
	EstadoRetirado    = "retirado"
	EstadoTrasladado  = "trasladado"
)

// Estados de una credencial QR.
const (
	CredencialActiva   = "activa"
	CredencialRevocada = "revocada"
)

// Motivos de revocación de una credencial QR.
const (
	MotivoPerdida       = "perdida"
	MotivoRobo          = "robo"
	MotivoDesactivacion = "desactivacion"
	MotivoRetiro        = "retiro"
	MotivoOtro          = "otro"
)

// Errores de negocio del módulo.
var (
	ErrIdentificadorExistente   = errors.New("el identificador ya está registrado")
	ErrAlumnoNoEncontrado       = errors.New("el alumno no existe")
	ErrAlumnoNoActivo           = errors.New("el alumno no está activo")
	ErrQRYaActivo               = errors.New("el alumno ya tiene un QR vigente")
	ErrQRSinActivo              = errors.New("el alumno no tiene un QR vigente")
	ErrMotivoRequerido          = errors.New("el motivo de la revocación no es válido")
	ErrSeccionInvalida          = errors.New("la sección no es válida")
	ErrMismaSeccion             = errors.New("el alumno ya pertenece a esa sección")
	ErrFechaInvalida            = errors.New("la fecha no es válida")
	ErrHistorialYaCerrado       = errors.New("el tramo de historial ya está cerrado")
	ErrTipoSalidaInvalido       = errors.New("el tipo de salida no es válido")
	ErrFotoInvalida             = errors.New("la foto no es un archivo JPG o PNG válido")
	ErrFotoMuyGrande            = errors.New("la foto supera el tamaño máximo permitido")
	ErrSeleccionCarnetsInvalida = errors.New("debe elegir un grado o una sección para los carnés")
)

// DatosAlumno son los campos editables de un alumno.
type DatosAlumno struct {
	Identificador     string
	Nombres           string
	Apellidos         string
	SeccionID         int64
	ApoderadoNombre   string
	ApoderadoTelefono string
}

// ErrorValidacion agrupa los mensajes por campo para volver a mostrar el formulario.
type ErrorValidacion struct {
	Campos map[string]string
}

// Error cumple con la interfaz error.
func (e *ErrorValidacion) Error() string {
	return "los datos del alumno no son válidos"
}

// Validar comprueba los campos obligatorios y el formato del identificador.
// Devuelve nil cuando los datos son válidos.
func Validar(datos DatosAlumno) *ErrorValidacion {
	campos := make(map[string]string)

	identificador := LimpiarIdentificador(datos.Identificador)
	switch {
	case identificador == "":
		campos["identificador"] = "El DNI o código es obligatorio."
	case strings.IndexFunc(identificador, unicode.IsSpace) >= 0:
		campos["identificador"] = "El DNI o código no puede contener espacios."
	}

	if strings.TrimSpace(datos.Nombres) == "" {
		campos["nombres"] = "Los nombres son obligatorios."
	}
	if strings.TrimSpace(datos.Apellidos) == "" {
		campos["apellidos"] = "Los apellidos son obligatorios."
	}
	if datos.SeccionID <= 0 {
		campos["seccion_id"] = "El grado y la sección son obligatorios."
	}

	if len(campos) == 0 {
		return nil
	}
	return &ErrorValidacion{Campos: campos}
}

// MotivoRevocacionValido indica si el motivo pertenece al catálogo permitido.
func MotivoRevocacionValido(motivo string) bool {
	switch motivo {
	case MotivoPerdida, MotivoRobo, MotivoDesactivacion, MotivoRetiro, MotivoOtro:
		return true
	default:
		return false
	}
}

// Alumno es la ficha completa que se muestra en la interfaz.
type Alumno struct {
	ID                int64
	Identificador     string
	Nombres           string
	Apellidos         string
	Estado            string
	SeccionID         int64
	Grado             string
	Seccion           string
	Turno             string
	AnioEscolar       string
	FotoClave         string
	ApoderadoNombre   string
	ApoderadoTelefono string
	FechaSalida       *time.Time
	CreadoEn          time.Time
	ActualizadoEn     time.Time
}

// NombreCompleto devuelve nombres y apellidos en el orden en que se muestran.
func (a Alumno) NombreCompleto() string {
	return a.Apellidos + ", " + a.Nombres
}

// Activo indica si el alumno puede tener un QR vigente.
func (a Alumno) Activo() bool {
	return a.Estado == EstadoActivo
}

// AlumnoListado es una fila del listado de alumnos.
type AlumnoListado struct {
	ID            int64
	Identificador string
	Nombres       string
	Apellidos     string
	Estado        string
	SeccionID     int64
	Grado         string
	Seccion       string
	AnioEscolar   string
	TieneFoto     bool
}

// NombreCompleto devuelve nombres y apellidos en el orden en que se muestran.
func (a AlumnoListado) NombreCompleto() string {
	return a.Apellidos + ", " + a.Nombres
}

// FiltrosListado son los filtros opcionales del listado.
type FiltrosListado struct {
	Texto         string
	Estado        string
	SeccionID     int64
	AnioEscolarID int64
}

// CredencialQR es una credencial emitida para un alumno.
type CredencialQR struct {
	ID         int64
	Token      string
	Estado     string
	Motivo     string
	EmitidaEn  time.Time
	RevocadaEn *time.Time
}
