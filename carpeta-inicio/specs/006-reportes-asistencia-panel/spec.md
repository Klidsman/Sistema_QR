# Feature Specification: Reportes de asistencia diarios y mensuales y panel de indicadores

**Feature Branch**: `006-reportes-asistencia-panel` (no creada; el hook de git no está configurado)

**Created**: 2026-11-05

**Status**: Draft

**Input**: User description: "Reportes de asistencia diarios y mensuales y panel de indicadores. La dirección y los docentes necesitan ver el estado de asistencia de los alumnos: quién llegó puntual, quién llegó tarde, quién faltó y quién está justificado."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Reporte diario por grado, sección y turno (Priority: P1)

El director, el administrador o el docente de una sección abre el reporte de un día. Ve cada alumno con su estado (presente, tardanza, falta, falta justificada o tardanza justificada) y la hora de marcación, agrupado por turno, grado y sección. Puede filtrar por turno, grado, sección, alumno o estado.

**Why this priority**: Es el reporte más consultado en la operación diaria y la base para los demás reportes.

**Independent Test**: Se abre el reporte de un día con dos turnos, tres grados y varias secciones. Se verifica que cada alumno aparece una sola vez, con su estado y su hora, y que los filtros reducen la lista correctamente.

**Acceptance Scenarios**:

1. **Given** un día con registros de ambos turnos, **When** el director abre el reporte diario, **Then** ve los alumnos agrupados por turno, luego por grado y sección, con su estado y hora de marcación.
2. **Given** el reporte diario abierto, **When** el usuario filtra por estado "tardanza", **Then** solo aparecen alumnos con tardanza en ese día.
3. **Given** un docente asignado a "3.° A", **When** abre el reporte diario, **Then** ve solo los alumnos de "3.° A" y no ve alumnos de otras secciones.
4. **Given** un alumno con falta justificada aprobada, **When** aparece en el reporte diario, **Then** su estado es "falta justificada" y no "falta".

---

### User Story 2 - Panel del día con presentes, tardanzas y ausentes (Priority: P1)

El director y el docente ven un panel con el resumen del día: cuántos alumnos están presentes, cuántos tienen tardanza, cuántos están ausentes y cuántos aún están pendientes de registro. El panel se actualiza casi en tiempo real y muestra los mismos totales que el reporte diario.

**Why this priority**: Permite la reacción inmediata durante la mañana, cuando la portería y los docentes registran la asistencia.

**Independent Test**: Se registran ingresos por QR y registros del docente durante la mañana. Se verifica que el panel refleja cada cambio en pocos segundos y que sus totales coinciden con el reporte diario de ese mismo momento.

**Acceptance Scenarios**:

1. **Given** un día con 100 alumnos activos, **When** 60 ingresan puntuales y 10 llegan tarde, **Then** el panel muestra 70 presentes (incluidas tardanzas según la definición vigente), 10 tardanzas y 30 pendientes.
2. **Given** el panel abierto, **When** un alumno ingresa por QR, **Then** el panel se actualiza en menos de 30 segundos sin que el usuario recargue la página.
3. **Given** un alumno sin registro al cierre de la sesión, **When** el cierre ocurre, **Then** aparece como ausente en el panel y en el reporte diario.
4. **Given** el mismo día, **When** se compara el panel con el reporte diario, **Then** los totales coinciden.

---

### User Story 3 - Reporte mensual por alumno y por sección (Priority: P2)

El usuario consulta un mes y ve, por alumno y por sección, el total de asistencias, tardanzas, faltas, faltas justificadas y tardanzas justificadas, junto con el porcentaje de asistencia. El porcentaje excluye feriados y días no laborables, y considera solo los días en que el alumno estuvo matriculado en la sección.

**Why this priority**: Es el seguimiento periódico que la dirección necesita, aunque no se usa a diario como el reporte del día.

**Independent Test**: Se calcula el mes de un alumno que ingresó a mitad de mes, con dos feriados y una suspensión. Se verifica que el denominador de días laborables es correcto y que el porcentaje coincide con el cálculo esperado.

**Acceptance Scenarios**:

