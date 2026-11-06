-- +goose Up
CREATE TABLE cursos (
    id BIGSERIAL PRIMARY KEY,
    anio_escolar_id BIGINT NOT NULL REFERENCES anios_escolares (id),
    nombre TEXT NOT NULL CHECK (char_length(nombre) BETWEEN 2 AND 80),
    activo BOOLEAN NOT NULL DEFAULT true,
    creado_en TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX cursos_nombre_idx ON cursos (anio_escolar_id, lower(nombre));

CREATE TABLE cursos_grados (
    curso_id BIGINT NOT NULL REFERENCES cursos (id),
    grado_id BIGINT NOT NULL REFERENCES grados (id),
    PRIMARY KEY (curso_id, grado_id)
);

-- +goose Down
DROP TABLE cursos_grados;
DROP TABLE cursos;
