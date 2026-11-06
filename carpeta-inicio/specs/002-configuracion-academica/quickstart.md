# Guía de validación rápida: Configuración académica del colegio

**Fecha**: 2026-11-06 | **Plan**: [plan.md](plan.md)

Esta guía comprueba que la configuración académica funciona de extremo a extremo en un entorno
local. No describe implementación; eso queda en `tasks.md`.

## Requisitos previos

- Docker y Docker Compose.
- Go estable reciente, Node y `make`.
- Un navegador moderno.

## Preparar el entorno

1. Copiar `.env.ejemplo` a `.env` y completar `CLAVE_SESION`, `BASE_DATOS_URL` y `ADMIN_*`.
2. Levantar PostgreSQL: `docker compose up -d postgres` (puerto 5433).
3. Aplicar migraciones (incluye `0009` a `0017`): `make migrar`.
4. Crear el administrador y las secciones de demostración: `make sembrar`.
5. Compilar estilos y ejecutar: `make ejecutar`. Abrir `http://localhost:8099/ingreso`.
6. Ingresar con `admin@colegio.pe` y la contraseña definida en `ADMIN_CONTRASENIA`.

Para las pruebas del director hace falta un usuario con rol `director`. Crearlo desde la consola
del servidor o con una consulta de desarrollo; el manual del administrador de la feature 001 no
cubre alta de usuarios.

## Escenario 1: año escolar, grados, secciones y turnos (US1)

1. Abrir `/configuracion/anio`. Si no hay año activo, crear `2027` con fecha de inicio
   `2027-03-01` y fecha de fin `2027-12-15`.
2. Verificar que el año queda activo y que el panel `/configuracion` lo muestra.
3. Intentar crear otro año activo: el sistema debe responder que ya existe uno activo.
4. En `/configuracion/grados`, crear `3.°` y `4.°`.
5. En `/configuracion/secciones`, crear `A` y `B` para `3.°`, y asignar `A` a mañana y `B` a
   tarde. Repetir para `4.°`.
6. Verificar en `/configuracion/secciones` que cada sección muestra su turno y en
   `/configuracion/turnos` que cada turno lista sus secciones.
7. Intentar guardar un año con fecha de fin anterior a la de inicio y confirmar el rechazo.

Resultado esperado: estructura completa visible, un solo año activo y fechas inválidas rechazadas.

## Escenario 2: horarios y tolerancias con vigencia (US2)

1. Abrir `/configuracion/turnos`.
2. Configurar mañana: ingreso `07:30`, tolerancia `10`, límite de tardanza `08:00`, cierre
   `08:30`, vigencia desde hoy.
3. Configurar tarde: ingreso `12:45`, tolerancia `10`, límite `13:15`, cierre `13:45`.
4. Intentar guardar cierre `07:50` con límite `08:00` y confirmar el rechazo con motivo.
5. Intentar guardar tolerancia `40` con ingreso `07:30` y límite `08:00`; debe rechazarse.
6. Guardar un cambio de tolerancia a `5` con vigencia desde una fecha futura.
7. Abrir `/configuracion/consulta?fecha=<fecha anterior al cambio>` y verificar que muestra la
   tolerancia de `10`.
8. Abrir `/configuracion/consulta?fecha=<fecha del cambio>` y verificar que muestra `5`.
9. Ingresar como director y ajustar la hora de ingreso del turno de mañana. Verificar que el
   cambio se guarda y queda auditado. Intentar crear un año o un grado: `403`.

Resultado esperado: los cuatro valores se guardan con vigencia; las combinaciones inválidas se
rechazan con mensaje; la consulta histórica muestra el valor correcto en cada fecha.

## Escenario 3: feriados y suspensiones (US3)

1. Abrir `/configuracion/calendario`.
2. Registrar un feriado para todo el colegio el `2027-07-28` con motivo "Fiestas Patrias".
3. Registrar una suspensión del `2027-05-10` al `2027-05-12` solo para `4.°`.
4. Verificar que el calendario muestra ambos registros con su alcance.
5. Registrar un feriado en sábado o domingo y verificar que se muestra sin alterar días
   laborables.
6. Registrar una excepción que se superpone con otra del mismo alcance y confirmar que aparece la
   pantalla de confirmación antes de guardar.
7. Intentar registrar una suspensión para un grado inexistente y confirmar el rechazo.

Resultado esperado: excepciones guardadas con alcance correcto y solapamientos advertidos.

## Escenario 4: umbrales de alerta (US4)

1. Abrir `/configuracion/umbrales`.
2. Configurar `3` tardanzas por mes desde hoy.
3. Cambiarlo a `4` con vigencia desde una fecha futura.
4. Verificar que la historia muestra ambos valores y sus fechas.
5. Intentar guardar `0` o un valor negativo y confirmar el rechazo.

Resultado esperado: umbrales con vigencia e historia; valores no positivos rechazados.

## Escenario 5: consulta histórica (US5)

1. Abrir `/configuracion/consulta?fecha=<fecha dentro del año>` y revisar horarios, excepciones,
   umbrales y turno de cada sección.
2. Consultar una fecha anterior al inicio del año y verificar el mensaje de que no hay
   configuración vigente.
3. Cambiar el turno de una sección con vigencia desde una fecha futura y consultar antes y
   después del cambio: cada fecha debe mostrar el turno correspondiente.

Resultado esperado: la consulta devuelve la configuración vigente en cada fecha.

## Escenario 6: bloques de hora de clase (US6)

1. Abrir `/configuracion/bloques?turno=manana`.
2. Crear el bloque 1 de `07:45` a `08:30` con vigencia desde hoy.
3. Crear el bloque 2 de `08:20` a `09:05`; debe rechazarse por solapamiento.
4. Crear un bloque que termine a las `13:00` con cierre `12:30`; debe rechazarse por quedar fuera
   del turno.
