# Feature Specification: Gestión de alumnos y generación de códigos QR

**Feature Branch**: `001-gestion-alumnos-qr` (no creada; el hook de git no está configurado)

**Created**: 2026-11-05

**Status**: Draft

**Input**: User description: "Gestión de alumnos y generación de códigos QR para un colegio: registro, edición y desactivación de alumnos, importación masiva desde Excel/CSV, QR único por alumno, carnés imprimibles por lote, invalidación de QR por pérdida o robo, retiro o traslado con historial, y cambio de grado o sección."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Registrar y mantener alumnos (Priority: P1)

El administrador registra a un alumno nuevo con su DNI o código, nombres, apellidos, año escolar, grado, sección, foto y datos de referencia del apoderado. Después puede buscarlo, editar sus datos y desactivarlo cuando corresponda. El director puede consultar la misma información sin modificarla.

**Why this priority**: Sin alumnos registrados no existe ningún otro flujo del sistema. Es la base de la gestión.

**Independent Test**: Un administrador crea un alumno con todos sus datos, lo busca por nombre o DNI, edita su sección y luego lo desactiva. Se verifica que cada paso se refleja en la lista y que el alumno desactivado no se elimina.

**Acceptance Scenarios**:

1. **Given** un administrador con sesión iniciada, **When** registra un alumno con DNI único y datos completos, **Then** el alumno aparece en la lista del grado y sección indicados.
2. **Given** un alumno ya registrado, **When** el administrador intenta crear otro con el mismo DNI o código, **Then** el sistema rechaza la operación e indica que el identificador ya existe, sin crear el duplicado.
3. **Given** un alumno registrado, **When** el administrador lo desactiva, **Then** el alumno deja de figurar como activo, pero sigue visible en la búsqueda con estado "desactivado" y conserva todo su historial.
4. **Given** un director con sesión iniciada, **When** consulta la ficha de un alumno, **Then** puede ver sus datos, pero no encuentra opciones para editarlos, desactivarlos ni generar carnés.

---

### User Story 2 - Generar QR único y reemplazarlo ante pérdida o robo (Priority: P1)

Cada alumno activo tiene exactamente un código QR vigente, generado por el sistema y asociado solo a un identificador o token. Si el carné se pierde o es robado, el administrador invalida el QR actual y emite uno nuevo. El QR anterior deja de funcionar de inmediato.

**Why this priority**: El QR es el medio de identificación en el ingreso. Su unicidad y la posibilidad de invalidarlo son requisitos de seguridad centrales.

**Independent Test**: Se genera el QR de un alumno, se reemplaza, y se verifica que el QR anterior queda marcado como invalidado y que solo el nuevo figura como activo.

**Acceptance Scenarios**:

1. **Given** un alumno activo sin QR, **When** el administrador genera su QR, **Then** el alumno tiene exactamente un QR activo y su contenido no incluye nombre, DNI ni otros datos personales.
2. **Given** un alumno con un QR activo, **When** el administrador lo reemplaza indicando el motivo (pérdida o robo), **Then** el QR anterior queda invalidado de inmediato, el nuevo queda activo y el cambio queda registrado.
3. **Given** un QR invalidado, **When** se presenta en cualquier punto de lectura del sistema, **Then** el sistema lo rechaza como no válido.
4. **Given** un alumno con asistencias registradas, **When** se reemplaza su QR, **Then** todas sus asistencias anteriores siguen asociadas al alumno y no se modifican.

---

### User Story 3 - Generar carnés en lote para imprimir (Priority: P2)

El administrador selecciona un grado o una sección y genera los carnés de todos los alumnos activos de esa selección. Cada carné muestra foto, nombre completo, grado, sección y el código QR. El resultado se puede imprimir directamente desde el navegador. Los alumnos sin foto aparecen con un marcador visible de "sin foto" y se listan en un aviso previo a la impresión.

**Why this priority**: Es la salida física principal del sistema, pero depende de tener alumnos y QR ya registrados.

**Independent Test**: Se selecciona una sección con alumnos con y sin foto, se genera el lote y se verifica que cada carné contiene los datos esperados y que los alumnos sin foto aparecen señalados.

**Acceptance Scenarios**:

