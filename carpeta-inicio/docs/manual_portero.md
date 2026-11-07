# Manual de portería

## Ingreso

1. Abrir `/ingreso` en el navegador del equipo de portería.
2. Ingresar con el correo y la contraseña asignados por el administrador en **Porteros**
   (`/porteros`).
3. Si aparece un error de credenciales, verificar mayúsculas y consultar al administrador.

## Registrar un ingreso

1. Abrir `/porteria`.
2. Escanear el QR del carné con la cámara o pegar su contenido en **Código QR**.
3. Presionar **Registrar ingreso**.

La pantalla ofrece tres formas de leer el carné:

- **Activar cámara**: lectura en vivo con `BarcodeDetector` o, si el navegador no lo tiene, con
  `jsQR`. Requiere HTTPS o `localhost`.
- **Tomar foto del código QR**: abre la cámara del celular y decodifica la foto. Funciona también
  por HTTP en la red local y es la opción recomendada desde el celular.
- **Código QR**: pegar o digitar el código, útil con lectores USB o el código de respaldo.

La pantalla muestra el nombre del alumno, su sección, la hora y el estado: **Puntual**,
**Tardanza** o **Tardanza grave**. Si el escaneo se rechaza, se muestra el motivo y se puede
reintentar.

Rechazos frecuentes:

| Mensaje | Causa |
|---|---|
| El código QR no es válido. | Token inexistente, revocado o alumno inactivo. |
| El ingreso del día ya fue registrado. | El alumno ya tiene marcación de hoy. |
| Hoy no es un día laborable para la sección del alumno. | Feriado o suspensión. |
| Todavía no comienza el registro del turno. | Escaneo antes de la hora de ingreso. |
| El registro del turno ya cerró. | Escaneo después de la hora de cierre. |
| La sección no tiene un turno vigente. | Falta configurar el turno de la sección. |
| El turno no tiene un horario configurado. | Falta el horario del turno para la fecha. |

El administrador puede ver el **Código de respaldo** en la ficha del alumno para digitarlo
manualmente si la cámara no está disponible.

## Alertas de tardanza

Cada ingreso por QR que corresponde a una tardanza genera una alerta interna para el portero que
lo registró. La alerta aparece en el indicador **Alertas** de la barra superior y en el panel
`/alertas`.

- Solo se ven las tardanzas registradas por la propia cuenta.
- El panel permite filtrar por fecha, tipo y estado.
- Una alerta se marca como vista al abrirla o con el botón **Marcar vista**.
- El botón **Marcar las mostradas como vistas** actualiza únicamente las alertas del filtro
  activo.
- Las alertas no se eliminan.

## Situaciones que no generan alerta

- Los ingresos puntuales no generan alerta.
- Las inasistencias no generan alerta diaria.
- Las diferencias entre el registro del docente y el QR no generan alerta.

## Reglas importantes

- No se envían notificaciones a apoderados ni a canales externos.
- La alerta de tardanza llega solo a la cuenta que registró el ingreso.
- Las alertas de reiteración por umbrales llegan a los docentes de la sección, no a portería.

## Recomendaciones

- Mantener la sesión abierta solo durante el turno.
- Cerrar sesión al terminar, especialmente en equipos compartidos.
- Informar al administrador si un ingreso queda registrado con el usuario equivocado.
