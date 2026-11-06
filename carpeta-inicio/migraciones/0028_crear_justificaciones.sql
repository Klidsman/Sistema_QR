-- +goose Up
CREATE TABLE justificaciones (
    id BIGSERIAL PRIMARY KEY,
    alumno_id BIGINT NOT NULL REFERENCES alumnos (id),
    seccion_id BIGINT NOT NULL REFERENCES secciones (id),
    motivo TEXT NOT NULL CHECK (btrim(motivo) <> '' AND char_length(motivo) <= 500),
    estado TEXT NOT NULL DEFAULT 'pendiente'
        CHECK (estado IN ('pendiente', 'aprobada', 'rechazada', 'anulada', 'registro_no_aplica')),
    registrante_id BIGINT NOT NULL REFERENCES usuarios (id),
    resuelto_por BIGINT REFERENCES usuarios (id),
    resuelto_en TIMESTAMPTZ,
    comentario TEXT CHECK (comentario IS NULL OR char_length(comentario) <= 500),
    anulado_por BIGINT REFERENCES usuarios (id),
    anulado_en TIMESTAMPTZ,
    motivo_anulacion TEXT CHECK (motivo_anulacion IS NULL OR char_length(motivo_anulacion) <= 500),
    creada_en TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX justificaciones_alumno_idx ON justificaciones (alumno_id, creada_en DESC);
CREATE INDEX justificaciones_estado_idx ON justificaciones (estado, creada_en DESC);

CREATE TABLE justificacion_fechas (
    id BIGSERIAL PRIMARY KEY,
    justificacion_id BIGINT NOT NULL REFERENCES justificaciones (id) ON DELETE CASCADE,
    alumno_id BIGINT NOT NULL REFERENCES alumnos (id),
    fecha_evento DATE NOT NULL,
    tipo TEXT NOT NULL CHECK (tipo IN ('falta', 'tardanza')),
    estado TEXT NOT NULL DEFAULT 'pendiente'
        CHECK (estado IN ('pendiente', 'aprobada', 'rechazada', 'anulada')),
    UNIQUE (justificacion_id, fecha_evento)
);

CREATE UNIQUE INDEX justificacion_fechas_activa_unica
    ON justificacion_fechas (alumno_id, fecha_evento)
    WHERE estado IN ('pendiente', 'aprobada');

CREATE INDEX justificacion_fechas_busqueda_idx
    ON justificacion_fechas (alumno_id, fecha_evento);

CREATE TABLE justificacion_adjuntos (
    id BIGSERIAL PRIMARY KEY,
    justificacion_id BIGINT NOT NULL REFERENCES justificaciones (id) ON DELETE CASCADE,
    nombre TEXT NOT NULL CHECK (btrim(nombre) <> ''),
    tipo TEXT NOT NULL CHECK (tipo IN ('application/pdf', 'image/jpeg', 'image/png')),
    tamano BIGINT NOT NULL CHECK (tamano > 0 AND tamano <= 5242880),
    clave TEXT NOT NULL UNIQUE,
    creado_en TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX justificacion_adjuntos_idx ON justificacion_adjuntos (justificacion_id);

ALTER TABLE marcaciones ADD COLUMN justificada BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE parametros_asistencia
    ADD COLUMN plazo_justificacion_dias INTEGER NOT NULL DEFAULT 5
        CHECK (plazo_justificacion_dias BETWEEN 1 AND 60);

ALTER TABLE auditoria DROP CONSTRAINT auditoria_entidad_check;

ALTER TABLE auditoria ADD CONSTRAINT auditoria_entidad_check CHECK (entidad IN (
    'alumno', 'credencial_qr', 'historial_seccion', 'importacion',
    'anio_escolar', 'grado', 'seccion', 'turno', 'calendario', 'umbral',
    'curso', 'bloque_clase', 'horario_seccion',
    'docente', 'asignacion_docente', 'registro_docente', 'parametro_asistencia', 'marcacion',
    'alerta', 'cierre_alerta', 'justificacion'
));

-- +goose Down
ALTER TABLE auditoria DROP CONSTRAINT auditoria_entidad_check;

ALTER TABLE auditoria ADD CONSTRAINT auditoria_entidad_check CHECK (entidad IN (
    'alumno', 'credencial_qr', 'historial_seccion', 'importacion',
    'anio_escolar', 'grado', 'seccion', 'turno', 'calendario', 'umbral',
    'curso', 'bloque_clase', 'horario_seccion',
    'docente', 'asignacion_docente', 'registro_docente', 'parametro_asistencia', 'marcacion',
    'alerta', 'cierre_alerta'
));

ALTER TABLE parametros_asistencia DROP COLUMN plazo_justificacion_dias;
ALTER TABLE marcaciones DROP COLUMN justificada;
DROP TABLE justificacion_adjuntos;
DROP TABLE justificacion_fechas;
DROP TABLE justificaciones;