1. **Given** un mes con 20 días laborables, **When** un alumno tiene 18 asistencias y 2 faltas, **Then** el reporte muestra 18 asistencias, 2 faltas, y el porcentaje correspondiente sobre 20 días.
2. **Given** un mes con dos feriados, **When** se calcula el porcentaje, **Then** los feriados no entran en el denominador.
3. **Given** un alumno que ingresó el día 16 del mes, **When** se calcula el reporte, **Then** solo se cuentan los días desde su ingreso.
4. **Given** un alumno que se retiró el día 10 del mes, **When** se calcula el reporte, **Then** solo se cuentan los días hasta su salida.
5. **Given** un alumno con falta justificada aprobada, **When** se calcula el reporte mensual, **Then** la falta aparece como justificada en su propia columna y queda fuera del denominador del porcentaje (ver FR-006).

---

### User Story 4 - Reporte por rango de fechas personalizado (Priority: P2)

El usuario elige una fecha inicial y una final, y obtiene el mismo tipo de resumen que el reporte mensual, aplicado a ese rango. Puede limitar el rango a un turno, grado, sección o alumno.

**Why this priority**: Cubre necesidades puntuales, como un periodo de evaluación o una quincena, sin tener que esperar el cierre del mes.

**Independent Test**: Se consulta un rango de tres semanas con un feriado en medio. Se verifica que el resumen coincide con la suma de los reportes diarios del mismo rango.

**Acceptance Scenarios**:

1. **Given** un rango de tres semanas, **When** el usuario consulta el resumen por sección, **Then** los totales coinciden con la suma de los reportes diarios de ese rango.
2. **Given** un rango con fecha final anterior a la inicial, **When** el usuario lo consulta, **Then** el sistema lo rechaza.
3. **Given** un rango que cruza dos meses, **When** el usuario lo consulta, **Then** el resumen total incluye los días de ambos meses sin duplicarlos.

---

### User Story 5 - Reporte de alumnos con tardanzas o faltas reiteradas (Priority: P2)

El director y el administrador ven la lista de alumnos que superaron los umbrales de reiteración configurados (tardanzas por mes o faltas seguidas) en un periodo, con su sección y el número de casos. Los docentes ven la misma lista para sus secciones.

**Why this priority**: Concentra la atención en los casos que requieren acción, usando los umbrales ya definidos en la configuración académica y en las alertas.

**Independent Test**: Se configura un umbral de 3 tardanzas por mes. Se verifica que la lista incluye a los alumnos que lo superaron en el periodo, y no incluye a los que solo tienen una o dos.

**Acceptance Scenarios**:

1. **Given** un umbral de 3 tardanzas por mes, **When** se consulta el reporte del mes, **Then** aparecen los alumnos con 3 o más tardanzas y no los demás.
2. **Given** un alumno con tres faltas, una de ellas justificada aprobada, **When** se consulta el reporte de faltas reiteradas, **Then** el alumno no aparece como reiterado por esa falta justificada.
3. **Given** un docente, **When** consulta el reporte, **Then** solo ve alumnos de sus secciones.

---

### User Story 6 - Exportar reportes a PDF y Excel (Priority: P2)

Cualquier reporte visible puede exportarse a PDF, para imprimir o archivar, o a Excel, para trabajar con los datos. La exportación respeta los filtros aplicados y los totales del reporte.

**Why this priority**: La dirección entrega informes impresos y trabaja los datos en hojas de cálculo.

**Independent Test**: Se exporta el reporte mensual de una sección a PDF y a Excel. Se verifica que los totales del archivo son iguales a los de la pantalla y que los nombres con tildes y ñ se ven correctamente.

**Acceptance Scenarios**:

1. **Given** un reporte mensual con filtros aplicados, **When** el usuario lo exporta a Excel, **Then** el archivo contiene solo los datos filtrados y los mismos totales que la pantalla.
2. **Given** un reporte diario, **When** el usuario lo exporta a PDF, **Then** el PDF muestra los nombres con tildes y ñ sin alteración.
3. **Given** un docente, **When** exporta un reporte, **Then** el archivo contiene solo alumnos de sus secciones.

---

### Edge Cases

