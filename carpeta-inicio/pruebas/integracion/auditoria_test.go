package integracion

import (
	"context"
	"testing"

	"asistencia/internal/autenticacion/generado"
	"asistencia/internal/plataforma/auditoria"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuditoria_RegistraYRechazaCambios(t *testing.T) {
	pool := AbrirBaseDePruebas(t)
	ctx := context.Background()

	usuarioID, err := generado.New(pool).CrearUsuario(ctx, generado.CrearUsuarioParams{
		Correo:          "admin@colegio.pe",
		ContraseniaHash: "hash-de-prueba",
		NombreCompleto:  "Administrador de prueba",
		Rol:             "administrador",
	})
	require.NoError(t, err)

	err = auditoria.Registrar(ctx, pool, auditoria.Evento{
		UsuarioID:     usuarioID,
		Entidad:       "alumno",
		EntidadID:     1,
		Accion:        "crear",
		ValorNuevo:    map[string]string{"nombres": "José"},
		ValorAnterior: nil,
	})
	require.NoError(t, err)

	_, err = pool.Exec(ctx, "UPDATE auditoria SET motivo = 'alterado'")
	assert.ErrorContains(t, err, "solo escritura")

	_, err = pool.Exec(ctx, "DELETE FROM auditoria")
	assert.ErrorContains(t, err, "solo escritura")
}
