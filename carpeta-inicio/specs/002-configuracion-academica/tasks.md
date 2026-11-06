# Tareas: Configuración académica del colegio

**Input**: Documentos de diseño en `/specs/002-configuracion-academica/`

**Prerrequisitos**: [plan.md](plan.md), [spec.md](spec.md), [research.md](research.md), [data-model.md](data-model.md), [contracts/rutas.md](contracts/rutas.md)

**Pruebas**: obligatorias según la constitución (principio V). Se escriben antes de la implementación de cada regla.

**Organización**: tareas agrupadas por historia de usuario para que cada una sea implementable y verificable por separado.

## Formato: `[ID] [P?] [Historia] Descripción`

- **[P]**: puede ejecutarse en paralelo (archivos distintos, sin dependencias pendientes)
- **[Historia]**: US1 a US7 según `spec.md`
- Todas las descripciones incluyen la ruta exacta del archivo

## Convenciones de rutas

- Código: `internal/academico/`, `internal/plataforma/`, `web/plantillas/configuracion/`
- Migraciones: `migraciones/`
- Pruebas: `pruebas/unidad/`, `pruebas/integracion/`, `pruebas/manejadores/`, `pruebas/e2e/`

---

## Fase 1: Preparación

**Propósito**: constantes, utilidades y estructura compartida.

- [X] T001 [P] Crear `internal/academico/constantes.go` con `TurnoManana = "manana"`, `TurnoTarde = "tarde"`, `TipoFeriado`, `TipoNoLaborable`, `TipoSuspension`, `TipoUmbralTardanzas` y `TipoUmbralFaltas`
- [X] T002 [P] Ampliar `docs/glosario.md` con los términos `vigencia`, `bloque_clase`, `horario_seccion`, `excepcion_calendario`, `umbral`, `curso` y `turno_seccion`
- [X] T003 [P] Agregar helpers de hora en `internal/plataforma/fechas/fechas.go`: `ParsearHora`, `FormatearHora`, `SumarMinutos` y una función que compare horas del mismo día
- [X] T004 [P] Escribir pruebas de los helpers de hora en `pruebas/unidad/fechas_hora_test.go`
- [X] T005 [P] Crear la carpeta `web/plantillas/configuracion/` con `panel.templ` y la navegación entre secciones de configuración
- [X] T006 [P] Agregar el objetivo `templ` al `Makefile` con `templ generate ./web/plantillas` y documentarlo en `docs/despliegue.md`
- [X] T007 [P] Definir el tipo `Reloj func() time.Time` y su dependencia en `internal/academico/servicio_configuracion.go`, con valor por defecto `fechas.Hoy`

---

## Fase 2: Base bloqueante

**Propósito**: esquema, modelos, consultas, validaciones y rutas que necesitan todas las historias.

**⚠️ CRÍTICO**: ninguna historia puede empezar hasta completar esta fase.

### Migraciones

