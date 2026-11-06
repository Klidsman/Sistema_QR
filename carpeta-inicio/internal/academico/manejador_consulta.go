package academico

import (
	"net/http"

	"asistencia/internal/plataforma/fechas"
	"asistencia/internal/plataforma/intermediarios"
	vista "asistencia/web/plantillas/configuracion"
)

func (m *Manejador) consulta(respuesta http.ResponseWriter, peticion *http.Request) {
	ctx := peticion.Context()
	fecha := fechaDeConsulta(peticion)
	configuracion, err := m.configuracion.ConfiguracionEnFecha(ctx, fecha)
	if err != nil {
		m.registro.Error("no se pudo cargar la configuración histórica", "error", err)
		http.Error(respuesta, "No se pudo cargar la configuración", http.StatusInternalServerError)
		return
	}

	var grados []Grado
	if configuracion.Anio != nil {
		grados, err = m.configuracion.ListarGrados(ctx, configuracion.Anio.ID)
		if err != nil {
			m.registro.Error("no se pudieron listar los grados", "error", err)
			http.Error(respuesta, "No se pudo cargar la configuración", http.StatusInternalServerError)
			return
		}
	}

	m.renderizar(respuesta, peticion, http.StatusOK, vista.Consulta(
		m.nombreUsuario(ctx),
		intermediarios.RolUsuario(ctx),
		m.tokenCSRF(ctx),
		fechas.FormatearISO(fecha),
		aAnioVista(configuracion.Anio),
		aHorariosVista(configuracion.Horarios),
		aExcepcionesVista(configuracion.Excepciones, grados),
		aUmbralesVista(configuracion.Umbrales),
		aSeccionesVista(configuracion.Secciones),
		aBloquesVista(configuracion.Bloques),
	))
}
