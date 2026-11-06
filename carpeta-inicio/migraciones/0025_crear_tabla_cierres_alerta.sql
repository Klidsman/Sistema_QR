-- +goose Up
CREATE TABLE cierres_alerta (
    id BIGSERIAL PRIMARY KEY,
    fecha DATE NOT NULL,
    turno TEXT NOT NULL CHECK (turno IN ('manana', 'tarde')),
    ejecutado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (fecha, turno)
);

-- +goose Down
DROP TABLE cierres_alerta;