- [X] T008 [P] Crear `migraciones/0009_limitar_un_anio_activo.sql` con `cerrado_en timestamptz` y `CREATE UNIQUE INDEX un_anio_activo ON anios_escolares (estado) WHERE estado = 'activo'`
- [X] T009 [P] Crear `migraciones/0010_crear_tabla_turnos_seccion.sql` con `turnos_seccion (id, seccion_id FK, turno CHECK IN ('manana','tarde'), vigente_desde DATE, creado_en)`, `UNIQUE (seccion_id, vigente_desde)`; llenar con `vigente_desde = anios_escolares.fecha_inicio`; eliminar `secciones.turno`; el `Down` restaura la columna con el turno vigente
- [X] T010 [P] Crear `migraciones/0011_crear_tabla_horarios_turno.sql` con `horarios_turno (id, turno CHECK IN ('manana','tarde'), hora_ingreso TIME, tolerancia_minutos INTEGER CHECK >= 0, hora_limite_tardanza TIME, hora_cierre TIME, vigente_desde DATE, creado_en)`, `UNIQUE (turno, vigente_desde)`, `CHECK (hora_limite_tardanza > hora_ingreso)`, `CHECK (hora_cierre > hora_limite_tardanza)` y `CHECK (hora_ingreso + make_interval(mins => tolerancia_minutos) <= hora_limite_tardanza)`
- [X] T011 [P] Crear `migraciones/0012_crear_tablas_calendario.sql` con `excepciones_calendario (id, anio_escolar_id FK, tipo CHECK IN ('feriado','no_laborable','suspension'), fecha_inicio DATE, fecha_fin DATE, motivo, creado_por FK, creado_en)` con `CHECK (fecha_fin >= fecha_inicio)` y `excepciones_grados (excepcion_id FK, grado_id FK, PK compuesta)`
- [X] T012 [P] Crear `migraciones/0013_crear_tabla_umbrales_alerta.sql` con `umbrales_alerta (id, anio_escolar_id FK, tipo CHECK IN ('tardanzas','faltas'), valor INTEGER CHECK > 0, vigente_desde DATE, creado_en)` y `UNIQUE (anio_escolar_id, tipo, vigente_desde)`
- [X] T013 [P] Crear `migraciones/0014_crear_tablas_cursos.sql` con `cursos (id, anio_escolar_id FK, nombre, activo BOOLEAN DEFAULT true, creado_en)` e índice único `(anio_escolar_id, lower(nombre))`, y `cursos_grados (curso_id FK, grado_id FK, PK compuesta)`
- [X] T014 [P] Crear `migraciones/0015_crear_tabla_bloques_clase.sql` con `bloques_clase (id, turno CHECK IN ('manana','tarde'), numero INTEGER CHECK > 0, hora_inicio TIME, hora_fin TIME, vigente_desde DATE, creado_en)`, `UNIQUE (turno, numero, vigente_desde)` y `CHECK (hora_fin > hora_inicio)`
- [X] T015 [P] Crear `migraciones/0016_crear_tabla_horario_seccion.sql` con `horario_seccion (id, seccion_id FK, dia_semana SMALLINT CHECK BETWEEN 1 AND 5, bloque_numero INTEGER, curso_id FK NULL, vigente_desde DATE, creado_en)`, `UNIQUE (seccion_id, dia_semana, bloque_numero, vigente_desde)`
- [X] T016 [P] Crear `migraciones/0017_ampliar_entidades_auditoria.sql` que reemplaza el `CHECK` de `auditoria.entidad` agregando `anio_escolar`, `grado`, `seccion`, `turno`, `calendario`, `umbral`, `curso`, `bloque_clase` y `horario_seccion`
- [X] T017 Escribir la prueba de migraciones (subir y bajar todas las versiones) en `pruebas/integracion/migraciones_test.go`

### Modelos y consultas

- [X] T018 [P] Implementar los modelos `AnioEscolar`, `Grado` y `Seccion` en `internal/academico/modelo.go` (cierra T035 de la feature 001; `Seccion` conserva `Turno`)
- [X] T019 [P] Agregar los modelos `HorarioTurno`, `ExcepcionCalendario`, `UmbralAlerta`, `Curso`, `BloqueClase`, `HorarioSeccion` y `ConfiguracionFecha` en `internal/academico/modelo.go`
- [X] T020 [P] Escribir las consultas sqlc de años, grados, secciones y turnos en `internal/academico/consultas/academico.sql`
- [X] T021 [P] Escribir las consultas sqlc de horarios, excepciones, umbrales, cursos, bloques y horario semanal en `internal/academico/consultas/academico.sql`
- [X] T022 Regenerar el código de sqlc con `sqlc generate` y compilar `internal/academico/generado/`

### Repositorios

