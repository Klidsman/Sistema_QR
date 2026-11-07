package integracion

import (
	"testing"

	"asistencia/internal/autenticacion"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCrearCuentaPorteria_GuardaElRolYPermiteAutenticarse(t *testing.T) {
	pool, _ := AbrirBaseDePruebasConURL(t)
	crearCuentasDeLogin(t, pool)
	servicio := autenticacion.NuevoServicio(pool)

	var adminID int64
	require.NoError(t, pool.QueryRow(t.Context(),
		`SELECT id FROM usuarios WHERE correo = 'admin@login.pe'`).Scan(&adminID))

	id, err := servicio.CrearCuenta(t.Context(), adminID, "porteria", autenticacion.DatosCuenta{
		Nombre:      "Portero de Prueba",
		Correo:      "portero.nuevo@login.pe",
		Contrasenia: claveLoginPrueba,
	})
	require.NoError(t, err)
	require.Positive(t, id)

	cuentas, err := servicio.ListarCuentasPorRol(t.Context(), "porteria")
	require.NoError(t, err)
	require.Len(t, cuentas, 2)
	assert.Contains(t, cuentas, autenticacion.Cuenta{
		ID: id, Correo: "portero.nuevo@login.pe", Nombre: "Portero de Prueba", Activo: true, Rol: "porteria",
	})

	usuario, err := servicio.Autenticar(t.Context(), "portero.nuevo@login.pe", claveLoginPrueba)
	require.NoError(t, err)
	assert.Equal(t, "porteria", usuario.Rol)
}

func TestCrearCuenta_RechazaRolesFueraDeDocenteYPorteria(t *testing.T) {
	pool, _ := AbrirBaseDePruebasConURL(t)
	crearCuentasDeLogin(t, pool)
	servicio := autenticacion.NuevoServicio(pool)

	_, err := servicio.CrearCuenta(t.Context(), 1, "administrador", autenticacion.DatosCuenta{
		Nombre:      "Otro administrador",
		Correo:      "otro.admin@login.pe",
		Contrasenia: claveLoginPrueba,
	})
	assert.ErrorIs(t, err, autenticacion.ErrRolNoPermitido)
}
