-- name: InsertarAuditoria :exec
INSERT INTO auditoria (usuario_id, entidad, entidad_id, accion, valor_anterior, valor_nuevo, motivo)
VALUES ($1, $2, $3, $4, $5, $6, $7);
