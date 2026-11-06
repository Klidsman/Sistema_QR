# Investigación técnica: Gestión de alumnos y generación de códigos QR

**Fecha**: 2026-11-05 | **Plan**: [plan.md](plan.md)

Esta investigación no tenía marcadores `NEEDS CLARIFICATION` en la spec. Las decisiones
siguientes cubren las elecciones técnicas que la constitución y el stack aprobado dejaban abiertas.

## 1. Router HTTP: Chi

- **Decisión**: Chi.
- **Justificación**: el stack aprobado lo exige. Permite agrupar rutas por módulo, aplicar
  intermediarios por rol a grupos enteros (por ejemplo, todas las rutas de administración) y
  leer parámetros de ruta sin lógica repetida.
- **Alternativas consideradas**: `net/http` con `ServeMux` de Go 1.22+. Cubre métodos y
  parámetros, pero los intermediarios por grupo quedarían repetidos en cada ruta. Se descarta por
  la regla de permisos por rol.

## 2. Interfaz: templ + HTMX + Tailwind CSS

- **Decisión**: plantillas templ en `web/plantillas`, interacciones parciales con HTMX, estilos
  con Tailwind compilado a `web/estaticos`.
- **Justificación**: plantillas tipadas, errores de compilación en vez de errores en ejecución,
  y actualizaciones parciales (por ejemplo, la lista de alumnos al filtrar) sin escribir
  JavaScript propio.
- **Alternativas consideradas**: `html/template` con CSS propio. Se descarta por el costo de
  mantener responsividad en varias vistas.

## 3. Base de datos: PostgreSQL con pgx, sqlc y goose

- **Decisión**: pgx como driver, sqlc para consultas tipadas, goose para migraciones.
- **Justificación**: SQL explícito y tipado sin ORM; migraciones versionadas en `migraciones/`.
- **Alternativas consideradas**: GORM. Se descarta por el principio de no usar ORM.

## 4. Sesiones y autenticación

- **Decisión**: scs con almacenamiento en PostgreSQL (tabla `sesiones`). Contraseñas con
  argon2id (`golang.org/x/crypto/argon2`), parámetros fijados en configuración.
- **Justificación**: sesiones por cookie con Secure, HttpOnly y SameSite; argon2id es el
  algoritmo exigido por la constitución.
- **Límite de intentos**: contador en memoria por combinación de IP y usuario, con ventana de 15
  minutos. Es válido con una sola instancia, que es el supuesto del plan. Si se escala a varias
  instancias, el contador debe moverse a PostgreSQL. Queda registrado como pendiente.

## 5. Token QR

- **Decisión**: 32 bytes aleatorios de `crypto/rand`, codificados en base64 URL sin relleno (43
  caracteres). Se guardan en `credenciales_qr.token` con restricción de unicidad. El QR contiene
  solo ese token.
- **Justificación**: el token no revela nada del alumno y puede revocarse en la base de datos
  (constitución, principio IV).
- **Alternativas consideradas**: firmar el DNI con HMAC. Se descarta porque el DNI quedaría
  codificado en el QR (incumple FR-007) y no permitiría revocar sin lista de bloqueo.

## 6. Generación de QR

- **Decisión**: skip2/go-qrcode, salida PNG, tamaño fijo definido en configuración.
- **Justificación**: librería pequeña y sin dependencias pesadas; la spec pide QR únicos y sin
  datos personales.

## 7. Importación de CSV y XLSX

- **Decisión**: `encoding/csv` para CSV con detección de separador (coma o punto y coma) y de
  BOM UTF-8. excelize para XLSX, leyendo solo la primera hoja.
- **Límites**: tamaño máximo de archivo y número máximo de filas definidos en configuración
  (valores iniciales: 5 MB y 5,000 filas).
- **Reglas**: cada fila se valida por separado. Los duplicados dentro del archivo se detectan
  antes de insertar; los que ya existen en la base de datos se detectan con la restricción de
  unicidad y se reportan sin detener el proceso.
