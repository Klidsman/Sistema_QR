# Feature Specification: Registro de asistencia de alumnos por el docente

**Feature Branch**: `003-registro-asistencia-docente` (no creada; el hook de git no está configurado)

**Created**: 2026-11-05

**Status**: Draft

**Input**: User description: "Registro de asistencia de alumnos por el docente en su salón. Además del ingreso por QR en portería, el docente registra la asistencia de los alumnos de sus secciones asignadas, y el sistema contrasta ambos registros. Solo se registra asistencia de alumnos."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Registrar la asistencia del salón para uso del docente (Priority: P1)

El docente abre la lista de alumnos de una hora de clase de un curso y sección que tiene asignados. Para cada alumno ve si ya ingresó por QR en portería y a qué hora. Marca el estado de asistencia (presente, tardanza, falta o falta justificada) y, si lo necesita, agrega una observación. Puede marcar a todos los alumnos como presentes de una vez y luego corregir solo las excepciones. Este registro es una herramienta personal del docente: sirve para llevar su salón, facilitar su registro manual y exportar un reporte. No es la asistencia oficial.

**Why this priority**: Es el núcleo de la feature. Sin esta herramienta el docente no puede llevar su salón ni exportar su registro.

**Independent Test**: Un docente con una sección de 30 alumnos, algunos ya registrados por QR, marca a todos como presentes y luego corrige tres excepciones. Se verifica que cada estado queda guardado y que los alumnos con diferencias frente al QR aparecen destacados.

**Acceptance Scenarios**:

1. **Given** un docente asignado a Matemática en "3.° A", **When** abre la lista de una hora de clase de Matemática del día, **Then** ve solo los alumnos activos de esa sección con su estado de ingreso por QR (sí, con hora; o no ingresó).
2. **Given** la lista del día abierta sin registros del docente, **When** el docente pulsa "Marcar a todos presentes", **Then** todos los alumnos sin registro quedan como presentes, y los que ya tenían registro no se modifican.
3. **Given** un alumno marcado como presente, **When** el docente lo cambia a tardanza, **Then** el estado queda guardado y el registro del QR de ese alumno no cambia.
4. **Given** un alumno, **When** el docente agrega una observación de texto, **Then** la observación queda guardada junto al registro de ese alumno y del día.
5. **Given** un alumno que ingresó por QR a las 07:42, **When** el docente lo marca como falta en su registro personal, **Then** el registro del docente queda guardado y el ingreso por QR no cambia.

---

### User Story 2 - Modificar un registro dentro del plazo y, después del plazo, solo con administrador (Priority: P2)

El docente puede corregir sus propios registros del día durante un plazo configurable desde su registro original. Cada cambio queda auditado. Pasado el plazo, solo el administrador puede modificar el registro, y debe indicar el motivo.

**Why this priority**: Permite corregir errores reales sin abrir el registro indefinidamente, y mantiene la integridad de los datos.

**Independent Test**: Un docente registra la asistencia, la corrige dentro del plazo y verifica la auditoría. Luego simula que el plazo venció: el docente ya no puede editar, y el administrador sí puede hacerlo indicando un motivo.

**Acceptance Scenarios**:

1. **Given** un registro del docente creado dentro del plazo, **When** el docente lo cambia de falta a presente, **Then** el cambio se guarda y queda registrado con usuario, fecha, valor anterior y valor nuevo.
2. **Given** un registro cuyo plazo de edición venció, **When** el docente intenta modificarlo, **Then** el sistema no permite el cambio y explica que solo un administrador puede hacerlo.
3. **Given** un registro con plazo vencido, **When** el administrador lo modifica, **Then** el sistema exige un motivo y lo guarda junto con el cambio y la auditoría.
4. **Given** un administrador con sesión iniciada, **When** modifica un registro sin indicar motivo, **Then** el sistema rechaza el cambio.

---

### User Story 3 - Asignar docentes a cursos y secciones, incluidos reemplazos por un día (Priority: P2)

