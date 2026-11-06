// Package autenticacion gestiona los usuarios, sus contraseñas y el ingreso al sistema.
package autenticacion

import "errors"

// ErrCredencialesInvalidas se devuelve cuando el correo o la contraseña no coinciden.
// No distingue cuál de los dos es incorrecto, para no revelar cuentas existentes.
var ErrCredencialesInvalidas = errors.New("correo o contraseña incorrectos")

// Errores de gestión de docentes.
var (
	ErrDocenteExistente      = errors.New("ya existe una cuenta con ese correo")
	ErrDocenteNoEncontrado   = errors.New("el docente no existe")
	ErrDatosDocenteInvalidos = errors.New("los datos del docente no son válidos")
)

// DatosDocente son los datos para crear una cuenta de docente.
type DatosDocente struct {
	Correo      string
	Nombre      string
	Contrasenia string
}

// Usuario es la identidad autenticada que usa la aplicación.
type Usuario struct {
	ID             int64
	Correo         string
	NombreCompleto string
	Rol            string
	Activo         bool
}
