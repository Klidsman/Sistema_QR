// Package autenticacion gestiona los usuarios, sus contraseñas y el ingreso al sistema.
package autenticacion

import "errors"

// ErrCredencialesInvalidas se devuelve cuando el correo o la contraseña no coinciden.
// No distingue cuál de los dos es incorrecto, para no revelar cuentas existentes.
var ErrCredencialesInvalidas = errors.New("correo o contraseña incorrectos")

// Errores de gestión de cuentas.
var (
	ErrCuentaExistente      = errors.New("ya existe una cuenta con ese correo")
	ErrCuentaNoEncontrada   = errors.New("la cuenta no existe")
	ErrDatosCuentaInvalidos = errors.New("los datos de la cuenta no son válidos")
	ErrRolNoPermitido       = errors.New("solo se pueden crear cuentas de docente o de portería")
)

// Compatibilidad con el módulo de docentes.
var (
	ErrDocenteExistente      = ErrCuentaExistente
	ErrDocenteNoEncontrado   = ErrCuentaNoEncontrada
	ErrDatosDocenteInvalidos = ErrDatosCuentaInvalidos
)

// DatosCuenta son los datos para crear una cuenta de docente o de portería.
type DatosCuenta struct {
	Correo      string
	Nombre      string
	Contrasenia string
}

// DatosDocente se mantiene como alias de DatosCuenta.
type DatosDocente = DatosCuenta

// Usuario es la identidad autenticada que usa la aplicación.
type Usuario struct {
	ID             int64
	Correo         string
	NombreCompleto string
	Rol            string
	Activo         bool
}
