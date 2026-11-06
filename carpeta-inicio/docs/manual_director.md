# Manual del director

El director usa el sistema en modo consulta. Puede revisar la información de alumnos, sus códigos
QR, los carnés y el historial de importaciones, pero no puede modificar datos de alumnos. La única
excepción son los ajustes de horarios de turno y de horas de bloques existentes, descritos en la
sección de configuración académica.

## Ingreso

1. Abrir la dirección del sistema y entrar a `/ingreso`.
2. Escribir el correo y la contraseña asignados.
3. Presionar **Ingresar**.

La sesión dura 8 horas. Si la contraseña es incorrecta cinco veces desde la misma conexión, el
ingreso se bloquea durante 15 minutos.

## Listado de alumnos

En `/alumnos` el director ve el mismo listado que el administrador:

- **Buscar** por nombres, apellidos o DNI, sin distinguir tildes ni mayúsculas.
- **Filtrar** por sección y estado.
- **Ver** la ficha de cada alumno.

El director no ve los botones **Nuevo alumno**, **Importar alumnos**, **Editar** ni **Desactivar**.

## Ficha del alumno

La ficha muestra:

- Datos personales y de contacto del apoderado.
- Foto del alumno.
- Grado, sección, turno y año escolar.
- Estado actual.
- Código QR vigente y su historial.
- Historial de secciones, con fechas de inicio y fin de cada tramo.
- Fecha de salida si el alumno fue retirado o trasladado.

El director puede consultar esta información, pero no puede generar, reemplazar ni revocar QR.

## Carnés

1. Entrar a `/carnets`.
2. Elegir una sección o un grado.
3. Presionar **Generar carnés**.
4. Revisar los avisos de alumnos sin foto o sin QR vigente.
5. Presionar **Imprimir** si necesita una copia.

El director puede generar la vista imprimible de carnés, pero no puede modificar los datos de los
alumnos ni emitir QR desde esta pantalla.

## Historial de importaciones

En `/alumnos/importaciones` el director puede revisar las cargas masivas realizadas por el
administrador:

- Nombre del archivo.
- Fecha de la importación.
- Filas leídas, creadas y rechazadas.
- Detalle de cada fila rechazada con su número de fila y motivo.

El director no puede iniciar una importación nueva.

## Configuración académica

En `/configuracion` el director consulta toda la configuración del año escolar: año, grados,
secciones, turnos, calendario, umbrales, cursos, bloques, horario semanal y consulta histórica.

Además de consultar, el director puede:

- Ajustar la **hora de ingreso**, la **tolerancia**, la **hora límite de tardanza** y la **hora de
  cierre** de cada turno en `/configuracion/turnos`.
- Ajustar la **hora de inicio y de fin** de bloques de clase existentes en
  `/configuracion/bloques`.

El director no puede crear ni cerrar años, ni crear grados, secciones, feriados, umbrales, cursos
o bloques. Tampoco puede armar horarios semanales ni desactivar cursos.

## Asistencia del docente

En `/asistencia/consulta` el director consulta los registros del docente por fecha y sección:
alumnos, curso, bloque, estado, hora de ingreso QR, docente y observaciones. Los alumnos con
diferencias entre el QR y el registro aparecen resaltados como información.

El director no puede corregir registros ni cambiar el plazo de edición. También puede descargar
la exportación CSV por sección desde `/asistencia/exportar`.

## Permisos

El sistema aplica dos reglas:

1. El director puede usar métodos de lectura (`GET`) en todas las pantallas.
2. Las únicas escrituras permitidas son los ajustes de horarios de turno y de horas de bloques
   existentes. Cualquier otro intento de escritura responde `403`.
3. El director no puede acceder a las pantallas exclusivas del administrador, como el alta o la
   importación.

Si un enlace no aparece en la interfaz, tampoco se puede ejecutar escribiendo la dirección
directamente.

## Reportes

El director accede a todos los reportes del colegio: diario, panel, resumen y reiteración, con
exportación a Excel y PDF. Los reportes son de solo lectura; no modifican la asistencia.

## Justificaciones

El director puede aprobar o rechazar justificaciones pendientes desde `/justificaciones`. El
rechazo exige un comentario. No puede anular justificaciones ni resolver las que él mismo registró;
la anulación corresponde al administrador.

## Alertas

El director no recibe ni accede al panel de alertas en esta versión. Puede consultar las faltas y
tardanzas en los reportes de asistencia; las alertas internas están reservadas al administrador,
los docentes y portería.

## Recomendaciones

- Usar la búsqueda para localizar alumnos por apellido o DNI.
- Revisar el historial de secciones cuando un alumno cambie de grado.
- Consultar el reporte de importación para conocer las filas que requieren corrección.
- Cerrar la sesión al terminar, especialmente en equipos compartidos.
