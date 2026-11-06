package alertas

import (
	"context"
	"time"

	"asistencia/internal/academico"
)

// Tipos de alerta.
const (
	TipoTardanza            = "tardanza"
	TipoReiteracionTardanza = "reiteracion_tardanzas"
	TipoReiteracionFalta    = "reiteracion_faltas"
)

// Estados de una alerta.
const (
	EstadoNueva = "nueva"
	EstadoVista = "vista"
)

// Entidades de auditoría.
const (
	EntidadAlerta       = "alerta"
	EntidadCierreAlerta = "cierre_alerta"
)

// Acciones de auditoría.
const (
	AccionCrearAlerta             = "crear_alerta"
	AccionMarcarVista             = "marcar_vista"
	AccionMarcarVistas            = "marcar_vistas"
	AccionAsignarResponsable      = "asignar_responsable"
	AccionMarcarSituacionCambiada = "marcar_situacion_cambiada"
	AccionEjecutarCierre          = "ejecutar_cierre"
)

// EventoMarcacion es una marcación oficial que puede originar una alerta.
type EventoMarcacion struct {
	MarcacionID   int64
	AlumnoID      int64
	Fecha         time.Time
	Hora          time.Time
	Estado        string
	RegistradaPor *int64
}

// AlumnoEvento son los datos mínimos de un alumno para una alerta.
type AlumnoEvento struct {
	ID            int64
	Identificador string
	Nombres       string
	Apellidos     string
}

// Alerta es un aviso interno para un usuario.
type Alerta struct {
	ID                int64
	Tipo              string
	AlumnoID          int64
	SeccionID         int64
	CursoID           *int64
	BloqueNumero      *int
	FechaEvento       time.Time
	Mes               *time.Time
	MarcacionID       *int64
	DestinatarioID    int64
	Estado            string
	LeidaEn           *time.Time
	LeidaPor          *int64
	SinResponsable    bool
	SituacionCambiada bool
	MotivoCambio      string
	CreadaEn          time.Time
}

// FiltrosAlerta son los filtros del panel.
type FiltrosAlerta struct {
	Desde          *time.Time
	Hasta          *time.Time
	SeccionID      int64
	Tipo           string
	Estado         string
	DestinatarioID int64
	Todos          bool
}

// Dependencias reúne lo que el servicio de alertas necesita.
type Dependencias struct {
	Asistencia  FuenteAsistencia
	Alumnos     FuenteAlumnos
	Academico   FuenteAcademica
	Usuarios    FuenteUsuarios
	Repositorio *Repositorio
	Reloj       func() time.Time
}

// FuenteAsistencia expone marcaciones y asignaciones sin consultar sus tablas.
type FuenteAsistencia interface {
	MarcacionesDelRango(ctx context.Context, desde, hasta time.Time) ([]EventoMarcacion, error)
	ContarTardanzas(ctx context.Context, alumnoID int64, desde, hasta time.Time) (int, error)
	TieneMarcacion(ctx context.Context, alumnoID int64, fecha time.Time) (bool, error)
	DocentesDeSeccionEnFecha(ctx context.Context, seccionID int64, fecha time.Time) ([]int64, error)
}

// FuenteAlumnos expone alumnos activos y su sección en una fecha.
type FuenteAlumnos interface {
	AlumnosActivosDeSeccionEnFecha(ctx context.Context, seccionID int64, fecha time.Time) ([]AlumnoEvento, error)
	SeccionDeAlumnoEnFecha(ctx context.Context, alumnoID int64, fecha time.Time) (int64, error)
	AlumnoPorID(ctx context.Context, alumnoID int64) (AlumnoEvento, error)
}

// FuenteAcademica expone secciones, calendario y umbrales.
type FuenteAcademica interface {
	SeccionesActivas(ctx context.Context) ([]academico.Seccion, error)
	EsDiaNoLaborable(ctx context.Context, seccionID int64, fecha time.Time) (bool, error)
	UmbralVigente(ctx context.Context, tipo string, fecha time.Time) (int, bool, error)
	EtiquetaSeccion(ctx context.Context, seccionID int64) (string, error)
}

// FuenteUsuarios expone el administrador de respaldo.
type FuenteUsuarios interface {
	AdministradorActivo(ctx context.Context) (int64, error)
}
