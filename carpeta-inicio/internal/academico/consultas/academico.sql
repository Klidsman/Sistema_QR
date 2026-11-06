-- name: CrearAnioEscolar :one
INSERT INTO anios_escolares (nombre, fecha_inicio, fecha_fin)
VALUES ($1, $2, $3)
RETURNING id;

-- name: CrearGrado :one
INSERT INTO grados (anio_escolar_id, nombre)
VALUES ($1, $2)
RETURNING id;

-- name: CrearSeccion :one
INSERT INTO secciones (grado_id, nombre, turno)
VALUES ($1, $2, $3)
RETURNING id;

-- name: BuscarAnioEscolarPorNombre :one
SELECT id
FROM anios_escolares
WHERE nombre = $1;

-- name: BuscarGradoPorNombre :one
SELECT id
FROM grados
WHERE anio_escolar_id = $1
  AND nombre = $2;

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

-- name: ListarSecciones :many
SELECT s.id,
       s.nombre,
       s.turno,
       g.id AS grado_id,
       g.nombre AS grado,
       an.id AS anio_escolar_id,
       an.nombre AS anio_escolar
FROM secciones s
JOIN grados g ON g.id = s.grado_id
JOIN anios_escolares an ON an.id = g.anio_escolar_id
ORDER BY an.nombre, g.nombre, s.nombre;

-- name: BuscarSeccionPorID :one
SELECT s.id,
       s.nombre,
       s.turno,
       g.id AS grado_id,
       g.nombre AS grado,
       an.id AS anio_escolar_id,
       an.nombre AS anio_escolar
FROM secciones s
JOIN grados g ON g.id = s.grado_id
JOIN anios_escolares an ON an.id = g.anio_escolar_id
WHERE s.id = $1;

-- name: ListarSeccionesVigentes :many
SELECT s.id,
       s.nombre,
       s.turno,
       g.id AS grado_id,
       g.nombre AS grado,
       an.id AS anio_escolar_id,
       an.nombre AS anio_escolar
FROM secciones s
JOIN grados g ON g.id = s.grado_id
JOIN anios_escolares an ON an.id = g.anio_escolar_id
WHERE an.estado = 'activo'
ORDER BY an.nombre DESC, g.nombre, s.nombre;
