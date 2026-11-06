// Package auditoria registra los cambios sobre datos sensibles en la tabla auditoria.
package auditoria

import (
	"context"
	"encoding/json"
	"fmt"

	"asistencia/internal/plataforma/auditoria/generado"

	"github.com/jackc/pgx/v5/pgtype"
)

// Evento describe un cambio que debe quedar auditado.
type Evento struct {
	UsuarioID     int64
	Entidad       string
	EntidadID     int64
	Accion        string
	ValorAnterior any
	ValorNuevo    any
	Motivo        string
}

// Registrar guarda el evento usando la conexión o la transacción indicada.
// Debe llamarse dentro de la misma transacción que el cambio auditado.
func Registrar(ctx context.Context, db generado.DBTX, evento Evento) error {
	anterior, err := serializar(evento.ValorAnterior)
	if err != nil {
		return err
	}
	nuevo, err := serializar(evento.ValorNuevo)
	if err != nil {
		return err
	}

	err = generado.New(db).InsertarAuditoria(ctx, generado.InsertarAuditoriaParams{
		UsuarioID:     evento.UsuarioID,
		Entidad:       evento.Entidad,
		EntidadID:     evento.EntidadID,
		Accion:        evento.Accion,
		ValorAnterior: anterior,
		ValorNuevo:    nuevo,
		Motivo:        pgtype.Text{String: evento.Motivo, Valid: evento.Motivo != ""},
	})
	if err != nil {
		return fmt.Errorf("registrando la auditoría de %s %d: %w", evento.Entidad, evento.EntidadID, err)
	}
	return nil
}

func serializar(valor any) ([]byte, error) {
	if valor == nil {
		return nil, nil
	}
	datos, err := json.Marshal(valor)
	if err != nil {
		return nil, fmt.Errorf("serializando el valor de auditoría: %w", err)
	}
	return datos, nil
}
