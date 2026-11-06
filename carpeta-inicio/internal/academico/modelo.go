package academico

import "time"

// Estados del año escolar.
const (
	EstadoAnioActivo  = "activo"
	EstadoAnioCerrado = "cerrado"
)

// AnioEscolar es un periodo lectivo con fecha de inicio y fin.
type AnioEscolar struct {
	ID          int64
	Nombre      string
	FechaInicio time.Time
	FechaFin    time.Time
	Estado      string
	CerradoEn   *time.Time
}

// Grado es un nivel dentro del año escolar.
type Grado struct {
	ID            int64
	AnioEscolarID int64
	Nombre        string
}

// Seccion es una sección de un grado con los datos que la interfaz necesita para mostrarla.
type Seccion struct {
	ID            int64
	Nombre        string
	Turno         string
	GradoID       int64
	Grado         string
	AnioEscolarID int64
	AnioEscolar   string
}

// Etiqueta devuelve el texto que se muestra en los formularios.
func (s Seccion) Etiqueta() string {
	return s.Grado + " " + s.Nombre + " (" + s.AnioEscolar + ")"
}

// TurnoSeccion es una versión del turno de una sección desde una fecha.
type TurnoSeccion struct {
	ID           int64
	SeccionID    int64
	Turno        string
	VigenteDesde time.Time
}

// HorarioTurno es una versión del horario de un turno desde una fecha.
type HorarioTurno struct {
	ID                 int64
	Turno              string
	HoraIngreso        time.Time
	ToleranciaMinutos  int
	HoraLimiteTardanza time.Time
	HoraCierre         time.Time
	VigenteDesde       time.Time
}

// ExcepcionCalendario es un feriado, día no laborable o suspensión.
type ExcepcionCalendario struct {
	ID            int64
	AnioEscolarID int64
	Tipo          string
	FechaInicio   time.Time
	FechaFin      time.Time
	Motivo        string
	Grados        []int64
	CreadoPor     int64
	CreadoEn      time.Time
}

// UmbralAlerta es una versión del umbral mensual de tardanzas o faltas.
type UmbralAlerta struct {
	ID            int64
	AnioEscolarID int64
	Tipo          string
	Valor         int
	VigenteDesde  time.Time
}

// Curso es una asignatura del año escolar vinculada a uno o más grados.
type Curso struct {
	ID            int64
	AnioEscolarID int64
	Nombre        string
	Activo        bool
	Grados        []int64
	CreadoEn      time.Time
}

// BloqueClase es una versión de un bloque de hora de clase de un turno.
type BloqueClase struct {
	ID           int64
	Turno        string
	Numero       int
	HoraInicio   time.Time
	HoraFin      time.Time
	VigenteDesde time.Time
}

// HorarioSeccion es una celda del horario semanal de una sección con su vigencia.
type HorarioSeccion struct {
	ID           int64
	SeccionID    int64
	DiaSemana    int
	BloqueNumero int
	CursoID      *int64
	VigenteDesde time.Time
}

// ClaseProgramada es una hora de clase válida para una fecha, sección, curso y bloque.
type ClaseProgramada struct {
	SeccionID  int64
	CursoID    int64
	Bloque     int
	Turno      string
	HoraInicio time.Time
	HoraFin    time.Time
	Seccion    string
	Curso      string
}

// ConfiguracionFecha reúne la configuración vigente en una fecha determinada.
type ConfiguracionFecha struct {
	Fecha       time.Time
	Anio        *AnioEscolar
	Horarios    []HorarioTurno
	Excepciones []ExcepcionCalendario
	Umbrales    []UmbralAlerta
	Secciones   []Seccion
	Bloques     []BloqueClase
}
