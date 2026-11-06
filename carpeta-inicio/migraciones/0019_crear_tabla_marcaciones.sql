-- +goose Up
CREATE TABLE marcaciones (
    id BIGSERIAL PRIMARY KEY,
    alumno_id BIGINT NOT NULL REFERENCES alumnos (id),
    fecha DATE NOT NULL,
    hora TIMESTAMPTZ NOT NULL,
    estado TEXT NOT NULL CHECK (estado IN ('puntual', 'tardanza', 'tardanza_grave')),
    creada_en TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (alumno_id, fecha)
);

CREATE INDEX marcaciones_fecha_idx ON marcaciones (fecha);

-- +goose Down
DROP TABLE marcaciones;
