-- name: BuscarSesion :one
SELECT datos, expira_en
FROM sesiones
WHERE token = $1;

-- name: GuardarSesion :exec
INSERT INTO sesiones (token, datos, expira_en)
VALUES ($1, $2, $3)
ON CONFLICT (token) DO UPDATE
SET datos = EXCLUDED.datos,
    expira_en = EXCLUDED.expira_en;

-- name: EliminarSesion :exec
DELETE FROM sesiones
WHERE token = $1;