1. **Given** una sección con alumnos activos que tienen foto y QR vigente, **When** el administrador genera los carnés de esa sección, **Then** se muestra un carné por alumno con foto, nombre completo, grado, sección y QR.
2. **Given** una selección que incluye un alumno sin foto, **When** el administrador genera el lote, **Then** el carné de ese alumno muestra el marcador "sin foto" y el sistema avisa antes de imprimir cuántos alumnos no tienen foto y cuáles son.
3. **Given** un alumno activo sin QR vigente dentro de la selección, **When** se genera el lote, **Then** el sistema avisa de ese alumno y no imprime un carné con QR vacío.
4. **Given** un lote generado, **When** el administrador lo imprime, **Then** el diseño respeta el tamaño de carné definido y los nombres con tildes y ñ se muestran completos.

---

### User Story 4 - Importar alumnos desde Excel o CSV (Priority: P2)

El administrador sube un archivo Excel o CSV con alumnos. El sistema crea los alumnos válidos y produce un reporte que indica cada fila rechazada, con su número de fila y el motivo (campo obligatorio vacío, identificador repetido dentro del archivo, identificador ya existente, grado o sección inexistentes, formato inválido). Las filas inválidas o duplicadas no detienen la importación del resto.

**Why this priority**: Evita registrar alumno por alumno al inicio del año escolar, pero el sistema es usable sin ella.

**Independent Test**: Se importa un archivo con filas válidas, una con DNI repetido dentro del archivo, una con DNI ya existente y una con campo obligatorio vacío. Se verifica que solo las válidas se crean y que el reporte describe las tres rechazadas.

**Acceptance Scenarios**:

1. **Given** un archivo con 50 filas válidas, **When** se importa, **Then** los 50 alumnos se crean y el reporte indica 50 creados y 0 rechazados.
2. **Given** un archivo con dos filas que tienen el mismo DNI, **When** se importa, **Then** ambas filas se rechazan como duplicadas dentro del archivo, se reportan con sus números de fila, y el resto de filas válidas se importa.
3. **Given** un archivo con una fila cuyo DNI ya existe en el sistema, **When** se importa, **Then** esa fila no sobrescribe al alumno existente, se reporta como "ya existe" y el resto se procesa normalmente.
4. **Given** un archivo con una fila sin nombre, **When** se importa, **Then** esa fila se reporta con el motivo "campo obligatorio vacío" y las demás se procesan.
5. **Given** un archivo con nombres que contienen tildes, ñ o apóstrofes, **When** se importa, **Then** los nombres se guardan y se muestran sin alteraciones.

---

### User Story 5 - Retirar, trasladar, cambiar de grado o sección (Priority: P2)

El administrador marca a un alumno como retirado o trasladado, o lo cambia de grado o sección. En todos los casos el historial de asistencia del alumno se conserva, y se deja constancia del cambio de grado o sección y de la fecha en que ocurrió.

**Why this priority**: Es necesario para mantener datos correctos durante el año escolar, pero no bloquea el ingreso diario.

**Independent Test**: Se cambia un alumno de sección, luego se marca como trasladado; se verifica que su historial permanece y que el alumno ya no aparece en los listados activos de su sección anterior.

**Acceptance Scenarios**:

1. **Given** un alumno activo en "3.° A", **When** el administrador lo cambia a "3.° B", **Then** el alumno figura en "3.° B" desde esa fecha, y su historial previo se conserva con la sección anterior.
2. **Given** un alumno activo, **When** el administrador lo marca como retirado o trasladado indicando la fecha, **Then** el alumno deja de figurar como activo, su QR queda invalidado y su historial se conserva.
3. **Given** un alumno retirado o trasladado, **When** el administrador consulta su ficha, **Then** ve su estado, la fecha y el tipo de salida (retirado o trasladado).

---

### Edge Cases

