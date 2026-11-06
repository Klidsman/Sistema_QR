# Guía de validación rápida: Alertas internas

**Fecha**: 2026-11-06 | **Plan**: [plan.md](plan.md)

## Requisitos previos

- Features 001 a 003 aplicadas (migraciones hasta la 22).
- PostgreSQL, Go, Node y `make`.
- Docente asignado a una sección, alumnos activos y horario/bloques configurados.

## Preparar el entorno

1. `docker compose up -d postgres`.
2. `make migrar` (incluye `0023` a `0027`).
3. `make ejecutar` y abrir `http://localhost:8099/ingreso`.

## Escenario 1: alerta de tardanza al portero

1. Insertar una marcación de prueba con estado `tardanza` y `registrada_por` de un usuario con rol
   portero (o usar la futura pantalla de portería).
2. Esperar la ejecución del proceso de tardanzas o ejecutarlo manualmente.
3. Ingresar como ese portero y abrir `/alertas`.
4. Verificar que aparece una alerta de tipo tardanza con alumno, sección, fecha y hora.
5. Verificar que el docente de la sección no recibe esa alerta.

Resultado esperado: solo el portero que registró el ingreso ve la alerta.

## Escenario 2: sin alerta diaria de inasistencia

1. Insertar un alumno sin marcación y ejecutar el cierre del turno.
2. Verificar que no se crea ninguna alerta diaria de inasistencia para docentes ni administrador.
3. Verificar que la falta aparece en la consulta de asistencia.

## Escenario 3: reiteración por tardanzas

1. Configurar un umbral de 3 tardanzas por mes.
2. Registrar tres tardanzas del mismo alumno en el mes.
3. Verificar que se genera **una sola** alerta de reiteración, en la fecha de la tercera tardanza.
4. Registrar una cuarta tardanza y verificar que no se duplica.

## Escenario 4: reiteración por faltas seguidas

1. Configurar un umbral de 2 faltas seguidas.
2. Dejar al alumno sin marcación dos días laborables consecutivos.
3. Ejecutar el cierre del segundo día.
4. Verificar que el docente de la sección recibe una alerta de reiteración por faltas seguidas.
5. Repetir el cierre y verificar que no se duplica.
6. Si no hay docente asignado, verificar que la alerta llega al administrador con “sin responsable”.

## Escenario 5: panel, filtros y lectura

1. Como docente, abrir `/alertas` con alertas de dos secciones y tipos mixtos.
2. Filtrar por sección y tipo; verificar que solo aparecen las que cumplen ambos.
3. Marcar una alerta como vista y comprobar que el contador baja en uno.
4. Marcar todas las del filtro como vistas y verificar que las demás conservan su estado.
5. Comprobar que el administrador ve todas las secciones.

## Escenario 6: situación cambiada

1. Generar una alerta de reiteración.
2. Corregir o justificar la falta que la originó.
3. Verificar que la alerta se marca como “situación cambiada” y no se elimina.

## Pruebas automáticas

```sh
go test ./...
cd pruebas/e2e && npx playwright test
```

Cobertura esperada: reglas puras, repositorios, procesos idempotentes, panel, permisos, CSRF,
indicador y E2E.

## Resultados de rendimiento

Revisión ejecutada con `PRUEBA_RENDIMIENTO=1` y 2,000 alumnos en PostgreSQL 16:

| Consulta | Plan | Tiempo |
|---|---|---|
| Panel de alertas con filtros | `Index Scan` sobre `alertas_panel_idx` | 0.104 ms |
| Conteo mensual de tardanzas | `Index Scan` sobre `marcaciones_alumno_id_fecha_key` | 0.050 ms |

Ambas consultas usan índices y no requieren cambios.

## Criterios de aceptación

- 100% de tardanzas oficiales genera alerta al portero.
- Ninguna falta diaria genera alerta.
- Ninguna situación genera más de una alerta.
- 100% de cruces de umbral genera reiteración, incluido el último día del mes.
- Un docente nunca ve secciones no asignadas.
- 100% de alertas sin responsable llega al administrador.
