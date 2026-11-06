package autenticacion

// Docente son los datos mínimos de una cuenta con rol docente.
type Docente struct {
	ID     int64
	Correo string
	Nombre string
	Activo bool
	Rol    string
}
