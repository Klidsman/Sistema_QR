# Contrato de rutas HTTP: Registro de asistencia del docente

**Fecha**: 2026-11-06 | **Plan**: [plan.md](../plan.md)

Las rutas están en español, sin tildes ni espacios. Las respuestas son HTML renderizado en el
servidor; HTMX pide fragmentos cuando el encabezado `HX-Request` está presente.

## Permisos

- `docente`: solo sus cursos y secciones asignados en la fecha. No accede a configuración ni a
  pantallas de administración.
- `administrador`: todas las rutas de esta feature.
- `director`: solo rutas `GET` de consulta y exportación. Cualquier escritura responde `403`.
- `porteria`: sin acceso a esta feature.
- Sin sesión: redirección a `/ingreso`.
- Toda ruta que cambia estado exige token CSRF.

## Docente

| Método | Ruta | Descripción | Permiso |
|---|---|---|---|
| GET | /asistencia | Selección de fecha y lista de clases asignadas ese día | Docente, administrador |
| GET | /asistencia/clase | Lista de alumnos de `?fecha=&seccion_id=&curso_id=&bloque=` | Docente asignado, administrador |
| POST | /asistencia/clase | Guarda estados y observaciones de los alumnos enviados | Docente asignado, administrador |
| POST | /asistencia/clase/todos | Marca como presentes solo los alumnos sin registro | Docente asignado, administrador |
| GET | /asistencia/exportar | Formulario de rango de fechas | Docente, administrador, director |
| GET | /asistencia/exportar/archivo | Descarga CSV de `?desde=&hasta=&curso_id=&seccion_id=` | Docente, administrador, director |

Reglas de la lista:
- El docente solo ve clases cuya asignación está vigente en la fecha.
- La fecha no puede ser futura.
- Un día no laborable o una excepción de la sección responde `422` con el motivo.
- La clase debe existir en el horario semanal con ese curso y bloque.

## Corrección del administrador

| Método | Ruta | Descripción | Permiso |
|---|---|---|---|
| GET | /asistencia/consulta | Consulta de registros por fecha y sección | Administrador, director |
| POST | /asistencia/consulta/correccion | Corrige un registro con motivo obligatorio | Administrador |

La corrección del administrador se permite siempre dentro del año escolar y exige `motivo` de 5 a
300 caracteres. Queda auditada con el valor anterior y el nuevo.

## Asignaciones de docente

| Método | Ruta | Descripción | Permiso |
|---|---|---|---|
| GET | /asistencia/asignaciones | Listado de asignaciones y formulario | Administrador, director |
| POST | /asistencia/asignaciones | Crea una asignación con docente, curso, sección y fechas | Administrador |
| POST | /asistencia/asignaciones/{id}/cierre | Cierra una asignación en una fecha | Administrador |

## Cuentas de docente

| Método | Ruta | Descripción | Permiso |
|---|---|---|---|
| GET | /docentes | Listado de docentes y formulario | Administrador, director |
| POST | /docentes | Crea una cuenta con nombre, correo y contraseña | Administrador |
| POST | /docentes/{id}/clave | Restablece la contraseña | Administrador |
| POST | /docentes/{id}/desactivacion | Desactiva la cuenta sin borrarla | Administrador |

## Parámetro del plazo de edición

| Método | Ruta | Descripción | Permiso |
|---|---|---|---|
| GET | /configuracion/asistencia | Plazo actual e historial de cambios | Administrador, director |
| POST | /configuracion/asistencia | Configura el plazo entre 1 y 168 horas | Administrador |

## Respuestas de error relevantes

| Situación | Código | Mensaje |
|---|---|---|
| Docente sin asignación vigente | 403 | "No tiene acceso a esta clase en la fecha indicada." |
| Fecha futura | 422 | "No se puede registrar asistencia de una fecha futura." |
| Día no laborable o suspensión | 422 | "La fecha no es laborable para esta sección." |
| Clase inexistente en el horario | 422 | "La clase no corresponde al horario de la sección." |
| Bloque inexistente en el turno | 422 | "El bloque indicado no existe en el turno de la sección." |
| Registro fuera del plazo del docente | 403 | "El plazo de edición terminó. Solo un administrador puede corregir este registro." |
| Corrección del administrador sin motivo | 422 | "Indique el motivo de la corrección." |
| Observación mayor a 300 caracteres | 422 | "La observación no puede superar los 300 caracteres." |
| Estado desconocido | 422 | "El estado de asistencia no es válido." |
| Alumno fuera de la sección en la fecha | 422 | "El alumno no pertenece a la sección en la fecha indicada." |
| Exportación sin permisos sobre el curso | 403 | "No tiene acceso a este curso o sección." |
