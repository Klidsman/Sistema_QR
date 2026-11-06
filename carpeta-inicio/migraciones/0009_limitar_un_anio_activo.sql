-- +goose Up
ALTER TABLE anios_escolares ADD COLUMN cerrado_en TIMESTAMPTZ;

CREATE UNIQUE INDEX un_anio_activo ON anios_escolares (estado) WHERE estado = 'activo';

-- +goose Down
DROP INDEX un_anio_activo;

ALTER TABLE anios_escolares DROP COLUMN cerrado_en;
