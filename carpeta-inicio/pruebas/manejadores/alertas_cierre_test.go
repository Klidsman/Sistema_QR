package manejadores

import (
	"net/http"
	"net/url"
	"testing"

	"asistencia/internal/academico"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfiguracionUmbrales_PermiteFaltasSeguidas(t *testing.T) {
	var recibido academico.UmbralAlerta
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen: academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
		configurarUmbral: func(umbral academico.UmbralAlerta) error {
			recibido = umbral
			return nil
		},
	})
	peticion := peticionFormulario(t, sesiones, "/configuracion/umbrales", url.Values{
		"tipo":          {"faltas_seguidas"},
		"valor":         {"2"},
		"vigente_desde": {"2026-11-06"},
	}, "administrador")

	respuesta := ejecutarConfiguracion(manejador, peticion)

	require.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Equal(t, academico.TipoUmbralFaltasSeguidas, recibido.Tipo)
	assert.Equal(t, 2, recibido.Valor)
}

func TestConfiguracionUmbrales_MuestraLaTarjetaDeFaltasSeguidas(t *testing.T) {
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen: academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
		umbralesVigentes: []academico.UmbralAlerta{{
			ID: 1, AnioEscolarID: 1, Tipo: academico.TipoUmbralFaltasSeguidas, Valor: 2,
			VigenteDesde: umbralDePrueba(2).VigenteDesde,
		}},
	})
	peticion := peticionConSesion(t, sesiones, http.MethodGet, "/configuracion/umbrales", nil, "", "administrador")

	respuesta := ejecutarConfiguracion(manejador, peticion)

	require.Equal(t, http.StatusOK, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "Faltas seguidas")
}
