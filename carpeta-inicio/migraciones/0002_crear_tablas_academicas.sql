-- +goose Up
CREATE TABLE anios_escolares (
    id BIGSERIAL PRIMARY KEY,
    nombre TEXT NOT NULL UNIQUE,
    fecha_inicio DATE NOT NULL,
    fecha_fin DATE NOT NULL,
    estado TEXT NOT NULL DEFAULT 'activo' CHECK (estado IN ('activo', 'cerrado')),
    CHECK (fecha_fin > fecha_inicio)
);

CREATE TABLE grados (
    id BIGSERIAL PRIMARY KEY,
    anio_escolar_id BIGINT NOT NULL REFERENCES anios_escolares (id),
    nombre TEXT NOT NULL,
    UNIQUE (anio_escolar_id, nombre)
);

CREATE TABLE secciones (
    id BIGSERIAL PRIMARY KEY,
    grado_id BIGINT NOT NULL REFERENCES grados (id),
    nombre TEXT NOT NULL,
    turno TEXT NOT NULL CHECK (turno IN ('manana', 'tarde')),
    UNIQUE (grado_id, nombre)
);

-- +goose Down
DROP TABLE secciones;
DROP TABLE grados;
DROP TABLE anios_escolares;
