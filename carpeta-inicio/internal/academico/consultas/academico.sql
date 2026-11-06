-- name: CrearAnioEscolar :one
INSERT INTO anios_escolares (nombre, fecha_inicio, fecha_fin)
VALUES ($1, $2, $3)
RETURNING id;

-- name: BuscarAnioEscolarPorNombre :one
SELECT id
FROM anios_escolares
WHERE nombre = $1;

-- name: BuscarAnioEscolarPorID :one
SELECT id, nombre, fecha_inicio, fecha_fin, estado, cerrado_en
FROM anios_escolares
WHERE id = $1;

-- name: BuscarAnioActivo :one
SELECT id, nombre, fecha_inicio, fecha_fin, estado, cerrado_en
FROM anios_escolares
WHERE estado = 'activo';

-- name: AnioEscolarQueCubreFecha :one
SELECT id, nombre, fecha_inicio, fecha_fin, estado, cerrado_en
FROM anios_escolares
WHERE fecha_inicio <= $1
  AND fecha_fin >= $1
ORDER BY fecha_inicio DESC
LIMIT 1;

-- name: ListarAniosEscolares :many
SELECT id, nombre, fecha_inicio, fecha_fin, estado, cerrado_en
FROM anios_escolares
ORDER BY fecha_inicio DESC;

-- name: CerrarAnioEscolar :execrows
UPDATE anios_escolares
SET estado = 'cerrado', cerrado_en = now()
WHERE id = $1
  AND estado = 'activo';

-- name: CrearGrado :one
INSERT INTO grados (anio_escolar_id, nombre)
VALUES ($1, $2)
RETURNING id;

-- name: BuscarGradoPorNombre :one
SELECT id
FROM grados
WHERE anio_escolar_id = $1
  AND nombre = $2;

-- name: BuscarGradoPorID :one
SELECT g.id, g.anio_escolar_id, g.nombre, an.estado AS anio_estado
FROM grados g
JOIN anios_escolares an ON an.id = g.anio_escolar_id
WHERE g.id = $1;

-- name: ListarGradosPorAnio :many
SELECT id, anio_escolar_id, nombre
FROM grados
WHERE anio_escolar_id = $1
ORDER BY nombre;

-- name: CrearSeccion :one
INSERT INTO secciones (grado_id, nombre)
VALUES ($1, $2)
RETURNING id;

-- name: BuscarSeccionPorNombre :one
SELECT id
FROM secciones
WHERE grado_id = $1
  AND nombre = $2;

-- name: BuscarSeccion :one
SELECT s.id
FROM secciones s
JOIN grados g ON g.id = s.grado_id
WHERE g.anio_escolar_id = $1
  AND g.nombre = $2
  AND s.nombre = $3;

-- name: CrearTurnoSeccion :one
INSERT INTO turnos_seccion (seccion_id, turno, vigente_desde)
VALUES ($1, $2, $3)
ON CONFLICT (seccion_id, vigente_desde) DO UPDATE SET turno = EXCLUDED.turno
RETURNING id;

-- name: TurnoVigenteSeccion :one
SELECT turno
FROM turnos_seccion
WHERE seccion_id = $1
  AND vigente_desde <= $2
ORDER BY vigente_desde DESC
LIMIT 1;

-- name: ListarTurnosSeccion :many
SELECT id, seccion_id, turno, vigente_desde
FROM turnos_seccion
WHERE seccion_id = $1
ORDER BY vigente_desde DESC;

-- name: ListarSecciones :many
SELECT s.id,
       s.nombre,
       COALESCE(t.turno, ''::text) AS turno,
       g.id AS grado_id,
       g.nombre AS grado,
       an.id AS anio_escolar_id,
       an.nombre AS anio_escolar
FROM secciones s
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
ORDER BY an.nombre, g.nombre, s.nombre;

-- name: ListarSeccionesPorAnio :many
SELECT s.id,
       s.nombre,
       COALESCE(t.turno, ''::text) AS turno,
       g.id AS grado_id,
       g.nombre AS grado,
       an.id AS anio_escolar_id,
       an.nombre AS anio_escolar
FROM secciones s
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
WHERE an.id = $1
ORDER BY g.nombre, s.nombre;

