# Modelo de datos: Configuración académica del colegio

**Fecha**: 2026-11-06 | **Plan**: [plan.md](plan.md) | **Investigación**: [research.md](research.md)

Convenciones (constitución, principio IX): tablas en plural, columnas en snake_case, sin tildes ni
ñ en nombres de tablas y columnas. Fechas de vigencia en `date` (America/Lima), horas del día en
`time`, marcas de tiempo en `timestamptz` (UTC). Los textos visibles llevan tildes.

## Regla general de vigencia

Todas las tablas versionadas son **solo de agregado**. Cada fila representa una versión y lleva
`vigente_desde DATE NOT NULL`. Un cambio inserta una fila nueva; las filas anteriores no se
actualizan ni se borran. La versión vigente en una fecha se obtiene con:

```sql
ORDER BY vigente_desde DESC LIMIT 1  -- con WHERE vigente_desde <= :fecha
```

Índice obligatorio en cada tabla versionada: `(clave, vigente_desde DESC)`.

## Tablas existentes que se modifican

### anios_escolares

| Columna | Tipo | Reglas |
|---|---|---|
| id | bigint, PK | generado |
| nombre | text | único; no nulo |
| fecha_inicio | date | no nulo |
| fecha_fin | date | no nulo; mayor que `fecha_inicio` |
| estado | text | `activo` o `cerrado` |
| cerrado_en | timestamptz | nulo mientras esté activo |

Cambios:
- `CREATE UNIQUE INDEX un_anio_activo ON anios_escolares (estado) WHERE estado = 'activo'`
  (FR-002).
- Un año `cerrado` no admite escrituras; la comprobación vive en el servicio (FR-003).

### secciones

Se elimina la columna `turno`. El turno pasa a `turnos_seccion` y las consultas lo derivan. Los
campos `id`, `grado_id` y `nombre` no cambian.

## Tablas nuevas

### turnos_seccion (FR-005, FR-028)

| Columna | Tipo | Reglas |
|---|---|---|
| id | bigint, PK | generado |
| seccion_id | bigint, FK secciones | no nulo |
| turno | text | `manana` o `tarde` |
| vigente_desde | date | no nulo |
| creado_en | timestamptz | no nulo |

Restricciones: `UNIQUE (seccion_id, vigente_desde)`.
Índice: `(seccion_id, vigente_desde DESC)`.
Migración de datos: una fila por cada sección existente con `vigente_desde = fecha_inicio` del
año escolar de su grado.

### horarios_turno (FR-007 a FR-014)

| Columna | Tipo | Reglas |
|---|---|---|
| id | bigint, PK | generado |
| turno | text | `manana` o `tarde` |
| hora_ingreso | time | no nulo |
| tolerancia_minutos | integer | mayor o igual que 0; máximo 120 |
| hora_limite_tardanza | time | no nulo; posterior a `hora_ingreso` |
| hora_cierre | time | no nulo; posterior a `hora_limite_tardanza` |
| vigente_desde | date | no nulo |
| creado_en | timestamptz | no nulo |

Restricciones:
- `UNIQUE (turno, vigente_desde)`.
- `CHECK (hora_limite_tardanza > hora_ingreso)`.
- `CHECK (hora_cierre > hora_limite_tardanza)`.
- `CHECK (hora_ingreso + make_interval(mins => tolerancia_minutos) <= hora_limite_tardanza)`.
- `CHECK (vigente_desde >= fecha de hoy)` se valida en el servicio, no en la base, para poder
  probar con reloj inyectado.

Regla adicional: las ventanas de mañana y tarde no se superponen en la misma fecha (FR-011). Se
valida en el servicio contra la versión vigente del otro turno en la nueva `vigente_desde`.

### excepciones_calendario (FR-015 a FR-020)

