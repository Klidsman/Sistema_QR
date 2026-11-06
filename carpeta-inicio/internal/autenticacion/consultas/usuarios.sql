-- name: BuscarUsuarioPorCorreo :one
SELECT id, correo, contrasenia_hash, nombre_completo, rol, activo
FROM usuarios
WHERE correo = $1;

-- name: BuscarUsuarioPorID :one
SELECT id, correo, nombre_completo, rol, activo
FROM usuarios
WHERE id = $1;

-- name: BuscarAdministradorActivo :one
SELECT id
FROM usuarios
WHERE rol = 'administrador'
  AND activo
ORDER BY id
LIMIT 1;

-- name: ListarDocentes :many
SELECT id, correo, nombre_completo, rol, activo
FROM usuarios
WHERE rol = 'docente'
ORDER BY nombre_completo;

-- name: DesactivarDocente :execrows
UPDATE usuarios
SET activo = false
WHERE id = $1
  AND rol = 'docente'
  AND activo;

-- name: ActualizarClaveDocente :execrows
UPDATE usuarios
SET contrasenia_hash = $2
WHERE id = $1
  AND rol = 'docente';

-- name: CrearUsuario :one
INSERT INTO usuarios (correo, contrasenia_hash, nombre_completo, rol)
VALUES ($1, $2, $3, $4)
RETURNING id;
