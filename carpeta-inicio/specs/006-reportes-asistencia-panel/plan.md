# Plan de implementación: Reportes de asistencia y panel

**Branch**: `006-reportes-asistencia-panel` | **Fecha**: 2026-11-06 | **Spec**: [spec.md](spec.md)

## Resumen

Nuevo módulo `internal/reportes` que calcula vistas sobre las marcaciones oficiales, las
justificaciones aprobadas, el calendario y la matrícula por sección. No crea tablas de totales.

## Decisiones

- **Cálculo**: las consultas leen las tablas existentes y el servicio agrega en Go para respetar
  calendario, cambios de sección y porcentajes.
- **Estados**: presente, tardanza, falta, falta justificada, tardanza justificada y pendiente.
- **Panel**: presentes incluye tardanzas; ausentes cuenta solo faltas injustificadas; pendientes
  aplica hasta el cierre del turno.
- **Porcentaje**: asistencias (presentes y tardanzas) dividido entre días laborables, restando
  faltas y tardanzas justificadas del denominador.
- **Docente**: solo secciones con asignación vigente en el periodo.
- **Reiteración**: usa los umbrales de la configuración vigentes al final del periodo.
- **Exportación**: Excel con `excelize` y PDF con `go-pdf/fpdf` y la fuente Go con soporte UTF-8.

## Rutas

- `/reportes/diario`, `/reportes/panel`, `/reportes/resumen`, `/reportes/reiteracion`.
- Exportaciones `diario.xlsx`, `diario.pdf`, `resumen.xlsx`, `resumen.pdf`.
