-- name: CrearHistorialSeccion :one
INSERT INTO historial_secciones (alumno_id, seccion_id, fecha_inicio, fecha_fin)
VALUES ($1, $2, $3, $4)
RETURNING id;

-- name: BuscarHistorialActual :one
SELECT h.id,
       h.seccion_id,
       h.fecha_inicio,
       h.fecha_fin,
       s.nombre AS seccion_nombre,
       g.nombre AS grado_nombre,
       an.nombre AS anio_escolar_nombre
FROM historial_secciones h
JOIN secciones s ON s.id = h.seccion_id
JOIN grados g ON g.id = s.grado_id
JOIN anios_escolares an ON an.id = g.anio_escolar_id
WHERE h.alumno_id = $1
  AND h.fecha_fin IS NULL;

-- name: CerrarHistorialSeccion :exec
UPDATE historial_secciones
SET fecha_fin = $2
WHERE alumno_id = $1
  AND fecha_fin IS NULL;

-- name: ListarHistorialSecciones :many
SELECT h.id,
       h.seccion_id,
       h.fecha_inicio,
       h.fecha_fin,
       s.nombre AS seccion_nombre,
       g.nombre AS grado_nombre,
       an.nombre AS anio_escolar_nombre
FROM historial_secciones h
JOIN secciones s ON s.id = h.seccion_id
JOIN grados g ON g.id = s.grado_id
JOIN anios_escolares an ON an.id = g.anio_escolar_id
WHERE h.alumno_id = $1
ORDER BY h.fecha_inicio DESC, h.id DESC;

-- name: ActualizarSeccionAlumno :exec
UPDATE alumnos
SET seccion_id = $2,
    actualizado_en = now()
WHERE id = $1;

-- name: RegistrarSalidaAlumno :exec
UPDATE alumnos
SET estado = $2,
    fecha_salida = $3,
    actualizado_en = now()
WHERE id = $1;
