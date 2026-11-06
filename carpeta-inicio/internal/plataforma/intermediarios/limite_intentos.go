package intermediarios

import (
	"sync"
	"time"
)

type registroIntentos struct {
	fallos int
	inicio time.Time
}

// LimiteIntentos cuenta los fallos de inicio de sesión por clave dentro de una ventana de tiempo.
// Es válido mientras la aplicación corre en una sola instancia.
type LimiteIntentos struct {
	mu        sync.Mutex
	maximo    int
	ventana   time.Duration
	ahora     func() time.Time
	registros map[string]*registroIntentos
}

// NuevoLimiteIntentos crea un límite que bloquea una clave tras maximo fallos dentro de la ventana.
func NuevoLimiteIntentos(maximo int, ventana time.Duration, ahora func() time.Time) *LimiteIntentos {
	return &LimiteIntentos{
		maximo:    maximo,
		ventana:   ventana,
		ahora:     ahora,
		registros: make(map[string]*registroIntentos),
	}
}

// Permitir indica si la clave puede intentar iniciar sesión en este momento.
func (l *LimiteIntentos) Permitir(clave string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	registro, existe := l.registros[clave]
	if !existe {
		return true
	}
	if l.expirado(registro) {
		delete(l.registros, clave)
		return true
	}
	return registro.fallos < l.maximo
}

// RegistrarFallo cuenta un intento fallido para la clave.
func (l *LimiteIntentos) RegistrarFallo(clave string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	registro, existe := l.registros[clave]
	if !existe || l.expirado(registro) {
		l.registros[clave] = &registroIntentos{fallos: 1, inicio: l.ahora()}
		return
	}
	registro.fallos++
}

// Limpiar elimina el conteo de la clave, por ejemplo tras un ingreso correcto.
func (l *LimiteIntentos) Limpiar(clave string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	delete(l.registros, clave)
}

func (l *LimiteIntentos) expirado(registro *registroIntentos) bool {
	return l.ahora().Sub(registro.inicio) >= l.ventana
}
