<!--
SYNC IMPACT REPORT (temporal, para revisión humana; eliminar antes de commit)

Version change: plantilla sin ratificar → 1.0.0 (primera ratificación)
Modified principles: no aplica (primera versión); todos los placeholders reemplazados
Added sections:
  - Core Principles I-V
  - Restricciones Técnicas
  - Flujo de Desarrollo y Calidad
  - Governance
Removed sections: ninguna

Follow-up TODOs / supuestos a confirmar:
  - Motor de base de datos: PostgreSQL o SQLite (se decide en la especificación de despliegue).
  - RATIFICATION_DATE asumido como 2026-11-05 (fecha de primera adopción).
  - Idioma: textos de interfaz en español; identificadores de código en inglés.
  - Valores numéricos (2 s, 500 ms, 400 líneas) son metas iniciales ajustables por enmienda MINOR.
-->

# Sistema de Asistencia Estudiantil QR Constitution

## Core Principles

### I. Simplicidad sin sobreingeniería

La solución MUST ser la más simple que cumpla la especificación aprobada (YAGNI).
Un monolito modular en Go es la arquitectura base; el sistema MUST NOT introducir
microservicios, colas de mensajes, cachés distribuidas ni capas de abstracción sin un
consumidor real. Cada dependencia externa MUST justificarse en el plan de la
funcionalidad, y la biblioteca estándar de Go MUST preferirse siempre que cubra la necesidad.
Cualquier complejidad adicional MUST registrarse en la sección "Complejidad justificada"
del plan.

Rationale: el sistema lo mantendrá un equipo pequeño; cada capa extra aumenta el costo de
cambio y de depuración sin beneficio para los usuarios.

### II. Código Go mantenible

- El código MUST pasar `gofmt` y `go vet` sin advertencias.
- Los paquetes MUST organizarse por dominio (asistencia, aulas, cursos, docentes,
  estudiantes, reportes), no por capa técnica genérica. Las dependencias entre paquetes
  MUST ser unidireccionales y sin ciclos.
- Las interfaces MUST definirse en el paquete consumidor y solo cuando exista más de una
  implementación o se necesite un doble de prueba.
- Los errores MUST propagarse con contexto (`fmt.Errorf` con `%w`) y MUST NOT ignorarse.
  El flujo de negocio MUST NOT usar `panic`.
- El logging MUST ser estructurado con `log/slog`.
- Los cambios de esquema MUST hacerse mediante migraciones versionadas.

Rationale: la mantenibilidad depende de límites claros entre dominios y de errores que
se puedan diagnosticar sin reproducir el problema.

### III. Experiencia de usuario centrada en el registro rápido

- El escaneo de un QR en portería MUST responder en menos de 500 ms (p95) en red local, y el
  usuario MUST percibir la confirmación en menos de 2 segundos.
- Cada resultado MUST mostrar un mensaje inequívoco en español y con color distintivo:
  asistencia registrada, ya registrada, QR vencido, QR no válido o estudiante no
  matriculado en el aula o curso.
- Las pantallas de portería y de profesor MUST ser mobile-first y utilizables en navegadores
  móviles modernos, con conexión intermitente tolerada en lo posible.
- Iniciar una sesión de asistencia en un aula MUST requerir no más de tres acciones del
  profesor.
- La interfaz MUST cumplir contraste WCAG AA y tener etiquetas accesibles en todos los
  controles.
- La pantalla del profesor MUST mostrar el docente vigente del aula para el curso y la
  fecha consultados, evitando confusiones tras un cambio de docente.

Rationale: la asistencia se toma en filas y en minutos; la lentitud o la ambigüedad
provocan colas en portería y registros erróneos.

### IV. Integridad y seguridad de la asistencia

- Cada código QR MUST estar firmado por el servidor (HMAC o equivalente), MUST tener
  vigencia limitada y MUST NOT revelar datos sensibles ni ser predecible.
- Un estudiante MUST tener como máximo un registro de asistencia por sesión (aula, curso,
  fecha y bloque). Las operaciones de registro MUST ser idempotentes.
- Los registros de asistencia MUST NOT eliminarse. Las correcciones MUST registrarse como
  ajustes con usuario responsable, fecha, motivo y valor anterior.
- Las asignaciones docente-aula-curso MUST tener vigencia (fecha de inicio y de fin). Los
  reportes históricos MUST usar el docente vigente en la fecha de la sesión.
- La autorización MUST ser por rol: estudiante, portero, profesor y administrador. Un
  profesor MUST registrar asistencia solo en sus aulas asignadas; un portero MUST
  operar solo el escaneo y la consulta de su puerta.
