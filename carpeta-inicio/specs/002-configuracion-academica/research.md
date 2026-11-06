# Investigación técnica: Configuración académica del colegio

**Fecha**: 2026-11-06 | **Plan**: [plan.md](plan.md)

La especificación no tiene marcadores `NEEDS CLARIFICATION`. Esta investigación cubre las
decisiones técnicas que la constitución y el stack aprobado dejaban abiertas.

## 1. Modelo de vigencia por fecha

- **Decisión**: cada tabla configurable tiene una columna `vigente_desde DATE` y es **solo de
  agregado**: un cambio inserta una fila nueva; nunca se actualiza el valor de una fila anterior.
  La versión vigente en una fecha se obtiene con
  `WHERE vigente_desde <= $fecha ORDER BY vigente_desde DESC LIMIT 1`.
- **Justificación**: cumple FR-012, FR-013, FR-014, FR-023 y FR-038 sin tablas espejo ni
  librerías bitemporales. Insertar un cambio con una fecha anterior a otro ya programado funciona
  sin reescribir nada: la consulta siempre elige la versión más reciente que no supere la fecha.
  Una fecha nueva reemplaza a la anterior de esa misma fecha mediante un índice único
  `(clave, vigente_desde)`.
- **Alternativas consideradas**:
  - Columnas `vigente_desde` / `vigente_hasta` con cierre de la fila anterior: obliga a
    reescribir filas y complica los cambios programados fuera de orden.
  - Tabla de auditoría como fuente histórica: mezcla responsabilidades y obliga a interpretar
    JSON para reconstruir la configuración.
- **Regla de fecha**: `vigente_desde` no puede ser anterior a hoy en America/Lima. La validación
  se hace en el servicio con el reloj inyectado (constitución, principio V).

## 2. Turno vigente de una sección

- **Decisión**: nueva tabla `turnos_seccion (seccion_id, turno, vigente_desde)`, solo de
  agregado. Se elimina `secciones.turno`; las consultas derivan el turno con la versión vigente.
- **Justificación**: FR-028 exige saber el turno de una sección en cualquier fecha y FR-005 exige
  un único turno en un momento dado. Una sola fuente de verdad evita que la columna y el historial
  se desincronicen.
- **Alternativas consideradas**:
  - Conservar `secciones.turno` como valor actual y agregar historial: dos fuentes de verdad.
  - Reconstruir el turno desde la auditoría: lento y frágil.
- **Migración**: se crea la tabla y se inserta una fila por cada sección existente con
  `vigente_desde = fecha_inicio` de su año escolar. Luego se elimina la columna. El `Down` vuelve
  a crearla con el turno vigente en la fecha de reversión.

## 3. Turnos fijos

- **Decisión**: no hay tabla `turnos`. Los valores `manana` y `tarde` son constantes del dominio,
  con `CHECK` en las tablas que los usan y constantes Go exportadas (`TurnoManana`, `TurnoTarde`).
- **Justificación**: la spec fija exactamente dos turnos que no se crean ni eliminan (FR-006).
  Una tabla de dos filas sin operaciones de alta o baja es complejidad sin consumidor real
  (constitución, principio I).

## 4. Horarios de turno

- **Decisión**: tabla `horarios_turno` con `turno`, `hora_ingreso`, `tolerancia_minutos`,
  `hora_limite_tardanza`, `hora_cierre` y `vigente_desde`, única por `(turno, vigente_desde)`.
- **Validaciones** (funciones puras en `validaciones.go`):
  - `hora_cierre > hora_limite_tardanza` (FR-008).
  - `hora_ingreso + tolerancia <= hora_limite_tardanza` (FR-009).
  - `hora_ingreso < hora_limite_tardanza <= hora_cierre`, todas dentro del mismo día (FR-010).
  - Las ventanas `[hora_ingreso, hora_cierre]` de mañana y tarde no se superponen en la misma
    fecha de vigencia (FR-011). Como los horarios cambian con fechas distintas, la comprobación
    se hace contra la versión vigente del otro turno en la nueva `vigente_desde`.
- **Concurrencia**: la escritura se ejecuta en una transacción con
  `pg_advisory_xact_lock(hashtext('horarios_turno'))`. Con una sola instancia no es
  imprescindible, pero cuesta una línea y evita carreras si dos administradores guardan a la vez.
