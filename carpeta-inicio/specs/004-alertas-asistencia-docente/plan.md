# Plan de implementación: Alertas internas por tardanzas e inasistencias

**Branch**: `004-alertas-asistencia-docente` | **Fecha**: 2026-11-06 | **Spec**: [spec.md](spec.md)

**Input**: Especificación de la feature en `/specs/004-alertas-asistencia-docente/spec.md`

## Resumen

El sistema genera alertas internas a partir de la asistencia oficial por QR. Una tardanza genera
una alerta para el portero que la registró. Cuando un alumno alcanza el umbral mensual de
tardanzas o el de faltas injustificadas seguidas, cada docente asignado a su sección en la fecha
recibe una alerta de reiteración. Las alertas se consultan en un panel con filtros y estado de
lectura; no se envían por canales externos. No existen alertas diarias de inasistencia.

La generación de tardanzas se hace con un proceso periódico que lee las marcaciones QR existentes
(la pantalla de portería queda para una feature posterior). Las ausencias se evalúan al cierre del
turno con un proceso diario. Ambas tareas son idempotentes.

## Contexto técnico

**Lenguaje y versión**: Go estable reciente, el mismo de las features 001–003.

**Dependencias principales**: Chi, templ, HTMX, Tailwind, pgx, sqlc, goose, scs, slog, testing,
testify, testcontainers, Playwright y **robfig/cron** para los procesos periódicos.

**Almacenamiento**: PostgreSQL. Nuevas tablas `alertas` y `cierres_alerta`; `marcaciones` recibe
`registrada_por`; `umbrales_alerta` admite el tipo `faltas_seguidas`.

**Pruebas**:
- Reglas puras: conteo mensual, racha de faltas, deduplicación y selección del umbral vigente.
- Repositorios con PostgreSQL real.
- Manejadores con `httptest`, permisos y CSRF.
- E2E: panel, filtros, marcar vistas e indicador.
- Procesos: ejecución repetida sin duplicados.

**Plataforma objetivo**: contenedor Docker en Linux; navegadores actuales.

**Metas de rendimiento**:
- Alerta de tardanza visible en menos de 1 minuto (SC-002).
- Panel con 100 alertas filtrado y marcado masivo en menos de 30 segundos (SC-006).
- Procesos diarios completos en menos de 30 segundos para 2,000 alumnos.

**Restricciones**:
- Solo alertas internas; sin correo, SMS, WhatsApp ni apoderados (FR-020).
- Sin alertas diarias de inasistencia (FR-002).
- Una alerta por situación y destinatario (FR-007).
- No se eliminan alertas; se marca “situación cambiada” (FR-015).
- Los umbrales tienen vigencia y no reescriben alertas anteriores (FR-016, FR-017).
- El director no recibe alertas en esta versión.
- El portero solo ve las alertas de tardanza de ingresos que él registró.
- Las justificaciones pertenecen a la feature 005; hasta entonces toda ausencia oficial cuenta
  como injustificada y la integración queda preparada.

## Revisión de la constitución

| Principio | Estado | Observación |
|---|---|---|
| I. Simplicidad | Cumple | Módulo nuevo con capas simples; una sola dependencia nueva justificada (cron ya exigido por la constitución). |
| II. Código Go y monolito | Cumple | `alertas` usa interfaces de `asistencia`, `alumnos` y `academico`; no consulta sus tablas. |
| III. Experiencia de usuario | Cumple | Panel mobile-first, indicador de nuevas y marcado masivo. |
| IV. Integridad y seguridad | Cumple | Alertas inmutables, lectura auditada, permisos por rol, CSRF y deduplicación por índice único. |
| V. Pruebas | Cumple | Reglas puras con tablas de casos, PostgreSQL real, manejadores y Playwright. Reloj inyectado. |
| VI. Reglas configurables | Cumple | Umbrales en base de datos con vigencia; ningún valor en el código. |
| VII. Datos de menores | Cumple | Solo datos necesarios; panel accesible a roles autorizados. |
| VIII. Entrega operable | Cumple | Migraciones goose, cron interno y documentación. |
| IX. Idioma y nomenclatura | Cumple | Nombres en español sin tildes; textos visibles con tildes. |

