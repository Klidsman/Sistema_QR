# Tareas: Gestión de alumnos y generación de códigos QR

**Entrada**: documentos de diseño en `/specs/001-gestion-alumnos-qr/`

**Prerrequisitos**: plan.md (obligatorio), spec.md (obligatorio para historias), research.md,
data-model.md, contracts/rutas.md, quickstart.md

**Pruebas**: incluidas. La constitución (principio V) exige pruebas antes de la implementación de
cada regla. Cada historia escribe sus pruebas primero y verifica que fallan.

**Organización**: las tareas se agrupan por historia de usuario para implementar, probar y
entregar cada historia por separado.

## Formato

`- [ ] TNNN [P?] [USn?] Descripción con ruta exacta`

- **[P]**: puede ejecutarse en paralelo (archivos distintos, sin dependencias).
- **[USn]**: historia a la que pertenece la tarea.

## Decisiones tomadas para esta lista

- Las tablas `anios_escolares`, `grados` y `secciones` se crean en la migración `0002` con su
  estructura mínima. La feature 002 agrega después sus reglas y pantallas. Así las claves
  foráneas de la 001 existen desde el inicio.
- `make e2e` se agrega al Makefile además de `ejecutar`, `migrar`, `probar` y `lint`.
- Las carpetas `contracts/` y `checklists/` se mantienen en inglés por convención de Spec Kit.

---

## Fase 1: Preparación del proyecto

**Propósito**: estructura base, herramientas y entorno local.

- [ ] T001 Crear `go.mod` con el módulo del proyecto y la versión estable de Go en `go.mod`
- [ ] T002 Crear la estructura de carpetas de `cmd/servidor`, `internal/`, `migraciones/`, `web/`, `pruebas/` y `docs/`
- [ ] T003 [P] Crear `Makefile` con los objetivos `ejecutar`, `migrar`, `probar`, `lint` y `e2e` en `Makefile`
- [ ] T004 [P] Crear `.env.ejemplo` con `BASE_DATOS_URL`, `PUERTO`, `CLAVE_SESION`, `CARNET_ANCHO_MM`, `CARNET_ALTO_MM`, `FOTO_TAMANIO_MAXIMO_MB`, `IMPORTACION_MAX_FILAS` y `ALMACENAMIENTO_TIPO` en `.env.ejemplo`
- [ ] T005 [P] Crear `docker-compose.yml` con el servicio PostgreSQL para desarrollo en `docker-compose.yml`
- [ ] T006 [P] Configurar golangci-lint en `.golangci.yml`
- [ ] T007 [P] Crear `docs/glosario.md` con los términos de la constitución (alumno, seccion, grado, marcacion, token_qr, rol, etc.)
- [ ] T008 Crear `air.toml` para la recarga en desarrollo en `air.toml`

---

## Fase 2: Fundamentos (bloquean todas las historias)

**Propósito**: infraestructura compartida. Ninguna historia puede empezar antes de cerrar esta fase.

### Configuración, base de datos y registro

- [ ] T009 [P] Escribir prueba unitaria de validación de configuración (variables obligatorias, valores fuera de rango) en `pruebas/unidad/configuracion_test.go`
- [ ] T010 Implementar carga y validación de configuración al iniciar en `internal/plataforma/configuracion/configuracion.go`
- [ ] T011 Implementar conexión y pool de PostgreSQL con pgx en `internal/plataforma/basedatos/conexion.go`
- [ ] T012 [P] Implementar registro estructurado con slog y mensajes en español en `internal/plataforma/registro/registro.go`
- [ ] T013 Implementar endpoint `/salud` que verifica la aplicación y la base de datos en `internal/plataforma/salud/manejador.go`

### Migraciones base

- [ ] T014 Crear migración `0001_crear_tabla_usuarios.sql` en `migraciones/0001_crear_tabla_usuarios.sql`
- [ ] T015 Crear migración `0002_crear_tablas_academicas.sql` con `anios_escolares`, `grados` y `secciones` en `migraciones/0002_crear_tablas_academicas.sql`
- [ ] T016 Crear migración `0007_crear_tabla_auditoria.sql` con la tabla y el disparador que bloquea `UPDATE` y `DELETE` en `migraciones/0007_crear_tabla_auditoria.sql`
- [ ] T017 Configurar sqlc con consultas en español en `sqlc.yaml` y `internal/plataforma/basedatos/consultas/`

### Auditoría y almacenamiento

