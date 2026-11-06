package academico

import (
	"net/http"
	"strconv"
	"strings"

	"asistencia/internal/plataforma/fechas"
	"asistencia/internal/plataforma/intermediarios"
	vista "asistencia/web/plantillas/configuracion"
)

func (m *Manejador) umbrales(respuesta http.ResponseWriter, peticion *http.Request) {
	m.renderizarUmbrales(respuesta, peticion, http.StatusOK, "", peticion.URL.Query().Get("ok"),
		vista.FormularioUmbral{Tipo: TipoUmbralTardanzas})
}

func (m *Manejador) configurarUmbral(respuesta http.ResponseWriter, peticion *http.Request) {
	ctx := peticion.Context()
	activo, err := m.configuracion.AnioActivo(ctx)
	if err != nil || activo == nil {
		m.renderizarUmbrales(respuesta, peticion, http.StatusUnprocessableEntity,
			mensajeDeError(ErrAnioNoEncontrado), "", vista.FormularioUmbral{})
		return
	}
	umbral, formulario, err := umbralDeFormulario(activo.ID, peticion)
	if err != nil {
		m.renderizarUmbrales(respuesta, peticion, http.StatusUnprocessableEntity, mensajeDeError(err), "", formulario)
		return
	}
	if err := m.configuracion.ConfigurarUmbral(ctx, m.usuarioID(ctx), umbral); err != nil {
		m.renderizarUmbrales(respuesta, peticion, http.StatusUnprocessableEntity, mensajeDeError(err), "", formulario)
		return
	}
	http.Redirect(respuesta, peticion, "/configuracion/umbrales?ok=umbral", http.StatusSeeOther)
}

func (m *Manejador) renderizarUmbrales(respuesta http.ResponseWriter, peticion *http.Request, estado int, mensajeError, codigoCorrecto string, formulario vista.FormularioUmbral) {
	ctx := peticion.Context()
	activo, err := m.configuracion.AnioActivo(ctx)
	if err != nil {
		m.registro.Error("no se pudo buscar el año activo", "error", err)
		http.Error(respuesta, "No se pudo cargar la configuración", http.StatusInternalServerError)
		return
	}

	var tardanzas, faltas, faltasSeguidas *UmbralAlerta
	var historial []UmbralAlerta
	if activo != nil {
		vigentes, err := m.configuracion.UmbralesVigentes(ctx, activo.ID, fechas.Hoy())
		if err != nil {
			m.registro.Error("no se pudieron cargar los umbrales vigentes", "error", err)
			http.Error(respuesta, "No se pudo cargar la configuración", http.StatusInternalServerError)
			return
		}
		for i := range vigentes {
			switch vigentes[i].Tipo {
			case TipoUmbralTardanzas:
				tardanzas = &vigentes[i]
			case TipoUmbralFaltas:
				faltas = &vigentes[i]
			case TipoUmbralFaltasSeguidas:
				faltasSeguidas = &vigentes[i]
			}
		}
		historial, err = m.configuracion.ListarUmbrales(ctx, activo.ID)
		if err != nil {
			m.registro.Error("no se pudo listar el historial de umbrales", "error", err)
			http.Error(respuesta, "No se pudo cargar la configuración", http.StatusInternalServerError)
			return
		}
	}

	m.renderizar(respuesta, peticion, estado, vista.Umbrales(
		m.nombreUsuario(ctx),
		intermediarios.RolUsuario(ctx),
		m.tokenCSRF(ctx),
		aAnioVista(activo),
		aUmbralVista(tardanzas),
		aUmbralVista(faltas),
		aUmbralVista(faltasSeguidas),
		aUmbralesVista(historial),
		formulario,
		mensajeCorrectoUmbral(codigoCorrecto),
		mensajeError,
	))
}

func umbralDeFormulario(anioID int64, peticion *http.Request) (UmbralAlerta, vista.FormularioUmbral, error) {
	formulario := vista.FormularioUmbral{
		Tipo:  peticion.PostFormValue("tipo"),
		Valor: peticion.PostFormValue("valor"),
		Fecha: peticion.PostFormValue("vigente_desde"),
	}
	valor, err := strconv.Atoi(strings.TrimSpace(formulario.Valor))
	if err != nil {
		return UmbralAlerta{}, formulario, ErrValorUmbralInvalido
	}
	vigencia, err := fechas.ParsearDia(formulario.Fecha)
	if err != nil {
		return UmbralAlerta{}, formulario, err
	}
	return UmbralAlerta{
		AnioEscolarID: anioID,
		Tipo:          formulario.Tipo,
		Valor:         valor,
		VigenteDesde:  vigencia,
	}, formulario, nil
}

func mensajeCorrectoUmbral(codigo string) string {
	if codigo == "umbral" {
		return "El umbral de alerta se guardó correctamente."
	}
	return ""
}

func aUmbralVista(umbral *UmbralAlerta) *vista.UmbralVista {
	if umbral == nil {
		return nil
	}
	return &vista.UmbralVista{
		Tipo:         etiquetaTipoUmbral(umbral.Tipo),
		Valor:        umbral.Valor,
		VigenteDesde: umbral.VigenteDesde,
	}
}

func aUmbralesVista(umbrales []UmbralAlerta) []vista.UmbralVista {
	resultado := make([]vista.UmbralVista, 0, len(umbrales))
	for i := range umbrales {
		resultado = append(resultado, *aUmbralVista(&umbrales[i]))
	}
	return resultado
}

func etiquetaTipoUmbral(tipo string) string {
	switch tipo {
	case TipoUmbralFaltas:
		return "Faltas por mes"
	case TipoUmbralFaltasSeguidas:
		return "Faltas seguidas"
	default:
		return "Tardanzas por mes"
	}
}