- **Alternativas consideradas**: restricción de exclusión con rangos `tstzrange`. No cubre la
  comparación entre turnos cuyas vigencias cambian en fechas distintas y vuelve las migraciones
  menos legibles.

## 5. Feriados, no laborables y suspensiones

- **Decisión**: `excepciones_calendario (anio_escolar_id, tipo, fecha_inicio, fecha_fin, motivo)`
  y `excepciones_grados (excepcion_id, grado_id)` para el alcance por grado. Sin filas en
  `excepciones_grados` significa alcance a todo el colegio.
- **Justificación**: FR-015 y FR-016 distinguen tres tipos y dos alcances. Una tabla puente
  normaliza el alcance por grado y permite consultar con un `EXISTS`.
- **Solapamiento** (FR-020): el servicio detecta cruces con otra excepción del mismo alcance y
  devuelve un aviso que el manejador muestra como confirmación; si el usuario confirma, se guarda
  la nueva excepción. No se reemplaza ni se borra la anterior porque el historial se conserva.
- **Fines de semana** (FR-019): se guardan y se muestran; no alteran el conteo porque el cálculo
  de días laborables (feature 003) ya excluye sábado y domingo.
- **Alcance por grado**: se valida que cada grado pertenezca al año escolar activo (FR-016 y caso
  límite de la spec).

## 6. Umbrales de alerta

- **Decisión**: `umbrales_alerta (anio_escolar_id, tipo, valor, vigente_desde)` con
  `tipo IN ('tardanzas', 'faltas')`, `valor > 0` y único por `(anio_escolar_id, tipo,
  vigente_desde)`.
- **Justificación**: FR-021 y FR-022. El envío de alertas pertenece a la feature 004; aquí solo se
  guarda el parámetro y su historia.

## 7. Cursos y vínculo con grados

- **Decisión**: `cursos (anio_escolar_id, nombre, activo)` único por año y
  `cursos_grados (curso_id, grado_id)`. Un curso debe tener al menos un grado. No se borra: se
  desactiva (`activo = false`), conservando su historial (FR-030, FR-031).
- **Justificación**: un curso puede dictarse en varios grados (caso límite de la spec). La tabla
  puente es la forma mínima de expresarlo.
- **Alternativas consideradas**: columna `grados` como arreglo en `cursos`. Dificulta las
  validaciones de pertenencia y los índices; se descarta.

## 8. Bloques de hora de clase

- **Decisión**: `bloques_clase (turno, numero, hora_inicio, hora_fin, vigente_desde)` único por
  `(turno, numero, vigente_desde)`.
- **Validaciones**:
  - `hora_inicio < hora_fin` y sin cruzar la medianoche (FR-033).
  - El bloque cae dentro de `[hora_ingreso, hora_cierre]` del horario vigente del turno en esa
    fecha.
  - Los bloques del mismo turno no se solapan en la misma vigencia (FR-034). Igual que en los
    horarios, se comparan las versiones vigentes en la nueva fecha.
- **Alternativas consideradas**: guardar solo la duración de cada bloque a partir de la hora de
  ingreso. No permite recreos entre bloques ni horas distintas por día; la spec exige horas de
  inicio y fin explícitas.
- **Recreos**: no se modelan. Simplemente queda un hueco entre dos bloques (FR-043).

## 9. Horario semanal de sección

- **Decisión**: `horario_seccion (seccion_id, dia_semana, bloque_numero, curso_id, vigente_desde)`
  único por `(seccion_id, dia_semana, bloque_numero, vigente_desde)`. `curso_id` nulo significa
  "bloque sin clase" y también es una versión explícita.
- **Justificación**: FR-035 a FR-039. Se referencia el **número** de bloque, no su `id`, porque
  los bloques se versionan y un cambio de horas no debe romper los horarios ya armados. El día se
  guarda como entero ISO (1 = lunes a 5 = viernes).
- **Vigencia por celda**: cada fila es una celda del horario con su fecha de vigencia. La consulta
  toma la última versión por `(seccion_id, dia_semana, bloque_numero)` anterior o igual a la
  fecha. Guardar la grilla crea filas nuevas solo para las celdas que cambian; las celdas sin
  cambio conservan su vigencia anterior.