- [X] T018 [P] Escribir prueba de integración que verifica que `auditoria` rechaza `UPDATE` y `DELETE` en `pruebas/integracion/auditoria_test.go`
- [X] T019 Implementar registrador de auditoría en `internal/plataforma/auditoria/registrador.go`
- [X] T020 [P] Definir la interfaz `AlmacenFotos` en `internal/plataforma/almacenamiento/almacen.go`
- [X] T021 [P] Implementar `AlmacenDisco` en `internal/plataforma/almacenamiento/disco.go`
- [X] T022 [P] Implementar `AlmacenS3` compatible con Cloudflare R2 y Backblaze B2 en `internal/plataforma/almacenamiento/s3.go`

### Autenticación, sesiones y roles

- [X] T023 [P] Escribir prueba unitaria de verificación de contraseña con argon2id en `pruebas/unidad/autenticacion_test.go`
- [X] T024 Implementar hash y verificación con argon2id en `internal/autenticacion/servicio.go`
- [X] T025 Implementar el modelo `Usuario` y `Rol` en `internal/autenticacion/modelo.go`
- [X] T026 Implementar el repositorio de usuarios en `internal/autenticacion/repositorio.go`
- [X] T027 Configurar scs con almacenamiento en PostgreSQL (tabla `sesiones`) en `internal/plataforma/intermediarios/sesion.go`
- [X] T028 Implementar protección CSRF y cookies Secure, HttpOnly y SameSite en `internal/plataforma/intermediarios/csrf.go`
- [X] T029 [P] Escribir prueba unitaria del límite de intentos de inicio de sesión en `pruebas/unidad/limite_intentos_test.go`
- [X] T030 Implementar límite de intentos por IP y usuario, ventana de 15 minutos, en `internal/plataforma/intermediarios/limite_intentos.go`
- [X] T031 Implementar intermediario de roles (administrador, director, docente, porteria) en `internal/plataforma/intermediarios/roles.go`
- [X] T032 Implementar manejador de ingreso y salida en `internal/autenticacion/manejador.go`

### Académico (lectura para alumnos)

- [X] T033 Implementar la interfaz `BuscadorSeccion` y su consulta de sección por año, grado y nombre en `internal/academico/servicio.go`
- [X] T034 Implementar el repositorio de lectura de grados y secciones en `internal/academico/repositorio.go`
- [ ] T035 Implementar los modelos `AnioEscolar`, `Grado` y `Seccion` en `internal/academico/modelo.go`

### Web base

- [X] T036 Configurar Chi, grupos de rutas por rol y manejo de errores en `internal/plataforma/rutas/rutas.go`
- [X] T037 Crear la plantilla base templ con diseño responsivo mobile-first en `web/plantillas/paginas/base.templ`
- [X] T038 Configurar Tailwind CSS y su compilación en `web/estaticos/css/`
- [X] T039 Implementar `cmd/servidor/main.go` que arranca configuración, base de datos, rutas y servidor
- [X] T040 Crear el auxiliar de pruebas que levanta PostgreSQL con testcontainers en `pruebas/integracion/ayudas_test.go`

**Punto de control**: los fundamentos están listos. Las historias pueden empezar.

---

## Fase 3: Historia 1 - Registrar, editar y desactivar alumnos (P1) 🎯 MVP

**Objetivo**: el administrador registra alumnos con sus datos, los busca, los edita y los desactiva.

**Prueba independiente**: crear un alumno, buscarlo sin tildes, editarlo, desactivarlo y verificar que
su registro se conserva.

### Pruebas de la historia 1

- [X] T041 [P] [US1] Escribir prueba unitaria de validación de alumno (obligatorios, identificador sin espacios) en `pruebas/unidad/alumnos_validacion_test.go`
- [X] T042 [P] [US1] Escribir prueba unitaria de normalización de texto (sin tildes, minúsculas, ñ) en `pruebas/unidad/normalizacion_test.go`
- [X] T043 [P] [US1] Escribir prueba de integración de unicidad del identificador y búsqueda en `pruebas/integracion/alumnos_repositorio_test.go`
- [X] T044 [P] [US1] Escribir prueba de manejador para alta, identificador repetido y permiso de director en `pruebas/manejadores/alumnos_manejador_test.go`

### Implementación de la historia 1