El administrador asigna cada curso de cada sección a su docente, con fecha de inicio y de fin. Un reemplazante recibe una asignación temporal, por uno o más días, que le da acceso a ese curso y sección solo durante ese rango. Un docente con varios cursos o secciones ve y registra todos ellos.

**Why this priority**: Sin asignaciones correctas el docente no puede registrar. El reemplazo es un caso frecuente que no debe requerir cambios permanentes.

**Independent Test**: El administrador asigna a un docente titular a dos cursos y secciones, y a un reemplazante a uno de ellos por un día. Se verifica que cada docente ve solo lo que le corresponde en cada fecha.

**Acceptance Scenarios**:

1. **Given** un docente asignado a Matemática en "3.° A" y a Lenguaje en "4.° B", **When** entra en un día laborable, **Then** ve los alumnos de ambas secciones y puede registrar en las horas de clase de sus cursos.
2. **Given** un reemplazante asignado a Matemática en "3.° A" solo el 12 de mayo, **When** entra el 12 de mayo, **Then** ve y registra alumnos de Matemática en "3.° A"; el 13 de mayo no ve esa asignación.
3. **Given** un docente que no está asignado a un curso y sección, **When** intenta abrir su lista o registrar asistencia en ellos, **Then** el sistema no muestra los alumnos y rechaza el registro.
4. **Given** una asignación de reemplazo que se superpone con la del titular para el mismo día, **When** el administrador la guarda, **Then** el sistema la acepta y ambos docentes pueden ver la sección ese día.

---

### User Story 4 - Registrar por hora de clase (Priority: P3)

El docente registra la asistencia por hora de clase. Cada hora de su curso tiene su propio registro para cada alumno, de modo que el registro de una hora no reemplaza el de otra.

**Why this priority**: Es la única modalidad de registro del docente en esta versión, y sustenta la referencia de asistencia por curso.

**Independent Test**: Un docente con dos horas de clase en el día registra la asistencia de cada una por separado. Se verifica que cada hora conserva su propio registro por alumno.

**Acceptance Scenarios**:

1. **Given** una hora de clase de "3.° A" en el curso del docente, **When** el docente registra la asistencia, **Then** el registro queda asociado a esa fecha, sección, curso y hora, y no existe un registro único por turno.
2. **Given** un turno configurado para registro por hora de clase, **When** el docente registra la asistencia de una hora, **Then** el registro queda asociado a esa hora y no reemplaza a los de otras horas del día.

---

### Edge Cases

