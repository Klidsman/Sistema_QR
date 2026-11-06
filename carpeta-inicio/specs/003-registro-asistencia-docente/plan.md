# Plan de implementación: Registro de asistencia de alumnos por el docente

**Branch**: `003-registro-asistencia-docente` | **Fecha**: 2026-11-06 | **Spec**: [spec.md](spec.md)

**Input**: Especificación de la feature en `/specs/003-registro-asistencia-docente/spec.md`

## Resumen

El docente abre una de sus horas de clase asignadas, ve a los alumnos activos de la sección con
su ingreso por QR del día y registra un estado personal: presente, tardanza, falta o falta
justificada. Puede marcar a todos presentes y corregir excepciones. Cada registro se guarda por
fecha, curso, sección y bloque, no se mezcla con el QR y queda auditado. El administrador asigna
docentes a cursos y secciones, incluidos reemplazos por rango de días, crea cuentas de docente y
configura el plazo de edición. El director consulta los registros sin modificarlos. El docente
exporta su registro a CSV.

Enfoque técnico: nuevo módulo `internal/asistencia` con capas manejador, servicio, repositorio y
modelo. Se comunica con `academico` y `alumnos` mediante interfaces pequeñas, sin consultar sus
tablas. Las cuentas de docente viven en `autenticacion`. Se agrega una tabla mínima de
`marcaciones` QR que esta feature solo lee; la pantalla de portería queda para una feature
posterior. Sin modo sin conexión: queda diferido por decisión y por el principio III.

## Contexto técnico

**Lenguaje y versión**: Go estable reciente, el mismo de las features 001 y 002.

**Dependencias principales**: Chi, templ, HTMX, Tailwind, pgx, sqlc, goose, scs, argon2id, slog,
testing, testify, testcontainers y Playwright. **Sin dependencias nuevas**; la exportación usa
`encoding/csv` de la biblioteca estándar.

**Almacenamiento**: PostgreSQL. Registros por fecha, curso, sección y bloque. Marcaciones QR
mínimas para lectura. Parámetro de plazo de edición en tabla propia.

**Pruebas**:
- Reglas puras: diferencia QR/docente, plazo de edición, estados y validación de observación.
- Repositorios contra PostgreSQL real con testcontainers.
- Manejadores con `httptest`, incluidos permisos por rol y CSRF.
- E2E con Playwright: alta de docente, asignación, registro, corrección y exportación.

**Plataforma objetivo**: contenedor Docker en Linux; navegadores actuales en celular, tablet y
computadora.

**Tipo de proyecto**: aplicación web monolítica con renderizado en el servidor.

**Metas de rendimiento** (de la spec):
- Registrar 40 alumnos en menos de 2 minutos con "Marcar a todos presentes" (SC-001).
- Lista de clase y guardado en menos de 500 ms en red local.
- Exportación de un mes de un curso en menos de 3 segundos.

**Restricciones**:
- No se registra asistencia de docentes (FR-024).
- Solo horas de clase; no existe registro por turno (FR-022).
- El QR es la asistencia oficial; el registro docente es personal (FR-023).
- No se permiten fechas futuras (FR-012).
- Plazo de edición configurable, 1 a 168 horas, valor inicial 24 (FR-009, FR-028).
- Observación máxima de 300 caracteres (FR-006).
- Una fila por alumno, fecha, curso, sección y bloque (FR-008).
- Sin offline en esta versión (FR-017/018 diferidos).
- Fechas y horas del colegio en America/Lima; marcas de tiempo en UTC.

**Escala y alcance**: un colegio, unos 2,000 alumnos, hasta 20 secciones, hasta 8 bloques por
turno, decenas de docentes y alrededor de 40 registros por clase.

## Revisión de la constitución

*Compuerta: debe cumplirse antes de la investigación. Se revisa de nuevo después del diseño.*

| Principio | Estado | Observación |
|---|---|---|
| I. Simplicidad | Cumple | Módulo nuevo con capas simples y sin dependencias nuevas. Se reutilizan academico, alumnos y autenticacion mediante interfaces. |
| II. Código Go y monolito | Cumple | `asistencia` no consulta tablas de `academico` ni de `alumnos`; usa interfaces `ResolutorClase` y `ListadorAlumnos`. |
| III. Experiencia de usuario | Cumple | Lista mobile-first, acciones grandes y pocos pasos. El modo sin conexión queda diferido y documentado. |
| IV. Integridad y seguridad | Cumple | Registro único por clave natural, auditoría append-only, CSRF, roles y motivo obligatorio del administrador. |
| V. Pruebas | Cumple | Reglas puras con tablas de casos, PostgreSQL real, `httptest` y Playwright. Reloj inyectado en el plazo. |
| VI. Reglas configurables | Cumple | Plazo de edición en base de datos, no como constante. |
| VII. Datos de menores | Cumple | Solo datos necesarios; lista accesible únicamente al docente asignado, administrador y director. |
| VIII. Entrega operable | Cumple | Migraciones goose, `/salud` existente y documentación actualizada. |
| IX. Idioma y nomenclatura | Cumple | Tablas, rutas y tipos en español sin tildes; textos visibles con tildes. |

Resultado: sin violaciones no justificadas.

## Estructura del proyecto

### Documentación de esta feature

```text
specs/003-registro-asistencia-docente/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── rutas.md
└── checklists/
    └── requirements.md
```

### Código fuente

