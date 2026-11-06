# Tareas: Registro de asistencia de alumnos por el docente

**Input**: Documentos de diseño en `/specs/003-registro-asistencia-docente/`

**Prerrequisitos**: [plan.md](plan.md), [spec.md](spec.md), [research.md](research.md), [data-model.md](data-model.md), [contracts/rutas.md](contracts/rutas.md)

**Pruebas**: obligatorias según la constitución (principio V). Se escriben antes de la implementación de cada regla.

**Organización**: tareas agrupadas por historia de usuario. US1 depende de US3 porque sin docentes y asignaciones no hay clase que registrar.

## Formato: `[ID] [P?] [Historia] Descripción`

- **[P]**: puede ejecutarse en paralelo (archivos distintos, sin dependencias pendientes)
- **[Historia]**: US1 a US4 según `spec.md`
- Todas las descripciones incluyen la ruta exacta del archivo

---

## Fase 1: Preparación

- [X] T001 [P] Crear la estructura `internal/asistencia/{consultas,generado}`, `web/plantillas/asistencia/` y `web/plantillas/docentes/`
- [X] T002 [P] Agregar los estados, entidades y acciones de asistencia en `internal/asistencia/constantes.go`
- [X] T003 [P] Agregar el paquete de asistencia a `sqlc.yaml` con salida en `internal/asistencia/generado`
- [X] T004 [P] Ampliar `docs/glosario.md` con `registro_docente`, `asignacion_docente`, `marcacion`, `plazo_edicion` y `diferencia`
- [X] T005 [P] Crear los errores de negocio en `internal/asistencia/errores.go`
- [X] T006 [P] Crear los tipos `Reloj` y las interfaces de apoyo en `internal/asistencia/dependencias.go`

---

## Fase 2: Base bloqueante

### Migraciones

- [X] T007 [P] Crear `migraciones/0018_crear_parametros_asistencia.sql` con una fila inicial de 24 horas, rango 1 a 168 y auditoría de cambios
- [X] T008 [P] Crear `migraciones/0019_crear_tabla_marcaciones.sql` con `alumno_id`, `fecha`, `hora`, `estado` y `UNIQUE (alumno_id, fecha)`
- [X] T009 [P] Crear `migraciones/0020_crear_tabla_asignaciones_docente.sql` con `docente_id`, `curso_id`, `seccion_id`, `fecha_inicio`, `fecha_fin`, `titular`, `creado_por`; `UNIQUE (docente_id, curso_id, seccion_id, fecha_inicio)` y rango de fechas válido
- [X] T010 [P] Crear `migraciones/0021_crear_tabla_registros_docente.sql` con estado, observación de hasta 300 caracteres y `UNIQUE (alumno_id, fecha, curso_id, seccion_id, bloque_numero)`
- [X] T011 [P] Crear `migraciones/0022_ampliar_auditoria_asistencia.sql` con `docente`, `asignacion_docente`, `registro_docente`, `parametro_asistencia` y `marcacion`
- [X] T012 Actualizar la prueba de migraciones para subir y bajar hasta la versión 22 en `pruebas/integracion/migraciones_test.go`

### Modelos, consultas y repositorios

- [X] T013 [P] Implementar los modelos `AsignacionDocente`, `RegistroDocente`, `Marcacion`, `ParametrosAsistencia`, `ClaseAsignada` y `AlumnoEnClase` en `internal/asistencia/modelo.go`
- [X] T014 [P] Escribir las consultas sqlc de asignaciones, registros, marcaciones y parámetros en `internal/asistencia/consultas/asistencia.sql`
- [X] T015 Regenerar sqlc con `sqlc generate` y compilar `internal/asistencia/generado/`
- [X] T016 [P] Implementar `internal/asistencia/repositorio.go` con pool, consultas y `EnTransaccion`
- [X] T017 [P] Implementar el repositorio de asignaciones en `internal/asistencia/repositorio_asignaciones.go`
- [X] T018 [P] Implementar el repositorio de registros y exportación en `internal/asistencia/repositorio_registros.go`
- [X] T019 [P] Implementar el repositorio de marcaciones en `internal/asistencia/repositorio_marcaciones.go`
- [X] T020 [P] Implementar el repositorio de parámetros en `internal/asistencia/repositorio_parametros.go`

### Interfaces entre módulos

