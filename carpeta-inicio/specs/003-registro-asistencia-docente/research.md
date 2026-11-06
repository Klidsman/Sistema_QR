# Investigación técnica: Registro de asistencia de alumnos por el docente

**Fecha**: 2026-11-06 | **Plan**: [plan.md](plan.md)

La spec se ajustó antes de planificar: se difirió el modo sin conexión, se aclaró que las
diferencias se destacan como información y se agregaron las cuentas de docente, el parámetro del
plazo y la tabla mínima de marcaciones QR.

## 1. Dependencia del ingreso por QR

- **Decisión**: crear una tabla mínima `marcaciones` en esta feature y leerla. La pantalla de
  portería, el escaneo con cámara y la escritura de marcaciones quedan para una feature posterior.
- **Justificación**: FR-003 exige mostrar si el alumno ingresó por QR y a qué hora. Ninguna
  feature anterior creó esa tabla y no se puede consultar lo que no existe.
- **Alternativas consideradas**:
  - Omitir el contraste: incumple FR-003 y SC-002.
  - Implementar portería completa aquí: duplica una feature futura y amplía el alcance de forma
    desproporcionada.
- **Forma de prueba**: las pruebas de integración insertan marcaciones directamente para verificar
  la lista y las diferencias. Cuando exista portería, solo tendrá que escribir en la misma tabla.

## 2. Resolución de una hora de clase

- **Decisión**: una clase se identifica por `(fecha, curso_id, seccion_id, bloque_numero)`. El
  módulo `academico` expone `ResolverClase(ctx, seccionID, cursoID, bloque, fecha)`.
- **Reglas**:
  - La fecha no puede ser futura (comparación civil).
  - El día debe ser laborable (1 a 5) y no puede estar cubierto por una excepción de calendario
    que alcance a la sección (feriado, no laborable o suspensión).
  - El turno vigente de la sección en la fecha debe existir.
  - El bloque debe existir en el turno vigente y estar dentro del horario del turno.
  - La celda del horario semanal para ese día y bloque debe tener el curso solicitado, con la
    vigencia correspondiente.
- **Justificación**: centraliza en `academico` las reglas de horarios y evita que `asistencia`
  consulte sus tablas.
- **Alternativas**: replicar las consultas en `asistencia`. Se descarta por el principio II.

## 3. Asignaciones de docente

- **Decisión**: `asignaciones_docente` con `docente_id`, `curso_id`, `seccion_id`, `fecha_inicio`,
  `fecha_fin` (nula para titular vigente), `titular` y `creado_por`.
- **Reglas**:
  - El usuario debe existir, estar activo y tener rol `docente`.
  - El curso debe estar activo y vinculado al grado de la sección.
  - `fecha_fin` es nula o mayor o igual a `fecha_inicio`.
  - Se permiten asignaciones superpuestas para el mismo curso y sección, porque un reemplazo puede
    convivir con el titular el mismo día (caso límite de la spec).
  - Un reemplazo se modela como una asignación con `titular = false` y un rango corto.
- **Cierre**: una asignación con `fecha_fin` deja de dar acceso automáticamente al terminar el
  rango; no requiere proceso programado.
- **Auditoría**: `crear` y `cerrar` sobre `asignacion_docente`.

## 4. Cuentas de docente

- **Decisión**: usar la tabla `usuarios` existente con rol `docente`. Se agregan al módulo
  `autenticacion` las operaciones de alta, listado, desactivación y restablecimiento de contraseña.
- **Justificación**: las asignaciones necesitan usuarios con rol docente y hoy no hay pantalla de
  alta. Reutilizar `usuarios` evita una tabla de identidad paralela.
- **Reglas**: correo único y en minúsculas, contraseña con argon2id, nombre obligatorio. No se
  eliminan cuentas: se desactivan.
- **Alternativa**: crear docentes por consola. Se descarta porque el administrador debe poder
  operar el sistema sin intervención técnica.

## 5. Plazo de edición

- **Decisión**: tabla `parametros_asistencia` de una sola fila con `plazo_edicion_horas` (1 a 168,
  valor inicial 24) y auditoría de cada cambio.
- **Cálculo**: un docente puede modificar un registro mientras
  `ahora <= creado_en + plazo`. El reloj se inyecta en el servicio para pruebas deterministas.
- **Después del plazo**: solo el administrador modifica, y el motivo es obligatorio (FR-010).
- **Alternativa**: variable de entorno. Se descarta porque la spec pide que el administrador lo
  ajuste y el principio VI exige reglas configurables fuera del binario.