- [X] T023 [P] Implementar el repositorio de años, grados, secciones y turnos en `internal/academico/repositorio_configuracion.go`
- [X] T024 [P] Implementar el repositorio de horarios de turno, excepciones y umbrales en `internal/academico/repositorio_calendario.go`
- [X] T025 [P] Implementar el repositorio de cursos y vínculos con grados en `internal/academico/repositorio_cursos.go`
- [X] T026 [P] Implementar el repositorio de bloques y horario semanal en `internal/academico/repositorio_horarios.go`
- [X] T027 Actualizar `internal/academico/repositorio.go` para derivar el turno vigente en `ListarSecciones`, `ListarSeccionesVigentes` y `BuscarSeccionPorID` después de eliminar `secciones.turno`

### Validaciones puras

- [X] T028 [P] Escribir las pruebas unitarias de validación de horarios en `pruebas/unidad/validaciones_academico_test.go`: cierre posterior al límite, ingreso más tolerancia dentro del límite, sin cruzar la medianoche, superposición de mañana y tarde, vigencia no anterior a hoy
- [X] T029 [P] Escribir las pruebas unitarias de validación de bloques en `pruebas/unidad/validaciones_academico_test.go`: fin posterior a inicio, dentro del horario del turno, sin solapamiento, sin cruce de medianoche
- [X] T030 Implementar las funciones puras de validación en `internal/academico/validaciones.go`

### Integración transversal

- [X] T031 Implementar el registro de auditoría para las nuevas entidades en `internal/academico/servicio_auditoria.go`, usando la tabla `auditoria` existente
- [X] T032 Registrar las dependencias y rutas de configuración en `internal/plataforma/rutas/rutas.go` y `cmd/servidor/main.go`
- [X] T033 [P] Escribir la prueba de regresión de `academico.Seccion.Turno` con turno vigente en `pruebas/integracion/secciones_turno_test.go`
- [X] T034 [P] Escribir la prueba base de permisos de configuración (director puede GET; POST no permitidos responden 403) en `pruebas/manejadores/configuracion_permisos_test.go`

**Punto de control**: base lista; las historias pueden comenzar.

---

## Fase 3: Historia 1 — Año escolar, grados, secciones y turnos (P1) 🎯 MVP

**Objetivo**: crear el año escolar con fecha de inicio y fin, grados y secciones con turno, y cerrar el año sin perder historia.

**Prueba independiente**: crear un año 2027, dos grados con dos secciones cada uno, asignar turnos, cerrar el año y verificar que queda de solo lectura.

### Pruebas

- [X] T035 [P] [US1] Escribir pruebas unitarias de reglas de año, grado y sección en `pruebas/unidad/configuracion_academica_test.go`: fecha de fin posterior a inicio, un solo año activo, nombre de grado único por año, sección con turno único
- [X] T036 [P] [US1] Escribir pruebas de integración del repositorio de años, grados y secciones en `pruebas/integracion/academico_repositorio_test.go`

### Implementación

- [X] T037 [US1] Implementar `CrearAnio`, `CerrarAnio`, `CrearGrado`, `CrearSeccion`, `CambiarTurno` y los listados en `internal/academico/servicio_configuracion.go`
- [X] T038 [US1] Implementar el manejador de panel, año, grados y secciones en `internal/academico/manejador_configuracion.go` con las rutas de `contracts/rutas.md`
- [X] T039 [US1] Crear las plantillas `anio.templ`, `grados.templ` y `secciones.templ` en `web/plantillas/configuracion/`, con formularios accesibles y errores junto al campo
- [X] T040 [US1] Integrar auditoría en año, grado, sección y cambio de turno con las acciones `crear`, `cerrar` y `cambiar_turno`
- [X] T041 [US1] Escribir las pruebas de manejadores de US1 en `pruebas/manejadores/configuracion_anio_test.go`: alta válida, fechas inválidas, año activo duplicado, año cerrado y director en POST
- [X] T042 [US1] Escribir la prueba E2E de US1 en `pruebas/e2e/configuracion_anio_test.spec.ts`

**Punto de control**: US1 completamente funcional y verificable por separado.

