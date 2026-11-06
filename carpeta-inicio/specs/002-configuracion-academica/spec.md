# Feature Specification: Configuración académica del colegio

**Feature Branch**: `002-configuracion-academica` (no creada; el hook de git no está configurado)

**Created**: 2026-11-05

**Status**: Draft

**Input**: User description: "Configuración académica del colegio: año escolar, grados, secciones, turnos, horarios de ingreso, tolerancias y feriados, con cambios vigentes desde una fecha y consulta de la configuración histórica."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Crear el año escolar con grados, secciones y turnos (Priority: P1)

El administrador crea el año escolar con fecha de inicio y fin, define sus grados, crea las secciones de cada grado y asigna cada sección al turno de mañana o de tarde. Con esto el colegio queda listo para que los alumnos se asignen a secciones y para que la clasificación de asistencia tenga un turno al que aplicar horarios.

**Why this priority**: Todo lo demás (alumnos, carnés, asistencia, reportes) depende de tener un año escolar con su estructura de grados, secciones y turnos.

**Independent Test**: Se crea un año escolar, se agregan dos grados con dos secciones cada uno, se asigna una sección a mañana y otra a tarde, y se verifica que la estructura aparece completa y que no permite fechas inválidas.

**Acceptance Scenarios**:

1. **Given** no existe año escolar activo, **When** el administrador crea el año 2027 con fecha de inicio y fin, **Then** el año queda creado como activo y puede agregar grados y secciones.
2. **Given** un año escolar con fecha de inicio posterior a la de fin, **When** el administrador intenta guardarlo, **Then** el sistema rechaza el cambio e indica que la fecha de fin debe ser posterior a la de inicio.
3. **Given** un año escolar activo, **When** el administrador crea una sección "3.° A" y la asigna al turno de mañana, **Then** la sección aparece bajo su grado con el turno indicado.
4. **Given** una sección asignada a un turno, **When** el administrador consulta los turnos, **Then** puede ver qué secciones pertenecen a mañana y cuáles a tarde.

---

### User Story 2 - Configurar horarios y tolerancias por turno con vigencia desde una fecha (Priority: P1)

El administrador (o el director, para horarios) configura por cada turno la hora de ingreso, la tolerancia para puntualidad, la hora límite de tardanza y la hora de cierre del registro. Cada cambio indica desde qué fecha aplica y rige desde ese día hacia adelante, sin modificar la configuración de días anteriores.

**Why this priority**: La clasificación de asistencia depende directamente de estos valores. Deben poder cambiarse sin intervención técnica.

**Independent Test**: Se configura el turno de mañana con ingreso 07:30 y tolerancia 10 minutos. Luego se cambia la tolerancia a 5 minutos con vigencia desde una fecha futura. Se verifica que la configuración anterior sigue vigente antes de esa fecha y que la nueva aplica desde ella.

**Acceptance Scenarios**:

1. **Given** el turno de mañana con ingreso 07:30, tolerancia 10 minutos, límite de tardanza 08:00 y cierre 08:30, **When** el administrador consulta el turno, **Then** ve esos cuatro valores y la fecha desde la que están vigentes.
2. **Given** el turno de tarde con cierre anterior o igual a la hora límite de tardanza, **When** el administrador intenta guardarlo, **Then** el sistema rechaza el cambio e indica que la hora de cierre debe ser posterior a la hora límite de tardanza.
3. **Given** un cambio de tolerancia de 10 a 5 minutos vigente desde el 15 de agosto, **When** se consulta la configuración del 10 de agosto, **Then** se muestra la tolerancia de 10 minutos.
4. **Given** un horario de mañana que se superpone con el de tarde en el mismo periodo, **When** el administrador lo guarda, **Then** el sistema rechaza el cambio e indica el solapamiento.
5. **Given** un director con sesión iniciada, **When** modifica la hora de ingreso de un turno, **Then** el cambio queda aplicado y auditado con su usuario.
6. **Given** un director con sesión iniciada, **When** intenta crear un año escolar, un grado o un feriado, **Then** el sistema no ofrece esas opciones.

---

### User Story 3 - Registrar feriados, días no laborables y suspensiones (Priority: P2)

El administrador registra un día o un rango de días como feriado, no laborable o suspensión de clases. Cada excepción aplica a todo el colegio o solo a grados específicos. Un día marcado como no laborable no genera faltas para los alumnos alcanzados por la excepción.

