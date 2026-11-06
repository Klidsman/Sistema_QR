# Manual del administrador

Este manual describe el uso diario del sistema de asistencia. El administrador puede registrar,
editar, importar y desactivar alumnos, gestionar sus códigos QR y generar carnés.

## Ingreso

1. Abrir la dirección del sistema y entrar a `/ingreso`.
2. Escribir el correo y la contraseña.
3. Presionar **Ingresar**.

Si hay cinco intentos fallidos desde la misma conexión, el sistema bloquea el ingreso durante
15 minutos. La sesión dura 8 horas.

## Listado de alumnos

En `/alumnos` se muestra el listado de alumnos activos, desactivados, retirados y trasladados.

- **Buscar**: acepta nombres, apellidos o DNI. No distingue tildes, mayúsculas ni la ñ.
- **Sección**: filtra por grado y sección del año escolar.
- **Estado**: filtra por activo, desactivado, retirado o trasladado.
- **Filtrar**: aplica los filtros.
- **Ver**: abre la ficha del alumno.
- **Editar**: modifica los datos del alumno.
- **Desactivar**: deja al alumno inactivo e invalida su QR. Pide confirmación.

Botones superiores:

- **Carnés en lote**: genera carnés imprimibles por sección o grado.
- **Historial de importaciones**: muestra las cargas masivas ya procesadas.
- **Importar alumnos**: abre el formulario de carga de CSV o XLSX.
- **Nuevo alumno**: abre el formulario de alta manual.

## Alta de un alumno

1. Entrar a `/alumnos/nuevo`.
2. Completar:
   - **DNI o código**: obligatorio y único.
   - **Grado y sección**: obligatorio.
   - **Nombres** y **Apellidos**: obligatorios.
   - **Nombre del apoderado** y **Teléfono del apoderado**: opcionales.
   - **Foto**: opcional, JPG o PNG de hasta 2 MB.
3. Presionar **Guardar**.

Si el identificador ya existe, el sistema muestra el error y no crea un duplicado. Si la foto no
es válida o pesa demasiado, el alumno se guarda sin foto y se informa el motivo.

## Edición

1. En el listado, presionar **Editar**.
2. Cambiar los datos permitidos. El DNI o código no se puede modificar.
3. Presionar **Guardar**.

Cada cambio queda auditado con el usuario y la fecha.

## Código QR

En la ficha del alumno, sección **Código QR**:

- **Generar QR**: crea una credencial nueva cuando el alumno todavía no tiene una vigente.
- **Reemplazar QR**: invalida el QR actual y genera otro. Se usa cuando el carné se pierde o se
  roba. El motivo es obligatorio.
- La imagen del QR solo contiene el token; no incluye el nombre, el DNI ni la sección.

Reglas importantes:

- Un alumno activo puede tener un solo QR vigente.
- Al desactivar, retirar o trasladar a un alumno, su QR se invalida de inmediato.
- El historial de QR muestra cuáles fueron revocados y por qué motivo.

## Importación desde CSV o XLSX

1. Entrar a `/alumnos/importar`.
2. Elegir un archivo `.csv` o `.xlsx` de hasta 5 MB.
3. Presionar **Importar**.

La primera fila debe tener los encabezados. Columnas obligatorias:

| Columna | Alias aceptados |
|---|---|
| `identificador` | `dni`, `codigo`, `documento` |
| `nombres` | `nombre` |
| `apellidos` | `apellido` |
| `grado` | — |
| `seccion` | — |

Columnas opcionales: `apoderado_nombre` y `apoderado_telefono`.

Ejemplo:

```csv
identificador,nombres,apellidos,grado,seccion,apoderado_nombre
70000001,José María,Núñez Ríos,3.°,A,María Ríos
70000002,Luis,Pérez,3.°,B,
```

Reglas:

