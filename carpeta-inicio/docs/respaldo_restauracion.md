# Respaldo y restauración

El sistema guarda datos en PostgreSQL y, cuando se usa almacenamiento en disco, las fotos en una
carpeta persistente. Un respaldo completo necesita ambos componentes.

## Qué respaldar

| Componente | Contenido | Frecuencia sugerida |
|---|---|---|
| PostgreSQL | Usuarios, alumnos, secciones, QR, historial, importaciones y auditoría. | Diaria. |
| Fotos en disco | Archivos JPG y PNG de los alumnos. | Diaria, junto con la base. |
| Variables de entorno | `BASE_DATOS_URL`, `CLAVE_SESION`, credenciales S3. | Al cambiar la configuración, en un gestor de secretos. |
| Migraciones | Carpeta `migraciones/` del repositorio. | Con cada versión. |

Si se usa S3 para las fotos, el respaldo lo provee el proveedor. Aun así, configure versionado y
una política de retención en el bucket.

## Respaldo de PostgreSQL

Use `pg_dump` en formato personalizado, que permite restaurar tablas de forma selectiva:

```sh
pg_dump "$BASE_DATOS_URL" \
  --format=custom \
  --file=respaldo_asistencia_$(date +%Y%m%d_%H%M).dump
```

Verificar que el archivo no esté vacío y que se pueda listar:

```sh
pg_restore --list respaldo_asistencia_20261106_0300.dump | head
```

### Respaldo desde Docker

```sh
docker exec carpeta-inicio-postgres-1 pg_dump -U asistencia -d asistencia \
  --format=custom > respaldo_asistencia_$(date +%Y%m%d_%H%M).dump
```

### Programación

En un servidor Linux, agregar una tarea cron diaria a las 03:00:

```cron
0 3 * * * cd /ruta/al/proyecto && ./scripts/respaldar.sh >> /var/log/asistencia-respaldo.log 2>&1
```

El script debe:

1. Ejecutar `pg_dump` y fallar si el comando devuelve un error.
2. Comprimir o cifrar el archivo.
3. Copiarlo a un almacenamiento externo (otro servidor, bucket privado o disco de red).
4. Borrar respaldos locales con más de 30 días.
5. Verificar semanalmente que el respaldo más reciente se pueda listar con `pg_restore --list`.

## Respaldo de fotos

Con almacenamiento en disco, copiar la carpeta completa:

```sh
tar -czf fotos_$(date +%Y%m%d_%H%M).tar.gz -C /ruta/datos fotos
```

En Docker:

```sh
docker run --rm -v asistencia_fotos:/datos -v "$PWD":/respaldo alpine \
  tar -czf /respaldo/fotos_$(date +%Y%m%d_%H%M).tar.gz -C /datos .
```

## Restauración

### 1. Detener la aplicación

Detenga el servicio web para evitar escrituras durante la restauración. Los usuarios no podrán
ingresar mientras dure el proceso.

### 2. Restaurar la base de datos

Sobre una base vacía:

```sh
pg_restore --clean --if-exists --no-owner \
  --dbname="$BASE_DATOS_URL" respaldo_asistencia_20261106_0300.dump
```

Si la base es nueva, cree primero el usuario y la base:

```sh
createdb -h localhost -p 5433 -U asistencia asistencia
```

### 3. Restaurar las fotos

```sh
tar -xzf fotos_20261106_0300.tar.gz -C /ruta/datos
```

En Docker:

```sh
docker run --rm -v asistencia_fotos:/datos -v "$PWD":/respaldo alpine \
  sh -c "cd /datos && tar -xzf /respaldo/fotos_20261106_0300.tar.gz"
```

### 4. Aplicar migraciones pendientes

El respaldo puede ser de una versión anterior. Aplique las migraciones del binario desplegado:

```sh
./servidor migrar
```

### 5. Verificar

1. Abrir `/salud`: debe responder `200`.
2. Ingresar como administrador.
3. Revisar el listado de alumnos y abrir una ficha con foto y QR.
4. Abrir el historial de importaciones.
5. Revisar la tabla `auditoria` para confirmar que los registros históricos están presentes.

### 6. Reiniciar la aplicación

Levante el servicio y comunique a los usuarios que el sistema está disponible.

## Objetivos de recuperación

| Métrica | Objetivo sugerido |
|---|---|
| RPO (pérdida máxima de datos) | 24 horas con respaldo diario; 1 hora con respaldo por WAL. |
| RTO (tiempo de recuperación) | Menos de 2 horas para una base pequeña o mediana. |

Para reducir el RPO, habilite archivado de WAL en PostgreSQL y pruebe la restauración a un punto
en el tiempo al menos una vez por trimestre.

## Prueba de restauración

Un respaldo no es confiable hasta que se restaura. Una vez al mes:

1. Crear una base temporal.
2. Restaurar el respaldo más reciente.
3. Ejecutar `./servidor migrar` contra la base temporal.
4. Abrir la aplicación apuntando a esa base y revisar alumnos, fotos, QR e historial.
5. Documentar la fecha, el responsable y el resultado.

## Seguridad de los respaldos

- Cifrar los archivos antes de enviarlos fuera del servidor.
- Restringir el acceso a los respaldos a las personas administradoras.
- No incluir `CLAVE_SESION` ni credenciales en el mismo archivo que los datos.
- Los respaldos contienen datos personales de menores: aplicar la Ley 29733 y la política interna
  de retención del colegio.