5. Crear los bloques válidos hasta completar seis.
6. Ingresar como director, ajustar la hora de inicio del bloque 3 y verificar que se guarda. El
   director no debe ver las acciones de crear ni eliminar bloques.

Resultado esperado: bloques válidos guardados; solapamientos y bloques fuera de horario
rechazados; el director solo ajusta horas.

## Escenario 7: cursos y horario semanal (US7)

1. Abrir `/configuracion/cursos` y crear `Matemática`, `Lenguaje` y `Educación Física`,
   vinculando los dos primeros a `3.°` y el tercero a `4.°` y `5.°`.
2. Abrir `/configuracion/horarios?seccion_id=<3.° A>`.
3. Asignar `Matemática` al bloque 1 del lunes y `Lenguaje` al bloque 2 del martes.
4. Intentar asignar `Educación Física` (vinculada a `4.°` y `5.°`) a `3.° A`: debe rechazarse.
5. Asignar otro curso al bloque 2 del martes y confirmar el reemplazo.
6. Consultar `/configuracion/consulta?fecha=<lunes>` y verificar que el lunes bloque 1 muestra
   Matemática.
7. Desactivar `Matemática` y verificar que deja de poder asignarse pero conserva su historial.
8. Verificar que un día feriado para la sección no aparece como hora de clase para registro.

Resultado esperado: horario semanal correcto, curso de otro grado rechazado y desactivación sin
borrado.

## Pruebas automáticas

```sh
go test ./...
cd pruebas/e2e && npx playwright test
```

Cobertura esperada:
- `pruebas/unidad`: validaciones de horarios, solapamientos, vigencias y selección de versión.
- `pruebas/integracion`: repositorios con PostgreSQL real y migraciones `0009` a `0017`.
- `pruebas/manejadores`: permisos del director, CSRF y errores de validación.
- `pruebas/e2e`: año, secciones, horario, feriado, bloques, horario semanal y consulta histórica.

## Resultados de la validación

**Fecha**: 2026-11-06 | **Entorno**: Windows, Go 1.27.3, Docker Desktop, PostgreSQL 16 en el puerto
5433, Playwright con Chromium.

| Escenario | Resultado | Evidencia |
|---|---|---|
| Preparación y migraciones `0009`–`0017` | Cumple | Subida y reversión completa verificadas en `pruebas/integracion/migraciones_test.go`. |
| 1. Año, grados, secciones y turnos | Cumple | Pruebas de manejadores y E2E de alta de grado, creación de sección y cambio de turno. |
| 2. Horarios y tolerancias con vigencia | Cumple | Integración de versionado y E2E de guardado y rechazo de cierre inválido. |
| 3. Feriados y suspensiones | Cumple | Integración por alcance y E2E de alta y confirmación de solapamiento. |
| 4. Umbrales de alerta | Cumple | Integración de versionado y E2E de valor actual y cambio futuro. |
| 5. Consulta histórica | Cumple | Integración con consulta antes y después de cambios; E2E de fecha dentro y fuera del año. |
| 6. Bloques de clase | Cumple | Integración de solapamiento y fuera de horario; E2E de alta y rechazo. |
| 7. Cursos y horario semanal | Cumple | Integración de cursos, curso de otro grado y reemplazo; E2E de asignación de una celda. |
| Permisos del director | Cumple | `pruebas/manejadores/configuracion_permisos_completos_test.go`: 403 en toda escritura salvo turnos y horas de bloques. |
| CSRF | Cumple | `pruebas/manejadores/configuracion_seguridad_test.go`: sin token y token incorrecto responden 403. |
| Auditoría | Cumple | `pruebas/integracion/academico_auditoria_test.go` verifica las 13 acciones de configuración. |
| Compatibilidad con la feature 001 | Cumple | Suite completa y E2E de alumnos, importación y carnés sin cambios. |
| Pruebas automáticas | Cumple | `go test ./...` sin errores. |
| Pruebas E2E | Cumple | 16 pruebas de Playwright aprobadas. |
| Lint | Cumple | `golangci-lint run ./...` con 0 issues. |
| Rendimiento | Cumple | Consulta de año que cubre la fecha: 0.3 ms; listado de secciones del año activo: 0.2 ms (EXPLAIN ANALYZE). |
| Integridad de datos | Cumple | 0 secciones sin turno y 0 alumnos con más de un QR activo en la base de desarrollo. |

**Observaciones**:

- En este entorno Windows, una política de Control de Aplicaciones bloqueó la ejecución de
  binarios recién compilados. Las pruebas Go y el servidor se ejecutaron dentro de contenedores
  `golang:1.27`; las pruebas E2E se ejecutaron contra el servidor en Docker con el puerto 8099
  publicado.
- La migración `0010` traslada el turno de las secciones a `turnos_seccion`, pero conserva los
  datos de demostración de la feature 001 sin cambios funcionales.
- El director puede modificar únicamente horarios de turno y horas de bloques existentes.

## Criterios de aceptación

- Crear un año completo (dos turnos, grados y secciones) en menos de 60 minutos (SC-001).
- Los cambios de vigencia no alteran ninguna asistencia anterior (SC-002).
- El 100% de las combinaciones inválidas se rechaza con el motivo (SC-003).
- Registrar un feriado o suspensión en menos de 1 minuto (SC-004).
- La consulta histórica devuelve el valor correcto en el 100% de los casos (SC-005).
- El director ajusta una hora en menos de 1 minuto y no puede crear otros elementos (SC-006).
- Armar el horario semanal completo en menos de 10 minutos (SC-007).
- El 100% de bloques solapados, fuera de horario o con curso de otro grado se rechaza (SC-008).
