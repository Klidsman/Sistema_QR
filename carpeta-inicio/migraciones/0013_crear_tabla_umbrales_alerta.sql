-- +goose Up
CREATE TABLE umbrales_alerta (
    id BIGSERIAL PRIMARY KEY,
    anio_escolar_id BIGINT NOT NULL REFERENCES anios_escolares (id),
    tipo TEXT NOT NULL CHECK (tipo IN ('tardanzas', 'faltas')),
    valor INTEGER NOT NULL CHECK (valor > 0 AND valor <= 100),
    vigente_desde DATE NOT NULL,
    creado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (anio_escolar_id, tipo, vigente_desde)
);

CREATE INDEX umbrales_alerta_vigencia_idx
    ON umbrales_alerta (anio_escolar_id, tipo, vigente_desde DESC);

-- +goose Down
DROP TABLE umbrales_alerta;
