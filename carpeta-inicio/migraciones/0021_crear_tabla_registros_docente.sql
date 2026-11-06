-- +goose Up
CREATE TABLE registros_docente (
    id BIGSERIAL PRIMARY KEY,
    alumno_id BIGINT NOT NULL REFERENCES alumnos (id),
    fecha DATE NOT NULL,
    curso_id BIGINT NOT NULL REFERENCES cursos (id),
    seccion_id BIGINT NOT NULL REFERENCES secciones (id),
    bloque_numero INTEGER NOT NULL CHECK (bloque_numero > 0),
    estado TEXT NOT NULL CHECK (estado IN ('presente', 'tardanza', 'falta', 'falta_justificada')),
    observacion TEXT CHECK (observacion IS NULL OR char_length(observacion) <= 300),
    docente_id BIGINT NOT NULL REFERENCES usuarios (id),
    creado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
    actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (alumno_id, fecha, curso_id, seccion_id, bloque_numero)
);

CREATE INDEX registros_docente_clase_idx
    ON registros_docente (fecha, seccion_id, curso_id, bloque_numero);

CREATE INDEX registros_docente_docente_idx
    ON registros_docente (docente_id, fecha);

CREATE INDEX registros_docente_alumno_idx
    ON registros_docente (alumno_id, fecha);

-- +goose Down
DROP TABLE registros_docente;
