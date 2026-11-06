# Feature Specification: Justificación de faltas y tardanzas de alumnos

**Feature Branch**: `005-justificacion-faltas-tardanzas` (no creada; el hook de git no está configurado)

**Created**: 2026-11-05

**Status**: Draft

**Input**: User description: "Justificación de faltas y tardanzas de alumnos. Una falta o tardanza puede justificarse (enfermedad, trámite, etc.). La justificación cambia el estado del registro y evita que cuente como inasistencia injustificada. Docente o tutor registra la justificación; director o administrador aprueba o rechaza."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Registrar una justificación para una falta o tardanza (Priority: P1)

El docente o tutor de una sección registra una justificación para un alumno que tuvo falta o tardanza oficial, registrada por el ingreso por QR de la puerta principal. Indica el motivo, la fecha o el rango de fechas y, opcionalmente, adjunta un documento. La justificación queda pendiente de aprobación.

**Why this priority**: Es el punto de entrada del flujo. Sin registro de justificaciones no hay nada que aprobar ni historial que consultar.

**Independent Test**: Un docente registra una justificación por enfermedad para una falta del 12 de mayo con un certificado adjunto. Se verifica que queda en estado "pendiente", que el registro original sigue igual y que el docente no puede aprobarla.

**Acceptance Scenarios**:

1. **Given** un alumno de "3.° A" con falta oficial registrada el 12 de mayo, **When** el docente de "3.° A" registra una justificación por enfermedad para esa fecha, **Then** la justificación queda pendiente y el registro sigue mostrando "falta" hasta su aprobación.
2. **Given** la misma justificación, **When** el docente intenta aprobarla, **Then** el sistema no ofrece esa acción.
3. **Given** una justificación con un documento PDF de 2 MB, **When** la registra el docente, **Then** el documento queda adjunto y solo es visible para los roles autorizados.
4. **Given** un docente asignado a "4.° B", **When** intenta justificar una falta de "3.° A", **Then** el sistema no le permite hacerlo.

---

### User Story 2 - Aprobar o rechazar una justificación con comentario (Priority: P1)

El director o el administrador revisa las justificaciones pendientes. Puede aprobarlas o rechazarlas. Al aprobar, el registro cambia a estado "justificado" y deja de contar como inasistencia injustificada. Al rechazar, el registro no cambia. En ambos casos puede dejar un comentario, que es obligatorio al rechazar.

**Why this priority**: La aprobación es lo que produce el efecto real sobre la asistencia. Sin ella, la justificación no tiene valor.

**Independent Test**: Se aprueba una justificación de falta: el registro pasa a "justificado" y deja de contar para la alerta de reiteración. Se rechaza otra justificación: el registro sigue como "falta" y la justificación queda con el comentario de rechazo.

**Acceptance Scenarios**:

1. **Given** una justificación pendiente de una falta, **When** el director la aprueba con comentario, **Then** el registro cambia a "justificado" y el comentario y el aprobador quedan registrados.
2. **Given** una justificación pendiente, **When** el administrador la rechaza sin comentario, **Then** el sistema no permite completar el rechazo.
3. **Given** una justificación pendiente, **When** el administrador la rechaza con comentario, **Then** el registro original no cambia y la justificación queda en estado "rechazada".
4. **Given** una justificación rechazada, **When** el docente consulta el historial, **Then** ve el motivo del rechazo y puede presentar una nueva justificación para la misma fecha.
5. **Given** una justificación pendiente, **When** el docente que la registró intenta aprobarla, **Then** el sistema no lo permite.

---

### User Story 3 - Justificar un rango de fechas (Priority: P2)

El docente registra una justificación para un rango de fechas, por ejemplo una semana de reposo. El sistema aplica la justificación solo a los días que tienen falta o tardanza registrada, y reporta los días omitidos con su motivo.

**Why this priority**: Un reposo o trámite suele cubrir varios días. Justificar día por día no es práctico, pero el caso simple (un día) ya cubre la operación básica.

