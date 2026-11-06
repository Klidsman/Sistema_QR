-- +goose Up
CREATE TABLE credenciales_qr (
    id BIGSERIAL PRIMARY KEY,
    alumno_id BIGINT NOT NULL REFERENCES alumnos (id),
    token TEXT NOT NULL UNIQUE CHECK (length(token) = 43),
    estado TEXT NOT NULL DEFAULT 'activa' CHECK (estado IN ('activa', 'revocada')),
    motivo_revocacion TEXT CHECK (motivo_revocacion IN ('perdida', 'robo', 'desactivacion', 'retiro', 'otro')),
    emitida_en TIMESTAMPTZ NOT NULL DEFAULT now(),
    revocada_en TIMESTAMPTZ,
    CHECK ((estado = 'activa') = (revocada_en IS NULL)),
    CHECK ((estado = 'revocada') = (motivo_revocacion IS NOT NULL))
);

CREATE UNIQUE INDEX credenciales_qr_una_activa_idx
    ON credenciales_qr (alumno_id)
    WHERE estado = 'activa';

-- +goose Down
DROP TABLE credenciales_qr;