---

## Fase 4: Historia 2 — Horarios y tolerancias con vigencia (P1)

**Objetivo**: configurar por turno ingreso, tolerancia, límite de tardanza y cierre, con vigencia desde una fecha e historia consultable.

**Prueba independiente**: configurar mañana con tolerancia 10, cambiarla a 5 con vigencia futura y comprobar que cada fecha conserva su valor.

### Pruebas

- [X] T043 [P] [US2] Escribir pruebas unitarias de configuración de horarios en `pruebas/unidad/horarios_turno_test.go`, incluida la fecha de vigencia anterior a hoy
- [X] T044 [P] [US2] Escribir pruebas de integración de versionado y consulta por fecha de `horarios_turno` en `pruebas/integracion/horarios_turno_test.go`

### Implementación

- [X] T045 [US2] Implementar `ConfigurarHorario`, `HorarioVigente` y `ListarHorarios` en `internal/academico/servicio_configuracion.go`, con validación de superposición entre turnos y bloqueo `pg_advisory_xact_lock`
- [X] T046 [US2] Implementar el manejador `GET /configuracion/turnos` y `POST /configuracion/turnos/{turno}` en `internal/academico/manejador_configuracion.go`
- [X] T047 [US2] Crear la plantilla `turnos.templ` en `web/plantillas/configuracion/` con valores vigentes, versión programada e historia
- [X] T048 [US2] Permitir el `POST` de turnos al director y mantener el resto de escrituras solo para administrador en `internal/plataforma/intermediarios/roles.go`
- [X] T049 [US2] Registrar la auditoría de `configurar_horario` con valor anterior, valor nuevo y `vigente_desde`
- [X] T050 [US2] Escribir las pruebas de manejadores de US2 en `pruebas/manejadores/configuracion_turnos_test.go`: validaciones 422, director 200, administrador 200
- [X] T051 [US2] Escribir la prueba E2E de US2 en `pruebas/e2e/configuracion_turnos_test.spec.ts`

**Punto de control**: horarios versionados con permisos correctos.

---

## Fase 5: Historia 6 — Bloques de hora de clase (P1)

**Objetivo**: definir por turno los bloques de clase con número, hora de inicio y fin, vigencia y ajuste de horas por el director.

**Prueba independiente**: crear el bloque 1 de 07:45 a 08:30, rechazar uno solapado y uno fuera del horario, y ajustar la hora de un bloque como director.

### Pruebas

- [X] T052 [P] [US6] Escribir pruebas unitarias de bloques en `pruebas/unidad/bloques_clase_test.go`
- [X] T053 [P] [US6] Escribir pruebas de integración de versionado y consulta de bloques en `pruebas/integracion/bloques_clase_test.go`

### Implementación

- [X] T054 [US6] Implementar `CrearBloque`, `AjustarHorasBloque` y `BloquesVigentes` en `internal/academico/servicio_horarios.go`
- [X] T055 [US6] Implementar el manejador de bloques (`GET /configuracion/bloques`, `POST /configuracion/bloques`, `POST /configuracion/bloques/{id}/horas`) en `internal/academico/manejador_horarios.go`
- [X] T056 [US6] Crear la plantilla `bloques.templ` en `web/plantillas/configuracion/` con la grilla del turno y las horas vigentes
- [X] T057 [US6] Aplicar permisos: solo administrador crea bloques; administrador y director ajustan horas en `internal/plataforma/intermediarios/roles.go`
- [X] T058 [US6] Registrar la auditoría de `crear_bloque` y `ajustar_bloque`
- [X] T059 [US6] Escribir las pruebas de manejadores de US6 en `pruebas/manejadores/configuracion_bloques_test.go`
- [X] T060 [US6] Escribir la prueba E2E de US6 en `pruebas/e2e/configuracion_bloques_test.spec.ts`

**Punto de control**: bloques versionados y ajustables según rol.

---

