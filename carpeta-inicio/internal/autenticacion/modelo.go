// Package autenticacion gestiona los usuarios, sus contraseñas y el ingreso al sistema.
package autenticacion

import "errors"

// ErrCredencialesInvalidas se devuelve cuando el correo o la contraseña no coinciden.
// No distingue cuál de los dos es incorrecto, para no revelar cuentas existentes.
var ErrCredencialesInvalidas = errors.New("correo o contraseña incorrectos")

// Usuario es la identidad autenticada que usa la aplicación.
type Usuario struct {
	ID             int64
	Correo         string
	NombreCompleto string
	Rol            string
	Activo         bool
}
