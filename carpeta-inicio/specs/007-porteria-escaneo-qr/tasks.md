# Tareas: Registro de ingreso por QR en portería

**Input**: [spec.md](spec.md) y [plan.md](plan.md)

## Fase 1: Acceso a datos

- [x] T001 [P] Agregar `AlumnoPorTokenQR` en `internal/alumnos/consultas/alumnos.sql`
- [x] T002 [P] Exponer `TurnoDeSeccion` en `internal/academico/consulta_alertas.go`
- [x] T003 [P] Agregar `CrearMarcacionPorteria` en `internal/asistencia/consultas/asistencia.sql`
- [x] T004 Implementar `RegistrarMarcacionPorteria` con auditoría en `internal/asistencia/servicio_porteria.go`

## Fase 2: Módulo de portería

- [x] T005 [P] Crear modelos, errores e interfaces en `internal/porteria/modelo.go`
- [x] T006 [P] Implementar `ClasificarEstado` y `Registrar` en `internal/porteria/servicio.go`
- [x] T007 [P] Crear `manejador.go` con rutas `/porteria` y `/porteria/ingreso`
- [x] T008 [P] Crear la plantilla `web/plantillas/porteria/porteria.templ` con cámara y entrada manual
- [x] T009 Mostrar el token de respaldo en la ficha del alumno para administradores

## Fase 3: Integración

- [x] T010 Registrar el módulo en `internal/plataforma/rutas/rutas.go` y `cmd/servidor/main.go`
- [x] T011 Disparar `ProcesarTardanzas` después de una tardanza
- [x] T012 Agregar el enlace **Portería** en la barra para administrador y portería
- [x] T013 Actualizar `docs/manual_portero.md`

## Fase 4: Pruebas

- [x] T014 [P] Pruebas unitarias de clasificación en `pruebas/unidad/porteria_test.go`
- [x] T015 [P] Pruebas del servicio con fuentes falsas en `pruebas/integracion/porteria_test.go`
- [x] T016 [P] Pruebas de manejador, roles y CSRF en `pruebas/manejadores/porteria_test.go`
- [x] T017 [P] E2E de ingreso, rechazo y alerta en `pruebas/e2e/alertas_tardanza_test.spec.ts`
- [x] T018 [P] E2E de reiteración por tardanzas en `pruebas/e2e/alertas_reiteracion_test.spec.ts`
