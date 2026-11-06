# Tareas: Justificación de faltas y tardanzas

**Input**: [spec.md](spec.md) y [plan.md](plan.md)

## Fase 1: Datos

- [x] T001 Crear migración `0028_crear_justificaciones.sql` con tablas, índices y `marcaciones.justificada`
- [x] T002 Agregar `plazo_justificacion_dias` a `parametros_asistencia`
- [x] T003 Crear consultas sqlc de justificaciones y ampliar asistencia
- [x] T004 Implementar repositorio y modelos del módulo

## Fase 2: Reglas

- [x] T005 Implementar registro por fecha y rango con omisiones
- [x] T006 Implementar plazo configurable y rechazo automático
- [x] T007 Implementar aprobación, rechazo y autoresolución bloqueada
- [x] T008 Implementar anulación solo para administrador
- [x] T009 Implementar historial y permisos por sección
- [x] T010 Implementar adjuntos PDF/JPG/PNG de hasta 5 MB
- [x] T011 Auditar creación, aprobación, rechazo y anulación
- [x] T012 Implementar `EsFaltaJustificada` y marcar alertas como situación cambiada

## Fase 3: Interfaz

- [x] T013 Montar rutas y pantallas de panel, formulario y detalle
- [x] T014 Agregar enlaces desde la ficha del alumno
- [x] T015 Agregar el plazo de justificación a `/configuracion/asistencia`
- [x] T016 Actualizar manuales de docente, administrador y director

## Fase 4: Pruebas

- [x] T017 Pruebas de integración del flujo completo, rechazo, anulación y duplicados
- [x] T018 Pruebas de manejadores, permisos y CSRF
- [x] T019 Prueba E2E de registro pendiente y bloqueo de autoaprobación
- [x] T020 Actualizar la prueba de migraciones a la versión 28