- **Validación de curso**: el curso debe estar activo y vinculado al grado de la sección
  (FR-036). Se resuelve el grado con una consulta a `secciones` y `cursos_grados`.
- **Reemplazo** (FR-037): si la celda ya tiene un curso para esa fecha, el manejador pide
  confirmación antes de guardar.
- **Turno de la sección**: el bloque se interpreta con el turno vigente de la sección en la fecha
  consultada (FR-029). Si la sección cambia de turno, las mismas celdas pasan a los bloques del
  nuevo turno desde la fecha de vigencia; el horario anterior sigue visible en la consulta
  histórica.

## 10. Consulta histórica

- **Decisión**: `GET /configuracion/consulta?fecha=AAAA-MM-DD` devuelve, en una sola pantalla, el
  horario de cada turno, las excepciones del día, los umbrales vigentes, el turno de cada sección y
  los bloques vigentes.
- **Justificación**: FR-023 y FR-024. Para una fecha anterior al inicio del año consultado, la
  pantalla indica que no hay configuración vigente. La consulta elige el año escolar que cubre la
  fecha (`fecha_inicio <= fecha <= fecha_fin`), no solo el activo.
- **Rendimiento**: todas las consultas usan índices por `(clave, vigente_desde DESC)`; con la
  escala del colegio son decenas de filas por tabla.

## 11. Autorización del director

- **Decisión**: se mantiene el intermediario `RequerirRol`. Las rutas de solo lectura admiten
  `administrador` y `director`. Los POST de turnos y de horas de bloques admiten ambos roles. El
  resto de los POST exige `administrador`.
- **Justificación**: FR-025, FR-026 y FR-041. El director solo ajusta horas y tolerancias, y solo
  de elementos existentes: no puede crear ni eliminar bloques, ni tocar años, grados, secciones,
  feriados, umbrales, cursos ni horarios de sección.
- **Alternativas consideradas**: un permiso por campo dentro del mismo formulario. Se descarta
  porque el formulario de bloques permite crear y editar a la vez; separar rutas hace el control
  explícito y fácil de probar.

## 12. Auditoría

- **Decisión**: se reutiliza la tabla `auditoria` existente. La migración `0017` amplía el `CHECK`
  de `entidad` con `anio_escolar`, `grado`, `seccion`, `turno`, `calendario`, `umbral`, `curso`,
  `bloque_clase` y `horario_seccion`.
- **Justificación**: FR-027 y FR-042 piden usuario, fecha y hora, valor anterior, valor nuevo y
  vigencia. La tabla ya guarda JSON de valores y es inmutable por disparador.
- **Acciones**: `crear`, `cerrar`, `cambiar_turno`, `configurar_horario`, `registrar_excepcion`,
  `configurar_umbral`, `crear_curso`, `desactivar_curso`, `crear_bloque`, `ajustar_bloque`,
  `asignar_horario`.

## 13. Interfaz

- **Decisión**: panel `/configuracion` con tarjetas por sección; formularios con `input type=date`
  e `input type=time`; errores junto al campo y un resumen con `role="alert"`.
- **Justificación**: la spec pide validaciones con mensajes que indiquen el motivo (SC-003,
  SC-008) y uso sin ayuda técnica. HTMX se usa solo para la confirmación de reemplazos y para
  recargar la grilla de horario al cambiar de sección; el resto son formularios normales.
- **Responsividad**: la grilla de horario se muestra como tabla con desplazamiento horizontal en
  celular y como grilla completa desde tablet, siguiendo el patrón de la feature 001.

## 14. Cierre del año escolar

- **Decisión**: `POST /configuracion/anio/cierre` cambia `estado` a `cerrado` y guarda
  `cerrado_en`. El servicio rechaza cualquier escritura sobre un año cerrado.
- **Justificación**: FR-002 y FR-003. El cierre es una operación explícita del administrador, no
  automática, para que el colegio decida cuándo termina el periodo.
- **Un solo año activo**: índice único parcial
  `CREATE UNIQUE INDEX ... ON anios_escolares (estado) WHERE estado = 'activo'`. La base de datos
  impide dos activos incluso ante una carrera.
