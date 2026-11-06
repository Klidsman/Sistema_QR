// Package rutas arma el enrutador HTTP con los intermediarios comunes.
package rutas

import (
	"net/http"

	"asistencia/internal/academico"
	"asistencia/internal/alumnos"
	"asistencia/internal/autenticacion"
	"asistencia/internal/plataforma/intermediarios"
	"asistencia/internal/plataforma/salud"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
)

// Dependencias son los componentes que el enrutador necesita para montar las rutas.
type Dependencias struct {
	Sesiones      *scs.SessionManager
	Verificador   salud.Verificador
	Autenticacion *autenticacion.Manejador
	Alumnos       *alumnos.Manejador
	Academico     *academico.Manejador
}

// NuevoEnrutador crea el enrutador con sesiones, protección CSRF y las rutas de cada módulo.
func NuevoEnrutador(dependencias Dependencias) http.Handler {
	enrutador := chi.NewRouter()
	enrutador.Use(dependencias.Sesiones.LoadAndSave)
	enrutador.Use(intermediarios.ProteccionCSRF(dependencias.Sesiones))

	enrutador.Handle("/estaticos/*", http.StripPrefix("/estaticos/", http.FileServer(http.Dir("web/estaticos"))))
	enrutador.Get("/salud", salud.Manejador(dependencias.Verificador))
	enrutador.Get("/", func(respuesta http.ResponseWriter, peticion *http.Request) {
		http.Redirect(respuesta, peticion, "/ingreso", http.StatusSeeOther)
	})
	dependencias.Autenticacion.Montar(enrutador)
	dependencias.Alumnos.Montar(enrutador)
	dependencias.Academico.Montar(enrutador)

	return enrutador
}
