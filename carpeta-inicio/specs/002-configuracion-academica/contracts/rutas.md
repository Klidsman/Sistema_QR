# Contrato de rutas HTTP: Configuración académica del colegio

**Fecha**: 2026-11-06 | **Plan**: [plan.md](../plan.md)

Las rutas están en español, sin tildes ni espacios (constitución, principio IX). Las respuestas
son HTML renderizado en el servidor; HTMX pide fragmentos cuando el encabezado `HX-Request` está
presente.

## Permisos

- `administrador`: todas las rutas de esta feature.
- `director`: todas las rutas `GET` y únicamente los `POST` de horarios de turno y de horas de
  bloques existentes. Cualquier otro `POST` responde `403`.
- Sin sesión: redirección a `/ingreso`.
- Todas las rutas que cambian estado exigen token CSRF.

## Panel y consulta

| Método | Ruta | Descripción | Permiso |
|---|---|---|---|
| GET | /configuracion | Panel del año activo: estado, grados, secciones, turnos y avisos | Administrador, director |
| GET | /configuracion/consulta | Configuración vigente en `?fecha=AAAA-MM-DD` | Administrador, director |

## Año escolar

| Método | Ruta | Descripción | Permiso |
|---|---|---|---|
| GET | /configuracion/anio | Formulario y datos del año activo | Administrador, director |
| POST | /configuracion/anio | Crea el año escolar (nombre, fecha de inicio y fin) | Administrador |
| POST | /configuracion/anio/cierre | Cierra el año activo y lo deja de solo lectura | Administrador |

## Grados

| Método | Ruta | Descripción | Permiso |
|---|---|---|---|
| GET | /configuracion/grados | Grados del año activo | Administrador, director |
| POST | /configuracion/grados | Crea un grado con nombre único dentro del año | Administrador |

## Secciones

| Método | Ruta | Descripción | Permiso |
|---|---|---|---|
| GET | /configuracion/secciones | Secciones agrupadas por grado, con su turno vigente | Administrador, director |
| POST | /configuracion/secciones | Crea una sección en un grado con turno inicial | Administrador |
| POST | /configuracion/secciones/{id}/turno | Cambia el turno desde una fecha de vigencia | Administrador |

## Turnos

| Método | Ruta | Descripción | Permiso |
|---|---|---|---|
| GET | /configuracion/turnos | Horarios vigentes y programados de mañana y tarde | Administrador, director |
| POST | /configuracion/turnos/{turno} | Configura ingreso, tolerancia, límite de tardanza y cierre desde una fecha | Administrador, director |

Valores válidos de `{turno}`: `manana` y `tarde`.

## Calendario

| Método | Ruta | Descripción | Permiso |
|---|---|---|---|
| GET | /configuracion/calendario | Excepciones del año, filtrables por mes y grado | Administrador, director |
| POST | /configuracion/calendario | Registra feriado, no laborable o suspensión (día o rango) | Administrador |
| POST | /configuracion/calendario | Igual que el anterior con `confirmar=si` cuando hay solapamiento | Administrador |

Si la nueva excepción se superpone con otra del mismo alcance, la primera respuesta `200` muestra
la confirmación con los mismos campos y sin guardar. El segundo envío con `confirmar=si` guarda.

## Umbrales de alerta

| Método | Ruta | Descripción | Permiso |
|---|---|---|---|
| GET | /configuracion/umbrales | Umbrales vigentes y su historia | Administrador, director |
| POST | /configuracion/umbrales | Define el umbral de tardanzas o faltas desde una fecha | Administrador |

## Cursos

| Método | Ruta | Descripción | Permiso |
|---|---|---|---|
| GET | /configuracion/cursos | Cursos del año con sus grados y estado | Administrador, director |
| POST | /configuracion/cursos | Crea un curso y lo vincula a uno o más grados | Administrador |
| POST | /configuracion/cursos/{id}/desactivacion | Desactiva el curso sin borrar su historial | Administrador |

## Bloques de hora de clase

| Método | Ruta | Descripción | Permiso |
|---|---|---|---|
| GET | /configuracion/bloques | Bloques vigentes del turno `?turno=manana` o `?turno=tarde` | Administrador, director |
| POST | /configuracion/bloques | Crea un bloque nuevo con número, horas y vigencia | Administrador |
| POST | /configuracion/bloques/{id}/horas | Ajusta la hora de inicio y de fin de un bloque existente | Administrador, director |

El ajuste de horas inserta una versión nueva del bloque con la misma `numero` y una
`vigente_desde` nueva; no modifica la fila anterior.

## Horario semanal de sección

| Método | Ruta | Descripción | Permiso |
|---|---|---|---|
| GET | /configuracion/horarios | Grilla semanal de la sección `?seccion_id=` | Administrador, director |
| POST | /configuracion/horarios | Guarda las celdas modificadas con su fecha de vigencia | Administrador |

El formulario envía las cinco jornadas y todos los bloques del turno vigente. Las celdas sin curso
se guardan explícitamente como "sin clase" cuando el usuario las vacía. Si una celda ya tenía un
curso, responde con confirmación y el segundo envío incluye `confirmar=si`.

## Respuestas de error relevantes

| Situación | Código | Mensaje |
|---|---|---|
| Fechas inválidas (fin anterior o igual a inicio) | 422 | "La fecha de fin debe ser posterior a la fecha de inicio." |
| Ya existe un año activo | 409 | "Ya existe un año escolar activo. Cierre el año actual antes de crear otro." |
| Año cerrado | 409 | "El año escolar está cerrado y no admite cambios." |
| Fecha de vigencia en el pasado | 422 | "La fecha de vigencia debe ser hoy o una fecha futura." |
| Cierre anterior o igual al límite de tardanza | 422 | "La hora de cierre debe ser posterior a la hora límite de tardanza." |
| Tolerancia que supera el límite | 422 | "La hora de ingreso más la tolerancia no puede superar la hora límite de tardanza." |
| Horarios de turnos superpuestos | 422 | "Los horarios de mañana y tarde se superponen." |
| Bloque fuera del horario del turno | 422 | "El bloque debe estar dentro del horario del turno." |
| Bloques solapados | 422 | "El bloque se superpone con otro bloque del mismo turno." |
| Curso de otro grado | 422 | "El curso no está vinculado al grado de la sección." |
| Solapamiento de excepción o reemplazo de celda | 200 | Pantalla de confirmación con el detalle del conflicto. |
| Acción de director no permitida | 403 | "No tiene permiso para esta acción." |

Todas las respuestas de error conservan los datos ingresados en el formulario, salvo las
contraseñas, que no existen en esta feature.
