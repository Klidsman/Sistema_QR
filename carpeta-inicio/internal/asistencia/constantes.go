package asistencia

import "time"

// Estados del registro personal del docente.
const (
	EstadoPresente         = "presente"
	EstadoTardanza         = "tardanza"
	EstadoFalta            = "falta"
	EstadoFaltaJustificada = "falta_justificada"
)

// Estados de una marcación oficial de ingreso por QR.
const (
	MarcacionPuntual       = "puntual"
	MarcacionTardanza      = "tardanza"
	MarcacionTardanzaGrave = "tardanza_grave"
)

// Entidades de auditoría de la feature.
const (
	EntidadDocente             = "docente"
	EntidadAsignacionDocente   = "asignacion_docente"
	EntidadRegistroDocente     = "registro_docente"
	EntidadParametroAsistencia = "parametro_asistencia"
	EntidadMarcacion           = "marcacion"
)

// Acciones de auditoría de la feature.
const (
	AccionCrearDocente       = "crear_docente"
	AccionDesactivarDocente  = "desactivar_docente"
	AccionRestablecerClave   = "restablecer_clave"
	AccionCrearAsignacion    = "crear_asignacion"
	AccionCerrarAsignacion   = "cerrar_asignacion"
	AccionCrearRegistro      = "crear_registro"
	AccionCorregirRegistro   = "corregir_registro"
	AccionMarcarTodos        = "marcar_todos_presentes"
	AccionConfigurarPlazo    = "configurar_plazo"
	AccionRegistrarMarcacion = "registrar_marcacion"
)

// EstadosRegistro devuelve los estados del docente válidos.
func EstadosRegistro() []string {
	return []string{EstadoPresente, EstadoTardanza, EstadoFalta, EstadoFaltaJustificada}
}

// Reloj devuelve la fecha y hora actuales. Se inyecta para pruebas deterministas.
type Reloj func() time.Time