- [X] T021 [P] Implementar `ResolutorClase` en `internal/academico/consulta_clase.go`: valida fecha, día laborable, turno, bloque y celda del horario, y expone el día no laborable
- [X] T022 [P] Implementar `ListadorAlumnosSeccion` en `internal/alumnos/servicio_consulta.go` y su consulta por historial de secciones en `internal/alumnos/consultas/alumnos.sql`
- [X] T023 [P] Implementar el buscador de docentes en `internal/autenticacion/servicio_docentes.go` (rol, estado y datos mínimos)
- [X] T024 Implementar la auditoría de las nuevas entidades en `internal/asistencia/servicio_auditoria.go`
- [X] T025 Registrar el módulo, dependencias y rutas base en `internal/plataforma/rutas/rutas.go` y `cmd/servidor/main.go`

### Reglas puras

- [X] T026 [P] Escribir pruebas unitarias de estados, observación y diferencia QR/docente en `pruebas/unidad/asistencia_reglas_test.go`
- [X] T027 [P] Escribir pruebas unitarias del plazo de edición y fechas futuras en `pruebas/unidad/asistencia_plazo_test.go`
- [X] T028 Implementar las reglas puras en `internal/asistencia/validaciones.go`
- [X] T029 Escribir la prueba de migraciones y parámetro inicial en `pruebas/integracion/asistencia_migraciones_test.go`
- [X] T030 Escribir la prueba base de permisos (docente, director, portería) en `pruebas/manejadores/asistencia_permisos_test.go`

---

## Fase 3: Historia 3 — Docentes y asignaciones (P2) — habilitante de US1

**Objetivo**: crear cuentas de docente y asignarlas a curso y sección con vigencia, incluidos reemplazos.

**Prueba independiente**: crear dos docentes, asignar uno como titular y otro como reemplazo por un día; verificar acceso por fecha.

- [X] T031 [P] [US3] Escribir pruebas unitarias de fechas de asignación y superposición permitida en `pruebas/unidad/asignaciones_test.go`
- [X] T032 [P] [US3] Escribir pruebas de integración de asignaciones y acceso por fecha en `pruebas/integracion/asignaciones_test.go`
- [X] T033 [US3] Implementar `CrearDocente`, `ListarDocentes`, `DesactivarDocente` y `RestablecerClave` en `internal/autenticacion/servicio_docentes.go`
- [X] T034 [US3] Implementar las rutas `GET/POST /docentes`, `POST /docentes/{id}/clave` y `POST /docentes/{id}/desactivacion` en `internal/autenticacion/manejador_docentes.go`
- [X] T035 [US3] Crear las plantillas `listado.templ` y `formulario.templ` en `web/plantillas/docentes/`
- [X] T036 [US3] Auditar `crear_docente`, `desactivar_docente` y `restablecer_clave`
- [X] T037 [US3] Escribir las pruebas de manejadores de docentes en `pruebas/manejadores/docentes_test.go`
- [X] T038 [US3] Escribir la prueba E2E de alta de docente en `pruebas/e2e/docentes_test.spec.ts`
- [X] T039 [US3] Implementar `CrearAsignacion`, `CerrarAsignacion` y `ListarAsignaciones` en `internal/asistencia/servicio_asignaciones.go`
- [X] T040 [US3] Validar rol docente, curso vinculado al grado y fechas en `internal/asistencia/servicio_asignaciones.go`
- [X] T041 [US3] Implementar `GET /asistencia/asignaciones`, `POST /asistencia/asignaciones` y `POST /asistencia/asignaciones/{id}/cierre` en `internal/asistencia/manejador_asignaciones.go`
- [X] T042 [US3] Crear la plantilla `web/plantillas/asistencia/asignaciones.templ`
- [X] T043 [US3] Auditar `crear_asignacion` y `cerrar_asignacion`
- [X] T044 [US3] Escribir las pruebas de manejadores de asignaciones en `pruebas/manejadores/asignaciones_test.go`
- [X] T045 [US3] Implementar `ClasesAsignadas` para resolver las clases de un docente en una fecha en `internal/asistencia/servicio_asignaciones.go`
- [X] T046 [US3] Escribir la prueba E2E de asignación y acceso del reemplazante en `pruebas/e2e/asignaciones_test.spec.ts`
- [X] T047 [US3] Escribir la prueba de acceso denegado para docente no asignado en `pruebas/manejadores/asistencia_acceso_test.go`

**Punto de control**: cuentas y asignaciones funcionando; US1 puede comenzar.

---

## Fase 4: Historia 1 — Registrar la asistencia de la clase (P1)

