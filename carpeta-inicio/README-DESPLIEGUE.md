# README de despliegue paso a paso

Guía recomendada para publicar el sistema de asistencia de la **IEP María Parado de Bellido — Ayacucho** con HTTPS.

> Antes de desplegar, revisar `README-PRODUCCION.md` para confirmar los pendientes y criterios de aceptación.

---

## 1. Elegir dónde desplegar

| Opción | Cuándo conviene | HTTPS | Dificultad | Costo |
|---|---|---|---|---|
| **A. Render (recomendada)** | El colegio tiene internet estable y se quiere evitar mantenimiento de servidor. | Automático | Baja | Plan pago pequeño + base |
| **B. VPS con Docker + Caddy** | Se quiere control total o los datos deben quedar en el país. | Automático con dominio | Media | VPS mensual |
| **C. Solo red local** | Prueba o contingencia sin internet. | No | Baja | Ninguno |

**Recomendación:** opción **A** si hay internet; opción **B** si se necesita control y presupuesto para un VPS.

> La cámara en vivo del celular requiere HTTPS. En red local por HTTP solo funciona "Tomar foto del código QR" y el lector USB.

---

## 2. Opción A — Render paso a paso

### 2.1 Requisitos

- Cuenta en GitHub con el repositorio del proyecto.
- Cuenta en Render.
- Dominio propio (opcional pero recomendado).
- PostgreSQL de Render o externo.

### 2.2 Crear la base de datos

1. En Render, **New → PostgreSQL**.
2. Elegir plan y región cercana (por ejemplo, Oregon o Virginia).
3. Crear la base `asistencia`.
4. Copiar la **Internal Database URL**. Debe verse así:

   ```text
   postgres://usuario:clave@host/asistencia
   ```

   Si no trae SSL, agregar `?sslmode=require`.

### 2.3 Crear el servicio web

1. **New → Web Service** y conectar el repositorio.
2. En **Runtime**, elegir **Docker**. Render detecta el `Dockerfile` de la raíz.
3. Configurar:

   | Campo | Valor |
   |---|---|
   | Name | `asistencia-mpb` |
   | Region | la misma de la base |
   | Branch | `main` |
   | Health Check Path | `/salud` |

4. En **Environment**, agregar:

   ```env
   BASE_DATOS_URL=postgres://usuario:clave@host/asistencia?sslmode=require
   CLAVE_SESION=<clave aleatoria de 32 o más caracteres>
   COOKIE_SEGURA=true
   PUERTO=8080
   ALMACENAMIENTO_TIPO=disco
   ALMACENAMIENTO_RUTA=/app/datos/fotos
   ```

5. En **Pre-Deploy Command**:

   ```sh
   ./servidor migrar
   ```

6. En **Disks**, agregar un disco persistente de al menos 1 GB montado en:

   ```text
   /app/datos/fotos
   ```

   > Si el plan no permite discos, usar S3 y configurar `ALMACENAMIENTO_TIPO=s3`.

7. Desplegar. Render construye la imagen y expone una URL `https://...onrender.com`.

### 2.4 Conectar el dominio (opcional)

1. En Render, **Settings → Custom Domains → Add Custom Domain**.
2. Copiar el registro `CNAME` que indica Render.
3. En el proveedor del dominio, crear ese `CNAME`.
4. Esperar la verificación y el certificado automático.
5. Abrir `https://asistencia.mi-colegio.pe`.

### 2.5 Crear el primer administrador

En Render, abrir **Shell** del servicio y ejecutar:

```sh
ADMIN_CORREO='admin@colegio.pe' \
ADMIN_CONTRASENIA='<clave segura>' \
ADMIN_NOMBRE='Administrador' \
./servidor crear-admin
```

Luego ingresar en `https://tu-dominio/ingreso`.

No ejecutar `sembrar` sobre producción: crea datos de demostración.

### 2.6 Verificar

1. `https://tu-dominio/salud` debe responder `200`.
2. Ingresar como administrador.
3. Crear una cuenta de docente y una de portería.
4. Probar el registro desde portería.
5. Subir una foto de alumno, reiniciar el servicio y confirmar que la foto sigue visible.

---

## 3. Opción B — VPS con Docker + Caddy

Requiere un VPS con Ubuntu 22.04/24.04, dominio apuntando al VPS y acceso SSH.

### 3.1 Preparar el servidor

```sh
sudo apt update && sudo apt upgrade -y
sudo apt install -y docker.io docker-compose-plugin git
sudo usermod -aG docker $USER
# cerrar y volver a entrar por SSH
```

### 3.2 Abrir puertos

En el firewall del proveedor y en `ufw`:

