// Package configuracion lee y valida las variables de entorno al iniciar la aplicación.
package configuracion

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	tipoAlmacenamientoDisco = "disco"
	tipoAlmacenamientoS3    = "s3"
	longitudMinimaClave     = 32
)

// Almacenamiento agrupa la configuración de las fotos y los carnés.
type Almacenamiento struct {
	Tipo         string
	Ruta         string
	Bucket       string
	Endpoint     string
	ClaveAcceso  string
	ClaveSecreta string
}

// Configuracion contiene los valores de ejecución de la aplicación.
type Configuracion struct {
	BaseDatosURL        string
	Puerto              int
	ClaveSesion         string
	CookieSegura        bool
	CarnetAnchoMM       int
	CarnetAltoMM        int
	FotoTamanioMaximoMB int
	ImportacionMaxFilas int
	Almacenamiento      Almacenamiento
}

// CargarDesdeEntorno lee la configuración desde las variables de entorno del proceso.
func CargarDesdeEntorno() (Configuracion, error) {
	return Cargar(os.Getenv)
}

// Cargar construye la configuración usando la función obtener para leer cada variable.
// Devuelve todos los errores de validación juntos.
func Cargar(obtener func(string) string) (Configuracion, error) {
	var errores []error
	cfg := Configuracion{
		BaseDatosURL: obtener("BASE_DATOS_URL"),
		ClaveSesion:  obtener("CLAVE_SESION"),
		CookieSegura: booleano(obtener("COOKIE_SEGURA"), true),
		Almacenamiento: Almacenamiento{
			Tipo:         valorOTexto(obtener("ALMACENAMIENTO_TIPO"), tipoAlmacenamientoDisco),
			Ruta:         valorOTexto(obtener("ALMACENAMIENTO_RUTA"), "./datos/fotos"),
			Bucket:       obtener("ALMACENAMIENTO_BUCKET"),
			Endpoint:     obtener("ALMACENAMIENTO_ENDPOINT"),
			ClaveAcceso:  obtener("ALMACENAMIENTO_CLAVE_ACCESO"),
			ClaveSecreta: obtener("ALMACENAMIENTO_CLAVE_SECRETA"),
		},
	}

	if cfg.BaseDatosURL == "" {
		errores = append(errores, errors.New("falta la variable BASE_DATOS_URL"))
	}
	if len(cfg.ClaveSesion) < longitudMinimaClave {
		errores = append(errores, fmt.Errorf("CLAVE_SESION debe tener al menos %d caracteres", longitudMinimaClave))
	}

	var err error
	if cfg.Puerto, err = enteroEnRango(obtener, "PUERTO", 8080, 65535); err != nil {
		errores = append(errores, err)
	}
	if cfg.CarnetAnchoMM, err = enteroEnRango(obtener, "CARNET_ANCHO_MM", 85, 500); err != nil {
		errores = append(errores, err)
	}
	if cfg.CarnetAltoMM, err = enteroEnRango(obtener, "CARNET_ALTO_MM", 54, 500); err != nil {
		errores = append(errores, err)
	}
	if cfg.FotoTamanioMaximoMB, err = enteroEnRango(obtener, "FOTO_TAMANIO_MAXIMO_MB", 2, 10); err != nil {
		errores = append(errores, err)
	}
	if cfg.ImportacionMaxFilas, err = enteroEnRango(obtener, "IMPORTACION_MAX_FILAS", 5000, 20000); err != nil {
		errores = append(errores, err)
	}

	errores = append(errores, validarAlmacenamiento(cfg.Almacenamiento)...)

	if len(errores) > 0 {
		return Configuracion{}, errors.Join(errores...)
	}
	return cfg, nil
}

func validarAlmacenamiento(a Almacenamiento) []error {
	var errores []error
	switch a.Tipo {
	case tipoAlmacenamientoDisco:
	case tipoAlmacenamientoS3:
		if a.Bucket == "" {
			errores = append(errores, errors.New("ALMACENAMIENTO_BUCKET es obligatorio cuando ALMACENAMIENTO_TIPO es s3"))
		}
		if a.ClaveAcceso == "" || a.ClaveSecreta == "" {
			errores = append(errores, errors.New("ALMACENAMIENTO_CLAVE_ACCESO y ALMACENAMIENTO_CLAVE_SECRETA son obligatorias para s3"))
		}
	default:
		errores = append(errores, fmt.Errorf("ALMACENAMIENTO_TIPO debe ser disco o s3, no %q", a.Tipo))
	}
	return errores
}

func valorOTexto(valor, porDefecto string) string {
	if valor == "" {
		return porDefecto
	}
	return valor
}

// booleano interpreta valores como false, 0, no o desactivado como falso.
func booleano(valor string, porDefecto bool) bool {
	switch strings.ToLower(strings.TrimSpace(valor)) {
	case "":
		return porDefecto
	case "false", "0", "no", "desactivado":
		return false
	default:
		return true
	}
}

func enteroEnRango(obtener func(string) string, clave string, porDefecto, maximo int) (int, error) {
	texto := obtener(clave)
	if texto == "" {
		return porDefecto, nil
	}
	numero, err := strconv.Atoi(texto)
	if err != nil {
		return 0, fmt.Errorf("%s debe ser un número entero", clave)
	}
	if numero < 1 || numero > maximo {
		return 0, fmt.Errorf("%s debe estar entre 1 y %d", clave, maximo)
	}
	return numero, nil
}
