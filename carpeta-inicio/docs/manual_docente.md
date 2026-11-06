# Manual del docente

Este manual describe el registro personal de asistencia de tus cursos y secciones. El ingreso por
QR en portería sigue siendo la asistencia oficial del colegio; el registro del docente es una
herramienta de seguimiento y exportación.

## Ingreso

1. Abrir la dirección del sistema y entrar a `/ingreso`.
2. Escribir el correo y la contraseña entregados por el administrador.
3. Presionar **Ingresar**.

La sesión dura 8 horas. Si la contraseña es incorrecta cinco veces desde la misma conexión, el
ingreso se bloquea durante 15 minutos.

## Pantalla de asistencia

1. Abrir `/asistencia`.
2. Elegir la fecha y pulsar **Ver clases**.
3. Verás únicamente las clases de los cursos y secciones que tienes asignados ese día.
4. Abrir una clase.

Si no aparece ninguna clase, revisa con el administrador que:
- Tengas una asignación vigente para el curso y la sección.
- La sección tenga turno y bloques configurados.
- El curso esté en el horario semanal de ese día.
- La fecha no sea un feriado o suspensión.

## Registrar la asistencia

En la lista de la clase verás, por cada alumno:
- Nombre e identificador.
- **Ingreso QR**: hora si el alumno entró por portería, o “Sin ingreso por QR”.
- **Estado del docente**: presente, tardanza, falta o falta justificada.
- **Observación**: texto opcional de hasta 300 caracteres.

Pasos:

1. Usar **Marcar a todos presentes** para crear los registros faltantes de una sola vez.
2. Corregir solo las excepciones.
3. Presionar **Guardar asistencia**.

Los alumnos con diferencia entre el QR y tu registro aparecen resaltados con un aviso. Es solo
información: el QR no se modifica.

## Plazo de edición

Puedes corregir tus registros mientras no venza el plazo configurado por el colegio (24 horas por
defecto, entre 1 y 168 horas). Pasado el plazo, el sistema bloquea la edición y solo el
administrador puede corregir, indicando un motivo.

## Registro por hora de clase

Cada bloque del día tiene su propio registro. Registrar el bloque 1 no cambia el bloque 2, aunque
sea el mismo curso y la misma sección.

## Exportar

1. Abrir `/asistencia/exportar`.
2. Elegir el rango de fechas.
3. Presionar **Descargar CSV**.

El archivo incluye fecha, curso, sección, bloque, identificador, alumno, estado, observación, hora
QR y docente. Como docente solo se exportan tus propios registros.

## Reglas importantes

- No se puede registrar una fecha futura.
- Un día feriado, no laborable o suspendido para la sección no pide registro.
- Un alumno retirado o trasladado no aparece en fechas posteriores a su salida.
- Los registros anteriores se conservan aunque el alumno cambie de sección.
- El sistema no registra asistencia de docentes ni de auxiliares.
- No se envían notificaciones a apoderados.

## Recomendaciones

- Marcar a todos presentes y corregir solo las excepciones ahorra tiempo.
- Revisar las diferencias con el QR antes de guardar.
- Exportar al cierre de cada semana si el colegio lo solicita.
- Cerrar sesión al terminar, especialmente en equipos compartidos.