```sh
sudo ufw allow OpenSSH
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

No exponer PostgreSQL a internet.

### 3.3 Clonar el proyecto

```sh
git clone <URL_DEL_REPOSITORIO> /opt/asistencia
cd /opt/asistencia
```

### 3.4 Crear `.env` de producción

```env
DB_PASSWORD=<clave fuerte de la base>
CLAVE_SESION=<clave aleatoria de 32 o más caracteres>
COOKIE_SEGURA=true
```

Generar la clave de sesión:

```sh
openssl rand -base64 48
```

### 3.5 Levantar PostgreSQL y la app

El `docker-compose.yml` del proyecto ya incluye `postgres` y `app`:

```sh
docker compose up -d --build
docker compose run --rm app migrar
docker compose ps
```

> En un VPS con Caddy delante, es recomendable publicar la app solo en localhost. En
> `docker-compose.yml`, cambiar `"8080:8080"` por `"127.0.0.1:8080:8080"` y no exponer el puerto
> de PostgreSQL. Caddy se encarga del acceso público por 443.

### 3.6 Crear el administrador

```sh
docker compose run --rm \
  -e ADMIN_CORREO='admin@colegio.pe' \
  -e ADMIN_CONTRASENIA='<clave segura>' \
  -e ADMIN_NOMBRE='Administrador' \
  app crear-admin
```

### 3.7 Publicar con HTTPS usando Caddy

Instalar Caddy:

```sh
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https curl
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update && sudo apt install -y caddy
```

Crear `/etc/caddy/Caddyfile`:

```caddy
asistencia.mi-colegio.pe {
    encode gzip
    reverse_proxy 127.0.0.1:8080
}
```

Recargar:

```sh
sudo systemctl reload caddy
```

Caddy obtiene y renueva el certificado de Let's Encrypt automáticamente. El dominio debe apuntar al IP del VPS antes de este paso.

### 3.8 Verificar

```sh
curl -I https://asistencia.mi-colegio.pe/salud
docker compose logs -f app
```

---

## 4. Opción C — Solo red local (sin dominio)

Útil para probar o como contingencia.

1. Levantar el sistema en la PC servidor:

   ```powershell
   docker compose up -d --build
   docker compose run --rm app migrar
   ```

2. En `.env`, usar:

   ```env
   COOKIE_SEGURA=false
   ```

3. Averiguar el IP local de la PC:

   ```powershell
   ipconfig
   ```

4. Abrir desde la PC: `http://localhost:8080`.
5. Abrir desde el celular: `http://IP_LOCAL:8080` (misma red Wi-Fi).
6. Si el celular no conecta, permitir el puerto en el firewall (PowerShell como administrador):

   ```powershell
   New-NetFirewallRule -DisplayName "ASISTENCIA 8080" -Direction Inbound -Protocol TCP -LocalPort 8080 -Action Allow
   ```

Limitaciones:

- Sin HTTPS, la cámara en vivo del celular no funciona. Usar **Tomar foto del código QR** o lector USB.
- Si se reinicia la PC o cambia el IP, cambia la URL.
- No recomendado para uso diario si hay internet disponible.

---

## 5. Después del despliegue

### 5.1 Respaldo diario

Programar en el servidor (Linux):

```cron
0 3 * * * cd /opt/asistencia && docker compose exec -T postgres pg_dump -U asistencia -d asistencia --format=custom > /respaldos/asistencia_$(date +\%Y\%m\%d).dump
```

> Ajustar la ruta y el usuario si el proyecto o la base usan otros nombres. Verificar que `/respaldos` exista y que el usuario de cron tenga permisos.

Respaldar también el volumen de fotos. Ver `docs/respaldo_restauracion.md`.

### 5.2 Actualizaciones

```sh
git pull
docker compose up -d --build
docker compose run --rm app migrar
```

Las migraciones se ejecutan antes de reiniciar la nueva versión.

### 5.3 Rollback

1. Volver a la versión anterior del código (`git checkout <tag>`).
2. Reconstruir: `docker compose up -d --build`.
3. Restaurar el respaldo previo solo si la migración nueva no es compatible hacia atrás.

### 5.4 Monitoreo

- Verificar `/salud` cada minuto desde un monitor externo.
- Revisar logs JSON: `docker compose logs --tail=200 app`.
- Alertar si el contenedor se reinicia repetidamente.

---

## 6. Checklist final

- [ ] Dominio apunta al servidor o se usa la URL del proveedor.
- [ ] HTTPS activo y certificado válido.
- [ ] `COOKIE_SEGURA=true`.
- [ ] Base de datos de producción sin datos de prueba.
- [ ] Migraciones aplicadas (`./servidor migrar`).
- [ ] Primer administrador creado.
- [ ] Disco persistente de fotos o S3 configurado.
- [ ] Respaldo automático funcionando.
- [ ] Restauración probada.
- [ ] `/salud` monitoreado.
- [ ] Cámara en vivo probada desde celular por HTTPS.
- [ ] Manuales entregados al personal.
- [ ] Criterios de aceptación de `README-PRODUCCION.md` cumplidos.