-- name: BuscarSeccionPorID :one
SELECT s.id,
       s.nombre,
       COALESCE(t.turno, ''::text) AS turno,
       g.id AS grado_id,
       g.nombre AS grado,
       an.id AS anio_escolar_id,
       an.nombre AS anio_escolar
FROM secciones s
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
WHERE s.id = $1;

-- name: ListarSeccionesVigentes :many
SELECT s.id,
       s.nombre,
       COALESCE(t.turno, ''::text) AS turno,
       g.id AS grado_id,
       g.nombre AS grado,
       an.id AS anio_escolar_id,
       an.nombre AS anio_escolar
FROM secciones s
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
WHERE an.estado = 'activo'
ORDER BY an.nombre DESC, g.nombre, s.nombre;

-- name: GuardarHorarioTurno :one
INSERT INTO horarios_turno (
    turno, hora_ingreso, tolerancia_minutos, hora_limite_tardanza, hora_cierre, vigente_desde
)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (turno, vigente_desde) DO UPDATE SET
    hora_ingreso = EXCLUDED.hora_ingreso,
    tolerancia_minutos = EXCLUDED.tolerancia_minutos,
    hora_limite_tardanza = EXCLUDED.hora_limite_tardanza,
    hora_cierre = EXCLUDED.hora_cierre
RETURNING id;

-- name: HorarioTurnoVigente :one
SELECT id,
       turno,
       to_char(hora_ingreso, 'HH24:MI') AS hora_ingreso,
       tolerancia_minutos,
       to_char(hora_limite_tardanza, 'HH24:MI') AS hora_limite_tardanza,
       to_char(hora_cierre, 'HH24:MI') AS hora_cierre,
       vigente_desde
FROM horarios_turno
WHERE turno = $1
  AND vigente_desde <= $2
ORDER BY vigente_desde DESC
LIMIT 1;

-- name: ListarHorariosTurno :many
SELECT id,
       turno,
       to_char(hora_ingreso, 'HH24:MI') AS hora_ingreso,
       tolerancia_minutos,
       to_char(hora_limite_tardanza, 'HH24:MI') AS hora_limite_tardanza,
       to_char(hora_cierre, 'HH24:MI') AS hora_cierre,
       vigente_desde
FROM horarios_turno
WHERE turno = $1
ORDER BY vigente_desde DESC;

