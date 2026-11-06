-- +goose Up
CREATE TABLE horario_seccion (
    id BIGSERIAL PRIMARY KEY,
    seccion_id BIGINT NOT NULL REFERENCES secciones (id),
    dia_semana SMALLINT NOT NULL CHECK (dia_semana BETWEEN 1 AND 5),
    bloque_numero INTEGER NOT NULL CHECK (bloque_numero > 0),
    curso_id BIGINT REFERENCES cursos (id),
    vigente_desde DATE NOT NULL,
    creado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (seccion_id, dia_semana, bloque_numero, vigente_desde)
);

CREATE INDEX horario_seccion_vigencia_idx
    ON horario_seccion (seccion_id, dia_semana, bloque_numero, vigente_desde DESC);

-- +goose Down
DROP TABLE horario_seccion;
