# Guía de validación rápida: Registro de asistencia del docente

**Fecha**: 2026-11-06 | **Plan**: [plan.md](plan.md)

Esta guía comprueba el registro de asistencia del docente de extremo a extremo en un entorno
local. No describe implementación; eso queda en `tasks.md`.

## Requisitos previos

- Features 001 y 002 aplicadas (migraciones hasta la versión 17).
- PostgreSQL en Docker, Go, Node y `make`.
- Un año escolar activo con grados, secciones, turnos, bloques y horario semanal configurados.
- Al menos un curso vinculado al grado de la sección y alumnos activos en ella.

## Preparar el entorno

1. Copiar `.env.ejemplo` a `.env` y completar `CLAVE_SESION` y `BASE_DATOS_URL`.
2. Levantar PostgreSQL: `docker compose up -d postgres`.
3. Aplicar migraciones (incluye `0018` a `0022`): `make migrar`.
4. Ejecutar: `make ejecutar` y abrir `http://localhost:8099/ingreso`.
5. Ingresar como administrador.

## Escenario 1: crear un docente y asignarle un curso

1. Abrir `/docentes` y crear una cuenta con nombre, correo y contraseña.
2. Verificar que aparece en el listado como activa y con rol docente.
3. Abrir `/asistencia/asignaciones` y asignar el docente a un curso y sección con fecha de inicio
   hoy y sin fecha de fin (titular).
4. Verificar que la asignación aparece vigente.
5. Crear una segunda asignación para el mismo curso y sección con otro docente, con fecha de
   inicio y fin de un solo día (reemplazo). Verificar que se acepta aunque se superponga.

Resultado esperado: el docente titular y el reemplazante quedan asignados; el reemplazo solo
tendrá acceso en su rango.

## Escenario 2: registrar la asistencia de una clase (US1)

1. Ingresar con la cuenta del docente.
2. Abrir `/asistencia` y elegir la fecha de hoy.
3. Verificar que aparecen solo los cursos y secciones asignados.
4. Abrir la clase y comprobar que la lista muestra únicamente alumnos activos de la sección.
5. Verificar que cada alumno muestra su ingreso por QR (hora) o "sin ingreso".
6. Pulsar **Marcar a todos presentes**.
7. Cambiar tres alumnos a tardanza y dos a falta; agregar una observación a uno de ellos.
8. Guardar y recargar la página.

Resultado esperado: los estados y observaciones persisten; los alumnos con diferencia entre QR y
registro docente aparecen destacados; el QR no cambia.

## Escenario 3: corregir dentro del plazo (US2)

1. Con el mismo docente, cambiar un alumno de falta a presente dentro del plazo configurado.
2. Verificar que el cambio se guarda y queda auditado.
3. Cambiar el plazo en `/configuracion/asistencia` a 1 hora.
4. Alterar la fecha de creación de un registro de prueba (solo en entorno de pruebas) para simular
   que el plazo venció.
5. Intentar corregir como docente: debe responder que solo un administrador puede hacerlo.
6. Como administrador, corregir el registro sin motivo: debe rechazarse.
7. Repetir con un motivo: debe guardarse y auditarse con el motivo.

Resultado esperado: el plazo se respeta y la corrección del administrador exige motivo.

## Escenario 4: reemplazo por un día (US3)

1. Asignar un reemplazante a un curso y sección por el día de hoy.
2. Ingresar como reemplazante: debe ver y registrar esa clase.
3. Consultar la misma asignación con fecha de mañana: no debe aparecer.
4. Como titular, verificar que conserva su acceso el mismo día si su asignación sigue vigente.

Resultado esperado: el acceso del reemplazante cesa automáticamente al terminar su rango.

## Escenario 5: registro por hora de clase (US4)

1. Abrir una clase del bloque 1 y registrar asistencia.
2. Abrir otra clase del mismo curso y sección en el bloque 2 y registrar un estado distinto para
   el mismo alumno.
3. Verificar que cada bloque conserva su propio registro.

Resultado esperado: no existe un registro único por turno; cada bloque tiene su fila.

## Escenario 6: consulta del director

1. Ingresar como director.
2. Abrir `/asistencia/consulta`, elegir fecha y sección.
3. Verificar que ve los registros y las diferencias con el QR.
4. Intentar corregir un registro y confirmar que no tiene la acción disponible y que un `POST`
   responde `403`.

## Escenario 7: exportación