## Fase 6: Historia 7 — Cursos y horario semanal (P1)

**Objetivo**: crear cursos vinculados a grados y armar el horario semanal de cada sección, con un curso por día y bloque.

**Prueba independiente**: crear cursos, armar el horario de 3.° A, rechazar un curso de otro grado y consultar el curso de un día y bloque.

### Pruebas

- [X] T061 [P] [US7] Escribir pruebas unitarias de cursos y celdas de horario en `pruebas/unidad/cursos_horario_test.go`: nombre único por año, al menos un grado, curso activo, curso vinculado al grado de la sección
- [X] T062 [P] [US7] Escribir pruebas de integración de cursos y versionado por celda en `pruebas/integracion/cursos_horario_test.go`

### Implementación

- [X] T063 [US7] Implementar `CrearCurso`, `DesactivarCurso`, `ListarCursos` y la validación de vínculo con grados en `internal/academico/servicio_cursos.go`
- [X] T064 [US7] Implementar `ObtenerGrilla`, `GuardarCeldas` y `CursoDeBloque` en `internal/academico/servicio_horarios.go`
- [X] T065 [US7] Implementar el manejador de cursos en `internal/academico/manejador_cursos.go`
- [X] T066 [US7] Implementar el manejador de horario semanal (con confirmación de reemplazo) en `internal/academico/manejador_horarios.go`
- [X] T067 [US7] Crear las plantillas `cursos.templ` y `horario.templ` en `web/plantillas/configuracion/`
- [X] T068 [US7] Aplicar permisos de solo administrador para cursos y horario semanal en `internal/plataforma/intermediarios/roles.go`
- [X] T069 [US7] Registrar la auditoría de `crear_curso`, `desactivar_curso` y `asignar_horario`
- [X] T070 [US7] Escribir las pruebas de manejadores de US7 en `pruebas/manejadores/configuracion_cursos_test.go`
- [X] T071 [US7] Escribir la prueba E2E de US7 en `pruebas/e2e/configuracion_horario_test.spec.ts`

**Punto de control**: cursos y horario semanal funcionales.

---

## Fase 7: Historia 3 — Feriados, no laborables y suspensiones (P2)

**Objetivo**: registrar excepciones de calendario por día o rango, con alcance global o por grado y confirmación ante solapamientos.

**Prueba independiente**: registrar un feriado global y una suspensión solo para 4.°, verificar su alcance y la pantalla de confirmación ante un solapamiento.

### Pruebas

- [X] T072 [P] [US3] Escribir pruebas unitarias de excepciones en `pruebas/unidad/calendario_test.go`: rango válido, fin de semana, alcance global, alcance por grado, solapamiento
- [X] T073 [P] [US3] Escribir pruebas de integración de excepciones y consulta por sección y fecha en `pruebas/integracion/calendario_test.go`

### Implementación

- [X] T074 [US3] Implementar `RegistrarExcepcion`, `ExcepcionesVigentes` y `DetectarSolapamiento` en `internal/academico/servicio_calendario.go`
- [X] T075 [US3] Implementar el manejador de calendario con confirmación de solapamiento en `internal/academico/manejador_calendario.go`
- [X] T076 [US3] Crear la plantilla `calendario.templ` en `web/plantillas/configuracion/` con filtros por mes y grado
- [X] T077 [US3] Registrar la auditoría de `registrar_excepcion`
- [X] T078 [US3] Escribir las pruebas de manejadores de US3 en `pruebas/manejadores/configuracion_calendario_test.go`
- [X] T079 [US3] Escribir la prueba E2E de US3 en `pruebas/e2e/configuracion_calendario_test.spec.ts`

**Punto de control**: calendario con alcance correcto y solapamientos advertidos.

---

## Fase 8: Historia 4 — Umbrales de alerta (P3)

**Objetivo**: configurar umbrales mensuales de tardanzas y faltas con historia por vigencia.