| Columna | Tipo | Reglas |
|---|---|---|
| id | bigint, PK | generado |
| anio_escolar_id | bigint, FK anios_escolares | no nulo |
| tipo | text | `feriado`, `no_laborable` o `suspension` |
| fecha_inicio | date | no nulo |
| fecha_fin | date | no nulo; mayor o igual que `fecha_inicio` |
| motivo | text | no nulo, 3 a 200 caracteres |
| creado_por | bigint, FK usuarios | no nulo |
| creado_en | timestamptz | no nulo |

Índice: `(anio_escolar_id, fecha_inicio, fecha_fin)`.

### excepciones_grados (FR-016)

| Columna | Tipo | Reglas |
|---|---|---|
| excepcion_id | bigint, FK excepciones_calendario | PK compuesta |
| grado_id | bigint, FK grados | PK compuesta |

Sin filas para una excepción significa alcance a todo el colegio. Cada `grado_id` debe pertenecer
al año escolar de la excepción.

### umbrales_alerta (FR-021, FR-022)

| Columna | Tipo | Reglas |
|---|---|---|
| id | bigint, PK | generado |
| anio_escolar_id | bigint, FK anios_escolares | no nulo |
| tipo | text | `tardanzas` o `faltas` |
| valor | integer | mayor que 0; máximo 100 |
| vigente_desde | date | no nulo |
| creado_en | timestamptz | no nulo |

Restricciones: `UNIQUE (anio_escolar_id, tipo, vigente_desde)`.
Índice: `(anio_escolar_id, tipo, vigente_desde DESC)`.

### cursos (FR-030, FR-031)

| Columna | Tipo | Reglas |
|---|---|---|
| id | bigint, PK | generado |
| anio_escolar_id | bigint, FK anios_escolares | no nulo |
| nombre | text | único dentro del año; 2 a 80 caracteres |
| activo | boolean | por defecto verdadero |
| creado_en | timestamptz | no nulo |

Restricciones: `UNIQUE (anio_escolar_id, lower(nombre))` mediante índice único.

### cursos_grados (FR-030, FR-036)

| Columna | Tipo | Reglas |
|---|---|---|
| curso_id | bigint, FK cursos | PK compuesta |
| grado_id | bigint, FK grados | PK compuesta |

Regla: al crear o reactivar un curso debe tener al menos un grado. Al asignar un curso a una
sección, el grado de la sección debe estar en esta tabla.

### bloques_clase (FR-032 a FR-034, FR-041, FR-043)

| Columna | Tipo | Reglas |
|---|---|---|
| id | bigint, PK | generado |
| turno | text | `manana` o `tarde` |
| numero | integer | mayor que 0; máximo 20 |
| hora_inicio | time | no nulo |
| hora_fin | time | no nulo; posterior a `hora_inicio` |
| vigente_desde | date | no nulo |
| creado_en | timestamptz | no nulo |

Restricciones:
- `UNIQUE (turno, numero, vigente_desde)`.
- `CHECK (hora_fin > hora_inicio)`.
- `CHECK (numero > 0)`.

Reglas de servicio:
- `hora_inicio >= hora_ingreso` y `hora_fin <= hora_cierre` del horario vigente del turno en
  `vigente_desde`.
- Sin solapamiento con otros bloques del mismo turno vigentes en esa fecha.
- Solo el administrador crea y elimina; el director ajusta `hora_inicio` y `hora_fin` de bloques
  existentes (esto inserta una versión nueva con la misma `numero` y una `vigente_desde` nueva).

### horario_seccion (FR-035 a FR-039)

| Columna | Tipo | Reglas |
|---|---|---|
| id | bigint, PK | generado |
| seccion_id | bigint, FK secciones | no nulo |
| dia_semana | smallint | 1 = lunes a 5 = viernes |
| bloque_numero | integer | mayor que 0 |
| curso_id | bigint, FK cursos | nulo significa bloque sin clase |
| vigente_desde | date | no nulo |
| creado_en | timestamptz | no nulo |

