# Feature Specification: Alertas internas a docentes por tardanzas e inasistencias

**Feature Branch**: `004-alertas-asistencia-docente` (no creada; el hook de git no está configurado)

**Created**: 2026-11-05

**Status**: Draft

**Input**: User description: "Alertas internas a docentes por tardanzas e inasistencias de sus alumnos. Cuando un alumno llega tarde o falta, o acumula demasiados casos, el docente o tutor de su sección debe enterarse dentro del sistema. No se envían notificaciones externas ni a apoderados."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Recibir alerta de tardanza en portería (Priority: P1)

El portero recibe, dentro del sistema, una alerta cuando un ingreso por QR que él registra corresponde a una tardanza. Ningún docente recibe alerta diaria por tardanza ni por inasistencia: las faltas del día quedan en los reportes y cuentan para las reiteraciones. Las alertas aparecen en un indicador visible con la cantidad de alertas nuevas y en el panel de alertas del portero.

**Why this priority**: Es la necesidad principal: que el docente se entere del hecho del día sin tener que revisar la lista alumno por alumno.

**Independent Test**: Un portero registra el ingreso por QR de un alumno de "3.° A" con tardanza. Se verifica que ese portero recibe una alerta nueva con alumno, sección, fecha y tipo, y que el docente de "3.° A" no la recibe. Luego, un alumno de "3.° A" queda registrado con falta al cierre, y se verifica que ningún docente recibe una alerta diaria por esa falta.

**Acceptance Scenarios**:

1. **Given** un portero que registra el ingreso por QR de un alumno de "3.° A", **When** el ingreso queda registrado como tardanza, **Then** ese portero recibe una alerta nueva de tipo "tardanza" con el alumno, la fecha y la hora del ingreso, y el docente de "3.° A" no recibe esa alerta.
2. **Given** un alumno de "3.° A" registrado con falta injustificada al cierre del registro, **When** el cierre ocurre, **Then** ningún docente recibe alerta por esa falta, y la falta aparece en el reporte diario.
3. **Given** un alumno registrado con falta justificada, **When** el día cierra, **Then** no se genera alerta diaria de inasistencia.
4. **Given** un docente asignado a "4.° B", **When** un alumno de "3.° A" llega tarde, **Then** el docente de "4.° B" no recibe ninguna alerta de ese alumno.

---

### User Story 2 - Recibir alerta por reiteración al superar un umbral (Priority: P1)

Cuando un alumno supera un umbral configurable de reiteración, el docente o tutor de su sección recibe una alerta de reiteración. Los umbrales de ejemplo son tardanzas acumuladas en el mes y faltas seguidas.

**Why this priority**: Permite detectar patrones que las alertas diarias no muestran por sí solas, y es el motivo por el que el colegio quiere el seguimiento.

**Independent Test**: El administrador configura un umbral de 3 tardanzas por mes. Un alumno acumula tres tardanzas en el mes. Se verifica que se genera exactamente una alerta de reiteración en el día en que se cruza el umbral, y no otra por cada tardanza adicional del mismo mes.

**Acceptance Scenarios**:

1. **Given** un umbral de 3 tardanzas por mes, **When** un alumno acumula su tercera tardanza del mes, **Then** el docente de la sección recibe una alerta de reiteración por tardanzas.
2. **Given** la alerta de reiteración ya generada en el mes, **When** el alumno tiene una cuarta tardanza en el mismo mes, **Then** no se genera una alerta duplicada.
3. **Given** un umbral de 2 faltas seguidas, **When** un alumno acumula dos faltas injustificadas consecutivas, **Then** el docente recibe una alerta de reiteración por faltas seguidas.
4. **Given** un alumno con dos faltas, una de ellas justificada, **When** se evalúa el umbral de faltas seguidas, **Then** la falta justificada no cuenta y no se genera alerta de reiteración.

---

### User Story 3 - Panel de alertas con filtros y estado de lectura (Priority: P2)