**Why this priority**: Es necesario para que las faltas no se generen en días sin clases, pero el colegio puede operar la clasificación básica sin ella al inicio.

**Independent Test**: Se registra una suspensión para el 3.° grado en una fecha y un feriado para todo el colegio en otra. Se verifica que la consulta del calendario muestra ambas excepciones con su alcance correcto.

**Acceptance Scenarios**:

1. **Given** un año escolar activo, **When** el administrador registra un feriado para todo el colegio el 28 de julio, **Then** el día aparece en el calendario como no laborable para todos los grados.
2. **Given** un año escolar activo, **When** el administrador registra una suspensión del 10 al 12 de mayo solo para 4.° y 5.° grado, **Then** esos días aparecen como suspendidos para esos grados y como laborables para los demás.
3. **Given** un feriado que cae en sábado o domingo, **When** el administrador lo registra, **Then** el sistema lo guarda y lo muestra en el calendario, sin contar días de falta adicionales.
4. **Given** una excepción ya registrada que se superpone con otra del mismo alcance, **When** el administrador la guarda, **Then** el sistema indica el solapamiento y pide confirmar el reemplazo o rechaza el cambio.

---

### User Story 4 - Configurar umbrales de alerta (Priority: P3)

El administrador define umbrales de alerta por mes, por ejemplo cuántas tardanzas o faltas en un mes activan una alerta para el alumno. El director consulta esos umbrales y las alertas que generan.

**Why this priority**: Aporta valor de seguimiento, pero no bloquea el registro de asistencia ni la clasificación básica.

**Independent Test**: Se configura un umbral de 3 tardanzas por mes. Se verifica que el valor queda guardado con su vigencia y que se puede consultar.

**Acceptance Scenarios**:

1. **Given** un año escolar activo, **When** el administrador configura un umbral de 3 tardanzas por mes, **Then** el umbral queda guardado con la fecha desde la que aplica.
2. **Given** un umbral configurado, **When** el administrador lo cambia a 4, **Then** el cambio aplica desde la fecha indicada y el valor anterior queda registrado en la historia.
3. **Given** un umbral con valor cero o negativo, **When** el administrador intenta guardarlo, **Then** el sistema rechaza el valor.

---

### User Story 5 - Consultar la configuración vigente en una fecha pasada (Priority: P3)

El administrador o el director consulta qué configuración estaba vigente en una fecha determinada: horarios por turno, tolerancias, feriados, umbrales y turno de cada sección. Esta consulta permite explicar por qué una asistencia pasada se clasificó como se clasificó.

**Why this priority**: Es esencial para auditoría y reclamos, pero puede usarse después de que la configuración ya exista.

**Independent Test**: Se hacen dos cambios de tolerancia en fechas distintas. Se consulta la configuración en tres fechas: antes del primer cambio, entre ambos y después del segundo. Se verifica que cada consulta devuelve el valor correcto.

**Acceptance Scenarios**:

1. **Given** dos cambios de tolerancia vigentes desde fechas distintas, **When** se consulta la configuración de una fecha entre ambos, **Then** se muestra el valor del primer cambio.
2. **Given** una fecha anterior al inicio del año escolar, **When** se consulta la configuración, **Then** el sistema indica que no hay configuración vigente para esa fecha.
3. **Given** una sección que cambió de turno, **When** se consulta su turno en una fecha anterior al cambio, **Then** se muestra el turno que tenía en esa fecha.

---

### User Story 6 - Definir las horas de clase de cada turno (Priority: P1)

El administrador define, por cada turno, los bloques de clase: número, hora de inicio y hora de fin. Los bloques deben caer dentro del horario del turno y no pueden solaparse. Cada cambio tiene vigencia desde una fecha. El director puede ajustar las horas de inicio y fin de los bloques, pero no crearlos ni eliminarlos.

**Why this priority**: El registro del docente se hace por hora de clase (feature 003). Sin bloques definidos no existe la unidad de registro.

**Independent Test**: Se configuran seis bloques para el turno de mañana. Se intenta crear un bloque que se solapa con otro y uno que termina después de la hora de cierre. Se verifica que ambos son rechazados y que los bloques válidos quedan guardados con su vigencia.

