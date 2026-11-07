package cuentas

// CuentaVista es una cuenta de usuario preparada para mostrarse.
type CuentaVista struct {
	ID     int64
	Nombre string
	Correo string
	Activo bool
}

// Datos reúne lo necesario para pintar el listado de cuentas de un rol.
type Datos struct {
	Usuario       string
	RolSesion     string
	TokenCSRF     string
	Titulo        string
	RutaBase      string
	Singular      string
	BotonCrear    string
	Mensaje       string
	MensajeError  string
	Cuentas       []CuentaVista
	PuedeEscribir bool
}