El docente o tutor abre un panel con las alertas de sus secciones. Puede filtrar por fecha, sección y tipo, ver si la alerta es nueva o vista, y marcar una o varias alertas como vistas. El portero ve en su panel solo las alertas de tardanza de los ingresos que él registró. El administrador ve las alertas de todas las secciones.

**Why this priority**: Sin panel las alertas no se pueden gestionar cuando hay muchas acumuladas.

**Independent Test**: Un docente tiene 20 alertas de dos secciones y tipos mixtos. Se aplica un filtro por sección y tipo, y se marcan como vistas las alertas filtradas. Se verifica que el contador de nuevas baja en la cantidad correcta.

**Acceptance Scenarios**:

1. **Given** un docente con alertas de dos secciones, **When** filtra por una sección y un tipo, **Then** el panel muestra solo las alertas que cumplen ambos filtros.
2. **Given** un docente con alertas nuevas, **When** marca una alerta como vista, **Then** la alerta cambia a estado "vista" y el contador de nuevas disminuye en uno.
3. **Given** un listado filtrado con varias alertas nuevas, **When** el docente marca todas como vistas, **Then** solo las alertas del listado cambian de estado.
4. **Given** un administrador, **When** abre el panel, **Then** ve las alertas de todas las secciones con los mismos filtros.

---

### User Story 4 - Configurar umbrales de alerta (Priority: P3)

El administrador define los umbrales de reiteración: número de tardanzas por mes y número de faltas seguidas. Cada cambio de umbral tiene una fecha de vigencia y no altera las alertas ya generadas.

**Why this priority**: Los valores iniciales pueden fijarse con el administrador en la configuración académica; el ajuste fino es posterior.

**Independent Test**: El administrador cambia el umbral de tardanzas de 3 a 4 con vigencia desde el 1 de septiembre. Se verifica que las alertas generadas antes del 1 de septiembre no cambian y que el umbral nuevo aplica desde esa fecha.

**Acceptance Scenarios**:

1. **Given** un umbral de 3 tardanzas por mes, **When** el administrador lo cambia a 4 con vigencia futura, **Then** el umbral anterior sigue aplicando hasta la fecha de vigencia.
2. **Given** un umbral con valor cero o negativo, **When** el administrador intenta guardarlo, **Then** el sistema lo rechaza.
3. **Given** un cambio de umbral, **When** se consulta el historial, **Then** se muestran el valor anterior, el nuevo y la fecha de vigencia.

---

### Edge Cases

