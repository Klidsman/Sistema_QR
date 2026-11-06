-- +goose Up
CREATE TABLE historial_secciones (
    id BIGSERIAL PRIMARY KEY,
    alumno_id BIGINT NOT NULL REFERENCES alumnos (id),
    seccion_id BIGINT NOT NULL REFERENCES secciones (id),
    fecha_inicio DATE NOT NULL,
    fecha_fin DATE,
    CHECK (fecha_fin IS NULL OR fecha_fin >= fecha_inicio)
);

CREATE UNIQUE INDEX historial_secciones_actual_idx
    ON historial_secciones (alumno_id)
    WHERE fecha_fin IS NULL;

-- +goose Down
DROP TABLE historial_secciones;