- Las contraseñas MUST almacenarse con un algoritmo de hash adaptativo (bcrypt o argon2
  mediante `golang.org/x/crypto`). Los secretos MUST cargarse desde variables de entorno y
  MUST NOT versionarse.
- La zona horaria del sistema MUST ser explícita y única para calcular fechas de sesión y
  reportes.

Rationale: la asistencia tiene consecuencias académicas; un registro duplicado, alterado
en silencio o falsificado afecta a estudiantes y docentes.

### V. Pruebas verificables y reproducibles

- Las reglas de negocio (validación de QR, ventanas de tiempo, asignación docente-aula,
  cálculo de reportes) MUST tener pruebas unitarias escritas antes de la implementación y
  MUST verse fallar antes de codificar la regla.
- Los repositorios SQL MUST probarse contra una base de datos real de pruebas, no contra
  mocks.
- Los handlers HTTP MUST probarse con `net/http/httptest`.
- Los flujos críticos MUST tener prueba de extremo a extremo: escaneo en portería, toma de
  asistencia por el profesor, cambio de docente y generación de reportes diarios y mensuales.
- Las pruebas MUST ser deterministas: el reloj MUST inyectarse, y las pruebas MUST NOT
  depender de red externa ni del orden de ejecución.
- `go test ./...` MUST pasar antes de fusionar cualquier cambio.

Rationale: el registro de asistencia se valida en la vida real; las pruebas deben
demostrar el comportamiento sin depender del entorno.

## Restricciones Técnicas

- Lenguaje: Go, en la versión estable declarada en `go.mod`.
- Servidor HTTP: `net/http` de la biblioteca estándar. Un router externo MUST justificarse
  solo si el mux estándar resulta insuficiente.
- Base de datos: relacional (PostgreSQL o SQLite, según la especificación de despliegue),
  con acceso mediante `database/sql` y SQL explícito. MUST NOT usarse un ORM pesado.
- Frontend: plantillas `html/template` renderizadas por el servidor, con JavaScript mínimo.
  Se MUST usar una única biblioteca aprobada para el lector de QR en el navegador. Una SPA
  MUST justificarse en la especificación.
- Despliegue: un solo binario con configuración por variables de entorno. No se requieren
  orquestadores en la primera versión.
- Idioma: textos de interfaz y reportes en español; identificadores de código en inglés.
- Reportes: diarios y mensuales, exportables a CSV, generados desde la misma fuente de
  asistencia. MUST NOT mantenerse tablas de reporte duplicadas salvo que una medición de
  rendimiento lo justifique.

## Flujo de Desarrollo y Calidad

- Todo cambio funcional MUST iniciar con una especificación que incluya criterios de
  aceptación, antes de planificar o codificar.
- Los planes MUST verificar el cumplimiento de los Core Principles y documentar cualquier
  excepción en "Complejidad justificada".
- Los cambios SHOULD ser pequeños (idealmente menos de 400 líneas modificadas) y MUST ser
  revisados por al menos una persona distinta de su autor.
- Quality gates obligatorios: `gofmt`, `go vet`, `go test -race ./...`. Un linter
  (por ejemplo `golangci-lint`) MAY usarse si su configuración se versiona en el repositorio.
- Todo cambio de esquema MUST incluir su migración.
- Antes de cada despliegue MUST validarse manualmente el flujo de portería y el flujo del
  profesor en un navegador móvil.

## Governance

- Esta constitución prevalece sobre cualquier otra práctica, guía o convención del
  proyecto. Ante un conflicto, MUST aplicarse la constitución.
- Enmiendas: se propone un cambio por escrito (pull request que edita este archivo) con
  justificación, impacto y, si aplica, plan de migración. La enmienda MUST ser aprobada por
  el responsable técnico del proyecto antes de fusionarse.
- Política de versionado (SemVer):
  - MAJOR: eliminación o redefinición incompatible de un principio o de una restricción.
  - MINOR: nuevo principio o sección, o ampliación material de una guía existente.
  - PATCH: aclaraciones, correcciones de redacción o refinamientos no semánticos.
- Revisión de cumplimiento: cada pull request MUST verificar los principios I a V. Los
  planes MUST justificar cualquier complejidad adicional. Esta constitución SHOULD revisarse
  al cerrar cada fase de especificación y, como mínimo, una vez por semestre.

**Version**: 1.0.0 | **Ratified**: 2026-11-05 | **Last Amended**: 2026-11-05
