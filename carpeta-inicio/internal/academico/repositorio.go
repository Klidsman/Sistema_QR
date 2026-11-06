package academico

import (
	"context"

	"asistencia/internal/academico/generado"
)

// Repositorio accede a las tablas académicas mediante consultas sqlc.
type Repositorio struct {
	consultas *generado.Queries
}

// NuevoRepositorio crea el repositorio sobre la conexión o transacción indicada.
func NuevoRepositorio(db generado.DBTX) *Repositorio {
	return &Repositorio{consultas: generado.New(db)}
}

// BuscarSeccion devuelve el identificador de la sección del grado indicado dentro del año.
func (r *Repositorio) BuscarSeccion(ctx context.Context, anioID int64, grado, seccion string) (int64, error) {
	return r.consultas.BuscarSeccion(ctx, generado.BuscarSeccionParams{
		AnioEscolarID: anioID,
		Nombre:        grado,
		Nombre_2:      seccion,
	})
}

// ListarSecciones devuelve todas las secciones con su grado y año escolar.
func (r *Repositorio) ListarSecciones(ctx context.Context) ([]Seccion, error) {
	filas, err := r.consultas.ListarSecciones(ctx)
	if err != nil {
		return nil, err
	}
	secciones := make([]Seccion, 0, len(filas))
	for _, fila := range filas {
		secciones = append(secciones, Seccion{
			ID:            fila.ID,
			Nombre:        fila.Nombre,
			Turno:         fila.Turno,
			GradoID:       fila.GradoID,
			Grado:         fila.Grado,
			AnioEscolarID: fila.AnioEscolarID,
			AnioEscolar:   fila.AnioEscolar,
		})
	}
	return secciones, nil
}

// BuscarSeccionPorID devuelve la sección con ese identificador.
func (r *Repositorio) BuscarSeccionPorID(ctx context.Context, id int64) (Seccion, error) {
	fila, err := r.consultas.BuscarSeccionPorID(ctx, id)
	if err != nil {
		return Seccion{}, err
	}
	return Seccion{
		ID:            fila.ID,
		Nombre:        fila.Nombre,
		Turno:         fila.Turno,
		GradoID:       fila.GradoID,
		Grado:         fila.Grado,
		AnioEscolarID: fila.AnioEscolarID,
		AnioEscolar:   fila.AnioEscolar,
	}, nil
}

// ListarSeccionesVigentes devuelve las secciones del año escolar activo.
func (r *Repositorio) ListarSeccionesVigentes(ctx context.Context) ([]Seccion, error) {
	filas, err := r.consultas.ListarSeccionesVigentes(ctx)
	if err != nil {
		return nil, err
	}
	secciones := make([]Seccion, 0, len(filas))
	for _, fila := range filas {
		secciones = append(secciones, Seccion{
			ID:            fila.ID,
			Nombre:        fila.Nombre,
			Turno:         fila.Turno,
			GradoID:       fila.GradoID,
			Grado:         fila.Grado,
			AnioEscolarID: fila.AnioEscolarID,
			AnioEscolar:   fila.AnioEscolar,
		})
	}
	return secciones, nil
}
