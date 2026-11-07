package autenticacion

// Cuenta son los datos mínimos de una cuenta con rol docente o de portería.
type Cuenta struct {
	ID     int64
	Correo string
	Nombre string
	Activo bool
	Rol    string
}

// Docente se mantiene como alias de Cuenta para las pantallas de asistencia.
type Docente = Cuenta
