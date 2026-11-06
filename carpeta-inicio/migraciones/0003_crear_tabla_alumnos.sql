-- +goose Up
CREATE TABLE alumnos (
    id BIGSERIAL PRIMARY KEY,
    identificador TEXT NOT NULL UNIQUE CHECK (identificador = btrim(identificador) AND identificador <> ''),
    nombres TEXT NOT NULL CHECK (btrim(nombres) <> ''),
    apellidos TEXT NOT NULL CHECK (btrim(apellidos) <> ''),
    nombres_busqueda TEXT NOT NULL,
    apellidos_busqueda TEXT NOT NULL,
    seccion_id BIGINT NOT NULL REFERENCES secciones (id),
    estado TEXT NOT NULL DEFAULT 'activo' CHECK (estado IN ('activo', 'desactivado', 'retirado', 'trasladado')),
    foto_clave TEXT,
    apoderado_nombre TEXT,
    apoderado_telefono TEXT,
    fecha_salida DATE,
    creado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
    actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK ((estado IN ('retirado', 'trasladado')) = (fecha_salida IS NOT NULL))
);

CREATE INDEX alumnos_seccion_idx ON alumnos (seccion_id);
CREATE INDEX alumnos_nombres_busqueda_idx ON alumnos (nombres_busqueda);
CREATE INDEX alumnos_apellidos_busqueda_idx ON alumnos (apellidos_busqueda);

-- +goose Down
DROP TABLE alumnos;
