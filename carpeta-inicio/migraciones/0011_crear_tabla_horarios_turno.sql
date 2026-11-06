-- +goose Up
CREATE TABLE horarios_turno (
    id BIGSERIAL PRIMARY KEY,
    turno TEXT NOT NULL CHECK (turno IN ('manana', 'tarde')),
    hora_ingreso TIME NOT NULL,
    tolerancia_minutos INTEGER NOT NULL CHECK (tolerancia_minutos >= 0 AND tolerancia_minutos <= 120),
    hora_limite_tardanza TIME NOT NULL,
    hora_cierre TIME NOT NULL,
    vigente_desde DATE NOT NULL,
    creado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (turno, vigente_desde),
    CHECK (hora_limite_tardanza > hora_ingreso),
    CHECK (hora_cierre > hora_limite_tardanza),
    CHECK (hora_ingreso + make_interval(mins => tolerancia_minutos) <= hora_limite_tardanza)
);

CREATE INDEX horarios_turno_vigencia_idx ON horarios_turno (turno, vigente_desde DESC);

-- +goose Down
DROP TABLE horarios_turno;
