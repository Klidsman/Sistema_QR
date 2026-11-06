-- name: CrearAsignacion :one
INSERT INTO asignaciones_docente (
    docente_id, curso_id, seccion_id, fecha_inicio, fecha_fin, titular, creado_por
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id;

-- name: BuscarAsignacionPorID :one
SELECT id, docente_id, curso_id, seccion_id, fecha_inicio, fecha_fin, titular, creado_por, creado_en
FROM asignaciones_docente
WHERE id = $1;

-- name: CerrarAsignacion :execrows
UPDATE asignaciones_docente
SET fecha_fin = $2
WHERE id = $1
  AND (fecha_fin IS NULL OR fecha_fin > $2);

-- name: ListarAsignacionesPorDocenteFecha :many
SELECT id, docente_id, curso_id, seccion_id, fecha_inicio, fecha_fin, titular, creado_por, creado_en
FROM asignaciones_docente
WHERE docente_id = $1
  AND fecha_inicio <= $2
  AND (fecha_fin IS NULL OR fecha_fin >= $2)
ORDER BY fecha_inicio, id;

-- name: ListarAsignaciones :many
SELECT id, docente_id, curso_id, seccion_id, fecha_inicio, fecha_fin, titular, creado_por, creado_en
FROM asignaciones_docente
ORDER BY fecha_inicio DESC, id DESC;

-- name: DocenteAsignadoAClase :one
SELECT EXISTS (
    SELECT 1
    FROM asignaciones_docente
    WHERE docente_id = $1
      AND curso_id = $2
      AND seccion_id = $3
      AND fecha_inicio <= $4
      AND (fecha_fin IS NULL OR fecha_fin >= $4)
);

-- name: ListarRegistrosClase :many
SELECT id, alumno_id, fecha, curso_id, seccion_id, bloque_numero, estado,
       observacion, docente_id, creado_en, actualizado_en
FROM registros_docente
WHERE fecha = $1
  AND seccion_id = $2
  AND curso_id = $3
  AND bloque_numero = $4
ORDER BY alumno_id;

-- name: GuardarRegistro :one
INSERT INTO registros_docente (
    alumno_id, fecha, curso_id, seccion_id, bloque_numero, estado, observacion, docente_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (alumno_id, fecha, curso_id, seccion_id, bloque_numero)
DO UPDATE SET
    estado = EXCLUDED.estado,
    observacion = EXCLUDED.observacion,
    docente_id = EXCLUDED.docente_id,
    actualizado_en = now()
RETURNING id;

-- name: InsertarRegistroSiFalta :execrows
INSERT INTO registros_docente (
    alumno_id, fecha, curso_id, seccion_id, bloque_numero, estado, observacion, docente_id
)
VALUES ($1, $2, $3, $4, $5, 'presente', NULL, $6)
ON CONFLICT (alumno_id, fecha, curso_id, seccion_id, bloque_numero) DO NOTHING;

-- name: BuscarRegistroPorID :one
SELECT id, alumno_id, fecha, curso_id, seccion_id, bloque_numero, estado,
       observacion, docente_id, creado_en, actualizado_en
FROM registros_docente
WHERE id = $1;

-- name: CorregirRegistro :one
UPDATE registros_docente
SET estado = $2,
    observacion = $3,
    docente_id = $4,
    actualizado_en = now()
WHERE id = $1
RETURNING id;

-- name: ListarRegistrosConsulta :many
SELECT id, alumno_id, fecha, curso_id, seccion_id, bloque_numero, estado,
       observacion, docente_id, creado_en, actualizado_en
FROM registros_docente
WHERE fecha = $1
  AND seccion_id = $2
ORDER BY curso_id, bloque_numero, alumno_id;

-- name: ListarRegistrosExportacionDocente :many
SELECT id, alumno_id, fecha, curso_id, seccion_id, bloque_numero, estado,
       observacion, docente_id, creado_en, actualizado_en
FROM registros_docente
WHERE fecha BETWEEN $1 AND $2
  AND docente_id = $3
ORDER BY fecha, curso_id, bloque_numero, alumno_id;

-- name: ListarRegistrosExportacionSeccion :many
SELECT id, alumno_id, fecha, curso_id, seccion_id, bloque_numero, estado,
       observacion, docente_id, creado_en, actualizado_en
FROM registros_docente
WHERE fecha BETWEEN $1 AND $2
  AND seccion_id = $3
  AND ($4 = 0 OR curso_id = $4)
ORDER BY fecha, curso_id, bloque_numero, alumno_id;

-- name: ListarMarcacionesPorFechaYAlumnos :many
SELECT id, alumno_id, fecha, hora, estado, creada_en
FROM marcaciones
WHERE fecha = $1
  AND alumno_id = ANY($2::bigint[]);

-- name: ListarMarcacionesPorRangoYAlumnos :many
SELECT id, alumno_id, fecha, hora, estado, creada_en
FROM marcaciones
WHERE fecha BETWEEN $1 AND $2
  AND alumno_id = ANY($3::bigint[]);

-- name: CrearMarcacionPorteria :one
INSERT INTO marcaciones (alumno_id, fecha, hora, estado, registrada_por)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (alumno_id, fecha) DO NOTHING
RETURNING id;

-- name: ListarMarcacionesDelRango :many
SELECT id, alumno_id, fecha, hora, estado, creada_en, registrada_por
FROM marcaciones
WHERE fecha BETWEEN $1 AND $2
ORDER BY fecha, hora;

-- name: ContarTardanzasAlumnoRango :one
SELECT count(*)
FROM marcaciones
WHERE alumno_id = $1
  AND fecha BETWEEN $2 AND $3
  AND estado IN ('tardanza', 'tardanza_grave');

-- name: ExisteMarcacion :one
SELECT EXISTS (
    SELECT 1 FROM marcaciones WHERE alumno_id = $1 AND fecha = $2
);

-- name: ListarDocentesDeSeccionFecha :many
SELECT DISTINCT docente_id
FROM asignaciones_docente
WHERE seccion_id = $1
  AND fecha_inicio <= $2
  AND (fecha_fin IS NULL OR fecha_fin >= $2);

-- name: ObtenerParametrosAsistencia :one
SELECT plazo_edicion_horas, actualizado_por, actualizado_en
FROM parametros_asistencia
WHERE id = 1;

-- name: ActualizarPlazoEdicion :one
UPDATE parametros_asistencia
SET plazo_edicion_horas = $1,
    actualizado_por = $2,
    actualizado_en = now()
WHERE id = 1
RETURNING plazo_edicion_horas, actualizado_por, actualizado_en;
