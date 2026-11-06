-- +goose Up
ALTER TABLE marcaciones ADD COLUMN registrada_por BIGINT REFERENCES usuarios (id);

CREATE INDEX marcaciones_portero_idx ON marcaciones (fecha, registrada_por);

-- +goose Down
DROP INDEX marcaciones_portero_idx;
ALTER TABLE marcaciones DROP COLUMN registrada_por;
