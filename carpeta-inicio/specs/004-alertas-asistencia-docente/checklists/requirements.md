# Specification Quality Checklist: Alertas internas a docentes por tardanzas e inasistencias

**Purpose**: Validar la completitud y calidad de la especificación antes de planificar
**Created**: 2026-11-05
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No hay detalles de implementación (lenguajes, frameworks, APIs)
- [x] Enfocada en el valor para el usuario y las necesidades del colegio
- [x] Escrita para stakeholders no técnicos
- [x] Todas las secciones obligatorias completadas

## Requirement Completeness

- [x] No quedan marcadores [NEEDS CLARIFICATION]
- [x] Los requisitos son verificables y no ambiguos
- [x] Los criterios de éxito son medibles
- [x] Los criterios de éxito son agnósticos de tecnología
- [x] Todos los escenarios de aceptación están definidos
- [x] Los casos límite están identificados
- [x] El alcance está claramente delimitado
- [x] Dependencias y supuestos identificados

## Feature Readiness

- [x] Todos los requisitos funcionales tienen criterios de aceptación claros
- [x] Los escenarios de usuario cubren los flujos principales
- [x] La funcionalidad cumple los resultados medibles de los criterios de éxito
- [x] Ningún detalle de implementación se filtra a la especificación

## Notes

- Asistencia oficial: el ingreso por QR de la puerta principal (feature 003, FR-023).
- Tardanza por QR: la alerta llega al portero que registró el ingreso (FR-001).
- Inasistencia diaria: no genera alerta a ningún usuario (FR-002); aparece en reportes.
- Reiteración: llega al docente asignado al curso y sección (FR-008).
- La alerta de discrepancia entre QR y docente fue eliminada, porque el registro del docente es
  personal.
- Pendiente de la feature 002: agregar el umbral de faltas seguidas a la configuración de
  umbrales, que hoy solo contempla conteos mensuales.
- Decisión de alcance: el Director no recibe alertas en esta versión (ver Assumptions).
- Lista para `/speckit.clarify` (si se quieren validar los supuestos) o `/speckit.plan`.