## 6. Diferencia entre QR y registro docente

- **Decisión**: la diferencia es de **presencia**, no de clasificación:
  - Hay QR (cualquier estado) y el docente marca `falta` o `falta_justificada` → diferencia.
  - No hay QR y el docente marca `presente` o `tardanza` → diferencia.
  - En los demás casos no se destaca.
- **Justificación**: el QR tiene estados de puntualidad y el docente usa otra escala; comparar
  directamente `puntual` con `presente` generaría diferencias falsas. La diferencia se muestra
  solo como aviso informativo y nunca modifica el QR (FR-023).
- **Alternativa**: comparar estados uno a uno. Se descarta por la razón anterior.

## 7. Guardado e idempotencia

- **Decisión**: restricción única `(alumno_id, fecha, curso_id, seccion_id, bloque_numero)` y
  `INSERT ... ON CONFLICT DO UPDATE` para el guardado.
- **Justificación**: FR-008 y SC-003. Reenviar el mismo formulario no duplica aunque el docente
  recargue o reintente.
- **Auditoría**: solo se registra un evento cuando el estado o la observación cambian realmente.
  Si el valor es idéntico, no se escribe auditoría.
- **Concurrencia**: el guardado de una clase se ejecuta en una transacción con
  `pg_advisory_xact_lock` por `(fecha, seccion, curso, bloque)` para que "marcar a todos" y una
  corrección simultánea no se pisen.

## 8. "Marcar a todos presentes"

- **Decisión**: dentro de la misma transacción, insertar `presente` solo para alumnos sin registro
  en esa clase, usando `ON CONFLICT DO NOTHING`.
- **Justificación**: FR-007 y el caso límite de correcciones previas. No modifica registros
  existentes ni genera auditoría por cada alumno si no hubo cambio real; se registra un solo evento
  con la cantidad creada.

## 9. Lista de alumnos

- **Decisión**: `alumnos` expone `ListarAlumnosDeSeccionEnFecha(ctx, seccionID, fecha)`.
- **Regla**: alumno `activo` y tramo de `historial_secciones` que cubre la fecha
  (`fecha_inicio <= fecha` y `fecha_fin` nula o mayor que la fecha).
- **Justificación**: FR-001, FR-019 y FR-020. El historial de secciones ya existe en la feature
  001 y resuelve los cambios de sección y las salidas.
- **Alternativa**: usar `alumnos.seccion_id` actual. Se descarta porque la consulta histórica
  mostraría alumnos en su sección nueva y perdería los registros anteriores.

## 10. Exportación

- **Decisión**: CSV con `encoding/csv`, nombre `asistencia_<curso>_<seccion>_<desde>_<hasta>.csv`.
  Columnas: fecha, curso, sección, bloque, identificador, alumno, estado, observación, hora QR y
  docente.
- **Justificación**: FR-026 pide un archivo descargable; CSV abre en Excel y no requiere
  dependencias nuevas.
- **Permisos**: el docente exporta solo sus cursos y secciones; el administrador y el director
  pueden exportar cualquier sección.

## 11. Fechas y horas

- **Decisión**: fechas de clase en `date`; horas y marcas de tiempo en `timestamptz` UTC. Las
  comparaciones de día usan `fechas.CompararDia` para evitar el desplazamiento UTC/Lima ya
  detectado en la feature 002.
- **Justificación**: una clase del 6 de noviembre en Lima no debe convertirse en 5 de noviembre por
  la zona horaria.

## 12. Sin conexión

- **Decisión**: diferido. No se implementa Service Worker, IndexedDB ni cola de sincronización.
- **Justificación**: la constitución lo declara funcionalidad posterior y exige su propia
  especificación. La spec 003 se ajustó para reflejar la decisión. La idempotencia del servidor se
  mantiene por la restricción única.

## 13. Límites entre módulos y dependencias

- `asistencia` depende de:
  - `academico.ResolutorClase` para validar la hora de clase y el día no laborable.
  - `alumnos.ListadorAlumnosSeccion` para la lista.
  - `autenticacion.BuscadorDocente` para validar rol y estado del docente.
- `asistencia` es dueña de `asignaciones_docente`, `registros_docente`, `marcaciones` y
  `parametros_asistencia`.
- `autenticacion` es dueña de `usuarios`.
- `academico` y `alumnos` no dependen de `asistencia`.
- Sin ciclos de importación.
