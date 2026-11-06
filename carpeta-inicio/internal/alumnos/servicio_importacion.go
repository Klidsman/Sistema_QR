package alumnos

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"asistencia/internal/academico"
	"asistencia/internal/alumnos/generado"
	"asistencia/internal/plataforma/auditoria"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BuscadorSeccionesVigentes resuelve las secciones del año escolar activo.
type BuscadorSeccionesVigentes interface {
	ListarSeccionesVigentes(ctx context.Context) ([]academico.Seccion, error)
}

// ServicioImportacion procesa archivos CSV y XLSX y guarda su reporte.
type ServicioImportacion struct {
	repositorio *Repositorio
	alumnos     *Servicio
	secciones   BuscadorSeccionesVigentes
	maxFilas    int
}

// NuevoServicioImportacion crea el servicio de importación.
func NuevoServicioImportacion(pool *pgxpool.Pool, secciones BuscadorSeccionesVigentes, maxFilas int) *ServicioImportacion {
	return NuevoServicioImportacionConRepositorio(NuevoRepositorio(pool), NuevoServicio(pool), secciones, maxFilas)
}

// NuevoServicioImportacionConRepositorio crea el servicio sobre un repositorio existente.
func NuevoServicioImportacionConRepositorio(repositorio *Repositorio, alumnos *Servicio, secciones BuscadorSeccionesVigentes, maxFilas int) *ServicioImportacion {
	if maxFilas <= 0 {
		maxFilas = maximoFilasImportacionPorDefecto
	}
	return &ServicioImportacion{
		repositorio: repositorio,
		alumnos:     alumnos,
		secciones:   secciones,
		maxFilas:    maxFilas,
	}
}

// Importar procesa el archivo, crea los alumnos válidos y guarda el reporte de rechazos.
func (s *ServicioImportacion) Importar(ctx context.Context, usuarioID int64, archivoNombre string, contenido []byte) (Importacion, error) {
	nombre := filepath.Base(strings.TrimSpace(archivoNombre))
	if nombre == "" || nombre == "." {
		nombre = "importacion"
	}

	var filas []FilaImportada
	var err error
	switch strings.ToLower(filepath.Ext(nombre)) {
	case ".csv":
		filas, err = LeerCSV(contenido, s.maxFilas)
	case ".xlsx":
		filas, err = LeerXLSX(contenido, s.maxFilas)
	default:
		return Importacion{}, ErrFormatoArchivo
	}
	if err != nil {
		return Importacion{}, err
	}

	validas, rechazadas := ClasificarFilas(filas)

	secciones, err := s.secciones.ListarSeccionesVigentes(ctx)
	if err != nil {
		return Importacion{}, fmt.Errorf("listando las secciones vigentes: %w", err)
	}
	porNombre := make(map[string]int64, len(secciones))
	for _, seccion := range secciones {
		porNombre[claveSeccionImportacion(seccion.Grado, seccion.Nombre)] = seccion.ID
	}

	creadas := 0
	for _, fila := range validas {
		seccionID, existe := porNombre[claveSeccionImportacion(fila.Grado, fila.Seccion)]
		if !existe {
			rechazadas = append(rechazadas, FilaRechazada{
				NumeroFila:    fila.Numero,
				Identificador: fila.Identificador,
				Motivo:        MotivoSeccionInexistente,
			})
			continue
		}

		_, err := s.alumnos.Crear(ctx, usuarioID, DatosAlumno{
			Identificador:     fila.Identificador,
			Nombres:           fila.Nombres,
			Apellidos:         fila.Apellidos,
			SeccionID:         seccionID,
			ApoderadoNombre:   fila.ApoderadoNombre,
			ApoderadoTelefono: fila.ApoderadoTelefono,
		})
		if errors.Is(err, ErrIdentificadorExistente) {
			rechazadas = append(rechazadas, FilaRechazada{
				NumeroFila:    fila.Numero,
				Identificador: fila.Identificador,
				Motivo:        MotivoIdentificadorExistente,
			})
			continue
		}
		if err != nil {
			return Importacion{}, fmt.Errorf("importando la fila %d: %w", fila.Numero, err)
		}
		creadas++
	}

	sort.Slice(rechazadas, func(i, j int) bool {
		return rechazadas[i].NumeroFila < rechazadas[j].NumeroFila
	})

	importacion := Importacion{
		UsuarioID:       usuarioID,
		ArchivoNombre:   nombre,
		FilasLeidas:     len(filas),
		FilasCreadas:    creadas,
		FilasRechazadas: len(rechazadas),
	}
	err = s.repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		fila, err := consultas.CrearImportacion(ctx, generado.CrearImportacionParams{
			UsuarioID:       usuarioID,
			ArchivoNombre:   nombre,
			FilasLeidas:     int32(importacion.FilasLeidas),
			FilasCreadas:    int32(importacion.FilasCreadas),
			FilasRechazadas: int32(importacion.FilasRechazadas),
		})
		if err != nil {
			return fmt.Errorf("registrando la importación: %w", err)
		}
		importacion.ID = fila.ID
		importacion.CreadaEn = fila.CreadaEn.Time

		for _, rechazada := range rechazadas {
			if err := consultas.CrearImportacionFila(ctx, generado.CrearImportacionFilaParams{
				ImportacionID: importacion.ID,
				NumeroFila:    int32(rechazada.NumeroFila),
				Identificador: rechazada.Identificador,
				Motivo:        rechazada.Motivo,
			}); err != nil {
				return fmt.Errorf("guardando la fila rechazada %d: %w", rechazada.NumeroFila, err)
			}
		}

		return auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID: usuarioID,
			Entidad:   "importacion",
			EntidadID: importacion.ID,
			Accion:    "importar",
			ValorNuevo: map[string]any{
				"archivo":          importacion.ArchivoNombre,
				"filas_leidas":     importacion.FilasLeidas,
				"filas_creadas":    importacion.FilasCreadas,
				"filas_rechazadas": importacion.FilasRechazadas,
			},
		})
	})
	if err != nil {
		return Importacion{}, err
	}
	return importacion, nil
}

// ListarImportaciones devuelve el historial de cargas, de la más reciente a la más antigua.
func (s *ServicioImportacion) ListarImportaciones(ctx context.Context) ([]Importacion, error) {
	importaciones, err := s.repositorio.listarImportaciones(ctx)
	if err != nil {
		return nil, fmt.Errorf("listando las importaciones: %w", err)
	}
	return importaciones, nil
}

// ObtenerImportacion devuelve el resumen y las filas rechazadas de una importación.
func (s *ServicioImportacion) ObtenerImportacion(ctx context.Context, id int64) (ImportacionDetalle, error) {
	importacion, err := s.repositorio.buscarImportacion(ctx, id)
	if err != nil {
		return ImportacionDetalle{}, err
	}
	rechazadas, err := s.repositorio.listarImportacionFilas(ctx, id)
	if err != nil {
		return ImportacionDetalle{}, fmt.Errorf("listando las filas rechazadas: %w", err)
	}
	return ImportacionDetalle{Importacion: importacion, Rechazadas: rechazadas}, nil
}

func claveSeccionImportacion(grado, seccion string) string {
	return Normalizar(grado) + "|" + Normalizar(seccion)
}