- [X] T045 [P] [US1] Crear migración `0003_crear_tabla_alumnos.sql` en `migraciones/0003_crear_tabla_alumnos.sql`
- [X] T046 [P] [US1] Implementar normalización de texto en `internal/alumnos/normalizacion.go`
- [X] T047 [P] [US1] Implementar el modelo `Alumno` en `internal/alumnos/modelo.go`
- [X] T048 [US1] Implementar consultas sqlc de alumnos en `internal/alumnos/consultas/alumnos.sql`
- [X] T049 [US1] Implementar el repositorio de alumnos en `internal/alumnos/repositorio.go`
- [X] T050 [US1] Implementar el servicio de alumnos (alta, edición, desactivación y reglas de unicidad) en `internal/alumnos/servicio.go`
- [X] T051 [US1] Implementar carga y reemplazo de foto (JPG o PNG, tamaño máximo) en `internal/alumnos/foto.go`
- [X] T052 [US1] Implementar el manejador de listado, alta, ficha, edición y desactivación en `internal/alumnos/manejador.go`
- [X] T053 [US1] Crear plantillas de listado, formulario y ficha de alumno en `web/plantillas/alumnos/`
- [X] T054 [US1] Registrar la auditoría de alta, edición y desactivación en `internal/alumnos/servicio.go`
- [X] T055 [US1] Implementar la ruta de foto autenticada `/alumnos/{id}/foto` en `internal/alumnos/manejador.go`

**Punto de control**: la historia 1 funciona de forma independiente.

---

## Fase 4: Historia 2 - QR único y reemplazo ante pérdida o robo (P1)

**Objetivo**: cada alumno activo tiene un único QR, y el reemplazo invalida el anterior de inmediato.

**Prueba independiente**: emitir, reemplazar y revisar que solo queda activo el QR nuevo.

### Pruebas de la historia 2

- [X] T056 [P] [US2] Escribir prueba unitaria de reglas de QR (un activo, reemplazo exige motivo) en `pruebas/unidad/qr_reglas_test.go`
- [X] T057 [P] [US2] Escribir prueba unitaria de generación de token (43 caracteres, sin datos personales) en `pruebas/unidad/token_qr_test.go`
- [X] T058 [P] [US2] Escribir prueba de integración del índice único parcial y del reemplazo transaccional en `pruebas/integracion/credenciales_qr_test.go`
- [X] T059 [P] [US2] Escribir prueba de manejador de emisión, reemplazo y permiso de director en `pruebas/manejadores/qr_manejador_test.go`

### Implementación de la historia 2

- [X] T060 [P] [US2] Crear migración `0004_crear_tabla_credenciales_qr.sql` con el índice único parcial en `migraciones/0004_crear_tabla_credenciales_qr.sql`
- [X] T061 [P] [US2] Implementar el modelo `CredencialQR` en `internal/alumnos/modelo_qr.go`
- [X] T062 [US2] Implementar generación de token con `crypto/rand` en `internal/alumnos/token.go`
- [X] T063 [US2] Implementar consultas sqlc de credenciales QR en `internal/alumnos/consultas/credenciales_qr.sql`
- [X] T064 [US2] Implementar el servicio de QR (emitir, revocar y reemplazar en una transacción) en `internal/alumnos/servicio_qr.go`
- [X] T065 [US2] Generar la imagen PNG del QR con skip2/go-qrcode, solo con el token, en `internal/alumnos/imagen_qr.go`
- [X] T066 [US2] Implementar los manejadores de emisión, reemplazo e imagen en `internal/alumnos/manejador_qr.go`
- [X] T067 [US2] Crear la plantilla de sección de QR en la ficha del alumno en `web/plantillas/alumnos/qr.templ`
- [X] T068 [US2] Registrar la auditoría de emisión y revocación con motivo en `internal/alumnos/servicio_qr.go`

**Punto de control**: la historia 2 funciona de forma independiente.

---

## Fase 5: Historia 5 - Retirar, trasladar, cambiar de sección (P2)

**Objetivo**: marcar salidas y cambiar de grado o sección sin perder historial.

> Se implementa antes de la historia 3 y 4 porque la desactivación y la salida usan la revocación de QR de la historia 2.

**Prueba independiente**: cambiar de sección, marcar un traslado y verificar historial y QR revocado.

### Pruebas de la historia 5