- ¿Qué ocurre si un alumno supera el umbral justo el último día del mes? La alerta de reiteración se genera en ese día y pertenece al mes en que se cruzó el umbral, aunque el sistema la muestre al docente al día siguiente.
- ¿Qué ocurre si una falta se justifica después de haberse generado la alerta diaria o la de reiteración? La alerta diaria conserva su registro y se anota como "situación cambiada". Si al justificar la falta el conteo cae por debajo del umbral, la alerta de reiteración se anota como "situación cambiada"; no se elimina.
- ¿Qué ocurre si la sección no tiene docente o tutor asignado en la fecha del evento? Las alertas de esa sección se envían al administrador, marcadas como "sin responsable asignado", para que pueda asignar un responsable.
- ¿Qué ocurre si el docente asignado cambia después de generarse la alerta? La alerta permanece con el docente que estaba asignado en la fecha del evento. El nuevo docente recibe las alertas generadas desde su asignación.
- ¿Qué ocurre si un docente tiene muchas alertas acumuladas tras varios días sin ingresar? Las alertas se conservan sin límite durante el año escolar. El panel muestra las más recientes primero, agrupa por alumno y día, y permite marcar en bloque las del filtro activo.
- ¿Qué ocurre si un mismo evento genera más de una situación (por ejemplo, tardanza y reiteración el mismo día)? Cada situación genera su propia alerta; la reiteración no reemplaza a la alerta diaria.
- ¿Qué ocurre si el alumno se traslada o se retira? Sus alertas previas se conservan. No se generan alertas nuevas después de su fecha de salida.
- ¿Qué ocurre si el registro se corrige después del plazo de edición por un administrador? La corrección actualiza las alertas afectadas como "situación cambiada" y no genera alertas duplicadas.
- ¿Qué ocurre si un docente de reemplazo tiene asignación solo por un día? Recibe las alertas de ese día, y pierde el acceso a ellas cuando termina su asignación, salvo las que ya marcó como vistas.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: El sistema MUST generar una alerta de tipo "tardanza" cuando el ingreso por QR de la puerta principal registra una tardanza. La alerta MUST llegar a la cuenta del portero que registró ese ingreso.
- **FR-002**: El sistema MUST NOT generar alerta diaria de inasistencia a ningún usuario, incluido el administrador. Las faltas injustificadas solo aparecen en los reportes y cuentan para la reiteración por faltas seguidas (FR-005).
- **FR-003**: El sistema MUST NOT generar alerta diaria de inasistencia para una falta justificada.
- **FR-004**: El sistema MUST generar una alerta de tipo "reiteración por tardanzas" cuando el número de tardanzas de un alumno en un mes calendario alcanza el umbral configurado. Solo se genera una alerta por alumno, mes y umbral.
- **FR-005**: El sistema MUST generar una alerta de tipo "reiteración por faltas seguidas" cuando un alumno acumula el número configurado de faltas injustificadas consecutivas. Las faltas justificadas no cuentan para este umbral ni cortan la racha.
- **FR-006**: El sistema MUST NOT generar alertas por diferencias entre el registro del docente y el ingreso por QR. El registro del docente es personal y no es asistencia oficial.
- **FR-007**: El sistema MUST NOT generar más de una alerta por la misma situación. Una situación se identifica por el tipo, el alumno, la fecha y la sesión (o, para reiteraciones, el mes y el umbral).
- **FR-008**: Cada alerta de reiteración MUST llegar a la cuenta de cada docente con una asignación vigente a la sección del alumno en la fecha del evento. Cada docente recibe sus alertas en su propia cuenta; no existen buzones compartidos por sección. Si hay varios docentes asignados a la sección, todos reciben su propia alerta. Las alertas de tardanza por QR se rigen por FR-001.
- **FR-009**: Si la sección no tiene docente o tutor asignado en la fecha del evento, la alerta MUST dirigirse al administrador, marcada como "sin responsable asignado".
- **FR-010**: El sistema MUST marcar cada alerta como "nueva" al generarse y como "vista" cuando el destinatario la marca como vista. Registrar el momento y el usuario de la lectura.
- **FR-011**: El docente o tutor MUST poder marcar una o varias alertas como vistas, incluidas todas las del listado filtrado.
- **FR-012**: El panel de alertas MUST permitir filtrar por fecha (rango), sección, tipo y estado (nueva o vista).
- **FR-013**: El panel del docente MUST mostrar solo las alertas de los cursos y secciones a los que estuvo asignado en la fecha de cada evento.
- **FR-014**: El administrador MUST poder ver todas las alertas de todas las secciones con los mismos filtros, y MUST poder asignar un responsable a las alertas "sin responsable asignado".
- **FR-015**: Las alertas MUST NOT eliminarse. Si la situación que las originó cambia (por justificación de una falta o corrección de un registro), la alerta MUST marcarse como "situación cambiada" y conservar su historial.
- **FR-016**: Los umbrales de reiteración (tardanzas por mes y faltas seguidas) MUST ser configurables por el administrador, como números enteros mayores que cero, con fecha de vigencia.
- **FR-017**: Un cambio de umbral MUST aplicar desde su fecha de vigencia y MUST NOT modificar las alertas ya generadas.
- **FR-018**: Un evento que ocurre el último día de un mes MUST contarse en ese mes, aunque la alerta se muestre al día siguiente.
- **FR-019**: Las alertas MUST generarse solo para alumnos activos en la fecha del evento. Después de su retiro o traslado, no se generan alertas nuevas.
- **FR-020**: Las alertas MUST NOT enviarse por correo, SMS, WhatsApp ni ningún canal externo, ni a apoderados.
- **FR-021**: El sistema MUST mostrar en la interfaz del docente un indicador con la cantidad de alertas nuevas de sus secciones.
- **FR-022**: Las alertas de reiteración por faltas seguidas MUST generarse al cierre del registro de la sesión, no antes, para no alertar mientras el docente aún puede registrar.
- **FR-023**: Un registro corregido por un administrador después del plazo MUST actualizar las alertas afectadas sin crear alertas duplicadas.

