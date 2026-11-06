package asistencia

import (
	"context"
	"time"
)

// MarcacionDelDia expone la marcación oficial para el módulo de justificaciones.
func (s *Servicio) MarcacionDelDia(ctx context.Context, alumnoID int64, fecha time.Time) (Marcacion, bool, error) {
	return s.repositorio.MarcacionDeAlumnoFecha(ctx, alumnoID, fecha)
}

// MarcarTardanzaJustificada cambia la marca de justificación de una marcación.
func (s *Servicio) MarcarTardanzaJustificada(ctx context.Context, alumnoID int64, fecha time.Time, justificada bool) error {
	return s.repositorio.MarcarMarcacionJustificada(ctx, alumnoID, fecha, justificada)
}

// DocenteAsignadoASeccionEnFecha indica si el docente puede justificar esa sección.
func (s *Servicio) DocenteAsignadoASeccionEnFecha(ctx context.Context, docenteID, seccionID int64, fecha time.Time) (bool, error) {
	return s.repositorio.DocenteAsignadoASeccionFecha(ctx, docenteID, seccionID, fecha)
}
