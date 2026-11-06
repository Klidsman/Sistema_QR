# Guía de validación rápida: Gestión de alumnos y generación de códigos QR

**Fecha**: 2026-11-05 | **Plan**: [plan.md](plan.md)

Esta guía sirve para comprobar que la feature funciona de extremo a extremo en un entorno local.
No describe implementación; eso queda en `tasks.md`.

## Requisitos previos

- Docker y Docker Compose.
- Go estable reciente y `make`.
- Un navegador moderno.

## Preparar el entorno

1. Copiar `.env.ejemplo` a `.env` y completar `CLAVE_SESION`, `BASE_DATOS_URL` y `ADMIN_*`.
2. Levantar PostgreSQL: `docker compose up -d postgres` (el contenedor usa el puerto 5433).
3. Aplicar migraciones: `make migrar`.
4. Crear el administrador inicial y las secciones de demostración: `make sembrar`.
   Antes de ejecutarlo, definir `ADMIN_CORREO`, `ADMIN_CONTRASENIA` y `ADMIN_NOMBRE` en el entorno.
   La sección de demostración se reemplaza cuando exista la feature 002.
5. Ejecutar la aplicación: `make ejecutar` (usa air). Abrir `http://localhost:PUERTO/ingreso`.

## Escenario 1: alta manual de alumno (US1)

1. Ingresar como administrador y abrir `/alumnos/nuevo`.
2. Completar identificador, nombres, apellidos, grado y sección. Subir una foto JPG.
3. Guardar. El alumno aparece en `/alumnos` dentro de su sección.
4. Intentar crear otro con el mismo identificador: se rechaza con mensaje en español.
5. Escribir "jose" en la búsqueda: encuentra "José" (sin distinguir tildes ni mayúsculas).

**Resultado esperado**: alta en menos de 1 minuto; búsqueda correcta; el identificador es único.

## Escenario 2: QR y reemplazo (US2)

1. Abrir la ficha del alumno y emitir su QR. Aparece una sola credencial activa.
2. Revisar la imagen del QR: solo contiene el token, no el identificador ni el nombre.
3. Reemplazar el QR indicando "robo". La credencial anterior queda revocada de inmediato y la
   nueva queda activa.
4. Comprobar en la base de datos que no hay más de una credencial activa por alumno.

**Resultado esperado**: QR anterior inválido en el momento del reemplazo; sin borrado de datos.

## Escenario 3: importación con errores (US4)

1. Preparar un CSV con cinco filas: una válida, una con identificador repetido dentro del archivo,
   una con identificador ya existente, una sin nombre y una con sección inexistente.
2. Importar desde `/alumnos/importar`.
3. El reporte muestra 1 creado y 4 rechazados, cada uno con número de fila y motivo.
4. Repetir con un XLSX con tildes y ñ: los nombres se guardan y se muestran sin alteración.

**Resultado esperado**: las filas válidas se crean aunque otras fallen; el reporte es completo.

## Escenario 4: carnés en lote (US3)

1. Abrir `/carnets`, elegir una sección con alumnos con y sin foto.
2. El aviso previo indica cuántos no tienen foto y cuáles.
3. Abrir la vista del lote: cada carné muestra foto (o "sin foto"), nombre, grado, sección y QR.
4. Imprimir desde el navegador: el tamaño respeta `CARNET_ANCHO_MM` y `CARNET_ALTO_MM`.

**Resultado esperado**: lote completo, avisos correctos, impresión legible.

## Escenario 5: salidas y cambios de sección (US5)

1. Cambiar un alumno de "3.° A" a "3.° B" con fecha de hoy. Su historial conserva ambas
   secciones.
2. Marcar a un alumno como trasladado con fecha. Su QR queda revocado y su historial permanece.
3. Desactivar un alumno. Su registro no se borra.

**Resultado esperado**: historial completo; QR revocado en cada salida.

## Escenario 6: permisos del director

1. Ingresar como director.
2. Consultar `/alumnos` y una ficha: se muestra la información.
3. Intentar un `POST` (por ejemplo, desactivar): responde 403.

**Resultado esperado**: el director no puede modificar nada.

## Escenario 7: auditoría

1. Revisar la tabla `auditoria` después de los escenarios anteriores.
2. Intentar un `UPDATE` o `DELETE` manual sobre `auditoria`: el disparador lo rechaza.

**Resultado esperado**: cada cambio tiene usuario, fecha, valor anterior y nuevo; no es editable.

## Pruebas automáticas

- `make probar`: unitarias, integración con PostgreSQL real y manejadores.
- `make e2e`: Playwright para los escenarios 1, 3 y 4.
- `make lint`: golangci-lint y gofmt.

## Salud

- `GET /salud` debe responder 200 con el estado de la base de datos cuando el sistema está listo.

## Resultados de la validación

**Fecha**: 2026-11-06 | **Entorno**: Windows, Go 1.27.3, Docker Desktop, PostgreSQL 16 en el puerto
5433, navegador Chromium mediante Playwright.

| Escenario | Resultado | Evidencia |
|---|---|---|
| Preparación del entorno | Cumple | Migraciones aplicadas hasta la versión 8 y semilla con administrador y secciones. |
| 1. Alta manual de alumno | Cumple | Prueba E2E `el administrador registra un alumno y lo encuentra en el listado`; búsqueda sin tildes verificada. |
| 2. QR y reemplazo | Cumple | Pruebas de unidad e integración de credenciales; prueba E2E de emisión de QR; consulta a la base sin alumnos con más de un QR activo. |
| 3. Importación con errores | Cumple | Prueba E2E `importa filas válidas y reporta cada rechazo`; reporte con fila y motivo; XLSX con tildes y ñ cubierto por la prueba de integración. |
| 4. Carnés en lote | Cumple | Prueba E2E `genera el carné del alumno con QR vigente`; avisos y vista imprimible verificados. |
| 5. Salidas y cambios de sección | Cumple | Pruebas de integración `salidas_test.go` y `historial_seccion_test.go`; auditoría con acciones `cambiar_seccion` y `trasladar`. |
| 6. Permisos del director | Cumple | `pruebas/manejadores/permisos_test.go`: todas las rutas POST responden 403 y las rutas de lectura responden 200. |
| 7. Auditoría | Cumple | Registros presentes para alumno, credencial QR e importación; `UPDATE` manual sobre `auditoria` rechazado por el disparador. |
| Salud | Cumple | `GET /salud` respondió 200 con `base_datos: correcta`. |
| Token QR sin datos personales | Cumple | `pruebas/unidad/imagen_qr_test.go` decodifica el PNG y comprueba que el contenido es exactamente el token. |
| Pruebas automáticas | Cumple | `go build ./...`, `go vet ./...` y `go test ./...` sin errores. |
| Pruebas E2E | Cumple | 7 pruebas de Playwright aprobadas (alta, importación, carné y vista responsiva en celular, tablet y computadora). |
| Imagen Docker | Cumple | `docker build -t asistencia:local .` y contenedor con `/salud` en 200 y CSS servido desde la imagen. |

**Observaciones**: las pruebas E2E crean alumnos y secciones de prueba en la base local. Para
repetirlas en un entorno limpio, ejecutar `make migrar` y `make sembrar` antes de `npm run e2e`.
