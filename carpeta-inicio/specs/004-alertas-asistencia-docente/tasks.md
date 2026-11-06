# Tareas: Alertas internas por tardanzas e inasistencias

**Input**: Documentos de diseño en `/specs/004-alertas-asistencia-docente/`

**Prerrequisitos**: [plan.md](plan.md), [spec.md](spec.md), [research.md](research.md), [data-model.md](data-model.md), [contracts/rutas.md](contracts/rutas.md)

**Pruebas**: obligatorias según la constitución. Se escriben antes de cada regla.

## Formato: `[ID] [P?] [Historia] Descripción`

---

## Fase 1: Preparación

- [ ] T001 [P] Crear `internal/alertas/{consultas,generado}` y `web/plantillas/alertas/`
- [ ] T002 [P] Agregar el paquete de alertas a `sqlc.yaml` con salida en `internal/alertas/generado`
- [ ] T003 [P] Crear los tipos, estados y constantes de alerta en `internal/alertas/modelo.go`
- [ ] T004 [P] Ampliar `docs/glosario.md` con `alerta`, `reiteracion_tardanzas`, `faltas_seguidas` y `situacion_cambiada`
- [ ] T005 Agregar `github.com/robfig/cron/v3` a `go.mod` y justificarlo en `docs/despliegue.md`

---

## Fase 2: Base bloqueante

### Migraciones

- [ ] T006 [P] Crear `migraciones/0023_agregar_portero_a_marcaciones.sql` con `registrada_por` e índice por fecha
- [ ] T007 [P] Crear `migraciones/0024_crear_tabla_alertas.sql` con tipos, estados, lectura, situación cambiada e índices únicos parciales
- [ ] T008 [P] Crear `migraciones/0025_crear_tabla_cierres_alerta.sql` con `UNIQUE (fecha, turno)`
- [ ] T009 [P] Crear `migraciones/0026_ampliar_umbrales_faltas_seguidas.sql` con el nuevo tipo
- [ ] T010 [P] Crear `migraciones/0027_ampliar_auditoria_alertas.sql` con `alerta` y `cierre_alerta`
- [ ] T011 Actualizar la prueba de migraciones hasta la versión 27 en `pruebas/integracion/migraciones_test.go`

### Modelos, consultas y repositorios

- [ ] T012 [P] Implementar los modelos `Alerta`, `FiltrosAlerta`, `EventoMarcacion` y `EstadoAlerta` en `internal/alertas/modelo.go`
- [ ] T013 [P] Escribir las consultas sqlc de alertas, conteos y cierres en `internal/alertas/consultas/alertas.sql`
- [ ] T014 Regenerar sqlc y compilar `internal/alertas/generado/`
- [ ] T015 [P] Implementar `internal/alertas/repositorio.go` con pool, consultas y transacciones
- [ ] T016 [P] Implementar el repositorio de alertas, filtros y lectura en `internal/alertas/repositorio_alertas.go`
- [ ] T017 [P] Implementar el repositorio de cierres en `internal/alertas/repositorio_cierres.go`

### Interfaces entre módulos

- [ ] T018 [P] Implementar en `internal/asistencia/servicio_fuente_alertas.go`: marcaciones sin alerta y docentes de una sección en una fecha
- [ ] T019 [P] Implementar en `internal/alumnos/servicio_alertas.go`: alumnos activos de una sección y sección del alumno en una fecha
- [ ] T020 [P] Implementar en `internal/academico/consulta_alertas.go`: secciones activas, día laborable y umbral vigente
- [ ] T021 [P] Crear las interfaces de fuente en `internal/alertas/dependencias.go`
- [ ] T022 Implementar la auditoría de alertas y cierres en `internal/alertas/servicio_auditoria.go`

### Reglas puras

- [ ] T023 [P] Escribir pruebas unitarias de conteo mensual y racha de faltas en `pruebas/unidad/alertas_reglas_test.go`
- [ ] T024 [P] Escribir pruebas unitarias de deduplicación y destinatarios en `pruebas/unidad/alertas_destinatarios_test.go`
- [ ] T025 Implementar las reglas puras en `internal/alertas/validaciones.go`
- [ ] T026 Escribir la prueba de migraciones y parámetros en `pruebas/integracion/alertas_migraciones_test.go`
- [ ] T027 Escribir la prueba base de permisos por rol en `pruebas/manejadores/alertas_permisos_test.go`

### Integración y rutas

- [ ] T028 Registrar el módulo, dependencias y rutas base en `internal/plataforma/rutas/rutas.go` y `cmd/servidor/main.go`
- [ ] T029 Implementar el programador cron con advisory lock en `internal/alertas/programador.go`
- [ ] T030 Escribir la prueba del programador idempotente en `pruebas/integracion/alertas_programador_test.go`