**Acceptance Scenarios**:

1. **Given** el turno de mañana con ingreso 07:30 y cierre 12:30, **When** el administrador crea el bloque 1 de 07:45 a 08:30, **Then** el bloque queda guardado con su fecha de vigencia.
2. **Given** el bloque 1 de 07:45 a 08:30, **When** el administrador crea el bloque 2 de 08:20 a 09:05, **Then** el sistema rechaza el bloque e indica el solapamiento.
3. **Given** el turno de mañana con cierre a las 12:30, **When** el administrador crea un bloque que termina a las 13:00, **Then** el sistema rechaza el bloque por quedar fuera del horario del turno.
4. **Given** un director con sesión iniciada, **When** ajusta la hora de inicio del bloque 3, **Then** el cambio queda auditado y el director no puede crear ni eliminar bloques.

---

### User Story 7 - Definir cursos y armar el horario semanal de cada sección (Priority: P1)

El administrador crea los cursos del colegio (por ejemplo, Matemática o Educación Física), los vincula a uno o varios grados y arma el horario semanal de cada sección. Para cada día laborable y cada bloque de su turno, asigna un curso vinculado al grado de la sección, o deja el bloque sin curso. Con ese horario, el sistema sabe qué curso corresponde a cada hora.

**Why this priority**: Permite que la asistencia por hora de clase y la referencia por curso (features 003 y 006) tengan un curso concreto para cada registro.

**Independent Test**: Se crean tres cursos para 3.° grado y se arma el horario de "3.° A" para lunes a viernes. Se consulta qué curso corresponde a un martes en el bloque 2. Se verifica que la respuesta coincide con el horario armado y que un curso de otro grado es rechazado.

**Acceptance Scenarios**:

1. **Given** el curso Matemática del 3.° grado y la sección "3.° A" en el turno de mañana, **When** el administrador lo asigna al bloque 1 de los lunes, **Then** la consulta del lunes en el bloque 1 muestra Matemática.
2. **Given** el curso Educación Física vinculado solo a 4.° y 5.° grado, **When** el administrador intenta asignarlo a "3.° A", **Then** el sistema rechaza la asignación porque el curso no está vinculado al grado de la sección.
3. **Given** el bloque 2 del martes de "3.° A" ya asignado a Lenguaje, **When** el administrador asigna Historia a ese mismo bloque, **Then** el sistema pide confirmar el reemplazo antes de guardarlo.
4. **Given** un curso con registros de asistencia asociados, **When** el administrador intenta eliminarlo, **Then** el sistema no lo elimina: lo desactiva para que no pueda asignarse, y conserva su historial.
5. **Given** un bloque en un día que es feriado para la sección, **When** se consulta el horario de ese día, **Then** el bloque no aparece como hora de clase para registro.

---

### Edge Cases

