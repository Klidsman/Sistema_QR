package integracion

import (
	"context"
	"testing"
	"time"

	"asistencia/internal/academico"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConsultaHistorica_DevuelveLaConfiguracionDeCadaFecha(t *testing.T) {
	pool := AbrirBaseDePruebas(t)
	usuarioID := CrearUsuarioPrueba(t, pool)
	servicio := academico.NuevoServicioConfiguracion(pool)
	ctx := context.Background()

	inicio := time.Date(2027, time.March, 1, 0, 0, 0, 0, time.UTC)
	fin := time.Date(2027, time.December, 15, 0, 0, 0, 0, time.UTC)
	anioID, err := servicio.CrearAnio(ctx, usuarioID, academico.DatosAnio{
		Nombre: "2027", FechaInicio: inicio, FechaFin: fin,
	})
	require.NoError(t, err)
	gradoID, err := servicio.CrearGrado(ctx, usuarioID, academico.DatosGrado{
		AnioEscolarID: anioID, Nombre: "3.°",
	})
	require.NoError(t, err)
	seccionID, err := servicio.CrearSeccion(ctx, usuarioID, academico.DatosSeccion{
		GradoID: gradoID, Nombre: "A", Turno: academico.TurnoManana, VigenteDesde: inicio,
	})
	require.NoError(t, err)

	require.NoError(t, servicio.ConfigurarHorario(ctx, usuarioID, academico.HorarioTurno{
		Turno:              academico.TurnoManana,
		HoraIngreso:        horaDePrueba(t, "07:30"),
		ToleranciaMinutos:  10,
		HoraLimiteTardanza: horaDePrueba(t, "08:00"),
		HoraCierre:         horaDePrueba(t, "12:30"),
		VigenteDesde:       inicio,
	}))
	_, err = servicio.CrearBloque(ctx, usuarioID, academico.BloqueClase{
		Turno:        academico.TurnoManana,
		Numero:       1,
		HoraInicio:   horaDePrueba(t, "07:45"),
		HoraFin:      horaDePrueba(t, "08:30"),
		VigenteDesde: inicio,
	})
	require.NoError(t, err)
	require.NoError(t, servicio.ConfigurarUmbral(ctx, usuarioID, academico.UmbralAlerta{
		AnioEscolarID: anioID,
		Tipo:          academico.TipoUmbralTardanzas,
		Valor:         3,
		VigenteDesde:  inicio,
	}))

	cambio := inicio.AddDate(0, 2, 0)
	require.NoError(t, servicio.ConfigurarHorario(ctx, usuarioID, academico.HorarioTurno{
		Turno:              academico.TurnoManana,
		HoraIngreso:        horaDePrueba(t, "07:30"),
		ToleranciaMinutos:  5,
		HoraLimiteTardanza: horaDePrueba(t, "08:00"),
		HoraCierre:         horaDePrueba(t, "12:30"),
		VigenteDesde:       cambio,
	}))
	require.NoError(t, servicio.CambiarTurno(ctx, usuarioID, seccionID, academico.TurnoTarde, cambio))
	bloques, err := servicio.BloquesVigentes(ctx, academico.TurnoManana, inicio)
	require.NoError(t, err)
	require.Len(t, bloques, 1)
	require.NoError(t, servicio.AjustarHorasBloque(ctx, usuarioID, bloques[0].ID,
		horaDePrueba(t, "07:45"), horaDePrueba(t, "08:20"), cambio))
	require.NoError(t, servicio.ConfigurarUmbral(ctx, usuarioID, academico.UmbralAlerta{
		AnioEscolarID: anioID,
		Tipo:          academico.TipoUmbralTardanzas,
		Valor:         4,
		VigenteDesde:  cambio,
	}))

	fechaFeriado := inicio.AddDate(0, 1, 0)
	_, err = servicio.RegistrarExcepcion(ctx, usuarioID, academico.DatosExcepcion{
		AnioEscolarID: anioID,
		Tipo:          academico.TipoFeriado,
		FechaInicio:   fechaFeriado,
		FechaFin:      fechaFeriado,
		Motivo:        "Feriado de prueba",
	}, false)
	require.NoError(t, err)

	antes := inicio.AddDate(0, 1, 0)
	configuracion, err := servicio.ConfiguracionEnFecha(ctx, antes)
	require.NoError(t, err)
	require.NotNil(t, configuracion.Anio)
	assert.Equal(t, "2027", configuracion.Anio.Nombre)
	require.Len(t, configuracion.Horarios, 1)
	assert.Equal(t, 10, configuracion.Horarios[0].ToleranciaMinutos)
	require.Len(t, configuracion.Secciones, 1)
	assert.Equal(t, academico.TurnoManana, configuracion.Secciones[0].Turno)
	require.Len(t, configuracion.Bloques, 1)
	assert.Equal(t, "08:30", horasTexto(configuracion.Bloques[0].HoraFin))
	require.Len(t, configuracion.Umbrales, 1)
	assert.Equal(t, 3, configuracion.Umbrales[0].Valor)
	require.Len(t, configuracion.Excepciones, 1)

	despues := cambio.AddDate(0, 1, 0)
	configuracion, err = servicio.ConfiguracionEnFecha(ctx, despues)
	require.NoError(t, err)
	require.NotNil(t, configuracion.Anio)
	require.Len(t, configuracion.Horarios, 1)
	assert.Equal(t, 5, configuracion.Horarios[0].ToleranciaMinutos)
	require.Len(t, configuracion.Secciones, 1)
	assert.Equal(t, academico.TurnoTarde, configuracion.Secciones[0].Turno)
	require.Len(t, configuracion.Bloques, 1)
	assert.Equal(t, "08:20", horasTexto(configuracion.Bloques[0].HoraFin))
	require.Len(t, configuracion.Umbrales, 1)
	assert.Equal(t, 4, configuracion.Umbrales[0].Valor)
	assert.Empty(t, configuracion.Excepciones)

	fuera := inicio.AddDate(0, 0, -1)
	configuracion, err = servicio.ConfiguracionEnFecha(ctx, fuera)
	require.NoError(t, err)
	assert.Nil(t, configuracion.Anio)
}

func horasTexto(valor time.Time) string {
	return valor.Format("15:04")
}