- ¿Qué ocurre si el DNI o código contiene espacios al inicio o al final, o ceros a la izquierda? Se eliminan espacios sobrantes; los ceros a la izquierda se conservan como parte del identificador, y dos identificadores que solo difieren en espacios se consideran el mismo.
- ¿Qué ocurre si un DNI o código de un alumno desactivado se intenta registrar de nuevo? El sistema lo rechaza indicando que pertenece a un alumno existente (desactivado) y ofrece reactivarlo en lugar de crear uno nuevo.
- ¿Qué ocurre si el archivo importado no tiene las columnas obligatorias? La importación completa se rechaza y el reporte indica las columnas faltantes; no se crea ningún alumno.
- ¿Qué ocurre si la foto no tiene un formato permitido o pesa demasiado? La foto no se acepta, el alumno queda sin foto y se informa el motivo.
- ¿Qué ocurre si dos administradores reemplazan el QR del mismo alumno al mismo tiempo? Solo queda activo un QR; la segunda operación se rechaza o reemplaza el QR resultante, pero nunca quedan dos activos.
- ¿Qué ocurre si se desactiva un alumno que tiene QR vigente? Su QR queda invalidado de inmediato como parte de la desactivación.
- ¿Qué ocurre si un grado o sección se usa en una importación pero no existe en el año escolar activo? La fila se rechaza con el motivo "grado o sección inexistente".
- ¿Qué ocurre si se intenta generar carnés para una sección vacía? El sistema informa que no hay alumnos activos y no genera archivo.
- ¿Qué ocurre con nombres en mayúsculas, minúsculas o con tildes al buscar? La búsqueda no distingue mayúsculas ni tildes (por ejemplo, "Nuñez" y "NUNEZ" encuentran al mismo alumno).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: El sistema MUST permitir registrar un alumno con DNI o código, nombres, apellidos, año escolar, grado, sección, foto y datos de referencia del apoderado.
- **FR-002**: El sistema MUST exigir DNI o código, nombres, apellidos, grado y sección para registrar un alumno. La foto y los datos del apoderado son opcionales.
- **FR-003**: El sistema MUST garantizar que el DNI o código de cada alumno sea único en todo el sistema, sin importar el año escolar, grado o sección.
- **FR-004**: El sistema MUST permitir editar los datos de un alumno, excepto que el DNI o código no puede cambiarse una vez creado sin una corrección auditada.
- **FR-005**: El sistema MUST permitir desactivar a un alumno sin eliminar su registro ni su historial.
- **FR-006**: El sistema MUST generar un QR único por alumno activo, cuyo contenido sea únicamente un identificador o token opaco.
- **FR-007**: El sistema MUST NOT incluir nombre, DNI, código, grado ni otros datos personales dentro del contenido del QR.
- **FR-008**: Cada alumno MUST tener como máximo un QR activo en un momento dado.
- **FR-009**: El sistema MUST permitir reemplazar el QR de un alumno indicando el motivo (pérdida, robo u otro).
- **FR-010**: Al reemplazar un QR, el QR anterior MUST quedar invalidado de inmediato, y todo intento de uso posterior MUST ser rechazado como no válido.
- **FR-011**: El reemplazo de un QR MUST NOT modificar ni eliminar las asistencias registradas del alumno.
- **FR-012**: El sistema MUST permitir importar alumnos desde archivos Excel (.xlsx) y CSV con las mismas columnas que el registro manual.
- **FR-013**: La importación MUST procesar cada fila de forma independiente: las filas válidas se crean y las inválidas se rechazan sin detener la importación.
- **FR-014**: La importación MUST generar un reporte que liste cada fila rechazada con su número de fila, el identificador y el motivo del rechazo.
- **FR-015**: La importación MUST rechazar como duplicadas todas las filas cuyo DNI o código aparezca más de una vez dentro del mismo archivo, reportándolas como tales.
- **FR-016**: La importación MUST rechazar cualquier fila cuyo DNI o código ya exista en el sistema, sin modificar al alumno existente.
- **FR-017**: El sistema MUST permitir generar carnés en lote para un grado completo o para una sección específica de un grado.
- **FR-018**: Cada carné MUST mostrar foto, nombre completo, grado, sección y código QR vigente del alumno.
- **FR-019**: Antes de imprimir un lote, el sistema MUST avisar de los alumnos sin foto y de los alumnos sin QR vigente que están incluidos en la selección.
- **FR-020**: Los alumnos sin foto MUST aparecer en el carné con un marcador visible que indique "sin foto", sin impedir la impresión del resto del lote.
- **FR-021**: El sistema MUST permitir imprimir los carnés directamente desde el navegador, con un diseño que respete un tamaño de carné definido por configuración.
- **FR-022**: El sistema MUST permitir marcar a un alumno como retirado o trasladado, registrando la fecha del cambio y el tipo de salida.
- **FR-023**: Al marcar a un alumno como retirado o trasladado, el sistema MUST invalidar su QR vigente.
- **FR-024**: El sistema MUST permitir cambiar a un alumno de grado o sección, registrando la fecha del cambio y conservando la sección anterior en su historial.
- **FR-025**: Ningún cambio de grado, sección, estado o QR MUST eliminar el historial de asistencia del alumno.
- **FR-026**: Todo cambio sobre un registro de alumno o de QR MUST quedar auditado con el usuario responsable, la fecha y hora, y el valor anterior y nuevo.
- **FR-027**: El rol Administrador MUST poder crear, editar, desactivar, importar, generar carnés, reemplazar QR y cambiar estados de alumnos.
- **FR-028**: El rol Director MUST poder consultar alumnos, carnés ya generados y reportes de importación, sin poder modificar ningún dato.
- **FR-029**: El sistema MUST guardar y mostrar nombres y apellidos con tildes, ñ, diéresis y apóstrofes sin alteración.
- **FR-030**: Las búsquedas de alumnos MUST NOT distinguir entre mayúsculas y minúsculas ni entre letras con y sin tilde.
- **FR-031**: Los reportes exportados MUST estar codificados en UTF-8 de forma que los caracteres especiales se vean correctamente al abrirlos en Excel.
- **FR-032**: Al desactivar a un alumno, el sistema MUST invalidar su QR vigente de inmediato.

