# Modelo de datos: Registro de asistencia de alumnos por el docente

**Fecha**: 2026-11-06 | **Plan**: [plan.md](plan.md) | **Investigación**: [research.md](research.md)

Convenciones: tablas en plural, columnas en snake_case, sin tildes ni ñ en identificadores.
Fechas de clase en `date`; horas y marcas en `timestamptz` (UTC). Textos visibles con tildes.

## Tablas nuevas

### parametros_asistencia (FR-009, FR-028)

| Columna | Tipo | Reglas |
|---|---|---|
| id | smallint, PK | siempre 1 |
| plazo_edicion_horas | integer | entre 1 y 168; valor inicial 24 |
| actualizado_por | bigint, FK usuarios | no nulo |
| actualizado_en | timestamptz | no nulo |

Una sola fila. Cada cambio se audita con valor anterior y nuevo.

### marcaciones (FR-003, FR-029)

| Columna | Tipo | Reglas |
|---|---|---|
| id | bigint, PK | generado |
| alumno_id | bigint, FK alumnos | no nulo |
| fecha | date | no nulo |
| hora | timestamptz | no nulo |
| estado | text | `puntual`, `tardanza` o `tardanza_grave` |
| creada_en | timestamptz | no nulo |

Restricción: `UNIQUE (alumno_id, fecha)`. Índice `(alumno_id, fecha)`.
Esta feature solo lee la tabla. La escritura pertenece a la feature de portería.

### asignaciones_docente (FR-013 a FR-016)

| Columna | Tipo | Reglas |
|---|---|---|
| id | bigint, PK | generado |
| docente_id | bigint, FK usuarios | rol `docente`, activo |
| curso_id | bigint, FK cursos | activo y vinculado al grado de la sección |
| seccion_id | bigint, FK secciones | no nulo |
| fecha_inicio | date | no nulo |
| fecha_fin | date | nulo o mayor o igual que `fecha_inicio` |
| titular | boolean | verdadero para titular, falso para reemplazo |
| creado_por | bigint, FK usuarios | no nulo |
| creado_en | timestamptz | no nulo |

Restricciones:
- `UNIQUE (docente_id, curso_id, seccion_id, fecha_inicio)`.
- `CHECK (fecha_fin IS NULL OR fecha_fin >= fecha_inicio)`.
- Índice `(docente_id, fecha_inicio, fecha_fin)` y `(seccion_id, curso_id, fecha_inicio, fecha_fin)`.

Regla: se permiten superposiciones, incluso entre titular y reemplazo, para el mismo curso y
sección. El acceso en una fecha se resuelve si `fecha_inicio <= fecha` y
(`fecha_fin IS NULL` o `fecha <= fecha_fin`).

### registros_docente (FR-002 a FR-011, FR-019, FR-020, FR-022)

| Columna | Tipo | Reglas |
|---|---|---|
| id | bigint, PK | generado |
| alumno_id | bigint, FK alumnos | no nulo |
| fecha | date | no nulo; no futura |
| curso_id | bigint, FK cursos | no nulo |
| seccion_id | bigint, FK secciones | no nulo |
| bloque_numero | integer | mayor que 0; bloque vigente del turno |
| estado | text | `presente`, `tardanza`, `falta` o `falta_justificada` |
| observacion | text | máximo 300 caracteres; opcional |
| docente_id | bigint, FK usuarios | usuario que registró o corrigió |
| creado_en | timestamptz | no nulo |
| actualizado_en | timestamptz | no nulo |

Restricción clave: `UNIQUE (alumno_id, fecha, curso_id, seccion_id, bloque_numero)` (FR-008).
Índices:
- `(fecha, seccion_id, curso_id, bloque_numero)` para la lista de clase.
- `(docente_id, fecha)` para la exportación del docente.
- `(alumno_id, fecha)` para consultas por alumno.

La `seccion_id` es la sección del alumno en la fecha del registro, no necesariamente la sección
actual. El cambio de sección no reescribe registros anteriores (FR-019).

### Auditoría

Se amplía el `CHECK` de `auditoria.entidad` con:

- `docente`
- `asignacion_docente`
- `registro_docente`
- `parametro_asistencia`
- `marcacion`

