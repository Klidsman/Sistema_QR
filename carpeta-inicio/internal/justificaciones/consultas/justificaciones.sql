-- name: InsertarJustificacion :one
INSERT INTO justificaciones (alumno_id, seccion_id, motivo, registrante_id)
VALUES ($1, $2, $3, $4)
RETURNING id;

-- name: InsertarJustificacionFecha :exec
INSERT INTO justificacion_fechas (justificacion_id, alumno_id, fecha_evento, tipo, estado)
VALUES ($1, $2, $3, $4, 'pendiente');

-- name: InsertarAdjuntoJustificacion :exec
INSERT INTO justificacion_adjuntos (justificacion_id, nombre, tipo, tamano, clave)
VALUES ($1, $2, $3, $4, $5);

-- name: ObtenerJustificacion :one
SELECT id, alumno_id, seccion_id, motivo, estado, registrante_id, resuelto_por,
       resuelto_en, COALESCE(comentario, '') AS comentario,
       anulado_por, anulado_en, COALESCE(motivo_anulacion, '') AS motivo_anulacion,
       creada_en
FROM justificaciones
WHERE id = $1;

-- name: ListarFechasDeJustificacion :many
SELECT id, justificacion_id, alumno_id, fecha_evento, tipo, estado
FROM justificacion_fechas
WHERE justificacion_id = $1
ORDER BY fecha_evento;

-- name: ListarFechasDeJustificaciones :many
SELECT id, justificacion_id, alumno_id, fecha_evento, tipo, estado
FROM justificacion_fechas
WHERE justificacion_id = ANY($1::bigint[])
ORDER BY fecha_evento;

-- name: ListarJustificacionesDeAlumno :many
SELECT id, alumno_id, seccion_id, motivo, estado, registrante_id, resuelto_por,
       resuelto_en, COALESCE(comentario, '') AS comentario,
       anulado_por, anulado_en, COALESCE(motivo_anulacion, '') AS motivo_anulacion,
       creada_en
FROM justificaciones
WHERE alumno_id = $1
ORDER BY creada_en DESC;

-- name: ListarJustificacionesPendientes :many
SELECT id, alumno_id, seccion_id, motivo, estado, registrante_id, resuelto_por,
       resuelto_en, COALESCE(comentario, '') AS comentario,
       anulado_por, anulado_en, COALESCE(motivo_anulacion, '') AS motivo_anulacion,
       creada_en
FROM justificaciones
WHERE estado = 'pendiente'
ORDER BY creada_en;

-- name: ListarJustificacionesDelRegistrante :many
SELECT id, alumno_id, seccion_id, motivo, estado, registrante_id, resuelto_por,
       resuelto_en, COALESCE(comentario, '') AS comentario,
       anulado_por, anulado_en, COALESCE(motivo_anulacion, '') AS motivo_anulacion,
       creada_en
FROM justificaciones
WHERE registrante_id = $1
ORDER BY creada_en DESC;

-- name: ResolverJustificacion :execrows
UPDATE justificaciones
SET estado = $2,
    resuelto_por = $3,
    resuelto_en = now(),
    comentario = $4
WHERE id = $1
  AND estado = 'pendiente';

-- name: ActualizarEstadoFechas :exec
UPDATE justificacion_fechas
SET estado = $2
WHERE justificacion_id = $1;

-- name: ActualizarEstadoFechasPorAlumnoFecha :exec
UPDATE justificacion_fechas
SET estado = $4
WHERE alumno_id = $1
  AND fecha_evento = $2
  AND estado = $3;

-- name: AnularJustificacion :execrows
UPDATE justificaciones
SET estado = 'anulada',
    anulado_por = $2,
    anulado_en = now(),
    motivo_anulacion = $3
WHERE id = $1
  AND estado = 'aprobada';

-- name: ExisteJustificacionActiva :one
SELECT EXISTS (
    SELECT 1 FROM justificacion_fechas
    WHERE alumno_id = $1
      AND fecha_evento = $2
      AND estado IN ('pendiente', 'aprobada')
);

-- name: EsFaltaJustificada :one
SELECT EXISTS (
    SELECT 1 FROM justificacion_fechas
    WHERE alumno_id = $1
      AND fecha_evento = $2
      AND estado = 'aprobada'
);

-- name: ListarAdjuntosDeJustificacion :many
SELECT id, justificacion_id, nombre, tipo, tamano, clave, creado_en
FROM justificacion_adjuntos
WHERE justificacion_id = $1
ORDER BY id;

-- name: ObtenerAdjunto :one
SELECT id, justificacion_id, nombre, tipo, tamano, clave, creado_en
FROM justificacion_adjuntos
WHERE id = $1;