**Objetivo**: lista de alumnos de una clase con su QR y registro personal del docente.

**Prueba independiente**: marcar a todos presentes, corregir tres excepciones, guardar y recargar.

- [X] T048 [P] [US1] Escribir pruebas de integración de listar clase, alumnos y marcaciones en `pruebas/integracion/asistencia_clase_test.go`
- [X] T049 [US1] Implementar `ListarClase` con alumnos, marcaciones y diferencias en `internal/asistencia/servicio_registros.go`
- [X] T050 [US1] Implementar `GuardarRegistros` con upsert idempotente y auditoría en `internal/asistencia/servicio_registros.go`
- [X] T051 [US1] Implementar `MarcarTodosPresentes` solo para alumnos sin registro en `internal/asistencia/servicio_registros.go`
- [X] T052 [US1] Implementar las rutas `GET /asistencia`, `GET /asistencia/clase`, `POST /asistencia/clase` y `POST /asistencia/clase/todos` en `internal/asistencia/manejador_asistencia.go`
- [X] T053 [US1] Crear las plantillas `seleccion.templ` y `clase.templ` en `web/plantillas/asistencia/`
- [X] T054 [US1] Mostrar el estado y la hora del QR junto a cada alumno y destacar diferencias sin alterar el QR en `web/plantillas/asistencia/clase.templ`
- [X] T055 [US1] Auditar `crear_registro` y `marcar_todos_presentes`
- [X] T056 [US1] Escribir las pruebas de manejadores de la clase en `pruebas/manejadores/asistencia_clase_test.go`
- [X] T057 [US1] Escribir la prueba E2E de marcar todos y corregir excepciones en `pruebas/e2e/asistencia_clase_test.spec.ts`

**Punto de control**: US1 funcional y verificable con asignaciones existentes.

---

## Fase 5: Historia 2 — Correcciones y plazo de edición (P2)

**Objetivo**: docente corrige dentro del plazo; administrador corrige después con motivo; consulta para director.

**Prueba independiente**: corregir dentro del plazo, simular vencimiento, verificar bloqueo y corrección administrativa con motivo.

- [ ] T058 [P] [US2] Escribir pruebas de integración del plazo y la corrección administrativa en `pruebas/integracion/asistencia_correcciones_test.go`
- [ ] T059 [US2] Implementar la comprobación de plazo con `creado_en + plazo_edicion_horas` en `internal/asistencia/servicio_registros.go`
- [ ] T060 [US2] Implementar `CorregirRegistro` con motivo obligatorio para el administrador y auditoría
- [ ] T061 [US2] Implementar `GET /asistencia/consulta` y `POST /asistencia/consulta/correccion` en `internal/asistencia/manejador_consulta.go`
- [ ] T062 [US2] Crear la plantilla `consulta.templ` con registros, QR y diferencias
- [ ] T063 [US2] Implementar `PlazoEdicion` y `ConfigurarPlazo` en `internal/asistencia/servicio_parametros.go`
- [ ] T064 [US2] Implementar `GET/POST /configuracion/asistencia` en `internal/asistencia/manejador_parametros.go`
- [ ] T065 [US2] Crear la plantilla `parametros_asistencia.templ` en `web/plantillas/configuracion/`
- [ ] T066 [US2] Auditar `corregir_registro` y `configurar_plazo`
- [ ] T067 [US2] Escribir las pruebas de manejadores de correcciones y plazo en `pruebas/manejadores/asistencia_correcciones_test.go`
- [ ] T068 [US2] Escribir la prueba E2E de corrección dentro del plazo y configuración del plazo en `pruebas/e2e/asistencia_correcciones_test.spec.ts`

**Punto de control**: correcciones con plazo y consulta del director funcionando.

---

## Fase 6: Historia 4 — Registro por hora de clase (P3)

**Objetivo**: cada bloque conserva su propio registro por alumno.

**Prueba independiente**: registrar dos bloques del mismo curso con estados distintos y verificar independencia.

- [ ] T069 [P] [US4] Escribir pruebas unitarias de la clave fecha, curso, sección y bloque en `pruebas/unidad/asistencia_bloques_test.go`
- [ ] T070 [P] [US4] Escribir pruebas de integración de dos bloques independientes en `pruebas/integracion/asistencia_bloques_test.go`
- [ ] T071 [US4] Incluir el bloque en la clave natural y en la consulta de clase en `internal/asistencia/consultas/asistencia.sql`
- [ ] T072 [US4] Mostrar el bloque y su horario en la lista y el selector de clases en `web/plantillas/asistencia/`
- [ ] T073 [US4] Escribir las pruebas de manejadores de cambio de bloque en `pruebas/manejadores/asistencia_bloques_test.go`
- [ ] T074 [US4] Escribir la prueba E2E de dos bloques con estados distintos en `pruebas/e2e/asistencia_bloques_test.spec.ts`

