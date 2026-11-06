-- +goose Up
CREATE TABLE auditoria (
    id BIGSERIAL PRIMARY KEY,
    usuario_id BIGINT NOT NULL REFERENCES usuarios (id),
    entidad TEXT NOT NULL CHECK (entidad IN ('alumno', 'credencial_qr', 'historial_seccion', 'importacion')),
    entidad_id BIGINT NOT NULL,
    accion TEXT NOT NULL,
    valor_anterior JSONB,
    valor_nuevo JSONB,
    motivo TEXT,
    creada_en TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION impedir_cambios_auditoria() RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'La tabla auditoria es de solo escritura';
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER auditoria_sin_actualizacion
    BEFORE UPDATE OR DELETE ON auditoria
    FOR EACH ROW EXECUTE FUNCTION impedir_cambios_auditoria();

-- +goose Down
DROP TRIGGER auditoria_sin_actualizacion ON auditoria;
DROP FUNCTION impedir_cambios_auditoria();
DROP TABLE auditoria;