1. Como docente, abrir `/asistencia/exportar`.
2. Elegir un rango de fechas y un curso.
3. Descargar el CSV y verificar columnas: fecha, curso, sección, bloque, identificador, alumno,
   estado, observación, hora QR y docente.
4. Verificar que no incluye cursos o secciones ajenos a sus asignaciones.

## Escenario 8: reglas de fecha y calendario

1. Intentar registrar una fecha futura: debe rechazarse.
2. Configurar un feriado para la sección y abrir la clase de esa fecha: debe indicar que no es
   laborable y no generar faltas.
3. Retirar o trasladar a un alumno y verificar que no aparece en fechas posteriores, conservando
   sus registros anteriores.

## Pruebas automáticas

```sh
go test ./...
cd pruebas/e2e && npx playwright test
```

Cobertura esperada:
- Unitarias: diferencia QR/docente, plazo, estados, observación y CSV.
- Integración: asignaciones, resolución de clase, guardado idempotente, auditoría y exportación.
- Manejadores: permisos, CSRF, fecha futura, día no laborable y corrección con motivo.
- E2E: alta de docente, asignación, registro, corrección, reemplazo y exportación.

## Resultados de la validación

**Fecha**: 2026-11-06 | **Entorno**: Windows, Go 1.27.3, Docker Desktop, PostgreSQL 16 en el puerto
5433, Playwright con Chromium.

| Escenario | Resultado | Evidencia |
|---|---|---|
| Preparación y migraciones `0018`–`0022` | Cumple | Subida y reversión completa en `pruebas/integracion/migraciones_test.go`. |
| Cuentas de docente y asignaciones | Cumple | Integración y E2E de alta de docente, titular y reemplazo por un día. |
| 1. Registrar la asistencia de la clase | Cumple | Integración de lista, QR, marcado masivo y diferencias; E2E de clase completa. |
| 2. Correcciones y plazo | Cumple | Integración de plazo vencido, corrección con motivo y auditoría; E2E de plazo y corrección. |
| 3. Asignaciones y reemplazos | Cumple | Acceso por fecha verificado en `asignaciones_test.go`; el reemplazo cesa al terminar su rango. |
| 4. Registro por hora de clase | Cumple | Dos bloques con estados distintos y registros independientes en `asistencia_bloques_test.go`. |
| Consulta del director | Cumple | Manejadores y E2E; sin acciones de corrección para el director. |
| Exportación CSV | Cumple | Unitaria de columnas y BOM, integración de permisos y E2E de descarga. |
| Permisos | Cumple | Docente solo asignados; director solo lectura; portería sin acceso; administrador completo. |
| CSRF | Cumple | `pruebas/manejadores/asistencia_permisos_completos_test.go`: sin token responde 403. |
| Auditoría | Cumple | Altas, correcciones, marcado masivo, asignaciones, docentes y plazo auditados con motivo cuando aplica. |
| Idempotencia y concurrencia | Cumple | Envíos repetidos no duplican ni auditan de más; dos "marcar todos" concurrentes crean una sola fila por alumno. |
| Reglas de fecha | Cumple | Fecha futura, feriado y alumno retirado verificados en `asistencia_clase_test.go`. |
| Compatibilidad con 001 y 002 | Cumple | Suite completa y 22 pruebas E2E. |
| Pruebas automáticas | Cumple | `go test ./...` sin errores. |
| Lint | Cumple | `golangci-lint run ./...` con 0 issues. |
| Rendimiento | Cumple | Lista de clase: 0.058 ms; exportación por docente: 0.032 ms (EXPLAIN ANALYZE). |

**Observaciones**:

- El modo sin conexión de FR-017/FR-018 quedó diferido por decisión y por el principio III de la
  constitución.
- La tabla `marcaciones` se crea y se lee, pero la escritura pertenece a la feature de portería.
- El QR sigue siendo la asistencia oficial; el registro docente es personal.
- Las pruebas Go y el servidor se ejecutaron en contenedores `golang:1.27` por una política de
  Control de Aplicaciones de Windows; Playwright se ejecutó contra el servidor en Docker.

## Criterios de aceptación

- 40 alumnos registrados en menos de 2 minutos con "Marcar a todos presentes" (SC-001).
- 100% de diferencias destacadas (SC-002).
- Reenviar el mismo registro no duplica (SC-003).
- Toda modificación fuera del plazo exige administrador y motivo (SC-004).
- Ningún docente ve cursos o secciones no asignados (SC-005).
- El acceso del reemplazante cesa al terminar su rango (SC-006).
