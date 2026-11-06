package integracion

import (
	"context"
	"testing"

	"asistencia/internal/asistencia"
	"asistencia/internal/plataforma/auditoria"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsistenciaMigraciones_ParametroInicialYEntidadesDeAuditoria(t *testing.T) {
	pool := AbrirBaseDePruebas(t)
	usuarioID := CrearUsuarioPrueba(t, pool)
	ctx := context.Background()

	var horas int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT plazo_edicion_horas FROM parametros_asistencia WHERE id = 1`).Scan(&horas))
	assert.Equal(t, 24, horas, "el plazo inicial debe ser de 24 horas")

	require.NoError(t, auditoria.Registrar(ctx, pool, auditoria.Evento{
		UsuarioID: usuarioID,
		Entidad:   asistencia.EntidadRegistroDocente,
		EntidadID: 1,
		Accion:    asistencia.AccionCrearRegistro,
		ValorNuevo: map[string]any{
			"estado": asistencia.EstadoPresente,
		},
	}))

	var cantidad int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM auditoria
		 WHERE entidad = $1 AND accion = $2 AND usuario_id = $3`,
		asistencia.EntidadRegistroDocente, asistencia.AccionCrearRegistro, usuarioID,
	).Scan(&cantidad))
	assert.Equal(t, 1, cantidad)
}
