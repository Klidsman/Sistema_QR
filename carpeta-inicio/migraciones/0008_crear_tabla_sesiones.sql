-- +goose Up
CREATE TABLE sesiones (
    token TEXT PRIMARY KEY,
    datos BYTEA NOT NULL,
    expira_en TIMESTAMPTZ NOT NULL
);

CREATE INDEX sesiones_expira_en_idx ON sesiones (expira_en);

-- +goose Down
DROP TABLE sesiones;