**Independent Test**: Un docente justifica del 10 al 14 de mayo para un alumno que tuvo faltas el 10, 11 y 13, y tuvo ingreso puntual el 12. El 14 es sábado. Se verifica que la justificación cubre solo el 10, 11 y 13, y que el 12 y el 14 aparecen como omitidos.

**Acceptance Scenarios**:

1. **Given** un alumno con faltas el 10, 11 y 13 de mayo, **When** se justifica el rango del 10 al 14 de mayo, **Then** la justificación cubre el 10, 11 y 13, y el sistema informa que el 12 (presente) y el 14 (no laborable) fueron omitidos.
2. **Given** un rango que incluye días ya justificados por una justificación aprobada, **When** el docente lo registra, **Then** esos días se omiten con el motivo "ya justificado" y el resto se procesa.
3. **Given** un rango que incluye días con una justificación pendiente, **When** el docente lo registra, **Then** esos días se omiten con el motivo "justificación pendiente" para no duplicar la solicitud.

---

### User Story 4 - Consultar el historial de justificaciones por alumno (Priority: P2)

El docente, el director o el administrador consulta todas las justificaciones de un alumno: fechas, motivo, estado, quién la registró, quién la resolvió, comentarios y documento adjunto cuando el usuario tiene permiso para verlo.

**Why this priority**: Permite revisar patrones y resolver consultas, y se necesita para auditar decisiones.

**Independent Test**: Un alumno tiene tres justificaciones: una aprobada, una rechazada y una pendiente. Se verifica que el historial muestra las tres, con su estado y la trazabilidad de quién hizo cada paso.

**Acceptance Scenarios**:

1. **Given** un alumno con tres justificaciones en distintos estados, **When** se consulta su historial, **Then** se muestran las tres, ordenadas de la más reciente a la más antigua, con su estado.
2. **Given** una justificación con documento adjunto, **When** un docente de otra sección consulta el historial, **Then** no ve el documento ni el motivo detallado.

---

### User Story 5 - Anular una justificación aprobada por error (Priority: P3)

Si una justificación se aprobó por error, el administrador puede anularla indicando un motivo. El registro vuelve a su estado original, y la anulación queda auditada.

**Why this priority**: Es necesario para corregir errores sin borrar datos, pero ocurre con poca frecuencia.

**Independent Test**: Se aprueba una justificación, luego el administrador la anula con motivo. Se verifica que el registro vuelve a "falta" y que la auditoría muestra la aprobación y la anulación.

**Acceptance Scenarios**:

1. **Given** una justificación aprobada, **When** el administrador la anula con motivo, **Then** el registro vuelve a su estado original y la anulación queda auditada.
2. **Given** una justificación aprobada, **When** el director intenta anularla, **Then** el sistema no le ofrece esa acción.

---

### Edge Cases

