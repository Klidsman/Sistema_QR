-- name: ListarAlumnosDeSeccionEnFecha :many
SELECT al.id, al.identificador, al.nombres, al.apellidos
FROM alumnos al
JOIN historial_secciones h ON h.alumno_id = al.id
WHERE al.estado = 'activo'
  AND h.seccion_id = $1
  AND h.fecha_inicio <= $2
  AND (h.fecha_fin IS NULL OR h.fecha_fin > $2)
ORDER BY al.apellidos, al.nombres;

-- name: AlumnoPorTokenQR :one
SELECT a.id, a.identificador, a.nombres, a.apellidos, h.seccion_id
FROM credenciales_qr c
JOIN alumnos a ON a.id = c.alumno_id
JOIN historial_secciones h ON h.alumno_id = a.id
WHERE c.token = $1
  AND c.estado = 'activa'
  AND a.estado = 'activo'
  AND h.fecha_inicio <= $2
  AND (h.fecha_fin IS NULL OR h.fecha_fin > $2)
ORDER BY h.fecha_inicio DESC
LIMIT 1;

-- name: SeccionDeAlumnoEnFecha :one
SELECT h.seccion_id
FROM historial_secciones h
JOIN alumnos a ON a.id = h.alumno_id
WHERE h.alumno_id = $1
  AND a.estado = 'activo'
  AND h.fecha_inicio <= $2
  AND (h.fecha_fin IS NULL OR h.fecha_fin > $2)
ORDER BY h.fecha_inicio DESC
LIMIT 1;

-- name: ListarAlumnosPorIDs :many
SELECT id, identificador, nombres, apellidos
FROM alumnos
WHERE id = ANY($1::bigint[]);

-- name: CrearAlumno :one
INSERT INTO alumnos (
    identificador,
    nombres,
    apellidos,
    nombres_busqueda,
    apellidos_busqueda,
    seccion_id,
    estado,
    apoderado_nombre,
    apoderado_telefono,
    foto_clave
)
VALUES ($1, $2, $3, $4, $5, $6, 'activo', $7, $8, $9)
RETURNING id;

-- name: BuscarAlumnoPorID :one
SELECT a.id,
       a.identificador,
       a.nombres,
       a.apellidos,
       a.estado,
       a.foto_clave,
       a.apoderado_nombre,
       a.apoderado_telefono,
       a.fecha_salida,
       a.creado_en,
       a.actualizado_en,
       s.id AS seccion_id,
       s.nombre AS seccion_nombre,
       COALESCE(t.turno, ''::text) AS turno,
       g.id AS grado_id,
       g.nombre AS grado_nombre,
       an.id AS anio_escolar_id,
       an.nombre AS anio_escolar_nombre
FROM alumnos a
JOIN secciones s ON s.id = a.seccion_id
JOIN grados g ON g.id = s.grado_id
JOIN anios_escolares an ON an.id = g.anio_escolar_id
LEFT JOIN LATERAL (
    SELECT ts.turno
    FROM turnos_seccion ts
    WHERE ts.seccion_id = s.id
      AND ts.vigente_desde <= GREATEST(CURRENT_DATE, an.fecha_inicio)
    ORDER BY ts.vigente_desde DESC
    LIMIT 1
) t ON true
WHERE a.id = $1;

-- name: BuscarAlumnoPorIdentificador :one
SELECT id, estado
FROM alumnos
WHERE identificador = $1;

-- name: ListarAlumnos :many
SELECT a.id,
       a.identificador,
       a.nombres,
       a.apellidos,
       a.estado,
       a.foto_clave,
       s.id AS seccion_id,
       s.nombre AS seccion_nombre,
       g.nombre AS grado_nombre,
       an.nombre AS anio_escolar_nombre
FROM alumnos a
JOIN secciones s ON s.id = a.seccion_id
JOIN grados g ON g.id = s.grado_id
JOIN anios_escolares an ON an.id = g.anio_escolar_id
WHERE (sqlc.narg('estado')::text IS NULL OR a.estado = sqlc.narg('estado')::text)
  AND (sqlc.narg('seccion_id')::bigint IS NULL OR a.seccion_id = sqlc.narg('seccion_id')::bigint)
  AND (sqlc.narg('anio_escolar_id')::bigint IS NULL OR an.id = sqlc.narg('anio_escolar_id')::bigint)
  AND (
      sqlc.narg('texto')::text IS NULL
      OR a.nombres_busqueda LIKE '%' || sqlc.narg('texto')::text || '%'
      OR a.apellidos_busqueda LIKE '%' || sqlc.narg('texto')::text || '%'
      OR a.identificador LIKE '%' || sqlc.narg('texto')::text || '%'
  )
ORDER BY a.apellidos, a.nombres;

-- name: ActualizarAlumno :exec
UPDATE alumnos
SET nombres = $2,
    apellidos = $3,
    nombres_busqueda = $4,
    apellidos_busqueda = $5,
    seccion_id = $6,
    apoderado_nombre = $7,
    apoderado_telefono = $8,
    actualizado_en = now()
WHERE id = $1;

-- name: ActualizarFotoAlumno :exec
UPDATE alumnos
SET foto_clave = $2,
    actualizado_en = now()
WHERE id = $1;

-- name: CambiarEstadoAlumno :exec
UPDATE alumnos
SET estado = $2,
    fecha_salida = $3,
    actualizado_en = now()
WHERE id = $1;
