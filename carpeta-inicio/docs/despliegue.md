# Despliegue

Esta guía describe cómo ejecutar el sistema de asistencia en un contenedor y cómo publicarlo en
Render. La aplicación es un monolito Go que sirve HTML, CSS y JavaScript desde el mismo proceso.

## Requisitos

- Docker 24 o superior.
- PostgreSQL 16 (local, contenedor o servicio administrado).
- Un dominio con HTTPS para producción. Las cookies de sesión usan `Secure` y no viajan por HTTP.

## Variables de entorno

| Variable | Obligatoria | Descripción |
|---|---|---|
| `BASE_DATOS_URL` | Sí | Cadena de conexión PostgreSQL. |
| `CLAVE_SESION` | Sí | Clave de al menos 32 caracteres para firmar sesiones. |
| `PUERTO` | No | Puerto HTTP. Por defecto `8080`. |
| `ALMACENAMIENTO_TIPO` | No | `disco` (por defecto) o `s3`. |
| `ALMACENAMIENTO_RUTA` | No | Carpeta de fotos cuando se usa disco. |
| `CARNET_ANCHO_MM` / `CARNET_ALTO_MM` | No | Tamaño del carné en milímetros. Por defecto `85` × `54`. |
| `FOTO_TAMANIO_MAXIMO_MB` | No | Tamaño máximo de la foto. Por defecto `2`. |
| `IMPORTACION_MAX_FILAS` | No | Máximo de filas por importación. Por defecto `5000`. |
| `ADMIN_CORREO` / `ADMIN_CONTRASENIA` / `ADMIN_NOMBRE` | Solo para la semilla | Administrador inicial. |
| `ANIO_NOMBRE` | Solo para la semilla | Nombre del año escolar de demostración. |

Use `.env.ejemplo` como plantilla. No versione el archivo `.env` real.

## Ejecución local con Docker

1. Construir la imagen:

   ```sh
   docker build -t asistencia:local .
   ```

2. Levantar PostgreSQL de desarrollo:

   ```sh
   docker compose up -d postgres
   ```

3. Aplicar migraciones. El binario incluye las migraciones SQL:

   ```sh
   docker run --rm --network host \
     -e BASE_DATOS_URL='postgres://asistencia:cambiar_esta_clave@localhost:5433/asistencia?sslmode=disable' \
     -e CLAVE_SESION='cambiar_por_una_clave_aleatoria_de_al_menos_32_caracteres' \
     asistencia:local migrar
   ```

4. Crear el administrador y las secciones de demostración:

   ```sh
   docker run --rm --network host \
     -e BASE_DATOS_URL='postgres://asistencia:cambiar_esta_clave@localhost:5433/asistencia?sslmode=disable' \
     -e CLAVE_SESION='cambiar_por_una_clave_aleatoria_de_al_menos_32_caracteres' \
     -e ADMIN_CORREO='admin@colegio.pe' \
     -e ADMIN_CONTRASENIA='cambiar_esta_clave' \
     -e ADMIN_NOMBRE='Administrador' \
     -e ANIO_NOMBRE='2026' \
     asistencia:local sembrar
   ```

5. Iniciar la aplicación:

   ```sh
   docker run -d --name asistencia -p 8080:8080 \
     -e BASE_DATOS_URL='postgres://asistencia:cambiar_esta_clave@localhost:5433/asistencia?sslmode=disable' \
     -e CLAVE_SESION='cambiar_por_una_clave_aleatoria_de_al_menos_32_caracteres' \
     -v asistencia_fotos:/app/datos/fotos \
     asistencia:local
   ```

6. Abrir `http://localhost:8080/ingreso` e iniciar sesión con el administrador creado.

En Linux, `--network host` permite que el contenedor llegue al PostgreSQL del host. En Docker
Desktop (Windows o macOS) use `host.docker.internal` en lugar de `localhost`.

## Despliegue en Render

### 1. Base de datos

1. Crear un PostgreSQL administrado en Render.
2. Copiar la cadena de conexión interna. Agregar `?sslmode=require` si Render no la incluye.
3. Anotar usuario, contraseña, host y base de datos.

### 2. Servicio web

1. Crear un **Web Service** conectado al repositorio.
2. Elegir **Docker** como entorno. Render detectará el `Dockerfile` de la raíz.
3. Configurar las variables de entorno:
   - `BASE_DATOS_URL` con la cadena interna de PostgreSQL.
   - `CLAVE_SESION` con una clave aleatoria de al menos 32 caracteres.
   - `ALMACENAMIENTO_TIPO=disco` y `ALMACENAMIENTO_RUTA=/app/datos/fotos`.
4. Agregar un **disco persistente** montado en `/app/datos/fotos` (mínimo 1 GB).
5. Configurar el **Pre-Deploy Command**:

   ```sh
   ./servidor migrar
   ```

6. Configurar el **Health Check Path** en `/salud`.
7. Desplegar. Render expone HTTPS automáticamente.

### 3. Primer administrador

Ejecutar una sola vez desde la consola del servicio:

```sh
ADMIN_CORREO='admin@colegio.pe' \
ADMIN_CONTRASENIA='cambiar_esta_clave' \
ADMIN_NOMBRE='Administrador' \
./servidor crear-admin
```

Para crear también secciones de demostración, usar `./servidor sembrar` con `ANIO_NOMBRE`.
En producción normalmente las secciones se cargan con la feature 002; no ejecute `sembrar`
sobre datos reales sin revisar antes los nombres.

## Almacenamiento de fotos en S3

Para usar un servicio compatible con S3:

```env
ALMACENAMIENTO_TIPO=s3
ALMACENAMIENTO_BUCKET=asistencia-fotos
ALMACENAMIENTO_ENDPOINT=https://s3.amazonaws.com
ALMACENAMIENTO_CLAVE_ACCESO=...
ALMACENAMIENTO_CLAVE_SECRETA=...
```

Use un bucket privado. La aplicación entrega las fotos mediante rutas autenticadas; no configure
el bucket como público.

## Operación

- **Salud**: `GET /salud` responde `200` cuando la base de datos está disponible y `503` si no.
- **Migraciones**: se ejecutan antes de iniciar cada versión con `./servidor migrar`.
- **Registros**: el proceso escribe logs JSON en la salida estándar. Render los captura.
- **Respaldo**: seguir `docs/respaldo_restauracion.md`.
- **Migración de datos**: el sistema no borra historial de asistencia ni alumnos; los cambios de
  estado se registran con fecha.

## Solución de problemas

| Síntoma | Causa probable | Acción |
|---|---|---|
| `/salud` responde 503 | Base de datos inaccesible | Revisar `BASE_DATOS_URL` y la red del servicio. |
| El ingreso redirige en bucle | Cookies `Secure` sobre HTTP | Publicar con HTTPS o usar localhost. |
| Error al guardar la contraseña inicial | `CLAVE_SESION` corta | Usar 32 caracteres o más. |
| Las fotos desaparecen tras un despliegue | Falta el disco persistente | Montar `/app/datos/fotos` o configurar S3. |
| Importación rechaza todas las filas | Encabezados o sección inexistente | Revisar el reporte y el año escolar activo. |