- [ ] T069 [P] [US5] Escribir prueba unitaria de cambio de sección (cierra el registro actual y abre uno nuevo) en `pruebas/unidad/historial_seccion_test.go`
- [ ] T070 [P] [US5] Escribir prueba de integración de salida (retirado o trasladado) que revoca el QR en `pruebas/integracion/salidas_test.go`
- [ ] T071 [P] [US5] Escribir prueba de manejador de cambio de sección y salida en `pruebas/manejadores/salidas_manejador_test.go`

### Implementación de la historia 5

- [ ] T072 [P] [US5] Crear migración `0005_crear_tabla_historial_secciones.sql` en `migraciones/0005_crear_tabla_historial_secciones.sql`
- [ ] T073 [P] [US5] Implementar el modelo `HistorialSeccion` en `internal/alumnos/modelo_historial.go`
- [ ] T074 [US5] Implementar consultas sqlc de historial y de salidas en `internal/alumnos/consultas/historial.sql`
- [ ] T075 [US5] Implementar cambio de sección con fecha de vigencia en `internal/alumnos/servicio.go`
- [ ] T076 [US5] Implementar salida (retirado o trasladado) que revoca el QR vigente en `internal/alumnos/servicio.go`
- [ ] T077 [US5] Implementar desactivación que revoca el QR vigente en `internal/alumnos/servicio.go`
- [ ] T078 [US5] Implementar los manejadores de cambio de sección y salida en `internal/alumnos/manejador.go`
- [ ] T079 [US5] Mostrar el historial de secciones y salidas en la ficha del alumno en `web/plantillas/alumnos/ficha.templ`

**Punto de control**: la historia 5 funciona de forma independiente.

---

## Fase 6: Historia 3 - Carnés en lote para imprimir (P2)

**Objetivo**: generar carnés por grado o sección con foto, nombre, grado, sección y QR, listos para imprimir.

**Prueba independiente**: seleccionar una sección con y sin foto, revisar avisos y la vista imprimible.

### Pruebas de la historia 3

- [ ] T080 [P] [US3] Escribir prueba unitaria de avisos previos (sin foto, sin QR vigente) en `pruebas/unidad/carnets_avisos_test.go`
- [ ] T081 [P] [US3] Escribir prueba de manejador de la vista de lote y del marcador "sin foto" en `pruebas/manejadores/carnets_manejador_test.go`

### Implementación de la historia 3

- [ ] T082 [US3] Implementar la selección de alumnos por grado o sección y los avisos previos en `internal/alumnos/carnets.go`
- [ ] T083 [US3] Implementar consultas sqlc de la selección para carnés en `internal/alumnos/consultas/carnets.sql`
- [ ] T084 [US3] Implementar los manejadores de `/carnets` y `/carnets/lote` en `internal/alumnos/manejador_carnets.go`
- [ ] T085 [US3] Crear la plantilla del carné con tamaño `CARNET_ANCHO_MM` y `CARNET_ALTO_MM` en `web/plantillas/carnets/carnet.templ`
- [ ] T086 [US3] Crear la vista de lote con `@page` para impresión y avisos previos en `web/plantillas/carnets/lote.templ`
- [ ] T087 [US3] Crear la pantalla de selección de carnés en `web/plantillas/carnets/seleccion.templ`

**Punto de control**: la historia 3 funciona de forma independiente.

---

## Fase 7: Historia 4 - Importación masiva desde Excel o CSV (P2)

**Objetivo**: importar alumnos con un reporte por fila, sin detener la importación por errores parciales.

**Prueba independiente**: importar un archivo con filas válidas, duplicadas, existentes, incompletas y con sección inexistente.

### Pruebas de la historia 4

- [ ] T088 [P] [US4] Escribir prueba unitaria de detección de duplicados dentro del archivo y de motivos de rechazo en `pruebas/unidad/importacion_test.go`
- [ ] T089 [P] [US4] Escribir prueba unitaria de lectura de CSV con BOM y separador punto y coma en `pruebas/unidad/importacion_csv_test.go`
- [ ] T090 [P] [US4] Escribir prueba de integración de importación con PostgreSQL real en `pruebas/integracion/importacion_test.go`
- [ ] T091 [P] [US4] Escribir prueba de manejador de carga, límite de filas y reporte en `pruebas/manejadores/importacion_manejador_test.go`

### Implementación de la historia 4

