-- +goose Up
CREATE TABLE usuarios (
    id BIGSERIAL PRIMARY KEY,
    correo TEXT NOT NULL UNIQUE CHECK (correo = lower(correo)),
    contrasenia_hash TEXT NOT NULL,
    nombre_completo TEXT NOT NULL,
    rol TEXT NOT NULL CHECK (rol IN ('administrador', 'director', 'docente', 'porteria')),
    activo BOOLEAN NOT NULL DEFAULT TRUE,
    creado_en TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE usuarios;
