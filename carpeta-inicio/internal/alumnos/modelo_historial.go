package alumnos

import "time"

// HistorialSeccion es un tramo de permanencia de un alumno en una sección.
type HistorialSeccion struct {
	ID          int64
	SeccionID   int64
	Grado       string
	Seccion     string
	AnioEscolar string
	FechaInicio time.Time
	FechaFin    *time.Time
}

// Actual indica si es el tramo vigente del alumno.
func (h HistorialSeccion) Actual() bool {
	return h.FechaFin == nil
}

// CambioSeccion describe el cierre del tramo actual y la apertura del nuevo.
type CambioSeccion struct {
	SeccionAnterior int64
	SeccionNueva    int64
	FechaCierre     time.Time
	FechaApertura   time.Time
}

// PrepararCambioSeccion valida el cambio y calcula el cierre del tramo actual y la
// apertura del nuevo. actual puede ser nil cuando el alumno no tiene historial previo.
func PrepararCambioSeccion(actual *HistorialSeccion, nuevaSeccionID int64, fecha time.Time) (CambioSeccion, error) {
	if nuevaSeccionID <= 0 {
		return CambioSeccion{}, ErrSeccionInvalida
	}
	if fecha.IsZero() {
		return CambioSeccion{}, ErrFechaInvalida
	}

	cambio := CambioSeccion{
		SeccionNueva:  nuevaSeccionID,
		FechaCierre:   fecha,
		FechaApertura: fecha,
	}
	if actual == nil {
		return cambio, nil
	}

	cambio.SeccionAnterior = actual.SeccionID
	if actual.FechaFin != nil {
		return CambioSeccion{}, ErrHistorialYaCerrado
	}
	if actual.SeccionID == nuevaSeccionID {
		return CambioSeccion{}, ErrMismaSeccion
	}
	if fecha.Before(actual.FechaInicio) {
		return CambioSeccion{}, ErrFechaInvalida
	}
	return cambio, nil
}

// TipoSalidaValido indica si el tipo corresponde a una salida definitiva del alumno.
func TipoSalidaValido(tipo string) bool {
	return tipo == EstadoRetirado || tipo == EstadoTrasladado
}
