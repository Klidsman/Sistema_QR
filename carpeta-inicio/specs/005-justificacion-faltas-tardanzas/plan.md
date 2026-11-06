# Plan de implementación: Justificación de faltas y tardanzas

**Branch**: `005-justificacion-faltas-tardanzas` | **Fecha**: 2026-11-06 | **Spec**: [spec.md](spec.md)

## Resumen

Nuevo módulo `internal/justificaciones` para registrar, aprobar, rechazar, anular y consultar
justificaciones de faltas y tardanzas oficiales. Una justificación cubre una o varias fechas; se
aprueba por el director o administrador y cambia el estado del registro sin borrar el dato
original. Al aprobar o anular se actualizan las alertas de la feature 004 como “situación
cambiada”.

## Decisiones

- **Datos**: tablas `justificaciones`, `justificacion_fechas` y `justificacion_adjuntos`.
  `marcaciones` recibe `justificada` para tardanzas justificadas.
- **Falta**: se representa por ausencia de marcación en un día laborable. La justificación de
  falta se guarda en `justificacion_fechas`.
- **Tardanza**: además de la fecha justificada, se marca `marcaciones.justificada = true`.
- **Duplicados**: índice único parcial por alumno y fecha para estados `pendiente` y `aprobada`.
- **Plazo**: nueva columna `plazo_justificacion_dias` en `parametros_asistencia` (5 por defecto).
- **Adjuntos**: se reutiliza `almacenamiento.AlmacenFotos`; máximo 3, PDF/JPG/PNG, 5 MB.
- **Permisos**: docente asignado a la sección en la fecha; director y administrador resuelven;
  solo el administrador anula.
- **Alertas**: al aprobar se llama `MarcarSituacionCambiada` y `ContarTardanzas` excluye
  justificadas. La justificación aprobada también responde `EsFaltaJustificada` a la feature 004.

## Fases

1. Migración, consultas y modelos.
2. Servicio de registro, resolución y anulación.
3. Manejador, plantillas y adjuntos.
4. Integración con alertas y configuración del plazo.
5. Pruebas unitarias, integración, manejadores y E2E.
