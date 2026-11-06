package asistencia

import (
	"log/slog"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
)

// Dependencias reúne lo que el manejador de asistencia necesita.
type Dependencias struct {
	Sesiones *scs.SessionManager
	Registro *slog.Logger
}

// Manejador atiende las pantallas de asistencia del docente.
type Manejador struct {
	sesiones *scs.SessionManager
	registro *slog.Logger
}

// NuevoManejador crea el manejador de asistencia.
func NuevoManejador(dependencias Dependencias) *Manejador {
	return &Manejador{
		sesiones: dependencias.Sesiones,
		registro: dependencias.Registro,
	}
}

// Montar registra las rutas de asistencia. Las historias agregan sus rutas sobre este manejador.
func (m *Manejador) Montar(enrutador chi.Router) {
	_ = enrutador
}
