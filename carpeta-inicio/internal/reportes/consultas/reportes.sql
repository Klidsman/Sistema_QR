-- name: AlumnosEnFecha :many
SELECT a.id, a.identificador, a.nombres, a.apellidos,
       h.seccion_id, s.nombre AS seccion, g.nombre AS grado, g.id AS grado_id, t.turno,
       g.anio_escolar_id
FROM alumnos a
JOIN historial_secciones h
  ON h.alumno_id = a.id
 AND h.fecha_inicio <= $1
 AND (h.fecha_fin IS NULL OR h.fecha_fin > $1)
JOIN secciones s ON s.id = h.seccion_id
JOIN grados g ON g.id = s.grado_id
JOIN LATERAL (
    SELECT ts.turno
    FROM turnos_seccion ts
    WHERE ts.seccion_id = s.id AND ts.vigente_desde <= $1
    ORDER BY ts.vigente_desde DESC
    LIMIT 1
) t ON true
WHERE a.fecha_salida IS NULL OR a.fecha_salida >= $1
ORDER BY t.turno, g.nombre, s.nombre, a.apellidos, a.nombres;

-- name: MarcacionesDelDia :many
SELECT alumno_id, estado, hora, justificada
FROM marcaciones
WHERE fecha = $1;

-- name: MarcacionesDelRango :many
SELECT alumno_id, fecha, estado, justificada
FROM marcaciones
WHERE fecha BETWEEN $1 AND $2;

-- name: JustificacionesAprobadasDelRango :many
SELECT alumno_id, fecha_evento, tipo
FROM justificacion_fechas
WHERE estado = 'aprobada'
  AND fecha_evento BETWEEN $1 AND $2;

-- name: AlumnosConHistorialEnRango :many
SELECT a.id, a.identificador, a.nombres, a.apellidos,
       h.seccion_id, s.nombre AS seccion, g.nombre AS grado, g.id AS grado_id, g.anio_escolar_id,
       h.fecha_inicio, COALESCE(h.fecha_fin, $2::date) AS fecha_fin,
       t.turno
FROM alumnos a
JOIN historial_secciones h
  ON h.alumno_id = a.id
 AND h.fecha_inicio <= $2
 AND (h.fecha_fin IS NULL OR h.fecha_fin >= $1)
JOIN secciones s ON s.id = h.seccion_id
JOIN grados g ON g.id = s.grado_id
JOIN LATERAL (
    SELECT ts.turno
    FROM turnos_seccion ts
    WHERE ts.seccion_id = s.id AND ts.vigente_desde <= GREATEST(h.fecha_inicio, $1)
    ORDER BY ts.vigente_desde DESC
    LIMIT 1
) t ON true
WHERE a.fecha_salida IS NULL OR a.fecha_salida >= $1
ORDER BY a.apellidos, a.nombres, h.fecha_inicio;

-- name: ExcepcionesDelRango :many
SELECT e.fecha_inicio, e.fecha_fin, eg.grado_id
FROM excepciones_calendario e
LEFT JOIN excepciones_grados eg ON eg.excepcion_id = e.id
WHERE e.fecha_fin >= $1 AND e.fecha_inicio <= $2;

-- name: SeccionesDelDocenteEnRango :many
SELECT DISTINCT seccion_id
FROM asignaciones_docente
WHERE docente_id = $1
  AND fecha_inicio <= $3
  AND (fecha_fin IS NULL OR fecha_fin >= $2);