- ¿Qué ocurre si un alumno ingresó o se retiró a mitad de mes? El reporte mensual cuenta solo los días en que estuvo matriculado en la sección, y el denominador del porcentaje se ajusta a esos días.
- ¿Qué ocurre en un mes con feriados y suspensiones? Los feriados y los días no laborables se excluyen del denominador. Un día con suspensión para un grado no cuenta para los alumnos de ese grado.
- ¿Qué ocurre si un alumno cambió de sección durante el periodo? El reporte mensual muestra el detalle por sección de cada día, de modo que la suma por sección coincide con el total del alumno. El alumno aparece en ambas secciones con sus días respectivos.
- ¿Qué ocurre con un reporte de un año completo con miles de registros? El reporte se genera en un tiempo aceptable (ver criterios de éxito) y se pagina o se procesa por bloques al exportar, sin perder filas.
- ¿Qué ocurre si el panel se consulta mientras un registro aún no está confirmado (pendiente de conexión)? El panel muestra el registro como pendiente y lo cuenta cuando se confirma.
- ¿Qué ocurre con un alumno sin registro antes del cierre de la sesión? Aparece como "pendiente" en el panel durante la sesión, y como "ausente" después del cierre.
- ¿Qué ocurre si un día fue corregido después de generarse un reporte? El reporte vuelve a calcularse con los datos vigentes. La versión exportada anteriormente no cambia, pero el reporte en pantalla refleja la corrección.
- ¿Qué ocurre si el usuario es docente y consulta un alumno de otra sección por su identificador? El sistema no muestra datos de ese alumno.
- ¿Qué ocurre si el usuario abre el panel en un celular en orientación horizontal o vertical? Los totales y la lista se muestran completos, sin recortes.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: El sistema MUST mostrar un reporte diario agrupado por turno, grado y sección, con el estado de cada alumno y su hora de marcación.
- **FR-002**: El reporte diario MUST mostrar solo los estados presente, tardanza, falta, falta justificada y tardanza justificada, y MUST distinguir las faltas y tardanzas justificadas de las injustificadas.
- **FR-003**: El sistema MUST mostrar un panel del día con los totales de presentes, tardanzas, ausentes y pendientes, actualizado en menos de 30 segundos después de cada registro o cambio.
- **FR-004**: El reporte mensual MUST mostrar, por alumno y por sección, el total de asistencias, tardanzas, faltas, faltas justificadas, tardanzas justificadas y el porcentaje de asistencia.
- **FR-005**: El porcentaje de asistencia MUST excluir los días feriados y los días no laborables del denominador, y MUST contar solo los días en que el alumno estuvo matriculado en la sección.
- **FR-006**: El porcentaje de asistencia MUST calcularse como asistencias (presentes y tardanzas) dividido entre los días laborables del alumno en la sección. Las faltas y tardanzas justificadas aprobadas MUST excluirse de ese denominador: ni suman ni restan.
- **FR-007**: El reporte por rango de fechas MUST producir el mismo resumen que la suma de los reportes diarios del mismo rango, sin duplicar días.
- **FR-008**: Todos los reportes MUST permitir filtrar por turno, grado, sección, alumno y estado, y los filtros MUST combinarse.
- **FR-009**: El reporte de reiteración MUST listar los alumnos que superaron los umbrales configurados (tardanzas por mes o faltas seguidas) en el periodo, con su sección y el número de casos.
- **FR-010**: El reporte de reiteración MUST excluir las faltas justificadas aprobadas del conteo.
- **FR-011**: Los reportes MUST NOT comparar el registro de ingreso por QR con el registro personal del docente.
- **FR-012**: Los totales del reporte diario, del reporte mensual, del reporte por rango y del panel MUST coincidir para los mismos datos y el mismo periodo.
- **FR-013**: Un usuario con rol docente MUST ver solo los datos de las secciones a las que estuvo asignado en la fecha consultada.
- **FR-014**: El director y el administrador MUST ver los reportes de todo el colegio.
- **FR-015**: Las justificaciones aprobadas MUST mostrarse con su estado de justificado en todos los reportes, y MUST NOT figurar como faltas injustificadas.
- **FR-016**: Un alumno que cambió de sección MUST aparecer en el reporte de cada sección con los días en que perteneció a ella, sin duplicar sus días.
- **FR-017**: Los reportes MUST exportarse a PDF y a Excel, respetando los filtros y los totales de la pantalla.
- **FR-018**: Los archivos exportados MUST conservar los caracteres con tildes, ñ y diéresis sin alteración.
- **FR-019**: Los reportes y el panel MUST ser utilizables desde un celular, en orientación vertical y horizontal, sin desplazamiento horizontal para leer los totales principales.
- **FR-020**: Un alumno sin registro antes del cierre de la sesión MUST aparecer como "pendiente" en el panel. Después del cierre, MUST aparecer como "ausente".
- **FR-021**: Un registro pendiente de confirmación MUST contarse en el panel solo cuando se confirme, y mientras tanto MUST mostrarse como pendiente.
- **FR-022**: Los reportes MUST reflejar las correcciones vigentes de registros, justificaciones y alertas aprobadas. La exportación ya descargada no cambia.
- **FR-023**: Los reportes MUST generarse sobre los registros existentes, sin mantener tablas de totales duplicadas, salvo que una medición de rendimiento lo justifique.
- **FR-024**: Los reportes de reiteración MUST respetar la misma restricción de acceso por sección que los demás reportes.
- **FR-025**: El rol Director MUST poder consultar todos los reportes, pero MUST NOT modificar los datos de asistencia desde ellos.

