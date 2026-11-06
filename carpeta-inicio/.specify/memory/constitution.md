# Constitución del Sistema de Asistencia Estudiantil QR

## Core Principles

### I. Simplicidad sin sobreingeniería

La solución MUST ser la más simple que cumpla la especificación aprobada (YAGNI).
El sistema es un monolito modular en Go con un solo binario y un solo despliegue.
El sistema MUST NOT introducir microservicios, colas externas, Redis, Kubernetes, capas de
abstracción ni dependencias "por si acaso": cada elemento MUST tener un consumidor real y
justificarse en el plan. La biblioteca estándar de Go MUST preferirse siempre que cubra la
necesidad. Cualquier complejidad adicional MUST registrarse en la sección "Complejidad
justificada" del plan.

Rationale: el sistema lo mantendrá un tercero sin personal técnico interno; cada capa o
dependencia extra aumenta el costo de cambio, de despliegue y de depuración.

### II. Código Go mantenible y monolito modular

- El código MUST pasar `gofmt` y `go vet` sin advertencias.
- Los módulos MUST organizarse por dominio dentro de `/internal/<modulo>`: autenticacion,
  alumnos, academico, asistencia, aula, alertas, justificaciones y reportes. La infraestructura
  compartida MUST vivir en `/internal/plataforma`: configuracion, basedatos, registro,
  intermediarios, programador y eventos.
- Cada módulo MUST seguir capas simples: manejador (HTTP), servicio (reglas de negocio),
  repositorio (acceso a datos) y modelo. Las dependencias entre módulos MUST ser
  unidireccionales y sin ciclos.
- Un módulo MUST NOT consultar las tablas de otro. Los módulos MUST comunicarse solo mediante
  interfaces pequeñas o eventos internos.
- La lógica de negocio MUST vivir en el servicio, sin depender de HTTP ni de la base de datos,
  para probarla de forma aislada.
- Las interfaces MUST crearse solo cuando sean necesarias para pruebas o para desacoplar
  módulos.
- Los errores MUST propagarse con contexto (`fmt.Errorf` con `%w`) y MUST NOT ignorarse. El
  flujo de negocio MUST NOT usar `panic`.
- Los cambios de esquema MUST hacerse mediante migraciones versionadas con goose.

Rationale: los límites entre módulos evitan que un cambio en un dominio rompa otro, y las
capas fijas hacen predecible dónde buscar cada responsabilidad.

### III. Experiencia de usuario centrada en el registro rápido

- El escaneo de un QR en portería MUST responder en menos de 500 ms (p95) en red local, y el
  usuario MUST percibir la confirmación en menos de 2 segundos.
- La pantalla de escaneo MUST tener botones grandes, información legible a distancia y no
  MUST exigir pasos innecesarios en hora punta.
- El escaneo MUST admitir dos entradas con la misma ruta de validación: cámara del dispositivo
  con lectura continua (ZXing en el navegador), y lector USB o Bluetooth con un campo de
  captura con autofoco permanente que procesa la lectura al recibir Enter.
- Cada resultado MUST mostrar un mensaje inequívoco en español y con color distintivo:
  asistencia registrada, tardanza, ya registrada, QR vencido, QR no válido o alumno no
  matriculado en el aula o curso.
- La interfaz MUST ser responsiva y utilizable en computadora, tablet y celular, tanto en
  orientación vertical como horizontal, con diseño mobile-first.
- La interfaz MUST funcionar en las últimas dos versiones estables de Chrome, Edge, Firefox
  y Safari, sin instalar software adicional.
- Iniciar una sesión de asistencia en un aula MUST requerir no más de tres acciones del
  docente.
- La interfaz MUST cumplir contraste WCAG AA y tener etiquetas accesibles en todos los
  controles.
- El modo sin conexión (PWA con Service Worker e IndexedDB, con sincronización idempotente)
  es una funcionalidad posterior y MUST justificarse en su propia especificación.

Rationale: la asistencia se toma en filas y en minutos; la lentitud o la ambigüedad provocan
colas en portería y registros erróneos.

### IV. Integridad, auditoría y seguridad de la asistencia

- Cada código QR MUST contener únicamente un token aleatorio y no predecible. El token MUST
  almacenarse en la base de datos, MUST poder revocarse y MUST NOT contener nombres, DNI,
  identificadores internos ni otros datos personales.
- Un alumno MUST tener como máximo un QR activo a la vez. Revocar un QR MUST invalidarlo de
  inmediato.
- Un alumno MUST tener como máximo un registro de asistencia por sesión (aula, curso, fecha y
  bloque). Las operaciones de registro MUST ser idempotentes.
- Los registros de asistencia MUST NOT eliminarse. Las correcciones MUST registrarse como
  ajustes.
