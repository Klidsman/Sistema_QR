package asistencia

import (
	"context"

	"asistencia/internal/asistencia/generado"

	"github.com/jackc/pgx/v5/pgtype"
)

// ObtenerParametros devuelve el plazo de edición vigente.
func (r *Repositorio) ObtenerParametros(ctx context.Context) (ParametrosAsistencia, error) {
	fila, err := r.consultas.ObtenerParametrosAsistencia(ctx)
	if err != nil {
		return ParametrosAsistencia{}, err
	}
	parametros := ParametrosAsistencia{
		PlazoEdicionHoras: int(fila.PlazoEdicionHoras),
		ActualizadoEn:     fila.ActualizadoEn.Time,
	}
	if fila.ActualizadoPor.Valid {
		actualizadoPor := fila.ActualizadoPor.Int64
		parametros.ActualizadoPor = &actualizadoPor
	}
	return parametros, nil
}

// ActualizarPlazo guarda el plazo de edición.
func (r *Repositorio) ActualizarPlazo(ctx context.Context, horas int, usuarioID int64) (ParametrosAsistencia, error) {
	fila, err := r.consultas.ActualizarPlazoEdicion(ctx, generado.ActualizarPlazoEdicionParams{
		PlazoEdicionHoras: int32(horas),
		ActualizadoPor:    pgtype.Int8{Int64: usuarioID, Valid: true},
	})
	if err != nil {
		return ParametrosAsistencia{}, err
	}
	parametros := ParametrosAsistencia{
		PlazoEdicionHoras: int(fila.PlazoEdicionHoras),
		ActualizadoEn:     fila.ActualizadoEn.Time,
	}
	if fila.ActualizadoPor.Valid {
		actualizadoPor := fila.ActualizadoPor.Int64
		parametros.ActualizadoPor = &actualizadoPor
	}
	return parametros, nil
}
