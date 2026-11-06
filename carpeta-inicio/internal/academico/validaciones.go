package academico

import (
	"errors"
	"strings"
	"time"

	"asistencia/internal/plataforma/fechas"
)

// Errores de validación de la configuración académica.
var (
	ErrFechasAnioInvalidas   = errors.New("la fecha de fin debe ser posterior a la fecha de inicio")
	ErrVigenciaPasada        = errors.New("la fecha de vigencia debe ser hoy o una fecha futura")
	ErrToleranciaInvalida    = errors.New("la tolerancia debe estar entre 0 y 120 minutos")
	ErrCierreNoPosterior     = errors.New("la hora de cierre debe ser posterior a la hora límite de tardanza")
	ErrToleranciaExcedida    = errors.New("la hora de ingreso más la tolerancia no puede superar la hora límite de tardanza")
	ErrHorarioFueraDeOrden   = errors.New("las horas del turno deben ir de ingreso a límite de tardanza y a cierre")
	ErrCruzaMedianoche       = errors.New("los horarios deben estar dentro del mismo día")
	ErrHorariosSuperpuestos  = errors.New("los horarios de mañana y tarde se superponen")
	ErrBloqueNumeroInvalido  = errors.New("el número de bloque debe estar entre 1 y 20")
	ErrBloqueHorasInvalidas  = errors.New("la hora de fin del bloque debe ser posterior a la de inicio")
	ErrBloqueFueraDeHorario  = errors.New("el bloque debe estar dentro del horario del turno")
	ErrBloqueSolapado        = errors.New("el bloque se superpone con otro bloque del mismo turno")
	ErrDiaFueraDeRango       = errors.New("el día de la semana debe estar entre lunes y viernes")
	ErrValorUmbralInvalido   = errors.New("el valor del umbral debe ser un entero mayor que cero")
	ErrTipoInvalido          = errors.New("el tipo indicado no es válido")
	ErrMotivoExcepcionCorto  = errors.New("el motivo debe tener entre 3 y 200 caracteres")
	ErrRangoFechasExcepcion  = errors.New("la fecha de fin debe ser igual o posterior a la de inicio")
	ErrNombreInvalido        = errors.New("el nombre debe tener entre 2 y 80 caracteres")
	ErrNombreSeccionInvalido = errors.New("el nombre de la sección debe tener entre 1 y 10 caracteres")
	ErrTurnoInvalido         = errors.New("el turno debe ser mañana o tarde")
)

// NormalizarNombre recorta los extremos y colapsa los espacios internos.
func NormalizarNombre(nombre string) string {
	return strings.Join(strings.Fields(nombre), " ")
}

// ValidarNombre exige un nombre de 2 a 80 caracteres después de normalizarlo.
func ValidarNombre(nombre string) error {
	normalizado := NormalizarNombre(nombre)
	longitud := len([]rune(normalizado))
	if longitud < 2 || longitud > 80 {
		return ErrNombreInvalido
	}
	return nil
}

// ValidarNombreSeccion exige un nombre corto para la sección (por ejemplo A, B o C).
func ValidarNombreSeccion(nombre string) error {
	normalizado := NormalizarNombre(nombre)
	longitud := len([]rune(normalizado))
	if longitud < 1 || longitud > 10 {
		return ErrNombreSeccionInvalido
	}
	return nil
}

// ValidarTurno exige uno de los dos turnos válidos.
func ValidarTurno(turno string) error {
	if !TurnoValido(turno) {
		return ErrTurnoInvalido
	}
	return nil
}

// ValidarVigencia rechaza fechas de vigencia anteriores a hoy.
func ValidarVigencia(vigenteDesde, hoy time.Time) error {
	if fechaAnterior(vigenteDesde, hoy) {
		return ErrVigenciaPasada
	}
	return nil
}

// ValidarFechasAnio exige que la fecha de fin sea posterior a la de inicio.
func ValidarFechasAnio(inicio, fin time.Time) error {
	if !fin.After(inicio) {
		return ErrFechasAnioInvalidas
	}
	return nil
}

