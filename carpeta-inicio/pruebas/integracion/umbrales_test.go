package integracion

import (
	"context"
	"testing"
	"time"

	"asistencia/internal/academico"
	"asistencia/internal/plataforma/fechas"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUmbrales_VersionaConsultaYAudita(t *testing.T) {
	pool := AbrirBaseDePruebas(t)
	usuarioID := CrearUsuarioPrueba(t, pool)
	servicio := academico.NuevoServicioConfiguracion(pool)
	ctx := context.Background()
	hoy := fechas.Hoy()

	anio := time.Now().Year() + 1
	anioID, err := servicio.CrearAnio(ctx, usuarioID, academico.DatosAnio{
		Nombre:      time.Date(anio, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006"),
		FechaInicio: time.Date(anio, time.March, 1, 0, 0, 0, 0, time.UTC),
		FechaFin:    time.Date(anio, time.December, 15, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)

	require.NoError(t, servicio.ConfigurarUmbral(ctx, usuarioID, academico.UmbralAlerta{
		AnioEscolarID: anioID,
		Tipo:          academico.TipoUmbralTardanzas,
		Valor:         3,
		VigenteDesde:  hoy,
	}))

	vigentes, err := servicio.UmbralesVigentes(ctx, anioID, hoy)
	require.NoError(t, err)
	require.Len(t, vigentes, 1)
	assert.Equal(t, 3, vigentes[0].Valor)

	futuro := hoy.AddDate(0, 1, 0)
	require.NoError(t, servicio.ConfigurarUmbral(ctx, usuarioID, academico.UmbralAlerta{
		AnioEscolarID: anioID,
		Tipo:          academico.TipoUmbralTardanzas,
		Valor:         4,
		VigenteDesde:  futuro,
	}))

	vigentes, err = servicio.UmbralesVigentes(ctx, anioID, hoy)
	require.NoError(t, err)
	require.Len(t, vigentes, 1)
	assert.Equal(t, 3, vigentes[0].Valor, "el cambio futuro no debe aplicarse antes de su fecha")

	vigentes, err = servicio.UmbralesVigentes(ctx, anioID, futuro)
	require.NoError(t, err)
	require.Len(t, vigentes, 1)
	assert.Equal(t, 4, vigentes[0].Valor)

	historial, err := servicio.ListarUmbrales(ctx, anioID)
	require.NoError(t, err)
	assert.Len(t, historial, 2)

	require.NoError(t, servicio.ConfigurarUmbral(ctx, usuarioID, academico.UmbralAlerta{
		AnioEscolarID: anioID,
		Tipo:          academico.TipoUmbralTardanzas,
		Valor:         5,
		VigenteDesde:  hoy,
	}), "un cambio con la misma fecha debe reemplazar la versión")
	vigentes, err = servicio.UmbralesVigentes(ctx, anioID, hoy)
	require.NoError(t, err)
	require.Len(t, vigentes, 1)
	assert.Equal(t, 5, vigentes[0].Valor)

	for _, valor := range []int{0, -1, 101} {
		err := servicio.ConfigurarUmbral(ctx, usuarioID, academico.UmbralAlerta{
			AnioEscolarID: anioID,
			Tipo:          academico.TipoUmbralFaltas,
			Valor:         valor,
			VigenteDesde:  hoy,
		})
		assert.ErrorIs(t, err, academico.ErrValorUmbralInvalido, "valor %d", valor)
	}

	assert.ErrorIs(t, servicio.ConfigurarUmbral(ctx, usuarioID, academico.UmbralAlerta{
		AnioEscolarID: anioID,
		Tipo:          "otro",
		Valor:         3,
		VigenteDesde:  hoy,
	}), academico.ErrTipoInvalido)

	pasado := hoy.AddDate(0, 0, -1)
	assert.ErrorIs(t, servicio.ConfigurarUmbral(ctx, usuarioID, academico.UmbralAlerta{
		AnioEscolarID: anioID,
		Tipo:          academico.TipoUmbralFaltas,
		Valor:         3,
		VigenteDesde:  pasado,
	}), academico.ErrVigenciaPasada)

	var auditorias int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM auditoria WHERE entidad = $1 AND accion = $2`,
		academico.EntidadUmbral, academico.AccionConfigurarUmbral,
	).Scan(&auditorias))
	assert.GreaterOrEqual(t, auditorias, 3)

	require.NoError(t, servicio.CerrarAnio(ctx, usuarioID, anioID))
	err = servicio.ConfigurarUmbral(ctx, usuarioID, academico.UmbralAlerta{
		AnioEscolarID: anioID,
		Tipo:          academico.TipoUmbralFaltas,
		Valor:         3,
		VigenteDesde:  hoy,
	})
	assert.ErrorIs(t, err, academico.ErrAnioCerrado)
}
