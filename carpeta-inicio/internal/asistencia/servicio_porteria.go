package asistencia

import (
	"context"
	"errors"
	"fmt"
	"time"

	"asistencia/internal/asistencia/generado"
	"asistencia/internal/plataforma/auditoria"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// RegistrarMarcacionPorteria crea la marcación oficial del ingreso por QR.
// Devuelve falso si el alumno ya tenía marcación ese día.
func (s *Servicio) RegistrarMarcacionPorteria(
	ctx context.Context,
	usuarioID, alumnoID int64,
	fecha, hora time.Time,
	estado string,
) (int64, bool, error) {
	var marcacionID int64
	err := s.repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		id, err := consultas.CrearMarcacionPorteria(ctx, generado.CrearMarcacionPorteriaParams{
			AlumnoID:      alumnoID,
			Fecha:         pgtype.Date{Time: fecha, Valid: true},
			Hora:          pgtype.Timestamptz{Time: hora, Valid: true},
			Estado:        estado,
			RegistradaPor: pgtype.Int8{Int64: usuarioID, Valid: true},
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("registrando la marcación de portería: %w", err)
		}
		marcacionID = id
		return auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID: usuarioID,
			Entidad:   EntidadMarcacion,
			EntidadID: id,
			Accion:    AccionRegistrarMarcacion,
			ValorNuevo: map[string]any{
				"alumno_id":      alumnoID,
				"fecha":          fecha.Format("2006-01-02"),
				"hora":           hora.Format(time.RFC3339),
				"estado":         estado,
				"registrada_por": usuarioID,
			},
		})
	})
	if err != nil {
		return 0, false, err
	}
	return marcacionID, marcacionID > 0, nil
}
