-- +goose Up
CREATE TABLE importaciones (
    id BIGSERIAL PRIMARY KEY,
    usuario_id BIGINT NOT NULL REFERENCES usuarios (id),
    archivo_nombre TEXT NOT NULL,
    filas_leidas INTEGER NOT NULL CHECK (filas_leidas >= 0),
    filas_creadas INTEGER NOT NULL CHECK (filas_creadas >= 0),
    filas_rechazadas INTEGER NOT NULL CHECK (filas_rechazadas >= 0),
    creada_en TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE importacion_filas (
    id BIGSERIAL PRIMARY KEY,
    importacion_id BIGINT NOT NULL REFERENCES importaciones (id),
    numero_fila INTEGER NOT NULL CHECK (numero_fila > 0),
    identificador TEXT NOT NULL,
    motivo TEXT NOT NULL CHECK (motivo IN (
        'campo_obligatorio_vacio',
        'identificador_repetido_en_archivo',
        'identificador_existente',
        'seccion_inexistente',
        'formato_invalido'
    ))
);

-- +goose Down
DROP TABLE importacion_filas;
DROP TABLE importaciones;