- ¿Qué ocurre si se cambia la tolerancia a mitad de año? El cambio aplica desde la fecha indicada en adelante; las clasificaciones anteriores no cambian y la consulta histórica muestra ambos valores.
- ¿Qué ocurre si se decreta una suspensión el mismo día, después de que el turno ya abrió? La suspensión aplica a partir de ese momento: los alumnos del grado alcanzado no reciben falta por ese día, y los registros de ingreso ya hechos se conservan sin modificación.
- ¿Qué ocurre si una sección tiene un turno distinto al que suele tener su grado? El horario que aplica a una sección es el de su turno asignado, no el de su grado. Los grados no tienen horario propio.
- ¿Qué ocurre si un feriado cae en fin de semana? Se guarda y se muestra en el calendario, pero no altera el conteo de días laborables porque el fin de semana ya no es laborable.
- ¿Qué ocurre si la suspensión tiene alcance a un grado que no existe en el año activo? El sistema rechaza el grado inexistente y no guarda la excepción para ese alcance.
- ¿Qué ocurre si el administrador intenta cambiar un horario con fecha de vigencia en el pasado? El sistema rechaza la fecha; los cambios solo pueden aplicar desde hoy hacia adelante.
- ¿Qué ocurre si se intenta activar un segundo año escolar mientras otro sigue activo? El sistema impide dos años activos al mismo tiempo, e indica cuál debe cerrarse primero.
- ¿Qué ocurre si un horario de turno cruza la medianoche? El sistema rechaza la configuración: todos los horarios de un turno deben estar dentro del mismo día.
- ¿Qué ocurre si la hora de ingreso más la tolerancia supera la hora límite de tardanza? El sistema rechaza el cambio, porque la tolerancia no puede terminar después del límite de tardanza.
- ¿Qué ocurre si una sección cambia de turno? Sus bloques y su horario pasan a los del nuevo turno desde la fecha de vigencia. El horario anterior se conserva en el historial.
- ¿Qué ocurre si cambian los bloques de un turno a mitad de año? Aplican desde la fecha de vigencia, igual que las tolerancias. Las horas de clase ya registradas no cambian, y la consulta histórica muestra ambas versiones.
- ¿Qué ocurre si un curso se dicta en varios grados? Se permite. Cada sección tiene su propio horario, y el docente de cada una se asigna por separado.
- ¿Qué ocurre con un recreo o refrigerio? No se define como bloque de hora de clase y no genera registro de asistencia del docente.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: El sistema MUST permitir crear un año escolar con fecha de inicio y fecha de fin, validando que la de fin sea posterior a la de inicio.
- **FR-002**: El sistema MUST permitir que exista como máximo un año escolar activo a la vez.
- **FR-003**: El sistema MUST permitir cerrar un año escolar. Un año cerrado MUST conservar su configuración y su historial, y MUST NOT admitir cambios.
- **FR-004**: El sistema MUST permitir crear grados y secciones dentro del año escolar activo. Un grado MUST tener un nombre único dentro del año.
- **FR-005**: Cada sección MUST pertenecer a un grado y a un turno. Una sección MUST tener un único turno en un momento dado.
- **FR-006**: El sistema MUST ofrecer exactamente dos turnos: mañana y tarde.
- **FR-007**: El sistema MUST permitir configurar por turno: hora de ingreso, tolerancia de puntualidad en minutos, hora límite de tardanza y hora de cierre del registro.
- **FR-008**: El sistema MUST rechazar una configuración de turno en la que la hora de cierre no sea posterior a la hora límite de tardanza.
- **FR-009**: El sistema MUST rechazar una configuración de turno en la que la hora de ingreso más la tolerancia sea posterior a la hora límite de tardanza.
- **FR-010**: El sistema MUST rechazar horarios que crucen la medianoche. Todos los horarios de un turno MUST estar dentro del mismo día.
- **FR-011**: El sistema MUST rechazar cualquier combinación de horarios en la que los horarios de mañana y de tarde se superpongan en el periodo de vigencia.
- **FR-012**: Cada cambio de horario o tolerancia MUST indicar una fecha de vigencia. El cambio MUST aplicar desde esa fecha hacia adelante.
- **FR-013**: El sistema MUST rechazar fechas de vigencia anteriores a la fecha actual para cambios de configuración.
- **FR-014**: Un cambio de horario o tolerancia MUST NOT recalcular ni modificar la clasificación de asistencias ya registradas.
- **FR-015**: El sistema MUST permitir registrar feriados, días no laborables y suspensiones de clases, tanto por un día como por un rango de fechas.
- **FR-016**: Cada excepción de calendario MUST indicar su alcance: todo el colegio o uno o más grados específicos del año activo.
- **FR-017**: Un día marcado como no laborable (feriado, no laborable o suspensión) para un alumno MUST NOT generar falta para ese alumno.
- **FR-018**: Una suspensión registrada después de que el turno ya abrió MUST aplicar desde el momento de su registro: los registros de ingreso ya hechos se conservan, y no se generan faltas por ese día para los grados alcanzados.
- **FR-019**: El sistema MUST permitir registrar excepciones de calendario que caigan en fin de semana. Estas MUST mostrarse en el calendario sin alterar el conteo de días laborables.
- **FR-020**: El sistema MUST advertir cuando una nueva excepción de calendario se superpone con otra del mismo alcance, y MUST exigir confirmación antes de reemplazarla.
- **FR-021**: El sistema MUST permitir configurar umbrales de alerta por mes para tardanzas y faltas. Cada umbral MUST ser un número entero mayor que cero.
- **FR-022**: El sistema MUST registrar la historia de los cambios de umbrales de alerta, con su fecha de vigencia.
- **FR-023**: El sistema MUST permitir consultar la configuración vigente en cualquier fecha del año escolar, incluyendo horarios por turno, tolerancias, excepciones de calendario, umbrales y turno de cada sección.
- **FR-024**: Para una fecha anterior al inicio del año escolar, la consulta MUST indicar que no hay configuración vigente.
- **FR-025**: El rol Administrador MUST poder crear y modificar año escolar, grados, secciones, turnos, horarios, excepciones de calendario y umbrales de alerta.
- **FR-026**: El rol Director MUST poder consultar toda la configuración y modificar únicamente las horas de ingreso, tolerancias, horas límite de tardanza y horas de cierre de los turnos.
- **FR-027**: Todo cambio de configuración MUST quedar auditado con el usuario responsable, la fecha y hora del cambio, el valor anterior, el valor nuevo y la fecha de vigencia.
- **FR-028**: El sistema MUST permitir cambiar el turno de una sección indicando una fecha de vigencia. La consulta histórica MUST mostrar el turno que tenía la sección en cada fecha.
- **FR-029**: El horario que aplica a una sección MUST ser el de su turno asignado en la fecha consultada, sin importar el horario habitual de su grado.
- **FR-030**: El sistema MUST permitir crear cursos dentro del año escolar activo. Cada curso MUST vincularse al menos a un grado del año activo, y su nombre MUST ser único dentro del año escolar.
- **FR-031**: Un curso MUST NOT eliminarse si tiene registros de asistencia asociados. En ese caso MUST desactivarse, dejando de poder asignarse sin perder su historial.
- **FR-032**: El sistema MUST permitir definir, por cada turno, bloques de hora de clase con número, hora de inicio y hora de fin, con fecha de vigencia.
- **FR-033**: Cada bloque de hora de clase MUST quedar dentro del horario del turno, desde la hora de ingreso hasta la hora de cierre, con hora de fin posterior a la de inicio, y MUST NOT cruzar la medianoche.
- **FR-034**: Los bloques de hora de clase de un mismo turno MUST NOT solaparse en el periodo de vigencia.
- **FR-035**: El sistema MUST permitir armar el horario semanal de cada sección: para cada día laborable y cada bloque de su turno, asignar un curso o dejar el bloque sin curso. El horario semanal MUST repetirse cada semana hasta que cambie con una nueva vigencia.
- **FR-036**: El curso asignado a un bloque MUST estar vinculado al grado de la sección. El sistema MUST rechazar cualquier otro curso.
- **FR-037**: Una sección MUST tener como máximo un curso por día laborable y bloque. Si el bloque ya tiene curso, el sistema MUST pedir confirmación antes de reemplazarlo.
- **FR-038**: Los cambios de bloques de hora de clase y de horario de sección MUST tener fecha de vigencia y aplicar desde esa fecha hacia adelante, sin modificar los registros de asistencia ya hechos.
- **FR-039**: El sistema MUST permitir consultar, para una fecha, una sección y un bloque, qué curso corresponde según la vigencia vigente en esa fecha, o indicar que el bloque no tiene clase.
- **FR-040**: Un bloque de hora de clase en un día que es no laborable o suspensión para la sección MUST NOT generar hora de clase para registro ni falta.
- **FR-041**: El administrador MUST poder crear y modificar cursos, bloques de hora de clase y horarios de sección. El director MUST poder únicamente ajustar la hora de inicio y de fin de los bloques existentes.
- **FR-042**: Todo cambio sobre cursos, bloques de hora de clase y horarios de sección MUST quedar auditado con usuario, fecha y hora, valor anterior, valor nuevo y fecha de vigencia.
- **FR-043**: Los recreos, refrigerios y otros tiempos no lectivos MUST NOT definirse como bloques de hora de clase, y MUST NOT generar registros de asistencia del docente.

