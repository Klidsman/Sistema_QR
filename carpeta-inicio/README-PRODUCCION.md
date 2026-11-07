# README de producción

Documento para el equipo que va a poner el sistema en producción. Explica **qué está listo**, **qué falta** y **qué debe verificarse antes de abrirlo a los usuarios**.

> Complemento: `README-DESPLIEGUE.md` contiene la guía paso a paso para elegir y configurar el hosting.

---

## 1. Qué es el sistema

Aplicación web de asistencia escolar para la **IEP María Parado de Bellido — Ayacucho**:

- Página pública de presentación en `/`.
- Ingreso por correo y contraseña en `/ingreso`.
- Registro del ingreso de alumnos por **código QR** desde portería (cámara, foto o lector USB).
- Asistencia por **hora de clase** registrada por docentes.
- Justificaciones de faltas y tardanzas con adjuntos.
- Alertas internas por tardanzas e inasistencias reiteradas.
- Reportes diarios, mensuales, panel de indicadores y exportación.
- Gestión de cuentas de **docentes** y **porteros** por parte del administrador.

Es un monolito Go que sirve HTML, CSS y JavaScript desde el mismo proceso, con PostgreSQL 16.

---

## 2. Estado actual

| Área | Estado |
|---|---|
| Funcionalidades 001–007 | Implementadas |
| Página central pública | Implementada |
| Cuentas de portería | Implementadas |
| Migraciones | 29, con subida y bajada probadas |
| Pruebas unitarias, de manejadores e integración | Pasando |
| Pruebas E2E de la página central y accesos | Pasando |
| Docker local (`docker compose up`) | Funcionando |
| Documentación de usuario por rol | Disponible en `docs/` |

La aplicación **funciona hoy en red local**. Lo que falta es la configuración de producción.

---

## 3. Qué falta antes de producción

### 3.1 Obligatorio

| # | Pendiente | Detalle | Responsable |
|---|---|---|---|
| 1 | **HTTPS con dominio o certificado** | Necesario para cookies seguras y para la cámara en vivo. Sin HTTPS, el celular solo puede usar "Tomar foto del QR". | Infraestructura |
| 2 | **Base de datos de producción** | PostgreSQL 16 administrado o con respaldo. No usar la base de desarrollo. | Infraestructura |
| 3 | **Variables de entorno reales** | `BASE_DATOS_URL`, `CLAVE_SESION` (32+ caracteres) y `COOKIE_SEGURA=true`. Ver sección 4. | Infraestructura |
| 4 | **Primer administrador** | Crear con `./servidor crear-admin` en el servidor de producción. No reutilizar cuentas de prueba. | Administrador del sistema |
| 5 | **Migraciones aplicadas** | Ejecutar `./servidor migrar` antes de iniciar cada versión. | Infraestructura |
| 6 | **Persistencia de fotos** | Disco persistente montado en `/app/datos/fotos` o almacenamiento S3. | Infraestructura |
| 7 | **Respaldo y restauración** | Configurar respaldo diario de PostgreSQL y fotos. Probar una restauración. Ver `docs/respaldo_restauracion.md`. | Infraestructura |
| 8 | **Health check** | Configurar `/salud` como ruta de salud del servicio. | Infraestructura |

### 3.2 Recomendado

| # | Pendiente | Detalle |
|---|---|---|
| 9 | **Correo institucional real** | Definir los correos de administrador y dirección antes de crear las cuentas. |
| 10 | **Limpieza de datos de prueba** | La base actual contiene decenas de cuentas/docentes generados por pruebas. Producción debe empezar limpia. |
| 11 | **Revisión legal** | Fotos de menores y documentos de salud requieren revisión de la Ley 29733 y política interna de retención. Hay un `LEGAL_REVIEW` pendiente en las especificaciones. |
| 12 | **Capacitación** | Entregar `docs/manual_administrador.md`, `manual_director.md`, `manual_docente.md` y `manual_portero.md`. |
| 13 | **Monitoreo y logs** | Centralizar logs JSON y alertar si `/salud` falla. |
| 14 | **Prueba de carga mínima** | Validar el registro de portería en el horario de mayor ingreso. |

### 3.3 Fuera de alcance de esta versión

- Modo sin conexión (offline/PWA): diferido por diseño.
- Notificaciones a apoderados o canales externos.
- Autorregistro público de cuentas: **no existe**. Las cuentas las crea el administrador.
- Registro de administradores desde la web: se crean por comando o por otro administrador.

---

## 4. Variables de entorno

| Variable | Obligatoria | Valor en producción |
|---|---|---|
| `BASE_DATOS_URL` | Sí | Cadena de PostgreSQL de producción con `?sslmode=require` si el proveedor lo exige. |
| `CLAVE_SESION` | Sí | Aleatoria, mínimo 32 caracteres. No reutilizar la de desarrollo. |
| `COOKIE_SEGURA` | No | **`true`** en producción con HTTPS. `false` solo para red local por HTTP. |
| `PUERTO` | No | `8080` por defecto. |
| `ALMACENAMIENTO_TIPO` | No | `disco` o `s3`. |
| `ALMACENAMIENTO_RUTA` | No | `/app/datos/fotos` cuando se usa disco. |
| `ADMIN_CORREO` / `ADMIN_CONTRASENIA` / `ADMIN_NOMBRE` | Solo al crear admin | Solo para `crear-admin` / `sembrar`. |
| `ANIO_NOMBRE` | Solo semilla | No usar `sembrar` sobre datos reales sin revisar. |

Nunca versionar el `.env` real ni las credenciales.

---

## 5. Criterios de aceptación antes de abrir

1. `/salud` responde `200` con `{"estado":"correcto"}`.
2. La página central carga en `https://dominio/`.
3. Se puede ingresar como administrador con la cuenta creada en producción.
4. El administrador puede crear una cuenta de docente y una de portería.
5. La cuenta de portería ingresa y registra un alumno de prueba por:
   - cámara en vivo (si hay HTTPS), y
   - lector USB o pegado manual.
6. Un docente registra asistencia de una clase y ve el historial.
7. Dirección consulta reportes y aprueba una justificación.
8. Las fotos cargadas persisten después de reiniciar el contenedor.
9. Se ejecutó y verificó una restauración de respaldo.
10. La base de producción no contiene datos de prueba.

---

## 6. Contactos y responsables

Completar antes de publicar:

| Rol | Nombre | Contacto |
|---|---|---|
| Responsable del sistema | | |
| Administrador del colegio | | |
| Infraestructura / hosting | | |
| Soporte técnico | | |

---

## 7. Referencias

- `README-DESPLIEGUE.md` — guía de despliegue paso a paso.
- `docs/despliegue.md` — referencia técnica de despliegue.
- `docs/respaldo_restauracion.md` — respaldos y restauración.
- `docs/manual_administrador.md`, `manual_director.md`, `manual_docente.md`, `manual_portero.md`.
- `.env.ejemplo` — plantilla de variables.