-- name: CrearExcepcionCalendario :one
INSERT INTO excepciones_calendario (
    anio_escolar_id, tipo, fecha_inicio, fecha_fin, motivo, creado_por
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id;

-- name: CrearExcepcionGrado :exec
INSERT INTO excepciones_grados (excepcion_id, grado_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: ListarExcepcionesPorAnio :many
SELECT id, anio_escolar_id, tipo, fecha_inicio, fecha_fin, motivo, creado_por, creado_en
FROM excepciones_calendario
WHERE anio_escolar_id = $1
ORDER BY fecha_inicio DESC, id DESC;

-- name: ListarGradosDeExcepciones :many
SELECT eg.excepcion_id, eg.grado_id
FROM excepciones_grados eg
JOIN excepciones_calendario e ON e.id = eg.excepcion_id
WHERE e.anio_escolar_id = $1;

-- name: CrearUmbralAlerta :one
INSERT INTO umbrales_alerta (anio_escolar_id, tipo, valor, vigente_desde)
VALUES ($1, $2, $3, $4)
RETURNING id;

-- name: UmbralVigente :one
SELECT id, anio_escolar_id, tipo, valor, vigente_desde
FROM umbrales_alerta
WHERE anio_escolar_id = $1
  AND tipo = $2
  AND vigente_desde <= $3
ORDER BY vigente_desde DESC
LIMIT 1;

-- name: ListarUmbralesPorAnio :many
SELECT id, anio_escolar_id, tipo, valor, vigente_desde
FROM umbrales_alerta
WHERE anio_escolar_id = $1
ORDER BY tipo, vigente_desde DESC;

-- name: CrearCurso :one
INSERT INTO cursos (anio_escolar_id, nombre)
VALUES ($1, $2)
RETURNING id;

-- name: CrearCursoGrado :exec
INSERT INTO cursos_grados (curso_id, grado_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: BuscarCursoPorNombre :one
SELECT id
FROM cursos
WHERE anio_escolar_id = $1
  AND lower(nombre) = lower($2);

-- name: ListarCursosPorAnio :many
SELECT id, anio_escolar_id, nombre, activo, creado_en
FROM cursos
WHERE anio_escolar_id = $1
ORDER BY nombre;

-- name: ListarCursosGradosPorAnio :many
SELECT cg.curso_id, cg.grado_id
FROM cursos_grados cg
JOIN cursos c ON c.id = cg.curso_id
WHERE c.anio_escolar_id = $1;

-- name: BuscarCursoPorID :one
SELECT id, anio_escolar_id, nombre, activo, creado_en
FROM cursos
WHERE id = $1;

-- name: DesactivarCurso :execrows
UPDATE cursos
SET activo = false
WHERE id = $1
  AND activo;

-- name: CursoVinculadoAGrado :one
SELECT EXISTS (
    SELECT 1
    FROM cursos c
    JOIN cursos_grados cg ON cg.curso_id = c.id
    WHERE c.id = $1
      AND c.activo
      AND cg.grado_id = $2
);

-- name: GradoDeSeccion :one
SELECT g.id
FROM secciones s
JOIN grados g ON g.id = s.grado_id
WHERE s.id = $1;

-- name: GuardarBloqueClase :one
INSERT INTO bloques_clase (turno, numero, hora_inicio, hora_fin, vigente_desde)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (turno, numero, vigente_desde) DO UPDATE SET
    hora_inicio = EXCLUDED.hora_inicio,
    hora_fin = EXCLUDED.hora_fin
RETURNING id;

-- name: BloquesVigentesTurno :many
SELECT DISTINCT ON (numero)
       id,
       turno,
       numero,
       to_char(hora_inicio, 'HH24:MI') AS hora_inicio,
       to_char(hora_fin, 'HH24:MI') AS hora_fin,
       vigente_desde
FROM bloques_clase
WHERE turno = $1
  AND vigente_desde <= $2
ORDER BY numero, vigente_desde DESC;

-- name: ListarBloquesTurno :many
SELECT id,
       turno,
       numero,
       to_char(hora_inicio, 'HH24:MI') AS hora_inicio,
       to_char(hora_fin, 'HH24:MI') AS hora_fin,
       vigente_desde
FROM bloques_clase
WHERE turno = $1
ORDER BY numero, vigente_desde DESC;

-- name: BuscarBloquePorID :one
SELECT id,
       turno,
       numero,
       to_char(hora_inicio, 'HH24:MI') AS hora_inicio,
       to_char(hora_fin, 'HH24:MI') AS hora_fin,
       vigente_desde
FROM bloques_clase
WHERE id = $1;

-- name: BloqueVigentePorNumero :one
SELECT id,
       turno,
       numero,
       to_char(hora_inicio, 'HH24:MI') AS hora_inicio,
       to_char(hora_fin, 'HH24:MI') AS hora_fin,
       vigente_desde
FROM bloques_clase
WHERE turno = $1
  AND numero = $2
  AND vigente_desde <= $3
ORDER BY vigente_desde DESC
LIMIT 1;

-- name: UpsertHorarioSeccion :exec
INSERT INTO horario_seccion (seccion_id, dia_semana, bloque_numero, curso_id, vigente_desde)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (seccion_id, dia_semana, bloque_numero, vigente_desde)
DO UPDATE SET curso_id = EXCLUDED.curso_id;

-- name: HorarioSeccionVigente :many
SELECT DISTINCT ON (dia_semana, bloque_numero)
       id,
       seccion_id,
       dia_semana,
       bloque_numero,
       curso_id,
       vigente_desde
FROM horario_seccion
WHERE seccion_id = $1
  AND vigente_desde <= $2
ORDER BY dia_semana, bloque_numero, vigente_desde DESC;

-- name: HorarioSeccionCelda :one
SELECT id,
       seccion_id,
       dia_semana,
       bloque_numero,
       curso_id,
       vigente_desde
FROM horario_seccion
WHERE seccion_id = $1
  AND dia_semana = $2
  AND bloque_numero = $3
  AND vigente_desde <= $4
ORDER BY vigente_desde DESC
LIMIT 1;
