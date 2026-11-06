# Modelo de datos: Alertas internas

**Fecha**: 2026-11-06 | **Plan**: [plan.md](plan.md) | **Investigación**: [research.md](research.md)

Convenciones: tablas en plural, columnas en snake_case, sin tildes. Fechas de evento en `date`,
marcas en `timestamptz` UTC.

## Cambios a tablas existentes

### marcaciones (0023)

| Columna nueva | Tipo | Reglas |
|---|---|---|
| registrada_por | bigint, FK usuarios | nulo; portero que registró el ingreso |

Índice: `(fecha, registrada_por)`.

### umbrales_alerta (0026)

Se amplía el `CHECK` de `tipo` para admitir `faltas_seguidas`. Los tipos quedan:

- `tardanzas`: umbral mensual de tardanzas (ya existente).
- `faltas`: umbral mensual de faltas (ya existente, no se usa en esta feature).
- `faltas_seguidas`: días laborables consecutivos sin marcación oficial.

## Tablas nuevas

### alertas (0024)

| Columna | Tipo | Reglas |
|---|---|---|
| id | bigint, PK | generado |
| tipo | text | `tardanza`, `reiteracion_tardanzas`, `reiteracion_faltas` |
| alumno_id | bigint, FK alumnos | no nulo |
| seccion_id | bigint, FK secciones | no nulo |
| curso_id | bigint, FK cursos | nulo salvo que el evento tenga curso |
| bloque_numero | integer | nulo salvo evento de clase |
| fecha_evento | date | no nulo |
| mes | date | primer día del mes; solo para `reiteracion_tardanzas` |
| marcacion_id | bigint, FK marcaciones | solo para `tardanza` |
| destinatario_id | bigint, FK usuarios | docente, portero o administrador |
| estado | text | `nueva` o `vista` |
| leida_en | timestamptz | nulo mientras esté nueva |
| leida_por | bigint, FK usuarios | nulo mientras esté nueva |
| sin_responsable | boolean | verdadero si se envió al administrador por falta de docente |
| situacion_cambiada | boolean | por defecto falso |
| motivo_cambio | text | nulo o hasta 300 caracteres |
| creada_en | timestamptz | no nulo |

Índices:
- `(destinatario_id, estado, creada_en DESC)` para el panel y el indicador.
- `(seccion_id, fecha_evento)` para filtros.
- `(alumno_id, fecha_evento)` para historial.

Índices únicos parciales (FR-007):
- `UNIQUE (marcacion_id, destinatario_id) WHERE tipo = 'tardanza'`.
- `UNIQUE (alumno_id, mes, destinatario_id) WHERE tipo = 'reiteracion_tardanzas'`.
- `UNIQUE (alumno_id, fecha_evento, destinatario_id) WHERE tipo = 'reiteracion_faltas'`.

Reglas:
- No se eliminan ni actualizan datos de la situación; solo cambia `estado` y la marca de lectura.
- `situacion_cambiada` es de una sola vía.
- La alerta conserva la sección y el destinatario vigentes en la fecha del evento.

### cierres_alerta (0025)

| Columna | Tipo | Reglas |
|---|---|---|
| id | bigint, PK | generado |
| fecha | date | no nulo |
| turno | text | `manana` o `tarde` |
| ejecutado_en | timestamptz | no nulo |

Restricción: `UNIQUE (fecha, turno)`. Evita repetir el cierre diario y sirve de bitácora.

## Auditoría (0027)

Se amplía el `CHECK` de `auditoria.entidad` con `alerta` y `cierre_alerta`.

Acciones:
- `crear_alerta`
- `marcar_vista`
- `marcar_vistas`
- `asignar_responsable`
- `marcar_situacion_cambiada`
- `ejecutar_cierre`

## Consultas derivadas

### Marcaciones sin alerta de tardanza

```sql
SELECT m.id, m.alumno_id, m.fecha, m.hora, m.estado, m.registrada_por
FROM marcaciones m
WHERE m.fecha BETWEEN $1 AND $2
  AND m.estado IN ('tardanza', 'tardanza_grave')
  AND NOT EXISTS (
    SELECT 1 FROM alertas a
    WHERE a.tipo = 'tardanza' AND a.marcacion_id = m.id
  );
```

### Docentes asignados a una sección en una fecha

```sql
SELECT DISTINCT docente_id
FROM asignaciones_docente
WHERE seccion_id = $1
  AND fecha_inicio <= $2
  AND (fecha_fin IS NULL OR fecha_fin >= $2);
```

### Conteo de tardanzas del mes

```sql
SELECT count(*)
FROM marcaciones m
JOIN historial_secciones h
  ON h.alumno_id = m.alumno_id
 AND h.fecha_inicio <= m.fecha
 AND (h.fecha_fin IS NULL OR h.fecha_fin > m.fecha)
WHERE m.alumno_id = $1
  AND m.fecha BETWEEN $2 AND $3
  AND m.estado IN ('tardanza', 'tardanza_grave')
  AND h.seccion_id = $4;
```

### Días sin marcación para la racha

Se calculan en el servicio con el calendario del módulo académico: se recorre hacia atrás desde
la fecha de cierre, se omiten fines de semana y excepciones que alcancen la sección, y se corta al
encontrar una marcación o una falta justificada.

### Alertas del panel

```sql
SELECT *
FROM alertas
WHERE destinatario_id = $1
  AND ($2::date IS NULL OR fecha_evento >= $2)
  AND ($3::date IS NULL OR fecha_evento <= $3)
  AND ($4::bigint = 0 OR seccion_id = $4)
  AND ($5::text = '' OR tipo = $5)
  AND ($6::text = '' OR estado = $6)
ORDER BY creada_en DESC
LIMIT 200;
```

El administrador usa la misma consulta sin filtro por destinatario.

## Migraciones

| Archivo | Contenido |
|---|---|
| `0023_agregar_portero_a_marcaciones.sql` | Columna `registrada_por` e índice. |
| `0024_crear_tabla_alertas.sql` | Tabla e índices únicos de deduplicación. |
| `0025_crear_tabla_cierres_alerta.sql` | Tabla de cierres diarios. |
| `0026_ampliar_umbrales_faltas_seguidas.sql` | Nuevo tipo de umbral. |
| `0027_ampliar_auditoria_alertas.sql` | Entidades de auditoría. |

Todas con `-- +goose Down` y prueba de subida y reversión.
