package unidad

import (
	"testing"

	"asistencia/internal/plataforma/configuracion"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func variablesValidas() map[string]string {
	return map[string]string{
		"BASE_DATOS_URL":         "postgres://usuario:clave@localhost:5432/asistencia",
		"CLAVE_SESION":           "clave_de_sesion_de_al_menos_32_caracteres_xx",
		"PUERTO":                 "9090",
		"CARNET_ANCHO_MM":        "85",
		"CARNET_ALTO_MM":         "54",
		"FOTO_TAMANIO_MAXIMO_MB": "2",
		"IMPORTACION_MAX_FILAS":  "5000",
		"ALMACENAMIENTO_TIPO":    "disco",
		"ALMACENAMIENTO_RUTA":    "./datos/fotos",
	}
}

func obtenerDesde(valores map[string]string) func(string) string {
	return func(clave string) string { return valores[clave] }
}

func TestCargar_Valida(t *testing.T) {
	cfg, err := configuracion.Cargar(obtenerDesde(variablesValidas()))
	require.NoError(t, err)

	assert.Equal(t, "postgres://usuario:clave@localhost:5432/asistencia", cfg.BaseDatosURL)
	assert.Equal(t, 9090, cfg.Puerto)
	assert.Equal(t, 85, cfg.CarnetAnchoMM)
	assert.Equal(t, 54, cfg.CarnetAltoMM)
	assert.Equal(t, 2, cfg.FotoTamanioMaximoMB)
	assert.Equal(t, 5000, cfg.ImportacionMaxFilas)
	assert.Equal(t, "disco", cfg.Almacenamiento.Tipo)
}

func TestCargar_FaltaBaseDatos(t *testing.T) {
	valores := variablesValidas()
	delete(valores, "BASE_DATOS_URL")

	_, err := configuracion.Cargar(obtenerDesde(valores))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "BASE_DATOS_URL")
}

func TestCargar_ClaveSesionCorta(t *testing.T) {
	valores := variablesValidas()
	valores["CLAVE_SESION"] = "corta"

	_, err := configuracion.Cargar(obtenerDesde(valores))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "CLAVE_SESION")
}

func TestCargar_ValoresPorDefecto(t *testing.T) {
	valores := variablesValidas()
	delete(valores, "PUERTO")
	delete(valores, "CARNET_ANCHO_MM")
	delete(valores, "FOTO_TAMANIO_MAXIMO_MB")
	delete(valores, "IMPORTACION_MAX_FILAS")

	cfg, err := configuracion.Cargar(obtenerDesde(valores))
	require.NoError(t, err)

	assert.Equal(t, 8080, cfg.Puerto)
	assert.Equal(t, 85, cfg.CarnetAnchoMM)
	assert.Equal(t, 2, cfg.FotoTamanioMaximoMB)
	assert.Equal(t, 5000, cfg.ImportacionMaxFilas)
}

func TestCargar_PuertoInvalido(t *testing.T) {
	valores := variablesValidas()
	valores["PUERTO"] = "abc"

	_, err := configuracion.Cargar(obtenerDesde(valores))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "PUERTO")
}

func TestCargar_TipoAlmacenamientoInvalido(t *testing.T) {
	valores := variablesValidas()
	valores["ALMACENAMIENTO_TIPO"] = "nube"

	_, err := configuracion.Cargar(obtenerDesde(valores))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ALMACENAMIENTO_TIPO")
}

func TestCargar_S3RequiereCredenciales(t *testing.T) {
	valores := variablesValidas()
	valores["ALMACENAMIENTO_TIPO"] = "s3"

	_, err := configuracion.Cargar(obtenerDesde(valores))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ALMACENAMIENTO_BUCKET")
}

func TestCargar_DimensionesDeCarnetPositivas(t *testing.T) {
	valores := variablesValidas()
	valores["CARNET_ANCHO_MM"] = "0"

	_, err := configuracion.Cargar(obtenerDesde(valores))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "CARNET_ANCHO_MM")
}
