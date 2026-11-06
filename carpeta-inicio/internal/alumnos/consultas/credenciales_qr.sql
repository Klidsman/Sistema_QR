-- name: CrearCredencialQR :one
INSERT INTO credenciales_qr (alumno_id, token)
VALUES ($1, $2)
RETURNING id;

-- name: BuscarCredencialActiva :one
SELECT id, alumno_id, token, estado, emitida_en
FROM credenciales_qr
WHERE alumno_id = $1
  AND estado = 'activa';

-- name: RevocarCredencialActiva :exec
UPDATE credenciales_qr
SET estado = 'revocada',
    motivo_revocacion = $2,
    revocada_en = now()
WHERE alumno_id = $1
  AND estado = 'activa';

-- name: ListarCredenciales :many
SELECT id, token, estado, motivo_revocacion, emitida_en, revocada_en
FROM credenciales_qr
WHERE alumno_id = $1
ORDER BY emitida_en DESC;