- ¿Qué ocurre si se justifica un rango que incluye días ya justificados? Esos días se omiten con el motivo "ya justificado". No se duplican ni se modifican.
- ¿Qué ocurre si la justificación se presenta fuera de plazo? El sistema la rechaza automáticamente al registrarla, indicando el plazo vigente. El plazo se cuenta desde la fecha del registro de falta o tardanza y es configurable en la configuración académica.
- ¿Qué ocurre si se aprueba una justificación que ya había generado una alerta? La alerta diaria o de reiteración conserva su historial y se marca como "situación cambiada". Si el conteo de faltas seguidas o de tardanzas del mes deja de superar el umbral, la alerta de reiteración también queda marcada como "situación cambiada".
- ¿Qué ocurre si la justificación se presenta cuando la alerta aún no existe? Solo se generan alertas por la situación vigente: una falta justificada y aprobada no genera alerta diaria de inasistencia.
- ¿Qué ocurre si se intenta justificar una fecha con asistencia "presente"? El sistema no lo permite, porque solo se justifican faltas y tardanzas registradas.
- ¿Qué ocurre si el registro de falta cambia después de presentarse la justificación pendiente (por ejemplo, el ingreso por QR se corrige a presente)? La justificación pendiente queda sin efecto y se marca como "registro ya no aplica"; no puede aprobarse.
- ¿Qué ocurre si el mismo docente es también director o administrador? No puede aprobar sus propias justificaciones. Otro director o administrador debe hacerlo.
- ¿Qué ocurre si el documento adjunto supera el tamaño permitido o tiene un formato no permitido? El sistema rechaza el adjunto e informa el límite y los formatos aceptados. La justificación puede enviarse sin adjunto, porque el adjunto es opcional.
- ¿Qué ocurre si se rechaza una justificación y el docente quiere presentar otra para la misma fecha? La nueva justificación es válida dentro del plazo; la rechazada conserva su historial.
- ¿Qué ocurre si el alumno cambia de sección o se retira después de presentarse la justificación? La justificación se resuelve normalmente, con la sección de la fecha del evento. Después de la salida del alumno, no se aceptan nuevas justificaciones.
- ¿Qué ocurre si dos personas intentan aprobar la misma justificación al mismo tiempo? Solo una resolución queda registrada; la segunda recibe el aviso de que la justificación ya fue resuelta.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: El docente o tutor asignado a la sección del alumno en la fecha del evento MUST poder registrar una justificación para una falta o tardanza registrada.
- **FR-002**: Cada justificación MUST indicar un motivo de texto y la fecha o el rango de fechas que cubre.
- **FR-003**: El sistema MUST permitir adjuntar un documento opcional a una justificación.
- **FR-004**: El sistema MUST aceptar únicamente los formatos PDF, JPG y PNG para los adjuntos, con un tamaño máximo de 5 MB por archivo.
- **FR-005**: El sistema MUST aceptar como máximo tres adjuntos por justificación.
- **FR-006**: Solo MUST justificarse fechas con falta o tardanza registrada oficialmente para el alumno. El sistema MUST NOT permitir justificar una fecha con asistencia presente.
- **FR-007**: Una justificación MUST presentarse dentro del plazo configurable, contado desde la fecha del registro de la falta o tardanza. El valor inicial del plazo es de 5 días calendario, definido en la configuración académica.
- **FR-008**: Una justificación presentada fuera de plazo MUST ser rechazada automáticamente al registrarla, indicando el plazo vigente.
- **FR-009**: Solo puede existir una justificación pendiente o aprobada por cada fecha y alumno. Una justificación rechazada o anulada no impide presentar una nueva dentro del plazo.
- **FR-010**: Cuando un rango incluye días ya justificados o con justificación pendiente, el sistema MUST omitir esos días, informar el motivo de la omisión y procesar los demás días.
- **FR-011**: Cuando un rango incluye días sin falta ni tardanza, o días no laborables, el sistema MUST omitir esos días e informarlo.
- **FR-012**: El director o el administrador MUST poder aprobar o rechazar una justificación pendiente.
- **FR-013**: Rechazar una justificación MUST exigir un comentario. Aprobar una justificación MAY incluir un comentario.
- **FR-014**: Al aprobar una justificación, el registro MUST cambiar a estado "justificado" manteniendo el tipo original (falta o tardanza) y sin borrar el dato original.
- **FR-015**: Al rechazar una justificación, el registro original MUST NOT cambiar.
- **FR-016**: Un usuario MUST NOT poder aprobar o rechazar una justificación que él mismo registró.
- **FR-017**: Una justificación aprobada MUST NOT contar como inasistencia injustificada para ninguna regla de alerta, reiteración ni reporte.
- **FR-018**: Toda creación, aprobación, rechazo, anulación y cambio de estado de una justificación MUST quedar auditado con usuario, fecha y hora, valor anterior y valor nuevo.
- **FR-019**: El administrador MUST poder anular una justificación aprobada indicando un motivo. La anulación MUST restaurar el estado original del registro.
- **FR-020**: El director MUST NOT poder anular justificaciones. Solo el administrador puede hacerlo.
- **FR-021**: Si un registro de falta o tardanza cambia mientras una justificación está pendiente, la justificación MUST marcarse como "registro ya no aplica" y MUST NOT poder aprobarse.
- **FR-022**: Cuando una justificación aprobada afecta una alerta existente, la alerta MUST marcarse como "situación cambiada" conforme a la feature 004. Las alertas NO MUST eliminarse.
- **FR-023**: El historial de justificaciones de un alumno MUST mostrar todas sus justificaciones, ordenadas de la más reciente a la más antigua, con fechas, motivo, estado, registrante, resolutor y comentarios.
- **FR-024**: Los documentos adjuntos MUST ser visibles solo para el administrador, el director y el docente o tutor que registró la justificación o es titular de la sección en la fecha del evento.
- **FR-025**: Los documentos adjuntos MAY contener datos de salud del alumno. El sistema MUST tratarlos como datos sensibles de menores, conforme a la constitución del proyecto.
- **FR-026**: El sistema MUST impedir que dos resoluciones distintas sobre una misma justificación queden registradas. La segunda debe recibir el aviso de que la justificación ya fue resuelta.
- **FR-027**: El sistema MUST NOT aceptar nuevas justificaciones para un alumno después de su fecha de retiro o traslado.