### Key Entities *(include if feature involves data)*

- **Año escolar**: Periodo lectivo con fecha de inicio, fecha de fin y estado (activo o cerrado). Contiene grados, secciones, excepciones de calendario y umbrales.
- **Grado**: Nivel dentro del año escolar (por ejemplo, 3.°). Agrupa secciones. No tiene horario propio.
- **Sección**: División de un grado (por ejemplo, "A"). Pertenece a un grado y está asignada a un turno en una fecha de vigencia.
- **Turno**: Uno de los dos turnos fijos (mañana o tarde). Tiene horarios configurables por vigencia.
- **Horario de turno**: Conjunto de valores de un turno (hora de ingreso, tolerancia, hora límite de tardanza, hora de cierre) con fecha de vigencia desde la que aplica.
- **Excepción de calendario**: Feriado, día no laborable o suspensión, con fecha o rango, alcance (todo el colegio o grados específicos) y motivo.
- **Umbral de alerta**: Valor mensual de tardanzas o faltas que activa una alerta, con fecha de vigencia.
- **Registro de auditoría de configuración**: Entrada inmutable de cambio, con usuario, fecha, valor anterior, valor nuevo y fecha de vigencia.
- **Curso**: Asignatura del colegio dentro del año escolar (por ejemplo, Matemática o Educación Física). Se vincula a uno o varios grados, puede dictarse en varias secciones de esos grados y puede estar activo o desactivado.
- **Bloque de hora de clase**: Número de bloque de un turno con hora de inicio y fin, vigente desde una fecha. Es la unidad de registro de asistencia del docente.
- **Horario de sección**: Asignación de un curso, o de ningún curso, a una combinación de sección, día laborable y bloque, con fecha de vigencia.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Un administrador crea un año escolar completo (dos turnos, grados y secciones) en menos de 60 minutos sin ayuda técnica.
- **SC-002**: El 100% de los cambios de horario o tolerancia con fecha de vigencia se aplica desde esa fecha y no modifica ninguna asistencia anterior, verificado en pruebas de aceptación.
- **SC-003**: El 100% de las combinaciones de horarios inválidas (cierre antes o igual al límite, tolerancia que supera el límite, horarios que se superponen o cruzan la medianoche) es rechazado con un mensaje que indica el motivo.
- **SC-004**: Un administrador registra un feriado o una suspensión para un grado en menos de 1 minuto.
- **SC-005**: La consulta de configuración para cualquier fecha del año escolar devuelve los valores correctos en el 100% de los casos de prueba, incluidos los cambios a mitad de año.
- **SC-006**: Un director puede ajustar una hora de ingreso y ver su vigencia en menos de 1 minuto, sin poder crear ni modificar otros elementos de la configuración.
- **SC-007**: Un administrador arma el horario semanal completo de una sección (cinco días, seis bloques) en menos de 10 minutos.
- **SC-008**: El 100% de los bloques solapados, fuera del horario del turno o con cursos de otro grado es rechazado con un mensaje que indica el motivo, verificado en pruebas de aceptación.