**Prueba independiente**: guardar 3 tardanzas, cambiarlo a 4 con vigencia futura y verificar la historia; rechazar cero y negativos.

### Pruebas

- [X] T080 [P] [US4] Escribir pruebas unitarias de umbrales en `pruebas/unidad/umbrales_test.go`
- [X] T081 [P] [US4] Escribir pruebas de integración de versionado de umbrales en `pruebas/integracion/umbrales_test.go`

### Implementación

- [X] T082 [US4] Implementar `ConfigurarUmbral` y `UmbralesVigentes` en `internal/academico/servicio_calendario.go`
- [X] T083 [US4] Implementar las rutas `GET` y `POST /configuracion/umbrales` en `internal/academico/manejador_calendario.go`
- [X] T084 [US4] Crear la plantilla `umbrales.templ` en `web/plantillas/configuracion/`
- [X] T085 [US4] Registrar la auditoría de `configurar_umbral`
- [X] T086 [US4] Escribir las pruebas de manejadores de US4 en `pruebas/manejadores/configuracion_umbrales_test.go`
- [X] T087 [US4] Escribir la prueba E2E de US4 en `pruebas/e2e/configuracion_umbrales_test.spec.ts`

**Punto de control**: umbrales con vigencia e historia.

---

## Fase 9: Historia 5 — Consulta de configuración histórica (P3)

**Objetivo**: consultar la configuración vigente en cualquier fecha: horarios, excepciones, umbrales, turno de cada sección, bloques y cursos.

**Prueba independiente**: con dos cambios de tolerancia, consultar tres fechas y comprobar el valor correcto; consultar una fecha fuera del año y ver el mensaje correspondiente.

### Pruebas

- [X] T088 [P] [US5] Escribir pruebas unitarias de selección de versión por fecha y del año que cubre la fecha en `pruebas/unidad/consulta_historica_test.go`
- [X] T089 [P] [US5] Escribir pruebas de integración de la consulta histórica completa en `pruebas/integracion/consulta_historica_test.go`

### Implementación

- [X] T090 [US5] Implementar `ConfiguracionEnFecha` en `internal/academico/consulta_historica.go`
- [X] T091 [US5] Implementar la ruta `GET /configuracion/consulta?fecha=` en `internal/academico/manejador_configuracion.go`
- [X] T092 [US5] Crear la plantilla `consulta.templ` en `web/plantillas/configuracion/` con el detalle por turno, sección, excepción y umbral
- [X] T093 [US5] Escribir las pruebas de manejadores de US5 en `pruebas/manejadores/configuracion_consulta_test.go`
- [X] T094 [US5] Escribir la prueba E2E de US5 en `pruebas/e2e/configuracion_consulta_test.spec.ts`

**Punto de control**: todas las historias funcionan y la consulta histórica responde con la versión correcta.

---

## Fase 10: Pulido y transversales

**Propósito**: permisos, seguridad, documentación y validación final.

- [X] T095 [P] Completar la prueba de permisos del director en `pruebas/manejadores/configuracion_permisos_test.go`: 403 en todos los POST salvo turnos y horas de bloques
- [X] T096 [P] Escribir la prueba de CSRF y validación de formularios de configuración en `pruebas/manejadores/configuracion_seguridad_test.go`
- [X] T097 Escribir la prueba de auditoría de todas las acciones de configuración en `pruebas/integracion/academico_auditoria_test.go`
- [X] T098 Verificar la compatibilidad con la feature 001: `go test ./...` y las pruebas E2E existentes de alumnos y carnés en `pruebas/e2e/`
- [X] T099 Revisar el rendimiento con `EXPLAIN ANALYZE` de la consulta histórica y los listados en `internal/academico/consultas/academico.sql`
- [X] T100 Regenerar plantillas y estilos: `templ generate`, `npm run css`, `gofmt -w`, `go vet ./...` y `golangci-lint run ./...`
- [X] T101 Actualizar `docs/manual_administrador.md` y `docs/manual_director.md` con las nuevas pantallas y permisos
- [X] T102 Actualizar `docs/despliegue.md` con el objetivo `templ` y confirmar que no hay variables de entorno nuevas
- [X] T103 Validar los escenarios de `specs/002-configuracion-academica/quickstart.md` y registrar los resultados en el mismo archivo
- [X] T104 Comprobar que ninguna sección quedó sin turno tras la migración 0010 y que un año cerrado no admite escrituras en `pruebas/integracion/migraciones_test.go`
- [X] T105 Actualizar `specs/001-gestion-alumnos-qr/quickstart.md` si la migración cambia algún dato de demostración