- El CSV puede usar coma o punto y coma como separador y admite BOM.
- El XLSX lee la primera hoja.
- Las tildes y la ñ se conservan en nombres y apellidos.
- Las filas con problemas se rechazan individualmente; el resto se importa.
- El máximo de filas se configura con `IMPORTACION_MAX_FILAS` (5000 por defecto).
- Si faltan columnas obligatorias, no se crea ningún alumno y se listan las columnas faltantes.

Motivos de rechazo:

| Motivo | Significado |
|---|---|
| Campo obligatorio vacío | Falta identificador, nombres, apellidos, grado o sección. |
| Identificador repetido en el archivo | El mismo identificador aparece dos o más veces. |
| El identificador ya existe | Ya hay un alumno registrado con ese identificador. |
| Grado o sección inexistente | La sección no existe en el año escolar activo. |
| Formato inválido | El identificador contiene espacios u otro formato no permitido. |

Al terminar, el sistema redirige al **Reporte de importación**, que muestra filas leídas, alumnos
creados y filas rechazadas con su número de fila y motivo.

## Carnés en lote

1. Entrar a `/carnets`.
2. Elegir una **Sección** o un **Grado completo**.
3. Presionar **Generar carnés**.
4. Revisar los avisos:
   - Alumnos sin foto: se imprimen con el marcador "Sin foto".
   - Alumnos sin QR vigente: no se imprimen y se listan en el aviso.
5. Presionar **Imprimir**.

Solo se incluyen alumnos activos. El tamaño del carné se controla con `CARNET_ANCHO_MM` y
`CARNET_ALTO_MM`. La vista está preparada para imprimir en A4.

## Cambio de sección

1. En la ficha del alumno, buscar **Cambiar de sección**.
2. Elegir la nueva sección y la fecha.
3. Presionar **Cambiar sección**.

El tramo anterior se cierra y se abre uno nuevo. El historial conserva todas las secciones por las
que pasó el alumno.

## Retiro o traslado

1. En la ficha, buscar **Registrar salida**.
2. Elegir **retirado** o **trasladado** y la fecha.
3. Presionar **Registrar salida** y confirmar.

La salida es definitiva: el alumno deja de estar activo, su QR se invalida y su historial se
conserva. No se borra ningún registro.

## Desactivación

Desde el listado o la ficha se puede **Desactivar** a un alumno. El registro no se elimina y el QR
deja de funcionar. Un alumno desactivado no aparece en los carnés.

## Auditoría

Todas las acciones de escritura quedan registradas con usuario, fecha, entidad, acción y valores
anterior y nuevo. La tabla `auditoria` es de solo escritura: no admite `UPDATE` ni `DELETE`, ni
siquiera desde la base de datos.

## Configuración académica

El módulo de configuración vive en `/configuracion`. Desde ahí el administrador gestiona el año
escolar y todo lo que necesita el registro de asistencia.

- **Año escolar** (`/configuracion/anio`): crea el año con fecha de inicio y fin, muestra el
  historial y permite cerrarlo. Solo puede existir un año activo; al cerrarlo queda de solo
  lectura.
- **Grados** (`/configuracion/grados`): niveles del año activo.
- **Secciones** (`/configuracion/secciones`): alta de secciones por grado, turno inicial y cambio
  de turno con fecha de vigencia.
- **Turnos** (`/configuracion/turnos`): hora de ingreso, tolerancia, límite de tardanza y cierre.
  Cada cambio indica desde qué fecha aplica y conserva el historial.
- **Calendario** (`/configuracion/calendario`): feriados, días no laborables y suspensiones, con
  alcance a todo el colegio o a grados específicos. Si se superpone con otra excepción del mismo
  alcance, el sistema pide confirmación.
- **Umbrales** (`/configuracion/umbrales`): cantidad mensual de tardanzas o faltas que activa una
  alerta. Los valores deben estar entre 1 y 100.
- **Cursos** (`/configuracion/cursos`): asignaturas vinculadas a uno o más grados. No se eliminan:
  se desactivan y conservan su historial.
- **Bloques de clase** (`/configuracion/bloques`): horas de clase por turno. No pueden salirse del
  horario del turno ni superponerse.
