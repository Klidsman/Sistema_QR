# Contrato de rutas HTTP: Gestión de alumnos y generación de códigos QR

**Fecha**: 2026-11-05 | **Plan**: [plan.md](../plan.md)

Las rutas están en español, sin tildes ni espacios (constitución, principio IX). Las respuestas
son HTML renderizado en el servidor; HTMX pide fragmentos cuando el encabezado `HX-Request` está
presente.

## Permisos

- `administrador`: todas las rutas de esta feature.
- `director`: solo rutas `GET`. Cualquier `POST` responde 403.
- Sin sesión: redirección a `/ingreso`.
- Todas las rutas que cambian estado exigen token CSRF.

## Autenticación

| Método | Ruta | Descripción | Permiso |
|---|---|---|---|
| GET | /ingreso | Formulario de ingreso | Público |
| POST | /ingreso | Inicia sesión; aplica límite de intentos | Público |
| POST | /salir | Cierra sesión | Autenticado |

## Alumnos

| Método | Ruta | Descripción | Permiso |
|---|---|---|---|
| GET | /alumnos | Listado con filtros de grado, sección, estado y búsqueda | Administrador, director |
| GET | /alumnos/nuevo | Formulario de alta | Administrador |
| POST | /alumnos | Crea un alumno; rechaza identificador repetido | Administrador |
| GET | /alumnos/{id} | Ficha del alumno, historial de secciones, QR y salidas | Administrador, director |
| GET | /alumnos/{id}/editar | Formulario de edición | Administrador |
| POST | /alumnos/{id} | Actualiza datos (identificador solo con corrección auditada) | Administrador |
| POST | /alumnos/{id}/desactivacion | Desactiva el alumno e invalida su QR | Administrador |
| POST | /alumnos/{id}/salida | Marca retirado o trasladado con fecha; invalida QR | Administrador |
| POST | /alumnos/{id}/seccion | Cambia grado o sección con fecha de vigencia | Administrador |
| GET | /alumnos/{id}/foto | Devuelve la foto por proxy autenticado | Administrador, director |
| POST | /alumnos/{id}/foto | Sube o reemplaza la foto (JPG o PNG, máximo configurado) | Administrador |

## Códigos QR

| Método | Ruta | Descripción | Permiso |
|---|---|---|---|
| POST | /alumnos/{id}/qr | Emite el QR de un alumno activo sin QR vigente | Administrador |
| POST | /alumnos/{id}/qr/reemplazo | Revoca el QR vigente (motivo obligatorio) y emite uno nuevo | Administrador |
| GET | /alumnos/{id}/qr/imagen | Imagen PNG del QR vigente (solo contiene el token) | Administrador, director |

## Importación

| Método | Ruta | Descripción | Permiso |
|---|---|---|---|
| GET | /alumnos/importar | Formulario de carga de CSV o XLSX | Administrador |
| POST | /alumnos/importar | Procesa el archivo; crea válidos y reporta rechazados | Administrador |
| GET | /alumnos/importaciones | Historial de importaciones | Administrador, director |
| GET | /alumnos/importaciones/{id} | Detalle con filas rechazadas y motivos | Administrador, director |

## Carnés

| Método | Ruta | Descripción | Permiso |
|---|---|---|---|
| GET | /carnets | Selección por grado o sección | Administrador, director |
| GET | /carnets/lote | Vista imprimible del lote (query: grado o seccion) con avisos previos | Administrador, director |

## Plataforma

| Método | Ruta | Descripción | Permiso |
|---|---|---|---|
| GET | /salud | Estado de la aplicación y de la base de datos | Público |

## Respuestas de error relevantes

- Identificador repetido en alta: 422 con mensaje en español; el formulario conserva los datos.
- QR revocado usado en otra feature: rechazado como no válido (la validación pertenece a la
  feature de portería).
- Acción de director sobre `POST`: 403 con mensaje "No tiene permiso para esta acción".