- Todo cambio sobre un registro de asistencia MUST quedar auditado en una tabla de auditoría
  propia, de solo escritura, con usuario, fecha y hora, valor anterior, valor nuevo y motivo.
- Las asignaciones docente-curso-sección MUST tener vigencia (fecha de inicio y de fin). Los
  reportes históricos MUST usar el docente vigente en la fecha de la sesión.
- La autenticación MUST usar sesiones por cookie (scs). Las cookies MUST ser Secure,
  HttpOnly y SameSite. Las contraseñas MUST almacenarse con argon2id.
- Todos los formularios que cambian estado MUST estar protegidos contra CSRF.
- El inicio de sesión MUST tener límite de intentos (rate limiting).
- La autorización MUST ser por rol mediante un intermediario propio: administrador, director,
  docente y porteria. Un docente MUST registrar solo en sus cursos y secciones asignados; un
  portero MUST operar solo el escaneo de la puerta principal.
- Los secretos MUST cargarse desde variables de entorno y MUST NOT versionarse.

Rationale: la asistencia tiene consecuencias académicas; un registro duplicado, alterado en
silencio o falsificado afecta a alumnos y docentes.

### V. Pruebas verificables y reproducibles

- Las reglas de negocio (clasificación puntual, tardanza o falta; duplicados; cierre de turno;
  validación de QR; ventanas de tiempo; cálculo de reportes) MUST tener pruebas unitarias
  escritas con `testing` y `testify`, con tablas de casos que incluyan los casos límite de la
  especificación. Las pruebas MUST escribirse antes de la implementación de la regla.
- Las reglas de negocio MUST poder probarse sin interfaz ni base de datos.
- Los repositorios MUST probarse contra PostgreSQL real, usando testcontainers o un contenedor
  de Docker Compose. No MUST usarse mocks para el acceso a datos.
- Los handlers HTTP MUST probarse con `net/http/httptest`.
- Los flujos críticos MUST tener pruebas de extremo a extremo con Playwright: escaneo en
  portería, registro de asistencia por el docente, cambio de docente y generación de reportes.
  Las vistas responsivas MUST verificarse en celular, tablet y computadora.
- Las pruebas MUST ser deterministas: el reloj MUST inyectarse, y las pruebas MUST NOT depender
  de red externa ni del orden de ejecución.
- `go test ./...` MUST pasar antes de fusionar cualquier cambio.

Rationale: el registro de asistencia se valida en la vida real; las pruebas deben demostrar el
comportamiento sin depender del entorno.

### VI. Reglas de negocio configurables y aisladas

- Toda regla configurable (horarios de entrada, tolerancias, duración de bloques, vigencia del
  QR, umbrales de alerta) MUST vivir en configuración externa al binario. Ningún valor de regla
  MUST escribirse como constante en el código.
- La lógica de reglas MUST estar en funciones o tipos puros que reciban su configuración y el
  instante de evaluación como parámetros. Estas reglas MUST NOT importar paquetes de interfaz,
  de base de datos ni de red.
- Cada cambio de configuración de reglas MUST quedar auditado y MUST NOT modificar
  retroactivamente la clasificación de asistencias ya registradas.

Rationale: el colegio cambia horarios y tolerancias con el tiempo; separar reglas de
infraestructura permite cambiarlas sin reescribir código y probarlas en segundos.

### VII. Protección de datos de menores de edad

- El sistema trata datos de menores de edad. Cada rol MUST tener solo los permisos necesarios
  para su función (mínimo privilegio), y todo acceso a datos personales MUST ser autenticado.
- El sistema MUST cumplir las obligaciones técnicas de la Ley N.° 29733, Ley de Protección de
  Datos Personales, y su Reglamento (D.S. N.° 003-2013-JUS), en particular finalidad,
  proporcionalidad, seguridad, confidencialidad y derechos de acceso, rectificación,
  cancelación y oposición.
- El sistema MUST recolectar solo los datos necesarios para registrar asistencia. Los
  reportes y las exportaciones MUST contener solo los datos requeridos por su finalidad.
- Toda comunicación MUST usar HTTPS en producción.
- Las fotos de alumnos y los carnés MUST almacenarse en un almacenamiento de objetos compatible
  con S3, detrás de una interfaz de almacenamiento. En desarrollo local MAY usarse una carpeta
  en disco. Los objetos MUST NOT ser públicos.
- Los datos MUST alojarse en una ubicación documentada. Si el alojamiento está fuera del Perú,
  la transferencia MUST quedar documentada y justificada ante el cliente.
- El cliente (colegio) es responsable del tratamiento de los datos; el proveedor del software
  actúa como encargado y MUST documentar las medidas técnicas aplicadas.