```text
internal/asistencia/
├── modelo.go                    # AsignacionDocente, RegistroDocente, Marcacion, Parametros
├── validaciones.go              # Funciones puras: diferencia, plazo, estados, observación
├── servicio_asignaciones.go
├── servicio_registros.go
├── servicio_consulta.go
├── servicio_exportacion.go
├── servicio_parametros.go
├── repositorio.go
├── repositorio_asignaciones.go
├── repositorio_registros.go
├── repositorio_marcaciones.go
├── repositorio_parametros.go
├── manejador_asistencia.go
├── manejador_asignaciones.go
├── manejador_consulta.go
├── manejador_parametros.go
├── consultas/
│   └── asistencia.sql
└── generado/

internal/autenticacion/
├── manejador_docentes.go        # Alta, listado, desactivación y clave de docentes
└── servicio_docentes.go

internal/academico/
├── consulta_clase.go            # ResolutorClase para asistencia
└── consultas/academico.sql      # Consultas de clase y día no laborable

internal/alumnos/
├── servicio_consulta.go         # ListadorAlumnos para asistencia
└── consultas/alumnos.sql        # Alumnos activos por sección en una fecha

web/plantillas/asistencia/
├── seleccion.templ              # Cursos y secciones del docente en una fecha
├── clase.templ                  # Lista de alumnos y registro
├── consulta.templ               # Consulta de administrador y director
└── exportacion.templ            # Filtro de fechas
web/plantillas/docentes/
├── listado.templ
└── formulario.templ
web/plantillas/configuracion/
└── parametros_asistencia.templ  # Plazo de edición

migraciones/
├── 0018_crear_parametros_asistencia.sql
├── 0019_crear_tabla_marcaciones.sql
├── 0020_crear_tabla_asignaciones_docente.sql
├── 0021_crear_tabla_registros_docente.sql
└── 0022_ampliar_auditoria_asistencia.sql
```

**Decisión de estructura**: `asistencia` es un módulo propio porque tiene reglas, datos y roles
distintos de `alumnos` y `academico`. Las cuentas de docente pertenecen a `autenticacion`.
`asistencia` no lee tablas ajenas: `academico` expone `ResolutorClase` y `alumnos` expone
`ListadorAlumnosSeccion`.

## Fases

### Fase 0: Investigación

Ver [research.md](research.md). Resuelve:

1. Dependencia del ingreso QR y tabla mínima de marcaciones.
2. Resolución de una hora de clase (fecha, sección, curso, bloque).
3. Asignaciones con titular y reemplazo por rango.
4. Plazo de edición y corrección del administrador.
5. Cálculo de diferencias sin destacar el QR como autoridad.
6. Exportación CSV e idempotencia.
7. Límites entre módulos.

### Fase 1: Diseño

Ver [data-model.md](data-model.md), [contracts/rutas.md](contracts/rutas.md) y
[quickstart.md](quickstart.md).

### Fase 2: Tareas

`tasks.md` se genera con `/speckit.tasks`. Orden previsto:

1. Migraciones, modelos y validaciones puras.
2. Cuentas de docente y parámetro de plazo.
3. Asignaciones.
4. Resolución de clase y lista de alumnos.
5. Registro docente, marcado masivo y correcciones.
6. Consulta del director y exportación.
7. E2E, documentación y pulido.

## Reglas de negocio clave

- **Clase válida**: fecha no futura, día laborable sin excepción para la sección, turno vigente de
  la sección, bloque vigente del turno, celda del horario semanal con el curso solicitado y
  asignación docente vigente en la fecha.
- **Alumno en la lista**: estado `activo` y tramo de `historial_secciones` que cubre la fecha.
- **Guardado**: un registro por alumno, fecha, curso, sección y bloque. Repetir el envío
  actualiza el registro existente si el usuario tiene permiso; nunca duplica.
- **Plazo**: el docente puede crear y corregir hasta `creado_en + plazo_edicion_horas`. Después
  solo el administrador corrige, con motivo obligatorio.
- **Marcar a todos presentes**: crea únicamente los registros faltantes; no toca los existentes.
- **Diferencia**: hay diferencia si el QR registra ingreso y el docente marca falta o falta
  justificada, o si no hay QR y el docente marca presente o tardanza. Se muestra como aviso
  informativo; el QR sigue siendo la asistencia oficial.
- **Auditoría**: cada alta o cambio guarda usuario, fecha y hora, valor anterior, valor nuevo y
  motivo cuando aplica, en la misma transacción.

## Complejidad justificada

| Elemento | Por qué se necesita | Alternativa simple descartada |
|---|---|---|
| Módulo `asistencia` separado | Tiene reglas, tablas y permisos propios; mezclarlo con `alumnos` rompería los límites de la constitución. | Agregar los registros al módulo `alumnos`: mezcla responsabilidades y crece el módulo base. |
| Interfaces `ResolutorClase` y `ListadorAlumnosSeccion` | Evitan que `asistencia` consulte tablas de otros módulos. | Consultas SQL cruzadas: violan el principio II. |
| Tabla `marcaciones` mínima | FR-003 y FR-029 necesitan leer el ingreso QR, y hoy no existe. | Mostrar siempre "sin ingreso": incumple la spec. |
| Parámetro en base de datos | FR-028 exige plazo configurable; una constante viola el principio VI. | Variable de entorno: no permite que el administrador lo ajuste. |
| CSV con `encoding/csv` | FR-026 pide descarga y no se necesita formato con estilos. | excelize: dependencia ya disponible pero innecesaria para CSV. |
