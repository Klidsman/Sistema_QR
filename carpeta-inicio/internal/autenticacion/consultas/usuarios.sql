-- name: BuscarUsuarioPorCorreo :one
SELECT id, correo, contrasenia_hash, nombre_completo, rol, activo
FROM usuarios
WHERE correo = $1;

-- name: BuscarUsuarioPorID :one
SELECT id, correo, nombre_completo, rol, activo
FROM usuarios
WHERE id = $1;

-- name: CrearUsuario :one
INSERT INTO usuarios (correo, contrasenia_hash, nombre_completo, rol)
VALUES ($1, $2, $3, $4)
RETURNING id;