---

## Fase 3: Historia 1 — Alerta de tardanza al portero (P1)

**Objetivo**: generar una alerta por cada marcación de tardanza para el portero que la registró.

**Prueba independiente**: insertar una marcación de tardanza con portero y verificar una única alerta para él.

- [ ] T031 [P] [US1] Escribir pruebas unitarias de destinatario y deduplicación de tardanza en `pruebas/unidad/alertas_tardanza_test.go`
- [ ] T032 [P] [US1] Escribir pruebas de integración de generación de alerta de tardanza en `pruebas/integracion/alertas_tardanza_test.go`
- [ ] T033 [US1] Implementar `ProcesarTardanzas` en `internal/alertas/servicio_procesos.go`
- [ ] T034 [US1] Asignar la alerta al portero de `registrada_por`, o al administrador si falta
- [ ] T035 [US1] Insertar con `ON CONFLICT DO NOTHING` para no duplicar
- [ ] T036 [US1] Auditar `crear_alerta` de tardanza
- [ ] T037 [US1] Implementar el indicador de alertas nuevas en `internal/alertas/manejador_indicador.go`
- [ ] T038 [US1] Mostrar el indicador en las plantillas base de los roles autorizados en `web/plantillas/paginas/base.templ`
- [ ] T039 [US1] Escribir pruebas de manejadores del indicador en `pruebas/manejadores/alertas_indicador_test.go`
- [ ] T040 [US1] Escribir la prueba E2E de alerta de tardanza para portero en `pruebas/e2e/alertas_tardanza_test.spec.ts`

---

## Fase 4: Historia 2 — Reiteraciones por umbral (P1)

**Objetivo**: alertar a los docentes cuando un alumno cruza el umbral de tardanzas mensuales o de faltas seguidas.

**Prueba independiente**: 3 tardanzas generan una sola alerta; 2 faltas seguidas generan una sola alerta; una falta justificada no cuenta.

- [ ] T041 [P] [US2] Escribir pruebas unitarias del umbral mensual en `pruebas/unidad/alertas_reiteracion_tardanzas_test.go`
- [ ] T042 [P] [US2] Escribir pruebas de integración de reiteración por tardanzas en `pruebas/integracion/alertas_reiteracion_tardanzas_test.go`
- [ ] T043 [US2] Implementar el conteo mensual de tardanzas y la selección del umbral vigente en `internal/alertas/servicio_alertas.go`
- [ ] T044 [US2] Crear la alerta de reiteración por tardanzas para cada docente asignado a la sección en `internal/alertas/servicio_alertas.go`
- [ ] T045 [US2] Auditar `crear_alerta` de reiteración por tardanzas
- [ ] T046 [P] [US2] Escribir pruebas unitarias de faltas seguidas en `pruebas/unidad/alertas_faltas_seguidas_test.go`
- [ ] T047 [P] [US2] Escribir pruebas de integración de faltas seguidas en `pruebas/integracion/alertas_faltas_seguidas_test.go`
- [ ] T048 [US2] Implementar el cálculo de la racha de días laborables sin marcación en `internal/alertas/servicio_alertas.go`
- [ ] T049 [US2] Excluir fines de semana, feriados y suspensiones del cálculo en `internal/alertas/servicio_alertas.go`
- [ ] T050 [US2] Implementar la interfaz `EsFaltaJustificada` con implementación inicial falsa y punto de extensión para la 005
- [ ] T051 [US2] Implementar `ProcesarCierre(fecha, turno)` con `cierres_alerta` en `internal/alertas/servicio_procesos.go`
- [ ] T052 [US2] Enviar al administrador las alertas sin docente asignado con `sin_responsable`
- [ ] T053 [US2] Auditar `ejecutar_cierre` y las alertas de faltas
- [ ] T054 [US2] Escribir pruebas de manejadores de cierre y umbrales en `pruebas/manejadores/alertas_cierre_test.go`
- [ ] T055 [US2] Escribir la prueba E2E de reiteración en `pruebas/e2e/alertas_reiteracion_test.spec.ts`

---

## Fase 5: Historia 3 — Panel, filtros y lectura (P2)

**Objetivo**: panel de alertas con filtros, contador de nuevas y marcado individual o masivo.

**Prueba independiente**: filtrar por sección y tipo, marcar una y luego todas las del filtro.

