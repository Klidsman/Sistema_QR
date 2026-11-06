-- +goose Up
CREATE TABLE alertas (
    id BIGSERIAL PRIMARY KEY,
    tipo TEXT NOT NULL CHECK (tipo IN ('tardanza', 'reiteracion_tardanzas', 'reiteracion_faltas')),
    alumno_id BIGINT NOT NULL REFERENCES alumnos (id),
    seccion_id BIGINT NOT NULL REFERENCES secciones (id),
    curso_id BIGINT REFERENCES cursos (id),
    bloque_numero INTEGER,
    fecha_evento DATE NOT NULL,
    mes DATE,
    marcacion_id BIGINT REFERENCES marcaciones (id),
    destinatario_id BIGINT NOT NULL REFERENCES usuarios (id),
    estado TEXT NOT NULL DEFAULT 'nueva' CHECK (estado IN ('nueva', 'vista')),
    leida_en TIMESTAMPTZ,
    leida_por BIGINT REFERENCES usuarios (id),
    sin_responsable BOOLEAN NOT NULL DEFAULT false,
    situacion_cambiada BOOLEAN NOT NULL DEFAULT false,
    motivo_cambio TEXT CHECK (motivo_cambio IS NULL OR char_length(motivo_cambio) <= 300),
    creada_en TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (tipo <> 'tardanza' OR marcacion_id IS NOT NULL),
    CHECK (tipo <> 'reiteracion_tardanzas' OR mes IS NOT NULL)
);

CREATE INDEX alertas_panel_idx ON alertas (destinatario_id, estado, creada_en DESC);
CREATE INDEX alertas_seccion_idx ON alertas (seccion_id, fecha_evento);
CREATE INDEX alertas_alumno_idx ON alertas (alumno_id, fecha_evento);

CREATE UNIQUE INDEX alertas_tardanza_unica
    ON alertas (marcacion_id, destinatario_id)
    WHERE tipo = 'tardanza';

CREATE UNIQUE INDEX alertas_reiteracion_tardanzas_unica
    ON alertas (alumno_id, mes, destinatario_id)
    WHERE tipo = 'reiteracion_tardanzas';

CREATE UNIQUE INDEX alertas_reiteracion_faltas_unica
    ON alertas (alumno_id, fecha_evento, destinatario_id)
    WHERE tipo = 'reiteracion_faltas';

-- +goose Down
DROP TABLE alertas;