### Key Entities *(include if feature involves data)*

- **Alumno**: Persona registrada en el colegio. Atributos: identificador único (DNI o código), nombres, apellidos, año escolar, grado, sección, estado (activo, desactivado, retirado, trasladado), foto y datos de referencia del apoderado. Se relaciona con sus QR emitidos y con sus cambios de grado o sección.
- **Año escolar**: Periodo lectivo al que pertenecen grados y secciones. Agrupa a los alumnos matriculados.
- **Grado y sección**: Organización del alumnado dentro de un año escolar. Un alumno pertenece a una sola sección a la vez.
- **Código QR**: Identificador o token opaco asociado a un alumno. Atributos: estado (activo, invalidado), motivo de invalidación, fecha de emisión y fecha de invalidación. Un alumno tiene como máximo un QR activo.
- **Historial de sección**: Registro de cada grado y sección que ha tenido un alumno, con fecha de inicio y fin.
- **Importación**: Operación de carga masiva. Atributos: archivo de origen, fecha, usuario, cantidad de filas creadas y rechazadas, y detalle de cada rechazo.
- **Registro de auditoría**: Entrada inmutable de cambio sobre un alumno o QR. Atributos: usuario, fecha y hora, entidad afectada, valor anterior y valor nuevo.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: El administrador registra un alumno nuevo con todos sus datos en menos de 1 minuto.
- **SC-002**: Una importación de 2,000 filas termina en menos de 2 minutos y el reporte lista el 100% de las filas rechazadas con su número de fila y motivo.
- **SC-003**: El 100% de los QR invalidados es rechazado en el momento de su reemplazo, sin ventana de tiempo en la que siga funcionando.
- **SC-004**: El administrador genera e imprime el lote de carnés de un grado completo (hasta 300 alumnos) en menos de 3 minutos.
- **SC-005**: El 100% de los cambios de grado, sección, estado o QR conserva el historial de asistencia previo intacto, verificado en las pruebas de aceptación.
- **SC-006**: El 100% de los nombres con tildes, ñ o apóstrofes se muestran, buscan e imprimen sin alteraciones en las pruebas de aceptación.
- **SC-007**: Inspeccionando cualquier QR generado, no se encuentra ningún dato personal dentro de su contenido.
- **SC-008**: El director puede consultar la información de alumnos sin encontrar en la interfaz ninguna acción de modificación.

## Assumptions

- Cada alumno tiene un DNI o un código institucional; si no tiene DNI (por ejemplo, un alumno extranjero o menor sin documento), el colegio usa un código. Ambos son válidos, y el campo obligatorio es uno de los dos.
- Las importaciones se hacen sobre el año escolar activo; los alumnos de años anteriores no se importan de nuevo.
- Los grados y secciones existen previamente en el sistema y son administrados por el colegio; esta funcionalidad solo los asigna a los alumnos.
- Los formatos de foto aceptados son JPG y PNG, con un tamaño máximo razonable definido por configuración.
- Los carnés se imprimen desde el navegador en impresora común; la impresión de carnés plastificados con equipo especializado está fuera de alcance.
- La importación con filas duplicadas rechaza todas las ocurrencias del identificador repetido, para evitar decidir arbitrariamente cuál es la correcta.
- Al desactivar, retirar o trasladar a un alumno, su QR se invalida automáticamente; si el alumno vuelve, se emite un QR nuevo.
- El cambio de grado o sección es una operación individual; no existe una promoción masiva de fin de año en esta funcionalidad.
- El registro de asistencia, la gestión de docentes, notas, pagos y cuentas para apoderados están fuera de alcance de esta especificación.
- Los datos de alumnos se tratan conforme a los principios de la constitución del proyecto (mínimo privilegio, datos necesarios y Ley 29733), y la validación legal sigue pendiente según lo registrado allí.