Rationale: los datos de menores exigen el nivel más alto de cuidado, y una filtración tendría
consecuencias legales y de confianza para el colegio.

### VIII. Entrega operable por terceros

- El sistema MUST desplegarse mediante una imagen Docker multietapa pequeña. La aplicación MUST
  exponer un endpoint `/salud` que verifique la disponibilidad de la aplicación y de la base
  de datos.
- La configuración MUST hacerse solo mediante variables de entorno en español (por ejemplo
  BASE_DATOS_URL, PUERTO, CLAVE_SESION), documentadas en un archivo `.env.ejemplo` sin valores
  reales. La aplicación MUST validar la configuración al iniciar.
- El sistema MUST NOT depender de un proveedor concreto de hosting. Cambiar de proveedor MUST
  requerir solo cambios de configuración y de documentación, no de código.
- La entrega MUST incluir: README de despliegue paso a paso, `.env.ejemplo`, guía de respaldo
  y restauración, glosario y manual de usuario por rol (portero, docente, director y
  administrador).
- El despliegue MUST hacerse en una plataforma administrada con PostgreSQL administrado y
  HTTPS, en cuentas a nombre del cliente. Las credenciales de administración MUST entregarse al
  cliente y MUST NOT quedar únicamente en poder del proveedor.
- Las migraciones de base de datos MUST ejecutarse de forma controlada y documentada, sin
  intervención manual sobre tablas.
- La operación MUST poder realizarse siguiendo solo la documentación entregada, sin personal
  técnico interno del colegio.

Rationale: el colegio no tiene personal técnico; si el sistema no puede desplegarse ni operarse
con la documentación, depende indefinidamente del proveedor.

### IX. Idioma y nomenclatura en español

- Todo el código propio del proyecto MUST nombrarse en español: carpetas, paquetes, archivos,
  tipos, estructuras, funciones, métodos, variables, constantes, interfaces, tablas, columnas,
  migraciones, rutas URL, plantillas, comentarios, mensajes de commit y documentación.
- Son excepciones aceptadas las impuestas por Go o por sus librerías: palabras reservadas
  (`func`, `if`, `for`, `return`, `struct`, `interface`), `package main`, `func main()`, el
  sufijo `_test.go`, el prefijo `Test` de las funciones de prueba, los tipos y funciones de
  librerías (`http.Request`, `context.Context`) y los nombres idiomáticos `err` y `ctx`.
- Los identificadores MUST NOT llevar tildes, ñ ni espacios. Ejemplos: `anio_escolar`, `seccion`,
  `contrasenia`, `ensenanza`. Los textos que ve el usuario MUST llevar tildes y ñ.
- Las carpetas y paquetes MUST escribirse en minúsculas, sin guiones ni guiones bajos
  (`alumnos`, `alertas`, `asistencia`). Los archivos MUST usar snake_case en español
  (`manejador.go`, `servicio.go`, `clasificador.go`).
- Los tipos exportados MUST usar PascalCase y los privados camelCase, ambos en español
  (`Alumno`, `Marcacion`, `ReglasTurno`, `clasificarMarcacion`).
- Las tablas MUST escribirse en plural y las columnas en snake_case (`alumnos`, `marcaciones`,
  `fecha_ingreso`). Las migraciones MUST numerarse con cuatro dígitos y un nombre descriptivo
  (`0001_crear_tabla_alumnos.sql`).
- Las rutas URL MUST estar en español y sin tildes (`/alumnos`, `/asistencia/escaneo`,
  `/reportes/mensual`).
- Cada concepto MUST usar un único término según el glosario de `docs/glosario.md`. Los valores
  almacenados MUST usar estas constantes en español: `puntual`, `tardanza`, `tardanza_grave`,
  `falta`, `falta_justificada`. Los roles MUST ser `administrador`, `director`, `docente` y
  `porteria`.

Rationale: un solo idioma y un solo término por concepto reduce la confusión del equipo de
mantenimiento y permite que el colegio entienda el código de su propio sistema.

## Alcance del Producto

- El sistema registra asistencia únicamente de alumnos. No registra asistencia de docentes ni
  de auxiliares.
- El sistema NO envía notificaciones a apoderados ni por correo, SMS o WhatsApp en esta
  versión. Cualquier notificación futura MUST requerir una enmienda de esta constitución.
- Los alumnos NO operan el sistema: presentan su código QR para ser escaneado.

## Restricciones Técnicas

- Lenguaje: Go en la versión estable reciente, declarada en `go.mod`. Monolito modular: un
  solo binario en `/cmd/servidor`.
- Router HTTP: Chi. Su uso se justifica por la necesidad de rutas agrupadas, intermediarios
  por rol y parámetros de ruta claros en un monolito con varios módulos.
