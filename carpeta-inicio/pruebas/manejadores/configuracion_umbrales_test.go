package manejadores

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"asistencia/internal/academico"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func umbralDePrueba(valor int) academico.UmbralAlerta {
	return academico.UmbralAlerta{
		ID:            2,
		AnioEscolarID: 1,
		Tipo:          academico.TipoUmbralTardanzas,
		Valor:         valor,
		VigenteDesde:  time.Date(2026, time.November, 6, 0, 0, 0, 0, time.UTC),
	}
}

func TestConfiguracionUmbrales_MuestraLosVigentesYElHistorial(t *testing.T) {
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen:          academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
		umbralesVigentes: []academico.UmbralAlerta{umbralDePrueba(3)},
		historialUmbrales: []academico.UmbralAlerta{
			umbralDePrueba(3),
			{ID: 1, AnioEscolarID: 1, Tipo: academico.TipoUmbralTardanzas, Valor: 2,
				VigenteDesde: time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)},
		},
	})
	peticion := peticionConSesion(t, sesiones, http.MethodGet, "/configuracion/umbrales", nil, "", "director")

	respuesta := ejecutarConfiguracion(manejador, peticion)

	assert.Equal(t, http.StatusOK, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "Tardanzas por mes")
	assert.Contains(t, respuesta.Body.String(), "Historial de umbrales")
	assert.Contains(t, respuesta.Body.String(), "3")
}

func TestConfiguracionUmbrales_AdministradorConfigura(t *testing.T) {
	var recibido academico.UmbralAlerta
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen: academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
		configurarUmbral: func(umbral academico.UmbralAlerta) error {
			recibido = umbral
			return nil
		},
	})
	peticion := peticionFormulario(t, sesiones, "/configuracion/umbrales", url.Values{
		"tipo":          {"tardanzas"},
		"valor":         {"4"},
		"vigente_desde": {"2026-11-06"},
	}, "administrador")

	respuesta := ejecutarConfiguracion(manejador, peticion)

	require.Equal(t, http.StatusSeeOther, respuesta.Code)
	assert.Equal(t, "/configuracion/umbrales?ok=umbral", respuesta.Header().Get("Location"))
	assert.Equal(t, academico.TipoUmbralTardanzas, recibido.Tipo)
	assert.Equal(t, 4, recibido.Valor)
}

func TestConfiguracionUmbrales_DirectorNoPuedeConfigurar(t *testing.T) {
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen: academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
	})
	peticion := peticionFormulario(t, sesiones, "/configuracion/umbrales", url.Values{
		"tipo":          {"tardanzas"},
		"valor":         {"4"},
		"vigente_desde": {"2026-11-06"},
	}, "director")

	respuesta := ejecutarConfiguracion(manejador, peticion)

	assert.Equal(t, http.StatusForbidden, respuesta.Code)
}

func TestConfiguracionUmbrales_RechazaValorInvalido(t *testing.T) {
	manejador, sesiones := nuevoManejadorConfiguracionConFalso(t, configuracionFalsa{
		resumen: academico.ResumenConfiguracion{Anio: anioActivoDePrueba()},
		configurarUmbral: func(academico.UmbralAlerta) error {
			return academico.ErrValorUmbralInvalido
		},
	})
	peticion := peticionFormulario(t, sesiones, "/configuracion/umbrales", url.Values{
		"tipo":          {"tardanzas"},
		"valor":         {"0"},
		"vigente_desde": {"2026-11-06"},
	}, "administrador")

	respuesta := ejecutarConfiguracion(manejador, peticion)

	assert.Equal(t, http.StatusUnprocessableEntity, respuesta.Code)
	assert.Contains(t, respuesta.Body.String(), "El valor del umbral debe ser un entero entre 1 y 100")
}
