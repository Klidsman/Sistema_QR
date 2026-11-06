package integracion

import (
	"bytes"
	"context"
	"testing"

	"asistencia/internal/academico"
	"asistencia/internal/alumnos"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func TestImportacion_ProcesaFilasValidasYReportaLosRechazos(t *testing.T) {
	pool := AbrirBaseDePruebas(t)
	seccionID := SembrarSeccion(t, pool, "3.°", "A")
	usuarioID := CrearUsuarioPrueba(t, pool)
	servicio := alumnos.NuevoServicio(pool)
	importador := alumnos.NuevoServicioImportacion(pool, academico.NuevoServicio(pool), 100)
	ctx := context.Background()

	_, err := servicio.Crear(ctx, usuarioID, alumnos.DatosAlumno{
		Identificador: "70009999",
		Nombres:       "Mario",
		Apellidos:     "Existente",
		SeccionID:     seccionID,
	})
	require.NoError(t, err)

	contenido := []byte("identificador,nombres,apellidos,grado,seccion\n" +
		"70000001,Ana,Ríos,3.°,A\n" +
		"70000002,Luis,Pérez,3.°,A\n" +
		"70000003,Elena,Ñuño,3.°,A\n" +
		"70000003,Elena,Duplicada,3.°,A\n" +
		"70009999,Mario,Existente,3.°,A\n" +
		"70000004,,Sin Nombre,3.°,A\n" +
		"70000005,Sin,Seccion,3.°,Z\n")

	resultado, err := importador.Importar(ctx, usuarioID, "alumnos.csv", contenido)
	require.NoError(t, err)
	assert.Equal(t, "alumnos.csv", resultado.ArchivoNombre)
	assert.Equal(t, 7, resultado.FilasLeidas)
	assert.Equal(t, 2, resultado.FilasCreadas)
	assert.Equal(t, 5, resultado.FilasRechazadas)

	detalle, err := importador.ObtenerImportacion(ctx, resultado.ID)
	require.NoError(t, err)
	require.Len(t, detalle.Rechazadas, 5)
	motivos := make(map[string]int)
	for _, fila := range detalle.Rechazadas {
		motivos[fila.Motivo]++
	}
	assert.Equal(t, 2, motivos[alumnos.MotivoIdentificadorRepetidoEnArchivo])
	assert.Equal(t, 1, motivos[alumnos.MotivoIdentificadorExistente])
	assert.Equal(t, 1, motivos[alumnos.MotivoCampoObligatorioVacio])
	assert.Equal(t, 1, motivos[alumnos.MotivoSeccionInexistente])

	listado, err := servicio.Listar(ctx, alumnos.FiltrosListado{Texto: "70000001"})
	require.NoError(t, err)
	require.Len(t, listado, 1)
	assert.Equal(t, "Ana", listado[0].Nombres)
	assert.Equal(t, "Ríos", listado[0].Apellidos)

	historial, err := importador.ListarImportaciones(ctx)
	require.NoError(t, err)
	require.Len(t, historial, 1)
	assert.Equal(t, resultado.ID, historial[0].ID)
}

func TestImportacion_LeerXLSXConTildesYN(t *testing.T) {
	pool := AbrirBaseDePruebas(t)
	SembrarSeccion(t, pool, "3.°", "A")
	usuarioID := CrearUsuarioPrueba(t, pool)
	servicio := alumnos.NuevoServicio(pool)
	importador := alumnos.NuevoServicioImportacion(pool, academico.NuevoServicio(pool), 100)
	ctx := context.Background()

	archivo := excelize.NewFile()
	defer func() { _ = archivo.Close() }()
	hoja := archivo.GetSheetName(0)
	require.NoError(t, archivo.SetSheetRow(hoja, "A1", &[]any{
		"identificador", "nombres", "apellidos", "grado", "seccion",
	}))
	require.NoError(t, archivo.SetSheetRow(hoja, "A2", &[]any{
		"70000010", "José María", "Núñez Ríos", "3.°", "A",
	}))
	var contenido bytes.Buffer
	require.NoError(t, archivo.Write(&contenido))

	resultado, err := importador.Importar(ctx, usuarioID, "alumnos.xlsx", contenido.Bytes())
	require.NoError(t, err)
	assert.Equal(t, 1, resultado.FilasCreadas)
	assert.Zero(t, resultado.FilasRechazadas)

	listado, err := servicio.Listar(ctx, alumnos.FiltrosListado{Texto: "70000010"})
	require.NoError(t, err)
	require.Len(t, listado, 1)
	assert.Equal(t, "José María", listado[0].Nombres)
	assert.Equal(t, "Núñez Ríos", listado[0].Apellidos)
}