- **Horario semanal** (`/configuracion/horarios`): curso por sección, día y bloque. Solo se puede
  asignar un curso vinculado al grado de la sección. Si la celda ya tiene curso, se pide
  confirmación antes de reemplazarlo.
- **Consulta histórica** (`/configuracion/consulta`): muestra la configuración vigente en una
  fecha: horarios, excepciones, umbrales, turno de cada sección y bloques.

Las fechas de vigencia deben ser hoy o posteriores. Los cambios futuros no alteran la
configuración anterior ni las asistencias ya registradas.

## Asistencia del docente

- **Docentes** (`/docentes`): crea cuentas con nombre, correo y contraseña, restablece la clave y
  desactiva cuentas sin borrarlas.
- **Asignaciones** (`/asistencia/asignaciones`): vincula un docente con un curso y una sección
  desde una fecha. Marca **Titular** o desmárcalo para un reemplazo con fecha de fin.
- **Consulta** (`/asistencia/consulta`): revisa los registros del docente por fecha y sección,
  con el ingreso QR y las diferencias. Puedes corregir un registro indicando estado, observación y
  **motivo obligatorio**.
- **Parámetros** (`/configuracion/asistencia`): define el plazo de edición del docente, entre 1 y
  168 horas.

El registro del docente es personal y no reemplaza la asistencia oficial del QR. Las correcciones
del administrador y los cambios de plazo quedan auditados.

## Reportes

- **Diario** (`/reportes/diario`): estado y hora de cada alumno, con filtros por fecha, turno,
  grado, sección, alumno y estado.
- **Panel** (`/reportes/panel`): presentes, tardanzas, ausentes y pendientes; se actualiza cada
  30 segundos.
- **Resumen** (`/reportes/resumen`): asistencias, tardanzas, faltas, justificadas y porcentaje por
  alumno y sección en un rango o mes.
- **Reiteración** (`/reportes/reiteracion`): alumnos que superan los umbrales vigentes.
- **Exportación**: cada reporte se descarga en Excel o PDF con los filtros aplicados.

Los feriados y días no laborables se excluyen del porcentaje. Los docentes solo ven sus secciones.

## Justificaciones

- **Registro**: el docente de la sección o el administrador registra una justificación con motivo,
  rango de fechas y hasta tres adjuntos PDF/JPG/PNG de 5 MB.
- **Resolución**: el director o el administrador aprueba o rechaza. El rechazo exige comentario.
  Nadie puede resolver una justificación que registró.
- **Anulación**: solo el administrador anula una justificación aprobada, indicando el motivo. El
  registro vuelve a su estado original.
- **Plazo**: se configura en `/configuracion/asistencia` (5 días por defecto). Fuera de plazo la
  justificación se rechaza al registrarse.
- **Alertas**: al aprobar, las alertas afectadas se marcan como “situación cambiada” y no se
  eliminan.

## Alertas internas

- **Panel** (`/alertas`): muestra todas las alertas del colegio con filtros por fecha, sección,
  tipo y estado. El administrador ve las alertas de todos los docentes y porteros.
- **Umbrales** (`/configuracion/umbrales`): define las tardanzas por mes, las faltas por mes y las
  faltas seguidas. Cada valor tiene fecha de vigencia y no modifica las alertas ya generadas.
- **Sin responsable**: si una sección no tiene docente asignado, la alerta de reiteración llega
  al administrador y se marca como “sin responsable”. Desde el panel se puede reasignar a un
  docente activo.
- **Lectura**: las alertas pasan de nuevas a vistas y guardan quién y cuándo las leyó. No se
  eliminan.

El sistema evalúa las tardanzas cada minuto y los cierres de turno a las 13:00 y 19:00. No se
envían notificaciones externas.

## Recomendaciones

- Verificar el año escolar activo antes de importar.
- Revisar el reporte de importación y corregir las filas rechazadas.
- Generar los QR antes de imprimir los carnés.
- Mantener las fotos en JPG o PNG y con un tamaño razonable.
- Cerrar la sesión al terminar, especialmente en computadoras compartidas.
