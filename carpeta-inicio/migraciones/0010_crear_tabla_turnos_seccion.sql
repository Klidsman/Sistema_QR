-- +goose Up
CREATE TABLE turnos_seccion (
    id BIGSERIAL PRIMARY KEY,
    seccion_id BIGINT NOT NULL REFERENCES secciones (id),
    turno TEXT NOT NULL CHECK (turno IN ('manana', 'tarde')),
    vigente_desde DATE NOT NULL,
    creado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (seccion_id, vigente_desde)
);

CREATE INDEX turnos_seccion_vigencia_idx ON turnos_seccion (seccion_id, vigente_desde DESC);

INSERT INTO turnos_seccion (seccion_id, turno, vigente_desde)
SELECT s.id, s.turno, an.fecha_inicio
FROM secciones s
JOIN grados g ON g.id = s.grado_id
JOIN anios_escolares an ON an.id = g.anio_escolar_id;

ALTER TABLE secciones DROP COLUMN turno;

-- +goose Down
ALTER TABLE secciones ADD COLUMN turno TEXT;

UPDATE secciones s
SET turno = t.turno
FROM (
    SELECT DISTINCT ON (seccion_id) seccion_id, turno
    FROM turnos_seccion
    ORDER BY seccion_id, vigente_desde DESC
) t
WHERE t.seccion_id = s.id;

UPDATE secciones SET turno = 'manana' WHERE turno IS NULL;

ALTER TABLE secciones ALTER COLUMN turno SET NOT NULL;
ALTER TABLE secciones ADD CONSTRAINT secciones_turno_check CHECK (turno IN ('manana', 'tarde'));

DROP TABLE turnos_seccion;
