# Investigación técnica: Alertas internas por tardanzas e inasistencias

**Fecha**: 2026-11-06 | **Plan**: [plan.md](plan.md)

## 1. Dependencia de portería

- **Decisión**: la feature lee la tabla `marcaciones` ya creada en la 003. Se agrega
  `registrada_por` para saber qué portero recibe la alerta de tardanza. La pantalla de escaneo
  sigue siendo de una feature posterior; mientras tanto, un proceso periódico convierte las
  marcaciones existentes en alertas.
- **Justificación**: la 004 no puede esperar a portería para existir; y no debe duplicar el
  escaneo. Cuando portería escriba marcaciones con `registrada_por`, las alertas fluirán sin
  cambios.
- **Alternativa**: generar la alerta desde el futuro manejador de escaneo. Se descarta porque
  acopla la 004 a una feature no planificada y no permite probar el flujo ahora.

## 2. Procesos periódicos

- **Decisión**: `robfig/cron` dentro de la aplicación.
  - Cada minuto: `ProcesarTardanzas` lee marcaciones de los últimos dos días sin alerta y crea
    las alertas de tardanza.
  - Cada día a la hora de cierre de cada turno: `ProcesarCierre(fecha, turno)` evalúa ausencias y
    reiteraciones.
- **Justificación**: la constitución exige cron interno; FR-022 pide evaluación al cierre y
  SC-002 una latencia menor a un minuto.
- **Concurrencia**: cada proceso toma `pg_advisory_xact_lock` con una clave fija y registra su
  ejecución en `cierres_alerta` (`fecha`, `turno`) para no repetir el cierre.
- **Alternativa**: evaluar todo al abrir el panel. Se descarta porque no genera alertas si nadie
  consulta y rompe SC-002.

## 3. Destinatarios

- **Tardanza**: el usuario `registrada_por` de la marcación. Si está vacío, el administrador.
- **Reiteraciones**: todos los docentes con asignación vigente a la sección del alumno en la fecha
  del evento. Si no hay ninguno, el administrador con `sin_responsable = true`.
- **Justificación**: FR-001 y FR-008. Varios docentes pueden estar asignados a la misma sección;
  cada uno recibe su propia alerta, sin buzones compartidos.
- **Alternativa**: un único “tutor” por sección. No existe ese perfil en la 003.

## 4. Faltas seguidas

- **Decisión**: al cierre del turno se cuentan los días laborables consecutivos sin marcación
  oficial, hacia atrás desde la fecha evaluada.
  - Se excluyen sábados, domingos, feriados, no laborables y suspensiones que alcancen a la
    sección.
  - Se corta la racha con una marcación oficial o con una falta justificada.
  - Se genera una alerta por alumno y fecha cuando la racha alcanza el umbral vigente.
- **Justificación**: FR-005, FR-018 y FR-022. Contar solo días laborables evita rachas falsas por
  fines de semana.
- **Justificaciones**: la feature 005 aún no existe. La 004 define la interfaz
  `EsFaltaJustificada(alumnoID, fecha)`; la implementación inicial devuelve falso y la 005 la
  reemplazará. Se documenta como limitación temporal.
- **Alternativa**: contar días calendario. Se descarta por generar rachas incorrectas.

## 5. Umbrales

- **Decisión**: reutilizar `umbrales_alerta` de la 002 y agregar el tipo `faltas_seguidas`.
  `tardanzas` sigue siendo mensual. El umbral vigente se elige por fecha con la regla de versión
  más reciente anterior o igual a la fecha.
- **Justificación**: FR-016 y FR-017. No se crea otra tabla ni otra pantalla de configuración.
- **Alternativa**: tabla nueva de umbrales. Duplica datos y reglas.

## 6. Deduplicación

- **Decisión**: índices únicos parciales en `alertas`:
  - `(marcacion_id, destinatario_id)` para `tardanza`.
  - `(alumno_id, mes, destinatario_id)` para `reiteracion_tardanzas`.
  - `(alumno_id, fecha_evento, destinatario_id)` para `reiteracion_faltas`.
- **Justificación**: FR-007 y SC-003. El proceso puede ejecutarse muchas veces sin duplicar.
- **Alternativa**: comprobar con `SELECT` antes de insertar. Tiene carreras; el índice lo resuelve.

## 7. Lectura y panel

- **Decisión**: estado `nueva`/`vista`, `leida_en` y `leida_por`. El panel filtra por rango de
  fechas, sección, tipo y estado. El marcado masivo actualiza solo los IDs del filtro activo.
- **Justificación**: FR-010 a FR-014, SC-006. Un `UPDATE ... WHERE id = ANY($ids)` evita marcar
  alertas fuera del filtro.
- **Indicador**: conteo de alertas nuevas del destinatario, consultado con HTMX cada 60 segundos
  y al cargar la página.
- **Alternativa**: marcar todo por usuario. Se descarta porque la spec exige respetar el filtro.

## 8. Situación cambiada

- **Decisión**: la 004 expone `MarcarSituacionCambiada(ctx, alumnoID, fecha)`; la llamará la 005
  al justificar y `asistencia` al corregir un registro administrativo. No se borra ninguna alerta.
- **Justificación**: FR-015 y FR-023.
- **Integración inicial**: la corrección administrativa de la 003 invocará esta interfaz cuando
  exista la conexión; se deja preparada y probada por servicio.

## 9. Límites entre módulos

- `alertas` no consulta tablas de otros módulos. Usa:
  - `asistencia`: marcaciones del rango y docentes asignados a una sección en una fecha.
  - `alumnos`: alumnos activos de una sección y sección del alumno en una fecha.
  - `academico`: secciones activas, calendario, umbrales vigentes.
- `alertas` es dueña de `alertas` y `cierres_alerta`.
- Sin ciclos de importación.

## 10. Rendimiento

- Índices por destinatario/estado/fecha y por sección/fecha.
- Procesos por lotes: una consulta de marcaciones y una de docentes por sección.
- Panel con 100 alertas: una consulta paginada y un `UPDATE` masivo por IDs.

## 11. Sin notificaciones externas

- **Decisión**: no hay correo, SMS, push ni WhatsApp. Todo ocurre en el panel interno.
- **Justificación**: FR-020 y alcance de la constitución.

## 12. Director

- **Decisión**: no recibe ni accede a alertas en esta versión.
- **Justificación**: supuesto de la spec. El director conserva sus pantallas de consulta de
  asistencia.
