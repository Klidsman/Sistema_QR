-- name: InsertarAlerta :one
INSERT INTO alertas (
    tipo, alumno_id, seccion_id, curso_id, bloque_numero, fecha_evento, mes,
    marcacion_id, destinatario_id, sin_responsable
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT DO NOTHING
RETURNING id;

-- name: BloquearProcesos :exec
SELECT pg_advisory_xact_lock($1);

-- name: ListarAlertas :many
SELECT id, tipo, alumno_id, seccion_id, curso_id, bloque_numero, fecha_evento, mes,
       marcacion_id, destinatario_id, estado, leida_en, leida_por, sin_responsable,
       situacion_cambiada, COALESCE(motivo_cambio, '') AS motivo_cambio, creada_en
FROM alertas
WHERE ($1::boolean OR destinatario_id = $2)
  AND ($3::date IS NULL OR fecha_evento >= $3)
  AND ($4::date IS NULL OR fecha_evento <= $4)
  AND ($5::bigint = 0 OR seccion_id = $5)
  AND ($6::text = '' OR tipo = $6)
  AND ($7::text = '' OR estado = $7)
ORDER BY creada_en DESC
LIMIT 200;

-- name: ContarAlertasNuevas :one
SELECT count(*)
FROM alertas
WHERE estado = 'nueva'
  AND ($1::boolean OR destinatario_id = $2);

-- name: MarcarAlertaVista :execrows
UPDATE alertas
SET estado = 'vista', leida_en = now(), leida_por = $2
WHERE id = $1
  AND estado = 'nueva'
  AND ($3::boolean OR destinatario_id = $2);

-- name: MarcarAlertasVistas :execrows
UPDATE alertas
SET estado = 'vista', leida_en = now(), leida_por = $2
WHERE id = ANY($1::bigint[])
  AND estado = 'nueva'
  AND ($3::boolean OR destinatario_id = $2);

-- name: AsignarResponsableAlerta :execrows
UPDATE alertas
SET destinatario_id = $2, sin_responsable = false
WHERE id = $1;

-- name: MarcarAlertasSituacionCambiada :execrows
UPDATE alertas
SET situacion_cambiada = true, motivo_cambio = $2
WHERE alumno_id = $1
  AND fecha_evento = $3
  AND situacion_cambiada = false;

-- name: RegistrarCierreAlerta :execrows
INSERT INTO cierres_alerta (fecha, turno)
VALUES ($1, $2)
ON CONFLICT (fecha, turno) DO NOTHING;
