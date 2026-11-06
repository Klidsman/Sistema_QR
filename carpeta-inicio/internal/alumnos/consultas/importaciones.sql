-- name: CrearImportacion :one
INSERT INTO importaciones (
    usuario_id,
    archivo_nombre,
    filas_leidas,
    filas_creadas,
    filas_rechazadas
)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, creada_en;

-- name: CrearImportacionFila :exec
INSERT INTO importacion_filas (
    importacion_id,
    numero_fila,
    identificador,
    motivo
)
VALUES ($1, $2, $3, $4);

-- name: ListarImportaciones :many
SELECT id,
       usuario_id,
       archivo_nombre,
       filas_leidas,
       filas_creadas,
       filas_rechazadas,
       creada_en
FROM importaciones
ORDER BY creada_en DESC, id DESC;

-- name: BuscarImportacionPorID :one
SELECT id,
       usuario_id,
       archivo_nombre,
       filas_leidas,
       filas_creadas,
       filas_rechazadas,
       creada_en
FROM importaciones
WHERE id = $1;

-- name: ListarImportacionFilas :many
SELECT numero_fila,
       identificador,
       motivo
FROM importacion_filas
WHERE importacion_id = $1
ORDER BY numero_fila;
