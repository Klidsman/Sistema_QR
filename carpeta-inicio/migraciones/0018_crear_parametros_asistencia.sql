-- +goose Up
CREATE TABLE parametros_asistencia (
    id SMALLINT PRIMARY KEY CHECK (id = 1),
    plazo_edicion_horas INTEGER NOT NULL DEFAULT 24
        CHECK (plazo_edicion_horas BETWEEN 1 AND 168),
    actualizado_por BIGINT REFERENCES usuarios (id),
    actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO parametros_asistencia (id) VALUES (1);

-- +goose Down
DROP TABLE parametros_asistencia;