## Assumptions

- El colegio tiene exactamente dos turnos, mañana y tarde. Los turnos no se crean ni eliminan.
- Los días laborables por defecto son lunes a viernes. Los sábados y domingos se tratan como no laborables salvo que el colegio registre una excepción específica.
- Los horarios se definen por turno, no por grado. Un grado puede tener secciones en turnos distintos, y cada sección usa el horario de su turno.
- La clasificación de asistencia (puntual, tardanza, falta) se define en la funcionalidad de registro de asistencia; esta especificación solo provee los valores y su vigencia. El orden de referencia es: ingreso, luego ingreso más tolerancia, luego límite de tardanza, y finalmente cierre.
- Las fechas de vigencia de los cambios de configuración deben ser hoy o posteriores. Las correcciones de configuración pasada se hacen mediante un cambio nuevo con vigencia desde hoy, sin reescribir la historia.
- Los umbrales de alerta se aplican por alumno y por mes calendario. La entrega de alertas a apoderados está fuera de alcance (el sistema no notifica a apoderados).
- Los feriados nacionales no se precargan automáticamente; los registra el administrador cada año escolar.
- Un año escolar cerrado no admite cambios. Los cambios de configuración de un año futuro se preparan dentro de su propio año escolar, no en el activo.
- Los cursos son asignaturas del colegio que pueden vincularse a uno o varios grados (por ejemplo, Educación Física para 1.° a 5.°). El horario semanal se repite cada semana hasta que cambie con una nueva vigencia (ver FR-035).
- Los recreos, refrigerios y otros tiempos no lectivos no son bloques de clase (ver FR-043).
- La asignación de docentes a cada curso y sección queda fuera de esta feature. Se define en la feature 003.
- Los valores numéricos de referencia (tolerancia en minutos, umbrales) son parámetros configurables; no se fijan valores por defecto en esta especificación.
