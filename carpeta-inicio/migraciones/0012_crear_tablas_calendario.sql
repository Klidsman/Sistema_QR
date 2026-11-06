-- +goose Up
CREATE TABLE excepciones_calendario (
    id BIGSERIAL PRIMARY KEY,
    anio_escolar_id BIGINT NOT NULL REFERENCES anios_escolares (id),
    tipo TEXT NOT NULL CHECK (tipo IN ('feriado', 'no_laborable', 'suspension')),
    fecha_inicio DATE NOT NULL,
    fecha_fin DATE NOT NULL,
    motivo TEXT NOT NULL CHECK (char_length(motivo) BETWEEN 3 AND 200),
    creado_por BIGINT NOT NULL REFERENCES usuarios (id),
    creado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (fecha_fin >= fecha_inicio)
);

CREATE INDEX excepciones_calendario_fechas_idx
    ON excepciones_calendario (anio_escolar_id, fecha_inicio, fecha_fin);

CREATE TABLE excepciones_grados (
    excepcion_id BIGINT NOT NULL REFERENCES excepciones_calendario (id),
    grado_id BIGINT NOT NULL REFERENCES grados (id),
    PRIMARY KEY (excepcion_id, grado_id)
);

-- +goose Down
DROP TABLE excepciones_grados;
DROP TABLE excepciones_calendario;