- ¿Qué ocurre si el alumno ingresó por QR pero el docente lo marca falta, o al revés? Ambos registros se conservan sin modificarse; el alumno aparece destacado como diferencia en la lista del docente, mostrando el estado del QR y el del docente.
- ¿Qué ocurre si el docente tiene varios cursos o secciones? Ve cada combinación de curso y sección por separado y registra cada una sin mezclar alumnos ni estados.
- ¿Qué ocurre si el docente de la sección se reemplaza por un día? El reemplazante registra con su propio usuario; el registro queda asociado al curso, la sección y la fecha, y el reemplazante no conserva acceso después del rango asignado.
- ¿Qué ocurre si la conexión se corta mientras el docente registra? El registro se guarda en el dispositivo, se muestra como "pendiente de confirmar" y se envía al reconectar. Ningún registro se pierde, y reenviar el mismo registro no crea duplicados.
- ¿Qué ocurre si el docente intenta marcar a todos presentes después de haber corregido alumnos? Solo se marcan los alumnos que aún no tienen registro del docente para ese día, curso y hora de clase; los ya registrados no cambian.
- ¿Qué ocurre si un alumno fue cambiado de sección durante el año? Los registros de fechas anteriores al cambio permanecen en la sección original. A partir de la fecha del cambio, el alumno aparece en la nueva sección y sus registros se generan allí.
- ¿Qué ocurre si el alumno se retira o se traslada durante el día? No aparece en la lista para registrar en días posteriores al cambio, pero sus registros anteriores se conservan.
- ¿Qué ocurre si el día es feriado o no laborable para la sección? La lista no solicita registro para ese día y no genera faltas (según la configuración académica).
- ¿Qué ocurre si el docente intenta registrar un día futuro? El sistema no permite registros para fechas futuras.
- ¿Qué ocurre si una observación supera la longitud máxima? El sistema la rechaza y conserva el texto que el docente escribió.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: El sistema MUST mostrar al docente únicamente los alumnos activos de las secciones en las que tiene asignado un curso, en la fecha consultada.
- **FR-002**: El sistema MUST permitir registrar para cada alumno uno de cuatro estados: presente, tardanza, falta o falta justificada.
- **FR-003**: El sistema MUST mostrar, junto a cada alumno de la lista, si ingresó por QR en portería ese día y a qué hora, o indicar que no tiene ingreso por QR.
- **FR-004**: El registro del docente MUST guardarse por separado del registro de ingreso por QR. Ninguno de los dos MUST sobrescribir al otro.
- **FR-005**: El sistema MUST NOT destacar diferencias entre el registro del docente y el ingreso por QR en la lista del docente.
- **FR-006**: El docente MUST poder agregar una observación de texto por alumno y por registro. La observación MUST tener una longitud máxima de 300 caracteres.
- **FR-007**: El sistema MUST ofrecer la acción "Marcar a todos presentes", que aplica solo a los alumnos que aún no tienen registro del docente para esa fecha, curso y hora de clase.
- **FR-008**: Un alumno MUST tener como máximo un registro del docente por fecha, curso y hora de clase.
- **FR-009**: El docente MUST poder modificar sus propios registros durante un plazo de edición configurable, contado desde la creación del registro original. El valor inicial del plazo es de 24 horas.
- **FR-010**: Pasado el plazo de edición, el docente MUST NOT poder modificar el registro. Solo un administrador MUST poder hacerlo, indicando un motivo obligatorio.
- **FR-011**: Todo cambio sobre un registro del docente MUST quedar auditado con el usuario responsable, la fecha y hora, el valor anterior, el valor nuevo y, si aplica, el motivo.
- **FR-012**: Un docente MUST NOT poder registrar asistencia para fechas futuras.
- **FR-013**: El administrador MUST poder asignar docentes a cada combinación de curso y sección, con fecha de inicio y de fin, y MUST poder asignar un reemplazante por un día o un rango de días.
- **FR-014**: Un docente asignado a varios cursos o secciones MUST poder ver y registrar en todos ellos, cada uno por separado.
- **FR-015**: Una asignación de reemplazo MUST dar acceso al curso y la sección solo durante su rango de fechas. Terminado el rango, el acceso MUST cesar automáticamente.
- **FR-016**: Cada registro del docente MUST guardar el usuario que lo creó, aun cuando fue un reemplazante, de forma que la responsabilidad del registro sea trazable.
- **FR-017**: El registro MUST guardarse en el dispositivo cuando la conexión falla, mostrar su estado como "pendiente de confirmar" y enviarse al servidor al recuperar la conexión, sin pérdida de datos.
- **FR-018**: El envío de un registro pendiente MUST ser idempotente: reenviar el mismo registro MUST NOT crear duplicados.
- **FR-019**: Cuando un alumno cambia de sección, los registros de fechas anteriores al cambio MUST permanecer asociados a la sección original, y los registros desde la fecha del cambio MUST asociarse a la nueva.
- **FR-020**: Un alumno retirado o trasladado MUST NOT aparecer en la lista para registrar en fechas posteriores a su salida, sin perder los registros anteriores.
- **FR-021**: En días marcados como no laborables o feriados para la sección, el sistema MUST NOT solicitar registro ni generar faltas, según la configuración académica.
- **FR-022**: El registro del docente MUST hacerse solo por hora de clase, identificada por fecha, sección, curso y hora. No existe registro por turno en esta versión. El docente MUST poder registrar únicamente en las horas de clase de los cursos y secciones que tiene asignados.
- **FR-023**: El registro oficial de asistencia es el ingreso por QR en la puerta principal. El registro del docente no es asistencia oficial ni se usa para calcular estados oficiales. Sirve solo para facilitar su registro manual y exportar un reporte.
- **FR-024**: El docente MUST registrar únicamente la asistencia de alumnos. El sistema MUST NOT ofrecer registro de asistencia de docentes ni de auxiliares.
- **FR-025**: El rol Director MUST poder consultar los registros del docente, sin poder modificarlos.
- **FR-026**: El docente MUST poder exportar el registro de sus cursos y secciones, por fecha o rango de fechas, a un archivo descargable.