Restricciones:
- `UNIQUE (seccion_id, dia_semana, bloque_numero, vigente_desde)`.
- `CHECK (dia_semana BETWEEN 1 AND 5)`.
- Índice: `(seccion_id, dia_semana, bloque_numero, vigente_desde DESC)`.

Reglas de servicio:
- El bloque referenciado debe existir en el turno vigente de la sección en esa fecha.
- El curso debe estar activo y vinculado al grado de la sección.
- Un día no laborable o suspendido para la sección no genera hora de clase para registro; esa
  regla se aplica en la consulta, no impide guardar el horario (FR-040, FR-039).

## Consultas derivadas

### Turno vigente de una sección en una fecha

```sql
SELECT turno
FROM turnos_seccion
WHERE seccion_id = $1 AND vigente_desde <= $2
ORDER BY vigente_desde DESC
LIMIT 1;
```

### Horario vigente de un turno en una fecha

```sql
SELECT *
FROM horarios_turno
WHERE turno = $1 AND vigente_desde <= $2
ORDER BY vigente_desde DESC
LIMIT 1;
```

### Bloques vigentes de un turno en una fecha

Una fila por `numero`, la versión más reciente anterior o igual a la fecha:

```sql
SELECT DISTINCT ON (numero) *
FROM bloques_clase
WHERE turno = $1 AND vigente_desde <= $2
ORDER BY numero, vigente_desde DESC;
```

### Curso de una sección en una fecha y bloque

```sql
SELECT curso_id
FROM horario_seccion
WHERE seccion_id = $1
  AND dia_semana = $2
  AND bloque_numero = $3
  AND vigente_desde <= $4
ORDER BY vigente_desde DESC
LIMIT 1;
```

### Excepciones que alcanzan una sección en una fecha

```sql
SELECT e.*
FROM excepciones_calendario e
WHERE e.anio_escolar_id = $1
  AND $2 BETWEEN e.fecha_inicio AND e.fecha_fin
  AND (
    NOT EXISTS (SELECT 1 FROM excepciones_grados eg WHERE eg.excepcion_id = e.id)
    OR EXISTS (
      SELECT 1 FROM excepciones_grados eg
      JOIN secciones s ON s.grado_id = eg.grado_id
      WHERE eg.excepcion_id = e.id AND s.id = $3
    )
  );
```

## Compatibilidad con la feature 001

- `academico.Seccion.Turno` se calcula con la consulta de turno vigente; el campo sigue existiendo.
- `BuscarSeccion(anioID, grado, seccion)` no cambia su firma ni su comportamiento.
- `ListarSecciones` y `ListarSeccionesVigentes` devuelven el turno vigente a la fecha actual.
- La importación de alumnos no requiere cambios.

## Migraciones

| Archivo | Contenido |
|---|---|
| `0009_limitar_un_anio_activo.sql` | Agrega `cerrado_en` e índice único parcial de año activo. |
| `0010_crear_tabla_turnos_seccion.sql` | Crea y llena `turnos_seccion`; elimina `secciones.turno`. |
| `0011_crear_tabla_horarios_turno.sql` | Crea `horarios_turno`. |
| `0012_crear_tablas_calendario.sql` | Crea `excepciones_calendario` y `excepciones_grados`. |
| `0013_crear_tabla_umbrales_alerta.sql` | Crea `umbrales_alerta`. |
| `0014_crear_tablas_cursos.sql` | Crea `cursos` y `cursos_grados`. |
| `0015_crear_tabla_bloques_clase.sql` | Crea `bloques_clase`. |
| `0016_crear_tabla_horario_seccion.sql` | Crea `horario_seccion`. |
| `0017_ampliar_entidades_auditoria.sql` | Amplía el `CHECK` de `auditoria.entidad`. |

Todas las migraciones incluyen su bloque `-- +goose Down` y una prueba de aplicación en
`pruebas/integracion`.
