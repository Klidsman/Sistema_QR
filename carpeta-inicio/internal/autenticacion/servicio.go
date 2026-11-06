package autenticacion

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"asistencia/internal/autenticacion/generado"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/argon2"
)

// Parámetros de argon2id. Son fijos para que cada hash pueda verificarse con los valores
// guardados en el propio hash.
const (
	argonMemoria       uint32 = 64 * 1024
	argonIteraciones   uint32 = 3
	argonParalelismo   uint8  = 2
	argonLongitudSal          = 16
	argonLongitudClave        = 32
)

// Servicio contiene las reglas de autenticación y de gestión de usuarios.
type Servicio struct {
	repositorio *Repositorio
}

// NuevoServicio crea el servicio de autenticación sobre la conexión indicada.
func NuevoServicio(db generado.DBTX) *Servicio {
	return &Servicio{repositorio: NuevoRepositorio(db)}
}

// Autenticar valida el correo y la contraseña y devuelve el usuario si son correctos.
func (s *Servicio) Autenticar(ctx context.Context, correo, contrasenia string) (Usuario, error) {
	fila, err := s.repositorio.BuscarPorCorreo(ctx, normalizarCorreo(correo))
	if errors.Is(err, pgx.ErrNoRows) {
		return Usuario{}, ErrCredencialesInvalidas
	}
	if err != nil {
		return Usuario{}, fmt.Errorf("buscando el usuario: %w", err)
	}

	coincide, err := VerificarContrasenia(fila.ContraseniaHash, contrasenia)
	if err != nil {
		return Usuario{}, fmt.Errorf("verificando la contraseña: %w", err)
	}
	if !coincide || !fila.Activo {
		return Usuario{}, ErrCredencialesInvalidas
	}
	return Usuario{
		ID:             fila.ID,
		Correo:         fila.Correo,
		NombreCompleto: fila.NombreCompleto,
		Rol:            fila.Rol,
		Activo:         fila.Activo,
	}, nil
}

// HashearContrasenia genera el hash argon2id de la contraseña con una sal aleatoria.
func HashearContrasenia(contrasenia string) (string, error) {
	sal := make([]byte, argonLongitudSal)
	if _, err := rand.Read(sal); err != nil {
		return "", fmt.Errorf("generando la sal: %w", err)
	}
	clave := argon2.IDKey([]byte(contrasenia), sal, argonIteraciones, argonMemoria, argonParalelismo, argonLongitudClave)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argonMemoria,
		argonIteraciones,
		argonParalelismo,
		base64.RawStdEncoding.EncodeToString(sal),
		base64.RawStdEncoding.EncodeToString(clave),
	), nil
}

// VerificarContrasenia compara la contraseña con un hash argon2id generado por HashearContrasenia.
func VerificarContrasenia(hash, contrasenia string) (bool, error) {
	partes := strings.Split(hash, "$")
	if len(partes) != 6 || partes[1] != "argon2id" {
		return false, errors.New("el hash de contraseña no tiene el formato argon2id")
	}

	var version int
	if _, err := fmt.Sscanf(partes[2], "v=%d", &version); err != nil || version != argon2.Version {
		return false, errors.New("la versión del hash de contraseña no es compatible")
	}

	var memoria, iteraciones uint32
	var paralelismo uint8
	if _, err := fmt.Sscanf(partes[3], "m=%d,t=%d,p=%d", &memoria, &iteraciones, &paralelismo); err != nil {
		return false, errors.New("los parámetros del hash de contraseña no son válidos")
	}

	sal, err := base64.RawStdEncoding.DecodeString(partes[4])
	if err != nil {
		return false, fmt.Errorf("decodificando la sal: %w", err)
	}
	esperada, err := base64.RawStdEncoding.DecodeString(partes[5])
	if err != nil {
		return false, fmt.Errorf("decodificando el hash: %w", err)
	}

	calculada := argon2.IDKey([]byte(contrasenia), sal, iteraciones, memoria, paralelismo, uint32(len(esperada)))
	return subtle.ConstantTimeCompare(calculada, esperada) == 1, nil
}

func normalizarCorreo(correo string) string {
	return strings.ToLower(strings.TrimSpace(correo))
}
