# Plan de implementación: Configuración académica del colegio

**Branch**: `002-configuracion-academica` | **Fecha**: 2026-11-06 | **Spec**: [spec.md](spec.md)

**Input**: Especificación de la feature en `/specs/002-configuracion-academica/spec.md`

## Resumen

El administrador crea el año escolar, sus grados y secciones, asigna turnos, configura los
horarios y tolerancias de cada turno con vigencia desde una fecha, registra feriados y
suspensiones, define umbrales de alerta, crea cursos, define los bloques de hora de clase y arma
el horario semanal de cada sección. El director consulta toda la configuración y solo puede
ajustar horas de turnos y de bloques existentes. Cualquier fecha del año escolar se puede
consultar con los valores que estaban vigentes en ese momento.

Enfoque técnico: se amplía el módulo `academico` existente con un modelo temporal **solo de
agregado** (`vigente_desde`), sin filas que se actualicen en el lugar. Cada cambio inserta una
versión nueva y las consultas toman la versión más reciente anterior o igual a la fecha pedida.
Esto cumple los requisitos de vigencia e historia sin duplicar tablas ni usar una librería
bitemporal. La interfaz se renderiza con templ + HTMX + Tailwind, los datos viven en PostgreSQL
con pgx y sqlc, y las migraciones se versionan con goose.

## Contexto técnico

**Lenguaje y versión**: Go estable reciente, el mismo de la feature 001 (`go.mod`).

**Dependencias principales**:
- Chi, templ, HTMX y Tailwind CSS (ya presentes).
- pgx y sqlc (consultas tipadas), goose (migraciones).
- scs y argon2id (sesiones y contraseñas, ya presentes).
- slog, testing, testify y testcontainers-go (ya presentes).
- Playwright para los flujos críticos (ya configurado en `pruebas/e2e`).

No se agregan dependencias nuevas.

**Almacenamiento**: PostgreSQL. Todas las tablas nuevas usan `date` para vigencias y `time` para
horas del día. Las marcas de creación y auditoría usan `timestamptz` en UTC.

**Pruebas**:
- Reglas puras (validación de horarios, solapamientos, selección de versión vigente) con tablas
  de casos en `pruebas/unidad`.
- Repositorios contra PostgreSQL real con testcontainers en `pruebas/integracion`.
- Manejadores con `net/http/httptest` en `pruebas/manejadores`.
- Flujo completo con Playwright: crear año, secciones, horario de turno, feriado, curso, bloques,
  horario semanal y consulta histórica.

**Plataforma objetivo**: contenedor Docker en Linux, navegadores actuales en computadora, tablet
y celular.

**Tipo de proyecto**: aplicación web monolítica con renderizado en el servidor.

**Metas de rendimiento** (de la spec):
- Crear un año escolar completo en menos de 60 minutos (SC-001).
- Registrar un feriado o suspensión en menos de 1 minuto (SC-004).
- Ajustar una hora de ingreso en menos de 1 minuto (SC-006).
- Armar el horario semanal de una sección (5 días × 6 bloques) en menos de 10 minutos (SC-007).
- Consulta histórica de configuración en menos de 500 ms con datos de un colegio de 2,000
  alumnos.
- Las páginas de configuración responden en menos de 1 segundo.

**Restricciones**:
- Fechas de vigencia en America/Lima; horas del día sin zona horaria.
- Solo dos turnos fijos: `manana` y `tarde`. No se crean ni eliminan.
- Un solo año escolar activo a la vez (FR-002).
- Un año cerrado no admite cambios (FR-003).
- Los cambios de vigencia solo pueden ser hoy o fechas futuras (FR-013).
- No se recalcula ni modifica ninguna asistencia ya registrada (FR-014, FR-038).
- Sin notificaciones a apoderados; los umbrales solo se configuran en esta feature.
- Monolito modular, sin microservicios, colas, Redis ni Kubernetes.

**Escala y alcance**: un colegio, unos 2,000 alumnos, un año activo, 5 grados con 2 secciones
cada uno (10 a 20 secciones), 2 turnos, hasta 8 bloques por turno, alrededor de 15 cursos.

## Revisión de la constitución

*Compuerta: debe cumplirse antes de la investigación. Se revisa de nuevo después del diseño.*

| Principio | Estado | Observación |
|---|---|---|
| I. Simplicidad | Cumple | Se amplía el módulo `academico` existente; ninguna dependencia nueva. El versionado se resuelve con columnas `vigente_desde` y consultas `ORDER BY ... LIMIT 1`, sin tablas espejo ni librerías bitemporales. |
| II. Código Go y monolito | Cumple | Capas manejador, servicio, repositorio y modelo dentro de `internal/academico`. `alumnos` sigue accediendo solo por la interfaz `BuscadorSeccion`. |
| III. Experiencia de usuario | Cumple | Vistas responsivas mobile-first, etiquetas accesibles y mensajes de error en español con el motivo. |
| IV. Integridad y seguridad | Cumple | CSRF en todos los POST, autorización por rol, restricción de un solo año activo y auditoría de cada cambio con valor anterior, nuevo y vigencia. |
| V. Pruebas | Cumple | Reglas puras con tablas de casos, repositorios con PostgreSQL real, manejadores con `httptest` y flujos Playwright. |
| VI. Reglas configurables | Cumple | Horarios, tolerancias, bloques y umbrales viven en la base de datos y se consultan por fecha; ninguna regla queda como constante en el código. |
| VII. Datos de menores | Cumple | La configuración no contiene datos personales de alumnos; solo estructura académica. |
| VIII. Entrega operable | Cumple | Migraciones goose, `.env.ejemplo` sin cambios y documentación de operación actualizada. |
| IX. Idioma y nomenclatura | Cumple | Tablas, columnas, rutas y tipos en español sin tildes; textos visibles con tildes. |

