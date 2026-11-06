# Manual de portería

## Ingreso

1. Abrir `/ingreso` en el navegador del equipo de portería.
2. Ingresar con el correo y la contraseña asignados por el administrador.
3. Si aparece un error de credenciales, verificar mayúsculas y consultar al administrador.

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
