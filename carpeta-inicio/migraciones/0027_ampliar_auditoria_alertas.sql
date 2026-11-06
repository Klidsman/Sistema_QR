-- +goose Up
ALTER TABLE auditoria DROP CONSTRAINT auditoria_entidad_check;

ALTER TABLE auditoria ADD CONSTRAINT auditoria_entidad_check CHECK (entidad IN (
    'alumno', 'credencial_qr', 'historial_seccion', 'importacion',
    'anio_escolar', 'grado', 'seccion', 'turno', 'calendario', 'umbral',
    'curso', 'bloque_clase', 'horario_seccion',
    'docente', 'asignacion_docente', 'registro_docente', 'parametro_asistencia', 'marcacion',
    'alerta', 'cierre_alerta'
));

-- +goose Down
ALTER TABLE auditoria DROP CONSTRAINT auditoria_entidad_check;

ALTER TABLE auditoria ADD CONSTRAINT auditoria_entidad_check CHECK (entidad IN (
    'alumno', 'credencial_qr', 'historial_seccion', 'importacion',
    'anio_escolar', 'grado', 'seccion', 'turno', 'calendario', 'umbral',
    'curso', 'bloque_clase', 'horario_seccion',
    'docente', 'asignacion_docente', 'registro_docente', 'parametro_asistencia', 'marcacion'
));