### Key Entities *(include if feature involves data)*

- **Reporte**: Vista calculada sobre los registros de asistencia. No se almacena como dato independiente. Tiene tipo (diario, mensual, rango, reiteración), periodo, filtros aplicados y el usuario que lo consulta.
- **Resumen de asistencia por alumno**: Totales de un alumno en un periodo: asistencias, tardanzas, faltas, justificadas y porcentaje, calculados por sección.
- **Indicador del día**: Totales de presentes, tardanzas, ausentes y pendientes de una fecha, calculados a partir de los registros en tiempo real.
- **Registro de asistencia**: Fuente de todos los cálculos. Incluye el estado, la hora de marcación, la sección en la fecha del evento y su origen (QR o docente).
- **Calendario de excepciones**: Feriados, días no laborables y suspensiones definidos en la configuración académica. Determinan los días que se excluyen del cálculo.
- **Matrícula de sección del alumno**: Fechas de inicio y fin de cada sección y de la matrícula del alumno. Determinan los días que cuentan para cada alumno y sección.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: El reporte diario de un colegio de 2,000 alumnos se muestra completo en menos de 5 segundos.
- **SC-002**: El reporte mensual de un año escolar completo, con miles de registros, se genera en menos de 10 segundos.
- **SC-003**: El 100% de los totales coincide entre el reporte diario, el mensual, el de rango y el panel para los mismos datos, verificado en pruebas de aceptación.
- **SC-004**: El panel refleja un nuevo ingreso en menos de 30 segundos sin recargar la página.
- **SC-005**: El 100% de los porcentajes excluye feriados y días no laborables, verificado en pruebas con meses de distintos calendarios.
- **SC-006**: Un docente nunca ve datos de alumnos de secciones a las que no está asignado, verificado en pruebas de acceso.
- **SC-007**: Los reportes principales se leen en un celular sin desplazamiento horizontal, verificado en orientación vertical y horizontal.
- **SC-008**: Las exportaciones a PDF y Excel contienen los mismos totales que la pantalla y conservan las tildes y la ñ en el 100% de las pruebas.

## Assumptions

- El reporte se calcula sobre la asistencia oficial, que es el ingreso por QR de la puerta principal (feature 003, FR-023). El registro personal del docente no entra en los totales.
- El panel de indicadores se actualiza mediante consulta periódica o actualización en vivo; el tiempo de 30 segundos es el máximo aceptado, no un valor fijo de implementación.
- "Presente" en el panel incluye las tardanzas, y se muestran por separado para que el director vea ambos totales.
- La exportación de un año completo puede entregarse por bloques si el archivo es demasiado grande, sin perder filas.
- Los reportes no se envían por correo ni por canales externos, según el alcance del proyecto.
- Los reportes de asistencia de docentes están fuera de alcance.
- Los reportes de calificaciones, comportamiento y pagos están fuera de alcance.
- El acceso a los reportes del docente usa la misma asignación por sección y fecha definida en la feature 003.
- Las alertas y justificaciones de las features 004 y 005 son la fuente de los estados "justificado" y de los casos de reiteración; esta feature no las redefine.