// ValidarHorarioTurno aplica las reglas de orden, tolerancia y mismo día.
func ValidarHorarioTurno(horario HorarioTurno) error {
	if horario.ToleranciaMinutos < 0 || horario.ToleranciaMinutos > 120 {
		return ErrToleranciaInvalida
	}
	if fechas.CompararHoras(horario.HoraIngreso, horario.HoraLimiteTardanza) >= 0 {
		return ErrHorarioFueraDeOrden
	}
	if fechas.CompararHoras(horario.HoraLimiteTardanza, horario.HoraCierre) >= 0 {
		return ErrCierreNoPosterior
	}
	finTolerancia := fechas.SumarMinutos(horario.HoraIngreso, horario.ToleranciaMinutos)
	if noEsMismoDia(horario.HoraIngreso, finTolerancia) {
		return ErrCruzaMedianoche
	}
	if fechas.CompararHoras(finTolerancia, horario.HoraLimiteTardanza) > 0 {
		return ErrToleranciaExcedida
	}
	return nil
}

// ValidarSuperposicionHorarios rechaza dos turnos cuyas ventanas de tiempo se cruzan.
func ValidarSuperposicionHorarios(primero, segundo HorarioTurno) error {
	empiezaAntes := fechas.CompararHoras(primero.HoraIngreso, segundo.HoraCierre) < 0
	terminaDespues := fechas.CompararHoras(segundo.HoraIngreso, primero.HoraCierre) < 0
	if empiezaAntes && terminaDespues {
		return ErrHorariosSuperpuestos
	}
	return nil
}

// ValidarBloque aplica las reglas de número, orden, horario del turno y solapamiento.
func ValidarBloque(bloque BloqueClase, horario HorarioTurno, otros []BloqueClase) error {
	if bloque.Numero < 1 || bloque.Numero > 20 {
		return ErrBloqueNumeroInvalido
	}
	if noEsMismoDia(bloque.HoraInicio, bloque.HoraFin) {
		return ErrCruzaMedianoche
	}
	if fechas.CompararHoras(bloque.HoraInicio, bloque.HoraFin) >= 0 {
		return ErrBloqueHorasInvalidas
	}
	if fechas.CompararHoras(bloque.HoraInicio, horario.HoraIngreso) < 0 ||
		fechas.CompararHoras(bloque.HoraFin, horario.HoraCierre) > 0 {
		return ErrBloqueFueraDeHorario
	}
	for _, otro := range otros {
		if otro.Turno != bloque.Turno || otro.Numero == bloque.Numero {
			continue
		}
		empiezaAntes := fechas.CompararHoras(bloque.HoraInicio, otro.HoraFin) < 0
		terminaDespues := fechas.CompararHoras(otro.HoraInicio, bloque.HoraFin) < 0
		if empiezaAntes && terminaDespues {
			return ErrBloqueSolapado
		}
	}
	return nil
}

// ValidarDiaSemana exige un día laborable entre lunes y viernes.
func ValidarDiaSemana(dia int) error {
	if dia < 1 || dia > 5 {
		return ErrDiaFueraDeRango
	}
	return nil
}

// ValidarUmbral exige un valor entero mayor que cero y un tipo conocido.
func ValidarUmbral(tipo string, valor int) error {
	if !TipoUmbralValido(tipo) {
		return ErrTipoInvalido
	}
	if valor <= 0 {
		return ErrValorUmbralInvalido
	}
	return nil
}

// ValidarExcepcion valida el motivo y el rango de fechas de una excepción.
func ValidarExcepcion(tipo, motivo string, fechaInicio, fechaFin time.Time) error {
	if !TipoExcepcionValido(tipo) {
		return ErrTipoInvalido
	}
	if len([]rune(motivo)) < 3 || len([]rune(motivo)) > 200 {
		return ErrMotivoExcepcionCorto
	}
	if fechaFin.Before(fechaInicio) {
		return ErrRangoFechasExcepcion
	}
	return nil
}

func fechaAnterior(primera, segunda time.Time) bool {
	if primera.Year() != segunda.Year() {
		return primera.Year() < segunda.Year()
	}
	if primera.Month() != segunda.Month() {
		return primera.Month() < segunda.Month()
	}
	return primera.Day() < segunda.Day()
}

func noEsMismoDia(primera, segunda time.Time) bool {
	return primera.Year() != segunda.Year() ||
		primera.Month() != segunda.Month() ||
		primera.Day() != segunda.Day()
}