### Key Entities *(include if feature involves data)*

- **Justificación**: Solicitud de justificación para una o varias fechas de un alumno. Atributos: alumno, sección en la fecha del evento, motivo, fechas cubiertas, estado (pendiente, aprobada, rechazada, anulada, registro ya no aplica), registrante, fecha de presentación, resolutor, fecha de resolución, comentario, y adjuntos.
- **Adjunto de justificación**: Documento asociado a una justificación. Atributos: nombre, tipo, tamaño y fecha de carga. Su contenido es sensible.
- **Registro de asistencia**: Falta o tardanza oficial registrada por el ingreso por QR para una fecha. Su estado puede cambiar a "justificado" mientras la justificación esté aprobada, conservando su tipo original.
- **Historial de auditoría de justificación**: Entrada inmutable de cada acción sobre una justificación, con usuario, fecha y hora, acción, valor anterior y valor nuevo.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Un docente registra una justificación de un día en menos de 1 minuto.
- **SC-002**: Un docente justifica un rango de una semana en menos de 2 minutos, con el sistema informando de todos los días omitidos.
- **SC-003**: El 100% de las justificaciones aprobadas cambia el estado del registro y deja de contar como inasistencia injustificada, verificado en pruebas de aceptación.
- **SC-004**: El 100% de las justificaciones rechazadas deja el registro original sin cambios, verificado en pruebas.
- **SC-005**: El 100% de las justificaciones presentadas fuera de plazo es rechazado automáticamente.
- **SC-006**: El 100% de las acciones sobre justificaciones queda auditado con usuario, fecha y hora, verificado en pruebas.
- **SC-007**: Ningún usuario que no sea administrador, director o docente de la sección puede ver los documentos adjuntos, verificado en pruebas de acceso.

## Assumptions

- Las justificaciones se aplican solo a la asistencia oficial, que es el ingreso por QR de la puerta principal (feature 003, FR-023). El registro personal del docente no se justifica, y sus faltas no alteran la asistencia oficial.
- El estado "justificado" conserva el tipo original: "falta justificada" o "tardanza justificada". Esto extiende la lista de estados de la feature 003, que solo contemplaba "falta justificada".
- El plazo de presentación es de 5 días calendario desde la fecha del evento, configurable en la feature 002. Es una decisión por defecto; el colegio puede cambiarla.
- Un director o administrador que aprueba o rechaza no puede ser al mismo tiempo el registrante de esa justificación.
- El director puede aprobar o rechazar justificaciones, lo que amplía su rol de consulta de las features 002 y 003. Esta ampliación proviene de la descripción de la feature.
- Los adjuntos se conservan durante el año escolar y después se conservan según la política de retención de datos del colegio, pendiente de validación legal (ver la constitución).
- La justificación no requiere validar el documento; el sistema solo verifica formato y tamaño. El director o administrador decide si lo acepta.
- El apoderado no tiene acceso para presentar justificaciones: queda fuera de alcance, según el alcance del proyecto.
- Las alertas de la feature 004 se actualizan al aprobar, rechazar o anular una justificación, sin eliminar el historial.
- El reporte de asistencia debe distinguir las faltas justificadas de las injustificadas; el diseño de ese reporte pertenece a otra feature.
