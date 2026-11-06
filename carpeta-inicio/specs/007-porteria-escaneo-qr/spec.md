# Feature Specification: Registro de ingreso por QR en portería

**Feature Branch**: `007-porteria-escaneo-qr`

**Created**: 2026-11-06

**Status**: Draft

**Input**: User description: "Registro oficial de ingreso por QR desde portería, con clasificación de puntualidad y generación de alertas"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Registrar el ingreso de un alumno por QR (Priority: P1)

El portero abre la pantalla de portería, escanea el QR del carné o pega su contenido y el sistema
registra la marcación oficial del día con el estado que corresponde al horario del turno. El
registro guarda quién lo realizó para que la alerta de tardanza llegue a esa cuenta.

**Why this priority**: Es el origen de la asistencia oficial y de las alertas de tardanza. Sin
esta pantalla, las marcaciones solo pueden crearse por procesos técnicos.

**Independent Test**: Con un turno configurado y un alumno con QR activo, escanear su token y
verificar que se crea una marcación con estado `puntual`, `tardanza` o `tardanza_grave` según la
hora, y que `registrada_por` es el usuario de portería.

**Acceptance Scenarios**:

1. **Given** un turno con ingreso 07:30 y tolerancia 10 minutos, **When** el portero escanea a las
   07:35, **Then** se registra una marcación `puntual`.
2. **Given** el mismo turno con límite de tardanza 08:00, **When** el portero escanea a las 07:45,
   **Then** se registra una marcación `tardanza`.
3. **Given** el mismo turno, **When** el portero escanea a las 08:15, **Then** se registra una
   marcación `tardanza_grave`.
4. **Given** una marcación de tardanza recién creada, **When** el portero abre su panel de alertas,
   **Then** ve la alerta de tardanza de ese alumno.

### User Story 2 - Rechazar ingresos que no corresponden (Priority: P1)

El sistema rechaza el escaneo cuando el token no existe, la credencial fue revocada, el alumno está
inactivo, el día no es laborable para su sección, el horario del turno no está configurado o el
alumno ya tiene marcación ese día.

**Why this priority**: Evita duplicados y datos oficiales incorrectos.

**Independent Test**: Escanear un token revocado, un token inexistente y un alumno ya marcado;
verificar que no se crea ninguna marcación y que se muestra el motivo.

**Acceptance Scenarios**:

1. **Given** un token inexistente, **When** el portero lo envía, **Then** el sistema responde con
   "El código QR no es válido." y no registra nada.
2. **Given** un alumno con marcación del día, **When** se vuelve a escanear, **Then** se informa
   que el ingreso ya fue registrado y no se duplica.
3. **Given** un día feriado para la sección, **When** se escanea, **Then** se rechaza el registro.

### User Story 3 - Consultar el resultado del escaneo (Priority: P2)

Después de cada escaneo, el portero ve el nombre del alumno, su sección, la hora y el estado
registrado, o el motivo del rechazo, y puede escanear al siguiente alumno sin recargar la pantalla.

**Why this priority**: Agiliza la fila de ingreso y reduce errores.

**Independent Test**: Escanear dos alumnos seguidos y verificar que el resultado se actualiza y que
el campo de escaneo vuelve a quedar listo.

**Acceptance Scenarios**:

1. **Given** un escaneo exitoso, **When** termina la petición, **Then** se muestra el resumen y el
   campo queda enfocado para el siguiente alumno.
2. **Given** un escaneo rechazado, **When** termina la petición, **Then** se muestra el motivo y se
   permite reintentar.

### Edge Cases

- Escaneo antes de la hora de ingreso: se rechaza con "Todavía no comienza el registro del turno."
- Escaneo después de la hora de cierre: se rechaza con "El registro del turno ya cerró."
- Token con espacios o saltos de línea: se recorta antes de buscar.
- Alumno sin sección vigente: se rechaza.
- Sección sin turno vigente: se rechaza.
- Horario del turno no configurado para la fecha: se rechaza.
- Varios escaneos simultáneos del mismo alumno: solo uno debe registrar la marcación.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: El sistema MUST ofrecer una pantalla de portería para escanear el QR o pegar su
  contenido.
- **FR-002**: El sistema MUST resolver el token contra las credenciales QR activas y rechazar
  tokens inexistentes, revocados o de alumnos inactivos.
- **FR-003**: El sistema MUST validar que el día sea laborable para la sección del alumno.
- **FR-004**: El sistema MUST obtener el turno vigente de la sección y el horario vigente del turno
  en la fecha.
- **FR-005**: El sistema MUST clasificar la marcación como `puntual`, `tardanza` o
  `tardanza_grave` según la tolerancia y el límite de tardanza del turno.
- **FR-006**: El sistema MUST rechazar escaneos antes de la hora de ingreso y después de la hora de
  cierre del turno.
- **FR-007**: El sistema MUST evitar más de una marcación por alumno y fecha, incluso con escaneos
  simultáneos.
- **FR-008**: El sistema MUST guardar en `registrada_por` el usuario que realizó el escaneo.
- **FR-009**: El sistema MUST auditar cada marcación registrada desde portería.
- **FR-010**: Después de una tardanza, el sistema MUST disparar la generación de alertas de
  tardanza sin esperar al proceso periódico, sin bloquear el registro si el proceso falla.
- **FR-011**: La pantalla MUST mostrar el resultado del escaneo o el motivo del rechazo.
- **FR-012**: Solo los roles `porteria` y `administrador` MUST poder registrar ingresos. El
  director y los docentes no acceden a portería.
- **FR-013**: Todo POST MUST exigir token CSRF.

### Key Entities

- **Marcación oficial**: alumno, fecha, hora, estado, usuario que registró.
- **Credencial QR**: token activo asociado a un alumno.
- **Horario de turno**: ingreso, tolerancia, límite de tardanza y cierre vigentes en la fecha.
- **Resultado de escaneo**: éxito con resumen o rechazo con motivo.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: El 100% de los escaneos válidos dentro de la ventana del turno crea una marcación.
- **SC-002**: El 100% de los escaneos duplicados se rechaza sin crear otra marcación.
- **SC-003**: El 100% de las tardanzas genera una alerta visible para el portero en menos de 5
  segundos.
- **SC-004**: Un escaneo válido se resuelve en menos de 2 segundos en condiciones normales.
- **SC-005**: El 100% de los rechazos muestra un motivo comprensible para el portero.

## Assumptions

- La cámara del navegador puede leer QR mediante `BarcodeDetector`; si no está disponible, el
  portero puede pegar el token manualmente.
- El carné contiene únicamente el token del QR, sin datos personales.
- La asistencia oficial es solo del alumno; no se registran docentes ni apoderados.