**Punto de control**: registro por hora completo.

---

## Fase 7: Exportación (transversal)

- [ ] T075 [P] Escribir pruebas unitarias del CSV y sus columnas en `pruebas/unidad/asistencia_exportacion_test.go`
- [ ] T076 [P] Escribir pruebas de integración de la exportación y sus permisos en `pruebas/integracion/asistencia_exportacion_test.go`
- [ ] T077 Implementar `ExportarCSV` en `internal/asistencia/servicio_exportacion.go`
- [ ] T078 Implementar `GET /asistencia/exportar` y `GET /asistencia/exportar/archivo` en `internal/asistencia/manejador_consulta.go`
- [ ] T079 Crear la plantilla `exportacion.templ` en `web/plantillas/asistencia/`
- [ ] T080 Escribir las pruebas de manejadores de exportación en `pruebas/manejadores/asistencia_exportacion_test.go`
- [ ] T081 Escribir la prueba E2E de descarga CSV en `pruebas/e2e/asistencia_exportacion_test.spec.ts`

**Punto de control**: exportación disponible para docente, administrador y director según permisos.

---

## Fase 8: Pulido y transversales

- [ ] T082 [P] Completar la prueba de permisos: docente solo asignados, director solo `GET`, portería sin acceso en `pruebas/manejadores/asistencia_permisos_test.go`
- [ ] T083 [P] Escribir la prueba de CSRF y validaciones de formularios en `pruebas/manejadores/asistencia_seguridad_test.go`
- [ ] T084 Escribir la prueba de auditoría de todas las acciones de la 003 en `pruebas/integracion/asistencia_auditoria_test.go`
- [ ] T085 Verificar idempotencia con envíos repetidos y concurrencia de "marcar todos" en `pruebas/integracion/asistencia_idempotencia_test.go`
- [ ] T086 Verificar reglas de fecha: futura, feriado, suspensión y alumno retirado en `pruebas/integracion/asistencia_fechas_test.go`
- [ ] T087 Revisar `EXPLAIN ANALYZE` de la lista de clase y la exportación con datos de 2,000 alumnos
- [ ] T088 Actualizar `docs/manual_administrador.md`, `docs/manual_director.md` y `docs/manual_docente.md` con las pantallas nuevas
- [ ] T089 Actualizar `docs/despliegue.md` y confirmar que no hay variables de entorno nuevas
- [ ] T090 Regenerar plantillas y estilos: `make templ`, `npm run css`, `gofmt -w`
- [ ] T091 Ejecutar `golangci-lint`, `go build`, `go vet`, `go test ./...` y Playwright completo
- [ ] T092 Validar los escenarios de `specs/003-registro-asistencia-docente/quickstart.md` y registrar resultados en el mismo archivo
- [ ] T093 Verificar compatibilidad con las features 001 y 002 (suite y E2E existentes) en `pruebas/e2e/`

---

## Dependencias y orden

| Historia | Depende de | Motivo |
|---|---|---|
| US3 (docentes y asignaciones) | Base | Habilita cualquier clase. |
| US1 (registrar) | US3 y Base | Sin docente asignado no hay lista que abrir. |
| US2 (correcciones y plazo) | US1 | Corrige registros existentes. |
| US4 (por hora) | US1 | El bloque es parte de la clave desde el inicio; la historia valida su independencia. |
| Exportación | US1 y US2 | Exporta registros y respeta permisos. |
| Pulido | Todas | Cierra la entrega. |

## Estrategia

1. Preparación y base.
2. US3: cuentas, asignaciones y acceso por fecha.
3. US1: lista, registro y marcado masivo.
4. US2: plazo, correcciones y consulta.
5. US4: independencia por bloque.
6. Exportación.
7. Pulido, documentación y validación.

## Notas

- No se implementa modo sin conexión: queda diferido y documentado.
- La tabla `marcaciones` se crea pero solo se lee; la escritura pertenece a portería.
- El QR sigue siendo la asistencia oficial; el registro docente es personal.
- Cada tarea o grupo lógico se confirma con un commit en español.