- [ ] T092 [P] [US4] Crear migración `0006_crear_tablas_importacion.sql` en `migraciones/0006_crear_tablas_importacion.sql`
- [ ] T093 [P] [US4] Implementar el modelo de importación y de fila rechazada en `internal/alumnos/modelo_importacion.go`
- [ ] T094 [US4] Implementar lectura de CSV (coma o punto y coma, BOM) en `internal/alumnos/importacion_csv.go`
- [ ] T095 [US4] Implementar lectura de XLSX con excelize, primera hoja, en `internal/alumnos/importacion_xlsx.go`
- [ ] T096 [US4] Implementar validación por fila y detección de duplicados dentro del archivo en `internal/alumnos/importacion.go`
- [ ] T097 [US4] Implementar consultas sqlc de importaciones en `internal/alumnos/consultas/importaciones.sql`
- [ ] T098 [US4] Implementar el servicio de importación con inserción por fila y reporte en `internal/alumnos/servicio_importacion.go`
- [ ] T099 [US4] Implementar los manejadores de carga, historial y detalle en `internal/alumnos/manejador_importacion.go`
- [ ] T100 [US4] Crear las plantillas de carga, historial y detalle con filas rechazadas en `web/plantillas/importacion/`
- [ ] T101 [US4] Registrar la auditoría de la importación en `internal/alumnos/servicio_importacion.go`

**Punto de control**: la historia 4 funciona de forma independiente.

---

## Fase 8: Pulido y transversales

**Propósito**: permisos, seguridad, entrega y documentación.

- [ ] T102 [P] Escribir prueba de permisos de director sobre todas las rutas `POST` (403) en `pruebas/manejadores/permisos_test.go`
- [ ] T103 [P] Escribir prueba de CSRF y cookies seguras en `pruebas/manejadores/seguridad_test.go`
- [ ] T104 Escribir pruebas E2E con Playwright para alta, importación y carné en `pruebas/e2e/alumnos_test.spec.ts`
- [ ] T105 Escribir prueba E2E de vista responsiva en celular, tablet y computadora en `pruebas/e2e/responsivo_test.spec.ts`
- [ ] T106 Crear `Dockerfile` multietapa en `Dockerfile`
- [ ] T107 Crear el workflow de GitHub Actions con lint, pruebas y construcción de imagen en `.github/workflows/integracion.yml`
- [ ] T108 Escribir el README de despliegue paso a paso en `docs/despliegue.md`
- [ ] T109 Escribir la guía de respaldo y restauración en `docs/respaldo_restauracion.md`
- [ ] T110 [P] Escribir el manual de usuario del administrador en `docs/manual_administrador.md`
- [ ] T111 [P] Escribir el manual de usuario del director en `docs/manual_director.md`
- [ ] T112 Ejecutar la validación de `quickstart.md` completa y registrar resultados en `specs/001-gestion-alumnos-qr/quickstart.md`
- [ ] T113 Revisar que el token QR no contenga datos personales en cualquier salida (pruebas y revisión manual)

---

## Dependencias y orden de ejecución

### Dependencias entre fases

- **Fase 1 (preparación)**: sin dependencias. Puede empezar de inmediato.
- **Fase 2 (fundamentos)**: depende de la fase 1 y bloquea todas las historias.
- **Fases 3 a 7 (historias)**: dependen de la fase 2.
  - La historia 5 depende de la historia 2 porque revoca QR.
  - Las historias 3 y 4 dependen de las historias 1 y 2.
- **Fase 8 (pulido)**: depende de las historias que se entreguen.

### Dependencias dentro de cada historia

- Las pruebas se escriben y fallan antes de la implementación.
- Modelos antes de servicios, servicios antes de manejadores, manejadores antes de plantillas.

### Oportunidades de paralelismo

- Tareas marcadas con [P] dentro de cada fase, si son de archivos distintos.
- Con más de una persona, la historia 1 y la historia 2 pueden avanzar en paralelo tras la fase 2.

---

## Estrategia de entrega

### MVP (historia 1 y 2)

1. Fase 1 y fase 2.
2. Historia 1 (alumnos) y historia 2 (QR).
3. Validar con `quickstart.md` escenarios 1 y 2.

### Entrega incremental

1. MVP.
2. Historia 5 (salidas y cambios de sección).
3. Historia 3 (carnés).
4. Historia 4 (importación).
5. Pulido y entrega documental.

---

## Notas

- Cada tarea nombra su archivo. Las rutas siguen `internal/<modulo>/` según el plan.
- Los nombres de archivo y de paquete están en español (constitución, principio IX).
- Commits en español, después de cada tarea o grupo lógico.
- El identificador `T` sigue la numeración secuencial de la lista.