- **Alternativas consideradas**: librerías de importación genéricas. Se descartan por complejidad.

## 8. Búsqueda sin tildes y sin distinguir mayúsculas

- **Decisión**: al guardar un alumno, el servicio calcula columnas normalizadas
  (`nombres_busqueda`, `apellidos_busqueda`) en minúsculas y sin tildes, y se indexan. Las
  búsquedas usan las mismas columnas normalizadas.
- **Justificación**: la extensión `unaccent` de PostgreSQL no está disponible en todos los
  proveedores administrados. Normalizar en la aplicación evita acoplar la solución a un proveedor
  (constitución, principio VIII).
- **Alternativas consideradas**: `unaccent` o índices funcionales. Se descartan por la
  dependencia del proveedor.

## 9. Fotos y almacenamiento de objetos

- **Decisión**: interfaz `AlmacenFotos` con métodos `Guardar`, `Obtener` y `Eliminar`. Dos
  implementaciones: `AlmacenDisco` para desarrollo y `AlmacenS3` con aws-sdk-go-v2 para
  producción (Cloudflare R2 o Backblaze B2). Las fotos se sirven por proxy autenticado
  (`/alumnos/{id}/foto`), nunca con URL pública.
- **Formatos**: JPG y PNG, máximo 2 MB (valor en configuración `FOTO_TAMANIO_MAXIMO_MB`).
- **Justificación**: el proveedor de hosting no garantiza disco persistente; los objetos quedan
  fuera del contenedor y de los respaldos de la base.
- **Alternativas consideradas**: guardar la foto en la base de datos. Se descarta por el tamaño de
  los respaldos.

## 10. Carnés: HTML imprimible

- **Decisión**: los carnés se renderizan como una página HTML con `@page` y dimensiones definidas
  por `CARNET_ANCHO_MM` y `CARNET_ALTO_MM`. El administrador imprime desde el navegador.
- **Justificación**: la spec pide imprimir directamente desde el navegador (FR-021). Evita una
  librería de PDF en esta feature.
- **Pendiente**: el PDF de carnés y reportes se evalúa en la feature 006, con maroto o gofpdf. Se
  descarta chromedp salvo que el diseño lo exija.

## 11. Auditoría

- **Decisión**: tabla `auditoria` de solo escritura. Un disparador bloquea `UPDATE` y `DELETE`
  sobre esa tabla. El registrador vive en `plataforma/auditoria`; `alumnos` lo usa mediante una
  interfaz pequeña.
- **Justificación**: FR-026 y el principio IV exigen auditoría inmutable. El disparador funciona
  en cualquier proveedor de PostgreSQL.

## 12. Fechas y zona horaria

- **Decisión**: columnas `timestamptz` en UTC; presentación y cálculos de turno en
  `America/Lima`. Las fechas de cambio de sección y retiro son `date` en la zona de Lima.
- **Justificación**: principio IV y restricciones técnicas.

## 13. Pruebas

- **Decisión**: unitarias con `testing` y `testify` para validación de alumnos, normalización,
  reglas de QR (un activo, reemplazo, desactivación) e importación. Integración con
  testcontainers-go y PostgreSQL real. Manejadores con `httptest`. Playwright para alta,
  importación y vista de carnés.
- **Justificación**: principio V.

## 14. Configuración

- **Decisión**: variables de entorno en español: `BASE_DATOS_URL`, `PUERTO`, `CLAVE_SESION`,
  `CARNET_ANCHO_MM`, `CARNET_ALTO_MM`, `FOTO_TAMANIO_MAXIMO_MB`, `IMPORTACION_MAX_FILAS`,
  `ALMACENAMIENTO_TIPO` (disco o s3), y las credenciales de almacenamiento. La configuración se
  valida al iniciar.
- **Justificación**: principios VI y VIII.

## Pendientes derivados

- Escalar a varias instancias requiere mover el límite de intentos a PostgreSQL.
- El PDF de carnés se decide en la feature 006.
- La validación legal de la retención de fotos (TODO LEGAL_REVIEW) sigue abierta.
