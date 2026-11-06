-- +goose Up
CREATE TABLE asignaciones_docente (
    id BIGSERIAL PRIMARY KEY,
    docente_id BIGINT NOT NULL REFERENCES usuarios (id),
    curso_id BIGINT NOT NULL REFERENCES cursos (id),
    seccion_id BIGINT NOT NULL REFERENCES secciones (id),
    fecha_inicio DATE NOT NULL,
    fecha_fin DATE,
    titular BOOLEAN NOT NULL DEFAULT true,
    creado_por BIGINT NOT NULL REFERENCES usuarios (id),
    creado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (docente_id, curso_id, seccion_id, fecha_inicio),
    CHECK (fecha_fin IS NULL OR fecha_fin >= fecha_inicio)
);

CREATE INDEX asignaciones_docente_vigencia_idx
    ON asignaciones_docente (docente_id, fecha_inicio, fecha_fin);

CREATE INDEX asignaciones_docente_clase_idx
    ON asignaciones_docente (seccion_id, curso_id, fecha_inicio, fecha_fin);

-- +goose Down
DROP TABLE asignaciones_docente;