- Frontend renderizado en el servidor: templ para componentes y plantillas, HTMX para
  interacciones parciales y Tailwind CSS para estilos. Se MUST usar poco JavaScript propio,
  solo donde no haya alternativa. Los archivos JavaScript propios MUST estar en `/web/estaticos`.
- Lector de QR en el navegador: ZXing como única biblioteca de lectura por cámara. No MUST
  depender solo de BarcodeDetector.
- Base de datos: PostgreSQL con pgx y sqlc para consultas SQL tipadas. Las consultas sqlc MUST
  usar nombres en español. No MUST usarse un ORM.
- Migraciones: goose, versionadas en `/migraciones`.
- Fechas y horas: almacenadas en UTC. Se presentan y se calculan turnos en la zona horaria
  America/Lima.
- Sesiones: scs. Contraseñas: argon2id.
- Logs: slog con mensajes en español.
- Tareas programadas: robfig/cron dentro de la aplicación, idempotente y protegida con advisory
  lock de PostgreSQL para evitar ejecuciones dobles con varias instancias.
- Tiempo real: SSE para las alertas, con polling como respaldo.
- Generación de documentos: skip2/go-qrcode para QR; excelize para Excel; maroto o gofpdf para
  PDF de reportes y carnés. chromedp solo si el diseño del carné lo exige, y MUST justificarse.
- Almacenamiento de fotos y carnés: interfaz de almacenamiento con implementación compatible
  con S3 (Cloudflare R2 o Backblaze B2) y carpeta en disco para desarrollo.
- Desarrollo local: Docker Compose para PostgreSQL; la aplicación se ejecuta con air. Un
  Makefile MUST ofrecer `ejecutar`, `migrar`, `probar` y `lint`.
- Calidad: golangci-lint configurado en el repositorio y formato con gofmt.
- Despliegue: imagen Docker multietapa. El proveedor concreto se documenta en el README de
  despliegue, no en el código.
- Integración continua: GitHub Actions con lint, pruebas y construcción de la imagen.
- Roles de cuenta: administrador, director, docente y porteria, con permisos mínimos
  documentados en el manual de cada rol.
- Reportes: diarios y mensuales, exportables a PDF y Excel, generados desde la misma fuente de
  asistencia. MUST NOT mantenerse tablas de reporte duplicadas salvo que una medición de
  rendimiento lo justifique.
- Simplicidad: sin Redis, colas externas ni Kubernetes. Una pieza nueva MUST justificarse en el
  plan de la funcionalidad que la requiere.

## Flujo de Desarrollo y Calidad

- Todo cambio funcional MUST iniciar con una especificación que incluya criterios de
  aceptación, antes de planificar o codificar.
- Los planes MUST verificar el cumplimiento de los Core Principles y documentar cualquier
  excepción en "Complejidad justificada".
- Los cambios SHOULD ser pequeños (idealmente menos de 400 líneas modificadas) y MUST ser
  revisados por al menos una persona distinta de su autor.
- Los mensajes de commit MUST escribirse en español.
- Quality gates obligatorios: `gofmt`, `go vet`, `golangci-lint`, `go test -race ./...` y la
  construcción de la imagen Docker en la integración continua.
- Todo cambio de esquema MUST incluir su migración de goose y su prueba de aplicación.
- Antes de cada entrega MUST validarse manualmente el flujo de portería y el flujo del docente
  en un celular, una tablet y una computadora, en cada navegador soportado.
- Cada cambio que afecte uso o despliegue MUST actualizar el README de despliegue, `.env.ejemplo`,
  el glosario o el manual por rol correspondiente.

## Governance

- Esta constitución prevalece sobre cualquier otra práctica, guía o convención del proyecto.
  Ante un conflicto, MUST aplicarse la constitución.
- Enmiendas: se propone un cambio por escrito (pull request que edita este archivo) con
  justificación, impacto y, si aplica, plan de migración. La enmienda MUST ser aprobada por el
  responsable técnico del proyecto. Cambios que afecten tratamiento de datos personales MUST
  además ser aprobados por el representante del colegio.
- Política de versionado (SemVer):
  - MAJOR: eliminación o redefinición incompatible de un principio o de una restricción.
  - MINOR: nuevo principio o sección, o ampliación material de una guía existente.
  - PATCH: aclaraciones, correcciones de redacción o refinamientos no semánticos.
- Revisión de cumplimiento: cada pull request MUST verificar los principios I a IX. Los planes
  MUST justificar cualquier complejidad adicional. Esta constitución SHOULD revisarse al cerrar
  cada fase de especificación, al menos una vez por año y antes de cada cambio de proveedor.

**Version**: 2.0.0 | **Ratified**: 2026-11-05 | **Last Amended**: 2026-11-05