Resultado: sin violaciones no justificadas.

## Estructura del proyecto

```text
internal/alertas/
├── modelo.go
├── validaciones.go
├── servicio_alertas.go
├── servicio_procesos.go
├── repositorio.go
├── repositorio_alertas.go
├── manejador_alertas.go
├── manejador_indicador.go
├── programador.go            # robfig/cron: tardanzas y cierres
├── consultas/alertas.sql
└── generado/

internal/asistencia/
├── servicio_fuente_alertas.go  # Marcaciones y docentes por sección
└── consultas/asistencia.sql

internal/alumnos/
├── servicio_alertas.go         # Alumnos activos y sección en fecha
└── consultas/alumnos.sql

internal/academico/
├── consulta_alertas.go         # Secciones activas, calendario y umbrales
└── consultas/academico.sql

web/plantillas/alertas/
├── panel.templ
└── indicador.templ

migraciones/
├── 0023_agregar_portero_a_marcaciones.sql
├── 0024_crear_tabla_alertas.sql
├── 0025_crear_tabla_cierres_alerta.sql
├── 0026_ampliar_umbrales_faltas_seguidas.sql
└── 0027_ampliar_auditoria_alertas.sql
```

## Reglas de negocio

- **Tardanza**: por cada marcación con estado `tardanza` o `tardanza_grave` se crea una alerta
  para el usuario que la registró. Si la marcación no tiene portero, la alerta va al
  administrador como “sin responsable”.
- **Reiteración por tardanzas**: al alcanzar el umbral mensual vigente en la fecha, se crea una
  alerta por alumno, mes y docente. Solo una vez por mes y umbral.
- **Reiteración por faltas seguidas**: al cierre diario se cuentan los días laborables consecutivos
  sin marcación oficial, descontando feriados y suspensiones. Si alcanza el umbral vigente, se
  crea una alerta por alumno y fecha para cada docente asignado a la sección; sin docentes, al
  administrador.
- **Destinatarios**: portero para tardanzas; todos los docentes con asignación vigente a la
  sección para reiteraciones; administrador cuando no hay responsable.
- **Deduplicación**: índices únicos parciales por situación y destinatario.
- **Lectura**: estado `nueva` o `vista`; se guarda usuario y momento. El marcado masivo solo afecta
  el filtro activo.
- **Situación cambiada**: una corrección administrativa o justificación posterior marca la alerta
  y conserva el registro.
- **Procesos**: cron cada minuto para tardanzas; cron diario por turno para cierres. Ambos
  idempotentes y protegidos con advisory lock.

## Fases

1. Migraciones, modelos, reglas puras y repositorios.
2. Fuentes entre módulos e integración con umbrales.
3. Alertas de tardanza y proceso periódico.
4. Reiteración por tardanzas.
5. Reiteración por faltas seguidas y cierre diario.
6. Panel, filtros, indicador y marcado de vistas.
7. Asignación de responsable y “situación cambiada”.
8. Pruebas E2E, documentación y pulido.

## Complejidad justificada

| Elemento | Por qué se necesita | Alternativa descartada |
|---|---|---|
| Módulo `alertas` | Tiene datos, reglas y permisos propios. | Guardar alertas en `asistencia`: mezcla responsabilidades. |
| Proceso cron | FR-022 exige evaluar faltas al cierre y SC-002 exige latencia de un minuto. | Evaluar al abrir el panel: no genera alertas si nadie entra. |
| Tabla `cierres_alerta` | Evita repetir el cierre diario y sus alertas. | Confiar solo en índices únicos: el cálculo de rachas se repetiría y costaría más. |
| `registrada_por` en marcaciones | FR-001 necesita saber qué portero recibe la alerta. | Enviar a todos los porteros: la spec pide el que registró. |
| Tipo `faltas_seguidas` | FR-005 no es un umbral mensual. | Reutilizar `faltas` mensual: cambia la regla aprobada. |
