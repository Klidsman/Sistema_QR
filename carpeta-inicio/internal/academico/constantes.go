package academico

import "time"

// Turnos fijos del colegio. No se crean ni se eliminan.
const (
	TurnoManana = "manana"
	TurnoTarde  = "tarde"
)

// Tipos de excepción de calendario.
const (
	TipoFeriado     = "feriado"
	TipoNoLaborable = "no_laborable"
	TipoSuspension  = "suspension"
)

// Tipos de umbral de alerta.
const (
	TipoUmbralTardanzas      = "tardanzas"
	TipoUmbralFaltas         = "faltas"
	TipoUmbralFaltasSeguidas = "faltas_seguidas"
)

// Turnos devuelve los dos turnos válidos en el orden en que se muestran.
func Turnos() []string {
	return []string{TurnoManana, TurnoTarde}
}

// TurnoValido indica si el texto corresponde a uno de los dos turnos.
func TurnoValido(turno string) bool {
	return turno == TurnoManana || turno == TurnoTarde
}

// TiposExcepcion devuelve los tipos de excepción de calendario válidos.
func TiposExcepcion() []string {
	return []string{TipoFeriado, TipoNoLaborable, TipoSuspension}
}

// TipoExcepcionValido indica si el texto corresponde a un tipo de excepción.
func TipoExcepcionValido(tipo string) bool {
	return tipo == TipoFeriado || tipo == TipoNoLaborable || tipo == TipoSuspension
}

// TiposUmbral devuelve los tipos de umbral de alerta válidos.
func TiposUmbral() []string {
	return []string{TipoUmbralTardanzas, TipoUmbralFaltas, TipoUmbralFaltasSeguidas}
}

// TipoUmbralValido indica si el texto corresponde a un tipo de umbral.
func TipoUmbralValido(tipo string) bool {
	return tipo == TipoUmbralTardanzas || tipo == TipoUmbralFaltas || tipo == TipoUmbralFaltasSeguidas
}

// DiaLaborable indica si un día de la semana genera clases por defecto.
func DiaLaborable(dia time.Weekday) bool {
	return dia >= time.Monday && dia <= time.Friday
}
