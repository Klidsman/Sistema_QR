# Contrato de rutas HTTP: Alertas internas

**Fecha**: 2026-11-06 | **Plan**: [plan.md](../plan.md)

## Permisos

- `docente`: solo alertas de secciones a las que estuvo asignado en la fecha del evento.
- `porteria`: solo alertas de tardanza de los ingresos que registró.
- `administrador`: todas las alertas y reasignación de responsables.
- `director`: sin acceso a alertas en esta versión.
- Sin sesión: redirección a `/ingreso`. Todo POST exige CSRF.

## Panel e indicador

| Método | Ruta | Descripción | Permiso |
|---|---|---|---|
| GET | /alertas | Panel con filtros de fecha, sección, tipo y estado | Docente, portería, administrador |
| GET | /alertas/indicador | Fragmento HTMX con la cantidad de alertas nuevas | Docente, portería, administrador |
| POST | /alertas/{id}/vista | Marca una alerta como vista | Destinatario o administrador |
| POST | /alertas/vistas | Marca como vistas las alertas del filtro activo | Destinatario o administrador |
| POST | /alertas/{id}/responsable | Asigna un docente responsable a una alerta sin responsable | Administrador |

## Configuración de umbrales

Se reutiliza `/configuracion/umbrales` de la feature 002 con el tipo adicional
`faltas_seguidas`.

| Método | Ruta | Descripción | Permiso |
|---|---|---|---|
| GET | /configuracion/umbrales | Umbrales vigentes e historial | Administrador, director |
| POST | /configuracion/umbrales | Guarda tardanzas, faltas o faltas seguidas con vigencia | Administrador |

## Respuestas de error

| Situación | Código | Mensaje |
|---|---|---|
| Alerta de otra sección o usuario | 403 | "No tiene acceso a esta alerta." |
| Alerta inexistente | 404 | "La alerta no existe." |
| Intento de eliminar | 405 | No se ofrece la acción; las alertas no se eliminan. |
| Umbral cero o negativo | 422 | "El valor del umbral debe ser un entero entre 1 y 100." |
| Sin responsable para reasignar | 422 | "Seleccione un docente activo." |
| Director en cualquier ruta | 403 | "No tiene permiso para esta acción." |