- [ ] T056 [P] [US3] Escribir pruebas unitarias de filtros y permisos en `pruebas/unidad/alertas_filtros_test.go`
- [ ] T057 [P] [US3] Escribir pruebas de integración del panel y marcado masivo en `pruebas/integracion/alertas_panel_test.go`
- [ ] T058 [US3] Implementar `ListarAlertas` con filtros y alcance por rol en `internal/alertas/servicio_alertas.go`
- [ ] T059 [US3] Implementar `MarcarVista` y `MarcarVistas` en `internal/alertas/servicio_alertas.go`
- [ ] T060 [US3] Implementar las rutas del panel y marcado en `internal/alertas/manejador_alertas.go`
- [ ] T061 [US3] Crear la plantilla `panel.templ` con filtros y acciones masivas
- [ ] T062 [US3] Crear la plantilla `indicador.templ` y su actualización HTMX
- [ ] T063 [US3] Implementar el contador de nuevas por destinatario en `internal/alertas/repositorio_alertas.go`
- [ ] T064 [US3] Auditar `marcar_vista` y `marcar_vistas`
- [ ] T065 [US3] Escribir pruebas de manejadores del panel en `pruebas/manejadores/alertas_panel_test.go`
- [ ] T066 [US3] Escribir la prueba E2E de panel, filtros y lectura en `pruebas/e2e/alertas_panel_test.spec.ts`

---

## Fase 6: Historia 4 — Umbrales configurables (P3)

**Objetivo**: configurar tardanzas mensuales y faltas seguidas con vigencia e historial.

- [ ] T067 [P] [US4] Escribir pruebas unitarias de umbrales y vigencia en `pruebas/unidad/alertas_umbrales_test.go`
- [ ] T068 [P] [US4] Escribir pruebas de integración de umbrales con vigencia en `pruebas/integracion/alertas_umbrales_test.go`
- [ ] T069 [US4] Ampliar `ValidarUmbral` y el servicio académico para el tipo `faltas_seguidas`
- [ ] T070 [US4] Agregar la opción “Faltas seguidas” a la plantilla `web/plantillas/configuracion/umbrales.templ`
- [ ] T071 [US4] Verificar que un cambio de umbral no modifica alertas anteriores en `pruebas/integracion/alertas_umbrales_test.go`
- [ ] T072 [US4] Escribir la prueba E2E de configuración de faltas seguidas en `pruebas/e2e/alertas_umbrales_test.spec.ts`

---

## Fase 7: Pulido y transversales

- [ ] T073 [P] Completar la prueba de permisos: docente, portero, administrador y director en `pruebas/manejadores/alertas_permisos_test.go`
- [ ] T074 [P] Escribir la prueba de CSRF de alertas en `pruebas/manejadores/alertas_seguridad_test.go`
- [ ] T075 Escribir la prueba de auditoría de todas las acciones en `pruebas/integracion/alertas_auditoria_test.go`
- [ ] T076 Verificar que las alertas no se eliminan y que “situación cambiada” conserva el registro en `pruebas/integracion/alertas_situacion_test.go`
- [ ] T077 Implementar `MarcarSituacionCambiada` y su integración con la corrección de asistencia en `internal/alertas/servicio_alertas.go`
- [ ] T078 Verificar la idempotencia de los dos procesos con ejecuciones repetidas y concurrentes en `pruebas/integracion/alertas_programador_test.go`
- [ ] T079 Revisar `EXPLAIN ANALYZE` del panel y de los conteos con datos de 2,000 alumnos
- [ ] T080 Actualizar `docs/manual_administrador.md`, `docs/manual_director.md` y `docs/manual_docente.md`; crear `docs/manual_portero.md`
- [ ] T081 Actualizar `docs/despliegue.md` con el cron, el tipo de umbral y la ausencia de variables nuevas
- [ ] T082 Regenerar plantillas, CSS y formato; ejecutar `golangci-lint`, `go test` y Playwright
- [ ] T083 Validar los escenarios de `quickstart.md` y registrar resultados
- [ ] T084 Verificar compatibilidad completa con las features 001, 002 y 003 en `pruebas/e2e/`

---

## Dependencias

| Historia | Depende de | Motivo |
|---|---|---|
| US1 tardanza | Base | Necesita marcaciones, alertas y programador. |
| US2 reiteraciones | US1 y Base | Usa marcaciones y destinatarios docentes. |
| US3 panel | US1 y US2 | Muestra las alertas generadas. |
| US4 umbrales | US2 | El umbral de faltas seguidas alimenta la reiteración. |
| Pulido | Todas | Cierra la entrega. |

## Estrategia

1. Base, migraciones e interfaces.
2. US1: tardanzas al portero e indicador.
3. US2: reiteraciones y cierre diario.
4. US3: panel y lectura.
5. US4: configuración de umbrales.
6. Pulido, documentación y validación.

## Notas

- No hay notificaciones externas.
- Las justificaciones quedan integradas por interfaz y se completan en la 005.
- Las alertas no se eliminan.
- Cada tarea o grupo lógico se confirma con un commit en español.
