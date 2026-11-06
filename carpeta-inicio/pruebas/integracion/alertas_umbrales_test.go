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

func TestUmbralesFaltasSeguidas_SeConfiguraYVersiona(t *testing.T) {
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
		Tipo:          academico.TipoUmbralFaltasSeguidas,
		Valor:         2,
		VigenteDesde:  hoy,
	}))

	vigentes, err := servicio.UmbralesVigentes(ctx, anioID, hoy)
	require.NoError(t, err)
	require.Len(t, vigentes, 1)
	assert.Equal(t, academico.TipoUmbralFaltasSeguidas, vigentes[0].Tipo)
	assert.Equal(t, 2, vigentes[0].Valor)

	require.NoError(t, servicio.ConfigurarUmbral(ctx, usuarioID, academico.UmbralAlerta{
		AnioEscolarID: anioID,
		Tipo:          academico.TipoUmbralFaltasSeguidas,
		Valor:         3,
		VigenteDesde:  hoy.AddDate(0, 0, 1),
	}))

	vigentes, err = servicio.UmbralesVigentes(ctx, anioID, hoy)
	require.NoError(t, err)
	require.Len(t, vigentes, 1)
	assert.Equal(t, 2, vigentes[0].Valor, "el cambio futuro no debe aplicarse antes de su fecha")

	vigentes, err = servicio.UmbralesVigentes(ctx, anioID, hoy.AddDate(0, 0, 1))
	require.NoError(t, err)
	require.Len(t, vigentes, 1)
	assert.Equal(t, 3, vigentes[0].Valor)
}