Acciones:
- `crear_docente`, `desactivar_docente`, `restablecer_clave`
- `crear_asignacion`, `cerrar_asignacion`
- `crear_registro`, `corregir_registro`, `marcar_todos_presentes`
- `configurar_plazo`
- `registrar_marcacion` (reservada para portería)

Cada evento guarda usuario, fecha y hora, valor anterior, valor nuevo y motivo cuando aplica
(FR-011).

## Tablas existentes que se consultan mediante interfaces

| Tabla | Módulo dueño | Uso en esta feature |
|---|---|---|
| usuarios | autenticacion | Docentes: identidad, rol, estado y contraseña |
| alumnos | alumnos | Alumnos activos de la sección en la fecha |
| historial_secciones | alumnos | Sección del alumno en la fecha |
| secciones, grados, anios_escolares | academico | Estructura de la clase |
| turnos_seccion | academico | Turno vigente de la sección |
| horarios_turno | academico | Ventana del turno |
| bloques_clase | academico | Bloques vigentes |
| horario_seccion | academico | Curso de la celda día/bloque |
| excepciones_calendario, excepciones_grados | academico | Días sin clase |
| cursos, cursos_grados | academico | Curso activo y vínculo con el grado |

Ninguna consulta SQL de `asistencia` cruza tablas de otro módulo: se usan las interfaces
`ResolutorClase` y `ListadorAlumnosSeccion`, implementadas por `academico` y `alumnos`.

## Consultas derivadas

### Asignaciones de un docente en una fecha

```sql
SELECT a.*
FROM asignaciones_docente a
WHERE a.docente_id = $1
  AND a.fecha_inicio <= $2
  AND (a.fecha_fin IS NULL OR a.fecha_fin >= $2)
ORDER BY a.fecha_inicio, a.id;
```

### Alumno activo en la sección de la fecha

```sql
SELECT al.id, al.identificador, al.nombres, al.apellidos
FROM alumnos al
JOIN historial_secciones h ON h.alumno_id = al.id
WHERE al.estado = 'activo'
  AND h.seccion_id = $1
  AND h.fecha_inicio <= $2
  AND (h.fecha_fin IS NULL OR h.fecha_fin > $2)
ORDER BY al.apellidos, al.nombres;
```

### Registros de una clase

```sql
SELECT r.*
FROM registros_docente r
WHERE r.fecha = $1
  AND r.seccion_id = $2
  AND r.curso_id = $3
  AND r.bloque_numero = $4;
```

### Marcaciones del día para una sección

```sql
SELECT m.alumno_id, m.hora, m.estado
FROM marcaciones m
JOIN historial_secciones h ON h.alumno_id = m.alumno_id
WHERE m.fecha = $1
  AND h.seccion_id = $2
  AND h.fecha_inicio <= $1
  AND (h.fecha_fin IS NULL OR h.fecha_fin > $1);
```

### Registro para exportación

```sql
SELECT r.fecha, c.nombre AS curso, s.nombre AS seccion, r.bloque_numero,
       al.identificador, al.apellidos, al.nombres, r.estado, r.observacion,
       m.hora AS hora_qr, u.nombre_completo AS docente
FROM registros_docente r
JOIN alumnos al ON al.id = r.alumno_id
JOIN cursos c ON c.id = r.curso_id
JOIN secciones s ON s.id = r.seccion_id
JOIN usuarios u ON u.id = r.docente_id
LEFT JOIN marcaciones m ON m.alumno_id = r.alumno_id AND m.fecha = r.fecha
WHERE r.fecha BETWEEN $1 AND $2
  AND r.docente_id = $3
ORDER BY r.fecha, c.nombre, r.bloque_numero, al.apellidos, al.nombres;
```

## Migraciones

| Archivo | Contenido |
|---|---|
| `0018_crear_parametros_asistencia.sql` | Tabla de parámetros con fila inicial de 24 horas. |
| `0019_crear_tabla_marcaciones.sql` | Tabla mínima de ingreso QR. |
| `0020_crear_tabla_asignaciones_docente.sql` | Asignaciones con vigencia y reemplazos. |
| `0021_crear_tabla_registros_docente.sql` | Registros por alumno, fecha, curso, sección y bloque. |
| `0022_ampliar_auditoria_asistencia.sql` | Nuevas entidades en el `CHECK` de auditoría. |

Todas con `-- +goose Down` y prueba de subida y reversión.
