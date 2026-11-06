# Manual del director

El director usa el sistema en modo consulta. Puede revisar la información de alumnos, sus códigos
QR, los carnés y el historial de importaciones, pero no puede modificar datos.

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

## Permisos

El sistema aplica dos reglas:

1. El director solo puede usar métodos de lectura (`GET`). Cualquier intento de escritura responde
   `403`.
2. El director no puede acceder a las pantallas exclusivas del administrador, como el alta o la
   importación.

Si un enlace no aparece en la interfaz, tampoco se puede ejecutar escribiendo la dirección
directamente.

## Recomendaciones

- Usar la búsqueda para localizar alumnos por apellido o DNI.
- Revisar el historial de secciones cuando un alumno cambie de grado.
- Consultar el reporte de importación para conocer las filas que requieren corrección.
- Cerrar la sesión al terminar, especialmente en equipos compartidos.
