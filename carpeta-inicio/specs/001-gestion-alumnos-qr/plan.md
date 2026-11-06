# Plan de implementación: Gestión de alumnos y generación de códigos QR

**Branch**: `001-gestion-alumnos-qr` | **Fecha**: 2026-11-05 | **Spec**: [spec.md](spec.md)

**Input**: Especificación de la feature en `/specs/001-gestion-alumnos-qr/spec.md`

## Resumen

El administrador registra alumnos, los importa desde Excel o CSV, les genera un QR con token
aleatorio, reemplaza ese QR cuando se pierde o se roba, y genera carnés imprimibles por grado o
sección. El director consulta lo mismo sin poder modificar datos. Todo cambio queda auditado y el
historial de asistencia nunca se borra.

Enfoque técnico: monolito modular en Go con el módulo `alumnos` como núcleo, `academico` como
fuente de grados y secciones (vía interfaz pequeña), `autenticacion` para roles, y
`plataforma` para configuración, base de datos, auditoría y almacenamiento de fotos. Interfaz
renderizada en el servidor con templ, HTMX y Tailwind. Datos en PostgreSQL con pgx y sqlc,
migraciones con goose.

## Contexto técnico

**Lenguaje y versión**: Go estable reciente, fijada en `go.mod` al iniciar el proyecto.

**Dependencias principales**:
- Chi (router), templ (plantillas), HTMX y Tailwind CSS (interfaz).
- pgx y sqlc (acceso a PostgreSQL), goose (migraciones).
- scs con almacenamiento en PostgreSQL (sesiones), argon2 de `golang.org/x/crypto` (contraseñas).
- skip2/go-qrcode (QR), excelize (lectura de .xlsx), `encoding/csv` de la biblioteca estándar.
- aws-sdk-go-v2 (almacenamiento de fotos compatible con S3; ver investigación).
- slog (registro), testing y testify (pruebas), testcontainers-go (PostgreSQL real).

**Almacenamiento**: PostgreSQL administrado para datos. Fotos en almacenamiento de objetos
compatible con S3 (Cloudflare R2 o Backblaze B2) detrás de la interfaz `AlmacenFotos`. En
desarrollo, carpeta en disco.

**Pruebas**: `testing` y `testify` para reglas (tablas de casos), testcontainers con PostgreSQL
real para repositorios, `net/http/httptest` para manejadores, Playwright para el flujo de alta,
importación y carné.

**Plataforma objetivo**: contenedor Docker en Linux. Navegadores: Chrome, Edge, Firefox y Safari
actuales, en computadora, tablet y celular.

**Tipo de proyecto**: aplicación web monolítica con renderizado en el servidor.

**Metas de rendimiento** (de la spec): alta de alumno en menos de 1 minuto; importación de 2,000
filas en menos de 2 minutos; lote de carnés de 300 alumnos en menos de 3 minutos; páginas de
listado responden en menos de 1 segundo con 2,000 alumnos.

**Restricciones**:
- Fechas y horas en UTC, presentadas en America/Lima.
- El QR nunca contiene datos personales (constitución, principio IV).
- Fotos y datos de menores tratados según la Ley 29733 (principio VII).
- Secretos solo por variables de entorno, validadas al iniciar (principio VIII).
- Una sola instancia de la aplicación en la primera versión, para que el límite de intentos de
  inicio de sesión sea consistente (ver investigación).

**Escala y alcance**: un colegio, unos 2,000 alumnos, un año escolar activo, unos pocos
administradores y directores.

## Revisión de la constitución

*Compuerta: debe cumplirse antes de la investigación. Se revisa de nuevo después del diseño.*

| Principio | Estado | Observación |
|---|---|---|
| I. Simplicidad | Cumple con justificación | Chi, templ, aws-sdk-go-v2 y excelize se justifican en Complejidad justificada. Sin microservicios ni colas. |
| II. Código Go y monolito | Cumple | Módulos `alumnos`, `academico`, `autenticacion` y `plataforma`. `alumnos` nunca lee tablas de `academico`; usa la interfaz `BuscadorSeccion`. |
| III. Experiencia de usuario | Cumple | Vistas responsivas y mobile-first. El escaneo es de la feature de portería, no de esta. |
| IV. Integridad y seguridad | Cumple | Token QR aleatorio, revocable, único activo por alumno. Auditoría append-only. Roles con intermediario. |
| V. Pruebas | Cumple | Reglas con tablas de casos; repositorios con PostgreSQL real; Playwright para el flujo. |
| VI. Reglas configurables | Cumple | Tamaño de carné, tamaño máximo de foto y plazos en variables de entorno. |
| VII. Datos de menores | Cumple | Foto privada en almacenamiento de objetos; acceso autenticado por proxy; datos mínimos. Validación legal pendiente (TODO LEGAL_REVIEW). |
| VIII. Entrega operable | Cumple | Docker multietapa, `/salud`, `.env.ejemplo`, migraciones goose. |
| IX. Idioma y nomenclatura | Cumple | Nombres en español; excepciones solo de Go y librerías. Glosario en `docs/glosario.md`. |

