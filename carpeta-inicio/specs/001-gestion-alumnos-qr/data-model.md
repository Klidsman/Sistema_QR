# Modelo de datos: Gestión de alumnos y generación de códigos QR

**Fecha**: 2026-11-05 | **Plan**: [plan.md](plan.md) | **Investigación**: [research.md](research.md)

Convenciones (constitución, principio IX): tablas en plural, columnas en snake_case, sin tildes
ni ñ en nombres de tablas y columnas. Fechas en `timestamptz` (UTC) y `date` (zona America/Lima).
Los textos que ve el usuario sí llevan tildes.

## Tablas de la feature

### usuarios (módulo autenticacion)

| Columna | Tipo | Reglas |
|---|---|---|
| id | bigint, PK | generado |
| correo | text | único, en minúsculas |
| contrasenia_hash | text | argon2id, no nulo |
| nombre_completo | text | no nulo |
| rol | text | `administrador` o `director` |
| activo | boolean | por defecto verdadero |
| creado_en | timestamptz | no nulo |

### anios_escolares, grados, secciones (módulo academico)

Estas tablas las administra la feature 002. En esta feature solo se leen, mediante la interfaz
`BuscadorSeccion`.

| Tabla | Columnas relevantes |
|---|---|
| anios_escolares | id, nombre, fecha_inicio, fecha_fin, estado (`activo` o `cerrado`) |
| grados | id, anio_escolar_id, nombre (único dentro del año) |
| secciones | id, grado_id, nombre (única dentro del grado), turno |

### alumnos (módulo alumnos)

| Columna | Tipo | Reglas |
|---|---|---|
| id | bigint, PK | generado |
| identificador | text | DNI o código; único global (FR-003); sin espacios al inicio o al final |
| nombres | text | no nulo |
| apellidos | text | no nulo |
| nombres_busqueda | text | calculado: minúsculas sin tildes; índice |
| apellidos_busqueda | text | calculado: minúsculas sin tildes; índice |
| seccion_id | bigint, FK secciones | no nulo mientras el estado sea `activo` |
| estado | text | `activo`, `desactivado`, `retirado` o `trasladado` |
| foto_clave | text | clave en almacenamiento de objetos; nulo si no tiene foto |
| apoderado_nombre | text | opcional |
| apoderado_telefono | text | opcional |
| fecha_salida | date | solo si el estado es `retirado` o `trasladado` |
| creado_en | timestamptz | no nulo |
| actualizado_en | timestamptz | no nulo |

Validaciones (FR-002, FR-029, FR-030):
- `identificador`, `nombres`, `apellidos` y `seccion_id` son obligatorios.
- Los nombres conservan tildes, ñ y apóstrofes; solo las columnas `*_busqueda` se normalizan.

### historial_secciones (módulo alumnos)

| Columna | Tipo | Reglas |
|---|---|---|
| id | bigint, PK | generado |
| alumno_id | bigint, FK alumnos | no nulo |
| seccion_id | bigint, FK secciones | no nulo |
| fecha_inicio | date | no nulo |
| fecha_fin | date | nulo mientras sea la sección actual |

Regla: para cada alumno, como máximo un registro con `fecha_fin` nulo. El cambio de sección cierra
el registro actual y abre uno nuevo (FR-024).

### credenciales_qr (módulo alumnos)

| Columna | Tipo | Reglas |
|---|---|---|
| id | bigint, PK | generado |
| alumno_id | bigint, FK alumnos | no nulo |
| token | text | 43 caracteres base64 URL; único global |
| estado | text | `activa` o `revocada` |
| motivo_revocacion | text | `perdida`, `robo`, `desactivacion`, `retiro`, `otro`; nulo si está activa |
| emitida_en | timestamptz | no nulo |
| revocada_en | timestamptz | nulo mientras esté activa |

Restricciones:
- Índice único parcial sobre `alumno_id` donde `estado = 'activa'`: un alumno tiene como máximo
  un QR activo (FR-008).
- Revocar pone `estado = 'revocada'` y `revocada_en`. Nunca se borra la fila (FR-011, FR-025).

### importaciones (módulo alumnos)

| Columna | Tipo | Reglas |
|---|---|---|
| id | bigint, PK | generado |
| usuario_id | bigint, FK usuarios | no nulo |
| archivo_nombre | text | nombre original |
| filas_leidas | integer | no nulo |
| filas_creadas | integer | no nulo |
| filas_rechazadas | integer | no nulo |
| creada_en | timestamptz | no nulo |

### importacion_filas (módulo alumnos)

| Columna | Tipo | Reglas |
|---|---|---|
| id | bigint, PK | generado |
| importacion_id | bigint, FK importaciones | no nulo |
| numero_fila | integer | número en el archivo original |
| identificador | text | tal como vino en el archivo |
| motivo | text | `campo_obligatorio_vacio`, `identificador_repetido_en_archivo`, `identificador_existente`, `seccion_inexistente`, `formato_invalido` |

### auditoria (módulo plataforma/auditoria)

| Columna | Tipo | Reglas |
|---|---|---|
| id | bigint, PK | generado |
| usuario_id | bigint, FK usuarios | no nulo |
| entidad | text | `alumno`, `credencial_qr`, `historial_seccion` |
| entidad_id | bigint | no nulo |
| accion | text | `crear`, `editar`, `desactivar`, `retirar`, `trasladar`, `cambiar_seccion`, `emitir_qr`, `revocar_qr`, `importar` |
| valor_anterior | jsonb | nulo al crear |
| valor_nuevo | jsonb | nulo al eliminar (no ocurre en esta feature) |
| motivo | text | opcional, obligatorio al revocar QR |
| creada_en | timestamptz | no nulo |

Regla: `UPDATE` y `DELETE` sobre `auditoria` están bloqueados por disparador (FR-026).

## Relaciones

- `alumnos.seccion_id` → `secciones.id`.
- `historial_secciones.alumno_id` → `alumnos.id`, con `seccion_id` → `secciones.id`.
- `credenciales_qr.alumno_id` → `alumnos.id`.
- `importacion_filas.importacion_id` → `importaciones.id`.
- `auditoria.usuario_id` → `usuarios.id`.

## Transiciones de estado

**Alumno**:
- `activo` → `desactivado` (desactivación; invalida QR activo, FR-032).
- `activo` → `retirado` o `trasladado` (indica fecha de salida; invalida QR activo, FR-023).
- `desactivado` → `activo` (reactivación; emite un QR nuevo, supuesto de la spec).
- `retirado` y `trasladado` no vuelven a `activo` en esta feature.

**Credencial QR**:
- Creación → `activa`.
- `activa` → `revocada` (motivo: pérdida, robo, desactivación, retiro u otro).
- Una sola `activa` por alumno; emitir un QR nuevo exige que la anterior esté revocada, en la
  misma transacción.

## Consultas principales

- Listado de alumnos por grado, sección, estado y texto de búsqueda (sobre columnas normalizadas).
- Alumnos incluidos en una selección de carnés, con su QR activo y su foto (si la tienen).
- Historial de secciones y de QR por alumno.
- Detalle de una importación con sus filas rechazadas.
