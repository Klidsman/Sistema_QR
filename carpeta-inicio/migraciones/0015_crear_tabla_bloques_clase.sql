-- +goose Up
CREATE TABLE bloques_clase (
    id BIGSERIAL PRIMARY KEY,
    turno TEXT NOT NULL CHECK (turno IN ('manana', 'tarde')),
    numero INTEGER NOT NULL CHECK (numero > 0 AND numero <= 20),
    hora_inicio TIME NOT NULL,
    hora_fin TIME NOT NULL,
    vigente_desde DATE NOT NULL,
    creado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (turno, numero, vigente_desde),
    CHECK (hora_fin > hora_inicio)
);

CREATE INDEX bloques_clase_vigencia_idx ON bloques_clase (turno, numero, vigente_desde DESC);

-- +goose Down
DROP TABLE bloques_clase;
