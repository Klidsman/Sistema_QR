package porteria

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"asistencia/internal/academico"
	"asistencia/internal/alumnos"
	"asistencia/internal/asistencia"
	"asistencia/internal/plataforma/fechas"

	"github.com/jackc/pgx/v5"
)

// ClasificarEstado aplica la tolerancia y el límite de tardanza del turno.
func ClasificarEstado(hora time.Time, horario academico.HorarioTurno) (string, error) {
	if fechas.CompararHoras(hora, horario.HoraIngreso) < 0 {
		return "", ErrAntesDeIngreso
	}
	if fechas.CompararHoras(hora, horario.HoraCierre) > 0 {
		return "", ErrRegistroCerrado
	}
	finTolerancia := fechas.SumarMinutos(horario.HoraIngreso, horario.ToleranciaMinutos)
	if fechas.CompararHoras(hora, finTolerancia) <= 0 {
		return asistencia.MarcacionPuntual, nil
	}
	if fechas.CompararHoras(hora, horario.HoraLimiteTardanza) <= 0 {
		return asistencia.MarcacionTardanza, nil
	}
	return asistencia.MarcacionTardanzaGrave, nil
}

// Registrar procesa un escaneo y crea la marcación oficial.
func (s *Servicio) Registrar(ctx context.Context, usuarioID int64, token string) (Resultado, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return Resultado{}, ErrTokenInvalido
	}

	ahora := s.dependencias.Reloj()
	fecha := fechas.Fecha(ahora)
	hora, err := fechas.ParsearHora(fechas.Hora(ahora))
	if err != nil {
		return Resultado{}, fmt.Errorf("interpretando la hora del escaneo: %w", err)
	}

	alumno, err := s.dependencias.Alumnos.AlumnoPorToken(ctx, token, fecha)
	if errors.Is(err, alumnos.ErrAlumnoNoEncontrado) {
		return Resultado{}, ErrTokenInvalido
	}
	if err != nil {
		return Resultado{}, err
	}
	if alumno.SeccionID == 0 {
		return Resultado{}, ErrSinSeccion
	}

	if fecha.Weekday() == time.Saturday || fecha.Weekday() == time.Sunday {
		return Resultado{}, ErrDiaNoLaborable
	}
	noLaborable, err := s.dependencias.Academico.EsDiaNoLaborable(ctx, alumno.SeccionID, fecha)
	if err != nil {
		return Resultado{}, err
	}
	if noLaborable {
		return Resultado{}, ErrDiaNoLaborable
	}

	turno, err := s.dependencias.Academico.TurnoDeSeccion(ctx, alumno.SeccionID, fecha)
	if errors.Is(err, pgx.ErrNoRows) {
		return Resultado{}, ErrSinTurno
	}
	if err != nil {
		return Resultado{}, err
	}
	horario, err := s.dependencias.Academico.HorarioVigente(ctx, turno, fecha)
	if errors.Is(err, pgx.ErrNoRows) {
		return Resultado{}, ErrSinHorario
	}
	if err != nil {
		return Resultado{}, err
	}

	estado, err := ClasificarEstado(hora, horario)
	if err != nil {
		return Resultado{}, err
	}

	_, creado, err := s.dependencias.Marcaciones.RegistrarMarcacionPorteria(
		ctx, usuarioID, alumno.ID, fecha, ahora, estado)
	if err != nil {
		return Resultado{}, err
	}
	if !creado {
		return Resultado{}, ErrIngresoDuplicado
	}

	if estado != asistencia.MarcacionPuntual && s.dependencias.Alertas != nil {
		// La alerta es interna: si falla, el ingreso oficial ya quedó registrado.
		_ = s.dependencias.Alertas.ProcesarTardanzas(ctx)
	}

	seccion, err := s.dependencias.Academico.EtiquetaSeccion(ctx, alumno.SeccionID)
	if err != nil {
		seccion = ""
	}
	return Resultado{
		Registrado:    true,
		Alumno:        alumno.NombreCompleto(),
		Identificador: alumno.Identificador,
		Seccion:       seccion,
		Hora:          fechas.Hora(ahora),
		Estado:        EtiquetaEstado(estado),
	}, nil
}

// EtiquetaEstado devuelve el texto visible del estado de la marcación.
func EtiquetaEstado(estado string) string {
	switch estado {
	case asistencia.MarcacionPuntual:
		return "Puntual"
	case asistencia.MarcacionTardanza:
		return "Tardanza"
	case asistencia.MarcacionTardanzaGrave:
		return "Tardanza grave"
	default:
		return estado
	}
}
