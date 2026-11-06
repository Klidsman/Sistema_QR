// Package porteria registra la asistencia oficial del alumno a partir del QR.
package porteria

import (
	"context"
	"errors"
	"time"

	"asistencia/internal/academico"
	"asistencia/internal/alumnos"
)

// Errores de validación del escaneo.
var (
	ErrTokenInvalido    = errors.New("el código QR no es válido")
	ErrDiaNoLaborable   = errors.New("hoy no es un día laborable para la sección")
	ErrAntesDeIngreso   = errors.New("todavía no comienza el registro del turno")
	ErrRegistroCerrado  = errors.New("el registro del turno ya cerró")
	ErrSinTurno         = errors.New("la sección no tiene un turno vigente")
	ErrSinHorario       = errors.New("el turno no tiene un horario configurado")
	ErrSinSeccion       = errors.New("el alumno no tiene una sección vigente")
	ErrIngresoDuplicado = errors.New("el ingreso del día ya fue registrado")
)

// Resultado describe el desenlace de un escaneo.
type Resultado struct {
	Registrado    bool
	Alumno        string
	Identificador string
	Seccion       string
	Hora          string
	Estado        string
	Motivo        string
}

// FuenteAlumnos resuelve el token del QR.
type FuenteAlumnos interface {
	AlumnoPorToken(ctx context.Context, token string, fecha time.Time) (alumnos.Alumno, error)
}

// FuenteAcademica aporta turno, horario y calendario.
type FuenteAcademica interface {
	TurnoDeSeccion(ctx context.Context, seccionID int64, fecha time.Time) (string, error)
	HorarioVigente(ctx context.Context, turno string, fecha time.Time) (academico.HorarioTurno, error)
	EsDiaNoLaborable(ctx context.Context, seccionID int64, fecha time.Time) (bool, error)
	EtiquetaSeccion(ctx context.Context, seccionID int64) (string, error)
}

// RegistradorMarcaciones crea la marcación oficial dentro del módulo de asistencia.
type RegistradorMarcaciones interface {
	RegistrarMarcacionPorteria(ctx context.Context, usuarioID, alumnoID int64, fecha, hora time.Time, estado string) (int64, bool, error)
}

// ProcesadorAlertas genera las alertas de tardanza sin esperar al cron.
type ProcesadorAlertas interface {
	ProcesarTardanzas(ctx context.Context) error
}

// Dependencias reúne lo que necesita el servicio de portería.
type Dependencias struct {
	Alumnos     FuenteAlumnos
	Academico   FuenteAcademica
	Marcaciones RegistradorMarcaciones
	Alertas     ProcesadorAlertas
	Reloj       func() time.Time
}

// Servicio aplica las reglas de registro de ingreso por QR.
type Servicio struct {
	dependencias Dependencias
}

// NuevoServicio crea el servicio de portería.
func NuevoServicio(dependencias Dependencias) *Servicio {
	if dependencias.Reloj == nil {
		dependencias.Reloj = time.Now
	}
	return &Servicio{dependencias: dependencias}
}