---

## Dependencias y orden de ejecución

### Dependencias de fases

- **Preparación (Fase 1)**: sin dependencias.
- **Base (Fase 2)**: depende de la Fase 1 y bloquea todas las historias.
- **Historias (Fases 3 a 9)**: dependen de la Fase 2.
- **Pulido (Fase 10)**: depende de todas las historias entregadas.

### Dependencias entre historias

| Historia | Depende de | Motivo |
|---|---|---|
| US1 (año, grados, secciones) | Base | Es el punto de partida del resto. |
| US2 (horarios de turno) | US1 | Necesita turnos y secciones para validar superposiciones. |
| US6 (bloques) | US2 | El bloque se valida contra el horario del turno. |
| US7 (cursos y horario) | US1, US6 | Necesita secciones y bloques vigentes. |
| US3 (calendario) | US1 | Necesita año y grados. |
| US4 (umbrales) | US1 | Necesita año escolar activo. |
| US5 (consulta histórica) | US1 a US4, US6, US7 | Consolida todas las versiones. |

### Dentro de cada historia

1. Pruebas primero y en rojo.
2. Modelos y repositorios antes que servicios.
3. Servicios antes que manejadores.
4. Manejadores antes que plantillas.
5. Auditoría y permisos junto con la ruta que los necesita.

### Oportunidades de paralelismo

- T001 a T007 (preparación) en paralelo.
- T008 a T017 (migraciones) en paralelo, salvo T017 que depende de todas.
- T018 a T021, T023 a T026 y T028 a T029 en paralelo.
- Dentro de cada historia, las pruebas `[P]` de unidad e integración en paralelo.
- US3 y US4 pueden ejecutarse en paralelo después de US1.
- US5 debe ir al final.

---

## Ejemplo paralelo: preparación

```text
Tarea: Crear `internal/academico/constantes.go`
Tarea: Ampliar `docs/glosario.md`
Tarea: Agregar helpers de hora en `internal/plataforma/fechas/fechas.go`
Tarea: Crear `web/plantillas/configuracion/panel.templ`
```

## Ejemplo paralelo: Historia 1

```text
Tarea: Pruebas unitarias de reglas de año, grado y sección
Tarea: Pruebas de integración del repositorio académico
```

---

## Estrategia de implementación

### MVP primero (US1)

1. Completar Fase 1 y Fase 2.
2. Completar US1 (registrar año, grados, secciones y turnos).
3. **Detenerse y validar**: crear el año completo y cerrarlo.
4. Desplegar o demostrar si está listo.

### Entrega incremental

1. Preparación y base.
2. US1 → validar (MVP).
3. US2 → validar horarios con vigencia.
4. US6 → validar bloques.
5. US7 → validar cursos y horario semanal.
6. US3 → validar calendario.
7. US4 → validar umbrales.
8. US5 → validar consulta histórica.
9. Pulido y documentación.

Cada historia agrega valor sin romper las anteriores.

---

## Notas

- `[P]` indica archivos distintos y sin dependencias pendientes.
- La etiqueta `[USx]` permite rastrear cada tarea contra la especificación.
- Las pruebas se escriben antes de la regla que cubren, según la constitución.
- Cada tarea o grupo lógico se confirma con un commit en español.
- La primera tarea de implementación cierra **T035** de la feature 001.
