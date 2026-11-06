# Glosario

Este glosario fija un único término por concepto. El código, las tablas, las rutas y la
documentación deben usar estos términos (constitución, principio IX).

| Concepto en inglés | Término del proyecto | Uso |
|---|---|---|
| handler | manejador | Capa HTTP de cada módulo |
| service | servicio | Reglas de negocio |
| repository | repositorio | Acceso a datos |
| middleware | intermediario | Sesión, CSRF, roles, límite de intentos |
| student | alumno | Persona registrada en el colegio |
| section | seccion | División de un grado (A, B, C) |
| grade | grado | Nivel dentro del año escolar (1.°, 2.°, 3.°) |
| attendance record | marcacion | Registro de ingreso de un alumno |
| on time | puntual | Ingreso dentro de la tolerancia |
| late | tardanza | Ingreso después de la tolerancia y antes del límite |
| late, severe | tardanza_grave | Ingreso después del límite de tardanza |
| absent | falta | Sin ingreso al cierre, sin justificación |
| justified absence | falta_justificada | Falta con justificación aprobada |
| shift | turno | Mañana o tarde (valores: manana, tarde) |
| validity date | vigencia | Fecha desde la que aplica un cambio de configuración (`vigente_desde`) |
| section shift | turno_seccion | Turno de una sección con fecha de vigencia |
| class block | bloque_clase | Hora de clase de un turno con hora de inicio y de fin |
| section schedule | horario_seccion | Curso asignado a una sección, día de la semana y bloque |
| calendar exception | excepcion_calendario | Feriado, día no laborable o suspensión de clases |
| threshold | umbral | Número de tardanzas o faltas por mes que activa una alerta |
| course | curso | Asignatura del colegio vinculada a uno o más grados |
| tolerance | tolerancia | Minutos de gracia después de la hora de ingreso |
| cutoff | hora_cierre | Hora en que se cierra el registro del turno |
| alert | alerta | Aviso interno a un usuario |
| audit | auditoria | Registro inmutable de cambios |
| role | rol | administrador, director, docente, porteria |
| school year | anio_escolar | Periodo lectivo |
| holiday | feriado | Día sin clases para todo el colegio |
| QR token | token_qr | Token aleatorio contenido en el QR |
| identifier | identificador | DNI o código institucional del alumno |
| attendance evaluation | asistencia | Conjunto de marcaciones y estados de un alumno |

## Roles

- **administrador**: gestiona alumnos, carnés, configuración y alertas.
- **director**: consulta todo; aprueba justificaciones; no modifica datos.
- **docente**: registra asistencia de sus cursos y secciones.
- **porteria**: opera el escaneo de la puerta principal.

## Estados de alumno

`activo`, `desactivado`, `retirado`, `trasladado`.

## Estados de marcación

`puntual`, `tardanza`, `tardanza_grave`, `falta`, `falta_justificada`.
