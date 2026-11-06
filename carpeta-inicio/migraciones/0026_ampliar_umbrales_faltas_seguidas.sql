-- +goose Up
ALTER TABLE umbrales_alerta DROP CONSTRAINT umbrales_alerta_tipo_check;

ALTER TABLE umbrales_alerta ADD CONSTRAINT umbrales_alerta_tipo_check
    CHECK (tipo IN ('tardanzas', 'faltas', 'faltas_seguidas'));

-- +goose Down
ALTER TABLE umbrales_alerta DROP CONSTRAINT umbrales_alerta_tipo_check;

ALTER TABLE umbrales_alerta ADD CONSTRAINT umbrales_alerta_tipo_check
    CHECK (tipo IN ('tardanzas', 'faltas'));
