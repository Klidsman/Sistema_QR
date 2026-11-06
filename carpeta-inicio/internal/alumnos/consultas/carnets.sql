-- name: ListarAlumnosParaCarnets :many
SELECT a.id,
       a.identificador,
       a.nombres,
       a.apellidos,
       a.foto_clave,
       s.id AS seccion_id,
       s.nombre AS seccion_nombre,
       g.id AS grado_id,
       g.nombre AS grado_nombre,
       an.nombre AS anio_escolar_nombre,
       c.token AS qr_token
FROM alumnos a
JOIN secciones s ON s.id = a.seccion_id
JOIN grados g ON g.id = s.grado_id
JOIN anios_escolares an ON an.id = g.anio_escolar_id
LEFT JOIN credenciales_qr c ON c.alumno_id = a.id AND c.estado = 'activa'
WHERE a.estado = 'activo'
  AND (sqlc.narg('seccion_id')::bigint IS NULL OR s.id = sqlc.narg('seccion_id')::bigint)
  AND (sqlc.narg('grado_id')::bigint IS NULL OR g.id = sqlc.narg('grado_id')::bigint)
ORDER BY a.apellidos, a.nombres;
