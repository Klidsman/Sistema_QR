# Plan de implementación: Registro de ingreso por QR en portería

**Branch**: `007-porteria-escaneo-qr` | **Fecha**: 2026-11-06 | **Spec**: [spec.md](spec.md)

## Resumen

Nuevo módulo `internal/porteria` que resuelve el token del QR, valida día, turno y horario,
clasifica la marcación y la registra en el módulo de asistencia con el usuario de portería. Si la
marcación es tardanza, dispara el proceso de alertas de la feature 004. La pantalla admite cámara
cuando el navegador soporta `BarcodeDetector` y entrada manual del token.

## Decisiones

- **Propiedad de datos**: `marcaciones` sigue siendo del módulo `asistencia`; portería usa
  `RegistrarMarcacionPorteria`.
- **Clasificación**: puntual hasta ingreso + tolerancia; tardanza hasta el límite; tardanza grave
  después del límite; fuera de la ventana ingreso–cierre se rechaza.
- **Duplicados**: el índice único `(alumno_id, fecha)` y `ON CONFLICT DO NOTHING` resuelven la
  concurrencia.
- **Alertas**: se ejecuta `ProcesarTardanzas` después de registrar; si falla, el ingreso queda
  registrado y el cron reintenta.
- **Cámara**: `BarcodeDetector` nativo; si no existe, se oculta el botón y se usa el campo manual.
- **Token de respaldo**: el administrador puede verlo en la ficha del alumno para digitarlo.

## Fases

1. Consulta de token y turno vigente.
2. Registro y auditoría de la marcación en asistencia.
3. Clasificación, servicio y pantalla de portería.
4. Rutas, navegación y pruebas.