Resultado: sin violaciones no justificadas.

## Estructura del proyecto

### Documentación de esta feature

```text
specs/002-configuracion-academica/
├── spec.md              # Especificación
├── plan.md              # Este archivo
├── research.md          # Decisiones técnicas
├── data-model.md        # Modelo de datos
├── quickstart.md        # Guía para validar la feature
├── contracts/
│   └── rutas.md         # Rutas HTTP y permisos
└── checklists/
    └── requirements.md
```

### Código fuente (raíz del repositorio)

```text
internal/academico/
├── modelo.go                    # AnioEscolar, Grado, Seccion, Turno, HorarioTurno, etc.
├── validaciones.go              # Funciones puras de validación de horarios y vigencias
├── servicio.go                  # BuscadorSeccion existente (feature 001)
├── servicio_configuracion.go    # Años, grados, secciones y turnos
├── servicio_calendario.go       # Feriados, suspensiones y umbrales
├── servicio_cursos.go           # Cursos y vínculos con grados
├── servicio_horarios.go         # Bloques y horario semanal
├── consulta_historica.go        # Configuración vigente en una fecha
├── repositorio.go               # Repositorio existente (feature 001)
├── repositorio_configuracion.go # Consultas de configuración
├── manejador_configuracion.go   # Rutas de año, grados, secciones y turnos
├── manejador_calendario.go      # Rutas de calendario y umbrales
├── manejador_cursos.go          # Rutas de cursos
├── manejador_horarios.go        # Rutas de bloques y horario semanal
├── consultas/
│   └── academico.sql            # Consultas sqlc (se amplía el archivo existente)
└── generado/                    # Salida de sqlc

web/plantillas/configuracion/
├── panel.templ                  # Resumen de configuración del año activo
├── anio.templ                   # Creación y cierre del año escolar
├── grados.templ                 # Grados del año
├── secciones.templ              # Secciones y turno vigente
├── turnos.templ                 # Horarios y tolerancias por turno
├── calendario.templ             # Feriados y suspensiones
├── umbrales.templ               # Umbrales de alerta
├── cursos.templ                 # Cursos y grados vinculados
├── bloques.templ                # Bloques de hora de clase
├── horario.templ                # Horario semanal por sección
└── consulta.templ               # Consulta histórica por fecha

migraciones/
├── 0009_limitar_un_anio_activo.sql
├── 0010_crear_tabla_turnos_seccion.sql
├── 0011_crear_tabla_horarios_turno.sql
├── 0012_crear_tablas_calendario.sql
├── 0013_crear_tabla_umbrales_alerta.sql
├── 0014_crear_tablas_cursos.sql
├── 0015_crear_tabla_bloques_clase.sql
├── 0016_crear_tabla_horario_seccion.sql
└── 0017_ampliar_entidades_auditoria.sql
```

**Decisión de estructura**: no se crea un módulo nuevo. Años, grados, secciones, turnos,
calendario, umbrales, cursos, bloques y horarios forman un único dominio académico y comparten
las mismas reglas de vigencia y auditoría. Mantenerlos en `internal/academico` evita ciclos de
importación y respeta el límite de módulos de la constitución.

## Fases

### Fase 0: Investigación

Ver [research.md](research.md). Resuelve:

1. Modelo de versionado temporal por fecha de vigencia.
2. Estrategia para el turno histórico de una sección sin duplicar la fuente de verdad.
3. Validación y concurrencia de horarios y bloques solapados.
4. Vínculo curso-grado-sección y desactivación sin borrado.
5. Autorización fina del director sobre dos tipos de edición.
6. Extensión de la auditoría existente.

### Fase 1: Diseño

Ver [data-model.md](data-model.md), [contracts/rutas.md](contracts/rutas.md) y
[quickstart.md](quickstart.md).

### Fase 2: Tareas

`tasks.md` se genera después con `/speckit.tasks`. La primera tarea debe ser cerrar **T035** de la
feature 001: agregar los modelos `AnioEscolar` y `Grado` en `internal/academico/modelo.go`.

## Integración con la feature 001

- Se elimina la columna `secciones.turno` y su valor se traslada a `turnos_seccion` con vigencia
  desde la fecha de inicio del año escolar.
- Las consultas de `ListarSecciones`, `ListarSeccionesVigentes` y `BuscarSeccionPorID` derivan el
  turno vigente de cada sección. El tipo `academico.Seccion` conserva el campo `Turno`, por lo que
  el módulo `alumnos` y los carnés no cambian su contrato.
- La importación de alumnos sigue usando `BuscarSeccion(anioID, grado, seccion)` sin cambios.
- La migración de datos es reversible: el `Down` vuelve a crear `secciones.turno` con el turno
  vigente hoy.

## Complejidad justificada

| Elemento | Por qué se necesita | Alternativa simple descartada |
|---|---|---|
| Versionado solo de agregado con `vigente_desde` | La spec exige historia exacta por fecha y prohibe recalcular asistencias pasadas (FR-012 a FR-014, FR-023, FR-038). | Actualizar filas en el lugar y guardar solo la última versión: pierde la consulta histórica. |
| Historial de turnos por sección en tabla aparte | FR-028 exige consultar el turno de una sección en cualquier fecha. | Mantener `secciones.turno` y una tabla de auditoría: la consulta histórica dependería de interpretar JSON de auditoría. |
| Validación de solapamientos en el servicio con bloqueo por turno | Los bloques y horarios cambian con fechas distintas; una restricción de exclusión en SQL no cubre todos los casos sin volver el esquema difícil de mantener. | Confiar solo en índices únicos: no detecta solapamientos con vigencias distintas. |

No se introduce ninguna dependencia, servicio ni infraestructura adicional.