### Key Entities *(include if feature involves data)*

- **Alerta**: Aviso generado por una situación de asistencia. Atributos: tipo (tardanza, reiteración por tardanzas, reiteración por faltas seguidas), alumno, sección en la fecha del evento, fecha, sesión (si aplica), destinatario (docente, tutor, portero o administrador), estado (nueva o vista), momento y usuario de lectura, y anotación de "situación cambiada" cuando corresponda.
- **Umbral de alerta**: Valor configurado para una reiteración (tardanzas por mes o faltas seguidas), con fecha de vigencia y valor anterior para su historial.
- **Asignación docente-curso-sección**: Vínculo entre un docente, un curso y una sección, con fechas de inicio y fin (definida en la feature 003). Determina la cuenta que recibe cada alerta.
- **Registro de asistencia**: Estado de un alumno en una fecha y sesión, tomado del docente o del QR (definidos en las features 001 y 003). Es la fuente de los eventos que generan alertas.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: El 100% de las tardanzas oficiales por QR genera una alerta al portero que registró el ingreso, y ninguna falta injustificada oficial genera alerta diaria a un docente, verificado en pruebas de aceptación.
- **SC-002**: Las alertas de tardanza aparecen para el portero en menos de 1 minuto desde el registro. No existen alertas diarias de inasistencia.
- **SC-003**: En el conjunto de pruebas, ninguna situación genera más de una alerta.
- **SC-004**: El 100% de los casos en que un alumno cruza un umbral genera la alerta de reiteración correspondiente, incluidos los que ocurren el último día del mes.
- **SC-005**: Un docente nunca ve alertas de secciones a las que no estuvo asignado en la fecha del evento, verificado en pruebas de acceso.
- **SC-006**: Un docente puede filtrar un listado de hasta 100 alertas y marcar todas como vistas en menos de 30 segundos.
- **SC-007**: El 100% de las alertas de una sección sin responsable llega al administrador, verificado en pruebas.

## Assumptions

- Las alertas se generan a partir de la asistencia oficial, que es el ingreso por QR de la puerta principal (feature 003, FR-023). Las tardanzas que solo registra el docente no generan alerta de tardanza en esta versión.
- El portero que recibe la alerta de tardanza es el que registró el ingreso por QR. Si el colegio quiere que todos los porteros de turno la vean, debe confirmarse.
- Las inasistencias diarias no generan alerta a ningún docente. Quedan en los reportes de la feature 006 y cuentan para la reiteración por faltas seguidas.
- "Tutor" es un docente asignado a una sección con ese rol; no existe un perfil separado. Si el colegio tiene un tutor distinto al docente de sección, deberá registrarse como asignación con ese rol (ver 003).
- Los umbrales de reiteración iniciales los define el administrador. La feature 002 ya aporta el umbral mensual de tardanzas; esta feature agrega el tipo `faltas_seguidas` a la misma tabla y pantalla de umbrales.
- La justificación de faltas pertenece a la feature 005. Mientras no exista, el cálculo de faltas seguidas considera injustificadas todas las ausencias oficiales; la interfaz de integración queda preparada para excluir justificadas cuando la 005 esté disponible.
- El Director no recibe alertas en esta versión. Solo consulta según su rol general; si se requiere que el director vea alertas, deberá agregarse en una enmienda.
- Las alertas son internas al sistema, sin notificaciones push ni correo; el docente las ve al ingresar al panel o en el indicador.
- La retención de alertas dura el año escolar activo y se conserva en el historial del año cerrado.
- La falta justificada se registra en la feature 003 y no requiere adjuntar documentos.
- Calificaciones, comportamiento, asistencia de docentes y comunicaciones a apoderados están fuera de alcance.