### Key Entities *(include if feature involves data)*

- **Asignación docente-curso-sección**: Vínculo entre un docente, un curso y una sección, con fecha de inicio y fin. Indica si es titular o reemplazante. Un docente puede tener varias asignaciones.
- **Registro de asistencia del docente**: Estado de un alumno en una fecha, curso y hora de clase. Atributos: alumno, sección en la fecha, estado, observación, docente que lo registró, fecha y hora de creación, y estado de confirmación (pendiente o confirmado).
- **Registro de ingreso por QR**: Ingreso registrado en portería. Es independiente del registro del docente; esta feature solo lo consulta y lo contrasta.
- **Corrección de registro**: Cambio posterior a la creación de un registro del docente. Atributos: usuario, fecha y hora, valor anterior, valor nuevo y motivo cuando corresponde.
- **Diferencia de asistencia**: Situación en la que el estado del docente y el estado del QR no coinciden para un alumno en una fecha. Se deriva de los dos registros y no se almacena como dato independiente.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Un docente registra la asistencia de un salón de 40 alumnos en menos de 2 minutos, usando "Marcar a todos presentes" y corrigiendo solo las excepciones.
- **SC-002**: El 100% de las diferencias entre el registro del docente y el del QR se muestra destacada en la lista, verificado en pruebas de aceptación.
- **SC-003**: El 100% de los registros hechos con la conexión interrumpida se conserva y se confirma al recuperarla, sin duplicados, verificado en pruebas de desconexión.
- **SC-004**: El 100% de las modificaciones posteriores al plazo de edición requiere intervención del administrador y un motivo registrado.
- **SC-005**: Un docente nunca ve ni registra alumnos de cursos y secciones a los que no está asignado en la fecha consultada, verificado en pruebas de acceso.
- **SC-006**: El acceso de un reemplazante cesa automáticamente al terminar su rango de fechas, sin intervención manual.

## Assumptions

- El plazo de edición del docente es un parámetro de la configuración académica (feature 002), con valor inicial de 24 horas. Esta feature asume que el administrador puede ajustarlo; la pantalla de configuración de 002 deberá incluirlo.
- El docente solo registra asistencia de alumnos de cursos y secciones asignados; no existe registro de asistencia de docentes en el sistema.
- El Director puede consultar los registros del docente y las diferencias con el QR, pero no modificarlos. Esta regla se toma del rol descrito en la funcionalidad de alumnos.
- La "falta justificada" es un estado más, sin adjuntar documentos de justificación en esta versión.
- Las observaciones son texto simple, sin adjuntos, y son visibles para el docente, el administrador y el director.
- El docente accede desde un celular o una computadora con navegador moderno; no se requiere una aplicación instalada.
- La asistencia del docente no genera notificaciones a apoderados, según el alcance del proyecto.
- La clasificación de puntualidad y tardanza que aparece en el QR es la definida por la configuración académica; el registro del docente no recalcula esa clasificación.
- La asignación docente-curso-sección se gestiona por el administrador. Los cursos y las horas de clase vienen de la feature 002. Esta feature no incluye la gestión de docentes como personas (datos, contratos, calificaciones).
- Calificaciones, comportamiento y pagos están fuera de alcance.