Resultado: sin violaciones no justificadas.

## Estructura del proyecto

### Documentación de esta feature

```text
specs/001-gestion-alumnos-qr/
├── spec.md              # Especificación
├── plan.md              # Este archivo
├── research.md          # Decisiones técnicas
├── data-model.md        # Modelo de datos
├── quickstart.md        # Guía para validar la feature
├── contracts/
│   └── rutas.md         # Rutas HTTP y permisos
└── checklists/
    └── requirements.md
```

### Código fuente (raíz del repositorio)

```text
cmd/
└── servidor/
    └── main.go                       # Arranque: configuración, base de datos, rutas

internal/
├── plataforma/
│   ├── configuracion/                # Variables de entorno y validación al iniciar
│   ├── basedatos/                    # Conexión pgx y pool
│   ├── registro/                     # slog con mensajes en español
│   ├── intermediarios/               # Sesión, CSRF, roles, límite de intentos
│   ├── auditoria/                    # Registrador de auditoría (tabla propia, solo escritura)
│   ├── almacenamiento/               # Interfaz AlmacenFotos; implementación en disco y S3
│   └── salud/                        # Endpoint /salud
├── autenticacion/
│   ├── manejador.go                  # Ingreso y salida
│   ├── servicio.go                   # Validación de contraseñas con argon2id
│   ├── repositorio.go
│   └── modelo.go                     # Usuario, Rol
├── academico/
│   ├── servicio.go                   # Consulta de grados y secciones (interfaz BuscadorSeccion)
│   ├── repositorio.go
│   └── modelo.go                     # Grado, Seccion, AnioEscolar (lectura en esta feature)
└── alumnos/
    ├── manejador.go                  # Pantallas y acciones de alumnos
    ├── servicio.go                   # Reglas: unicidad, estados, cambios de sección
    ├── servicio_qr.go                # Emisión, revocación y reemplazo de QR
    ├── importacion.go                # Lectura de CSV/XLSX y reporte por fila
    ├── carnets.go                    # Selección por grado o sección y avisos previos
    ├── repositorio.go                # Consultas sqlc envueltas
    ├── modelo.go                     # Alumno, CredencialQR, HistorialSeccion
    └── normalizacion.go              # Texto sin tildes para búsqueda y validación

migraciones/
├── 0001_crear_tabla_usuarios.sql
├── 0002_crear_tablas_academicas.sql  # anios_escolares, grados, secciones (si 002 no existe aún)
├── 0003_crear_tabla_alumnos.sql
├── 0004_crear_tabla_credenciales_qr.sql
├── 0005_crear_tabla_historial_secciones.sql
├── 0006_crear_tablas_importacion.sql
└── 0007_crear_tabla_auditoria.sql

web/
├── plantillas/                       # templ: paginas/, componentes/, carnets/
└── estaticos/                        # CSS generado por Tailwind y JavaScript mínimo

pruebas/
├── unidad/                           # Reglas (sin base de datos)
├── integracion/                      # Repositorios con PostgreSQL real
├── manejadores/                      # httptest
└── e2e/                              # Playwright

docs/
└── glosario.md
```

**Decisión de estructura**: monolito modular bajo `internal/`, un módulo por dominio, con
manejador, servicio, repositorio y modelo en cada uno. La auditoría y el almacenamiento de fotos
son infraestructura compartida y viven en `plataforma`. Las pruebas de integración viven fuera de
los paquetes para no mezclarlas con las unitarias.

## Seguimiento de complejidad

| Violación o adición | Por qué es necesaria | Alternativa más simple descartada porque |
|---|---|---|
| Chi como router | Rutas agrupadas por módulo, intermediarios por rol y parámetros de ruta claros | El `ServeMux` de Go 1.22+ cubre rutas y métodos, pero la agrupación con intermediarios por rol queda más difusa. Se evalúa en la investigación. |
| templ + HTMX + Tailwind | Componentes tipados en Go, actualizaciones parciales sin SPA, estilos responsivos | `html/template` con CSS escrito a mano requiere más código de interfaz y es más difícil de mantener con varios roles. |
| aws-sdk-go-v2 (S3) | El proveedor de hosting no ofrece disco persistente; las fotos deben persistir fuera del contenedor | Guardar fotos en la base de datos aumenta el tamaño de los respaldos y complica su restauración. |
| excelize | Requerido por la spec para importar .xlsx | Solo CSV no cubre el archivo que usa el colegio. |

## Notas de consistencia

- La spec menciona solo el "DNI o código". En el modelo, la columna se llama `identificador`, con
  unicidad global (FR-003).
- El reemplazo de QR y la desactivación usan la misma transacción de invalidación.
- Los carnés se generan como HTML imprimible; el PDF queda para la feature 006 (ver investigación).
