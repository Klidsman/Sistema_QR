// Package justificaciones gestiona las justificaciones de faltas y tardanzas.
package justificaciones

import (
	"context"
	"errors"
	"time"

	"asistencia/internal/academico"
	"asistencia/internal/alumnos"
	"asistencia/internal/asistencia"
	"asistencia/internal/plataforma/almacenamiento"
)

// Estados de una justificación y de sus fechas.
const (
	EstadoPendiente        = "pendiente"
	EstadoAprobada         = "aprobada"
	EstadoRechazada        = "rechazada"
	EstadoAnulada          = "anulada"
	EstadoRegistroNoAplica = "registro_no_aplica"
)

// Tipos de registro justificable.
const (
	TipoFalta    = "falta"
	TipoTardanza = "tardanza"
)

// Entidad y acciones de auditoría.
const (
	EntidadJustificacion        = "justificacion"
	AccionCrearJustificacion    = "crear_justificacion"
	AccionAprobarJustificacion  = "aprobar_justificacion"
	AccionRechazarJustificacion = "rechazar_justificacion"
	AccionAnularJustificacion   = "anular_justificacion"
)

// Errores de negocio.
var (
	ErrMotivoRequerido       = errors.New("indique el motivo de la justificación")
	ErrMotivoLargo           = errors.New("el motivo no puede superar los 500 caracteres")
	ErrSinFechas             = errors.New("indique al menos una fecha")
	ErrRangoInvalido         = errors.New("la fecha final debe ser igual o posterior a la inicial")
	ErrFueraDePlazo          = errors.New("la justificación está fuera del plazo permitido")
	ErrSinFechasAplicables   = errors.New("no hay fechas con falta o tardanza que justificar")
	ErrJustificacionNoExiste = errors.New("la justificación no existe")
	ErrYaResuelta            = errors.New("la justificación ya fue resuelta")
	ErrAutoresolucion        = errors.New("no puede resolver una justificación que usted registró")
	ErrComentarioRequerido   = errors.New("indique el comentario del rechazo")
	ErrSinAcceso             = errors.New("no tiene acceso a esta justificación")
	ErrSoloAdministrador     = errors.New("solo el administrador puede anular justificaciones")
	ErrNoAnulable            = errors.New("solo se puede anular una justificación aprobada")
	ErrMotivoAnulacion       = errors.New("indique el motivo de la anulación")
	ErrAdjuntosExcedidos     = errors.New("puede adjuntar como máximo tres documentos")
	ErrAdjuntoInvalido       = errors.New("solo se aceptan PDF, JPG o PNG de hasta 5 MB")
	ErrAlumnoInactivo        = errors.New("el alumno no está activo")
)

// FechaJustificada es una fecha cubierta por la justificación.
type FechaJustificada struct {
	ID              int64
	JustificacionID int64
	AlumnoID        int64
	Fecha           time.Time
	Tipo            string
	Estado          string
}

// Adjunto es un documento adjunto a la justificación.
type Adjunto struct {
	ID              int64
	JustificacionID int64
	Nombre          string
	Tipo            string
	Tamano          int64
	Clave           string
	CreadoEn        time.Time
}

// EntradaAdjunto es un documento recibido desde el formulario.
type EntradaAdjunto struct {
	Nombre    string
	Tipo      string
	Tamano    int64
	Contenido []byte
}

// Justificacion es una solicitud y sus fechas.
type Justificacion struct {
	ID              int64
	AlumnoID        int64
	SeccionID       int64
	Motivo          string
	Estado          string
	RegistranteID   int64
	ResueltoPor     *int64
	ResueltoEn      *time.Time
	Comentario      string
	AnuladoPor      *int64
	AnuladoEn       *time.Time
	MotivoAnulacion string
	CreadaEn        time.Time
	Fechas          []FechaJustificada
	Adjuntos        []Adjunto
}

// Omision describe un día del rango que no se justificó.
type Omision struct {
	Fecha  time.Time
	Motivo string
}

// ResultadoRegistro informa lo cubierto y lo omitido.
type ResultadoRegistro struct {
	JustificacionID int64
	Cubiertas       []FechaJustificada
	Omitidas        []Omision
}

// Dependencias reúne las fuentes del módulo.
type Dependencias struct {
	Alumnos    FuenteAlumnos
	Academico  FuenteAcademica
	Asistencia FuenteAsistencia
	Alertas    FuenteAlertas
	Almacen    almacenamiento.AlmacenFotos
	Repo       *Repositorio
	Reloj      func() time.Time
}

// FuenteAlumnos expone datos del alumno.
type FuenteAlumnos interface {
	Obtener(ctx context.Context, alumnoID int64) (alumnos.Alumno, error)
}

// FuenteAcademica expone calendario y etiquetas.
type FuenteAcademica interface {
	EsDiaNoLaborable(ctx context.Context, seccionID int64, fecha time.Time) (bool, error)
	EtiquetaSeccion(ctx context.Context, seccionID int64) (string, error)
}

// FuenteAsistencia expone marcaciones y permisos docentes.
type FuenteAsistencia interface {
	MarcacionDelDia(ctx context.Context, alumnoID int64, fecha time.Time) (asistencia.Marcacion, bool, error)
	MarcarTardanzaJustificada(ctx context.Context, alumnoID int64, fecha time.Time, justificada bool) error
	DocenteAsignadoASeccionEnFecha(ctx context.Context, docenteID, seccionID int64, fecha time.Time) (bool, error)
	PlazoJustificacion(ctx context.Context) (int, error)
}

// FuenteAlertas marca las alertas afectadas como situación cambiada.
type FuenteAlertas interface {
	MarcarSituacionCambiada(ctx context.Context, alumnoID int64, fecha time.Time, motivo string) error
}

// academico.HorarioTurno se usa solo para mantener el enlace del paquete.
var _ = academico.HorarioTurno{}
