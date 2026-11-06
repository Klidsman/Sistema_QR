package academico

import (
	"context"
	"fmt"
	"time"

	"asistencia/internal/academico/generado"
	"asistencia/internal/plataforma/fechas"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RepositorioConfiguracion accede a las tablas de configuración académica.
type RepositorioConfiguracion struct {
	pool      *pgxpool.Pool
	consultas *generado.Queries
}

// NuevoRepositorioConfiguracion crea el repositorio sobre el pool indicado.
func NuevoRepositorioConfiguracion(pool *pgxpool.Pool) *RepositorioConfiguracion {
	return &RepositorioConfiguracion{pool: pool, consultas: generado.New(pool)}
}

// EnTransaccion ejecuta la función dentro de una transacción y confirma si no hay error.
func (r *RepositorioConfiguracion) EnTransaccion(ctx context.Context, funcion func(*generado.Queries, generado.DBTX) error) error {
	transaccion, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("iniciando la transacción: %w", err)
	}
	defer func() {
		_ = transaccion.Rollback(ctx)
	}()

	if err := funcion(generado.New(transaccion), transaccion); err != nil {
		return err
	}
	if err := transaccion.Commit(ctx); err != nil {
		return fmt.Errorf("confirmando la transacción: %w", err)
	}
	return nil
}

// CrearAnio registra un año escolar y devuelve su identificador.
func (r *RepositorioConfiguracion) CrearAnio(ctx context.Context, nombre string, inicio, fin time.Time) (int64, error) {
	return r.consultas.CrearAnioEscolar(ctx, generado.CrearAnioEscolarParams{
		Nombre:      nombre,
		FechaInicio: parametroFecha(inicio),
		FechaFin:    parametroFecha(fin),
	})
}

// BuscarAnioActivo devuelve el año escolar activo.
func (r *RepositorioConfiguracion) BuscarAnioActivo(ctx context.Context) (AnioEscolar, error) {
	fila, err := r.consultas.BuscarAnioActivo(ctx)
	if err != nil {
		return AnioEscolar{}, err
	}
	return aAnioEscolar(fila), nil
}

// BuscarAnioPorID devuelve el año escolar con ese identificador.
func (r *RepositorioConfiguracion) BuscarAnioPorID(ctx context.Context, id int64) (AnioEscolar, error) {
	fila, err := r.consultas.BuscarAnioEscolarPorID(ctx, id)
	if err != nil {
		return AnioEscolar{}, err
	}
	return aAnioEscolar(fila), nil
}

// BuscarAnioPorNombre devuelve el identificador del año con ese nombre.
func (r *RepositorioConfiguracion) BuscarAnioPorNombre(ctx context.Context, nombre string) (int64, error) {
	return r.consultas.BuscarAnioEscolarPorNombre(ctx, nombre)
}

// AnioQueCubreFecha devuelve el año escolar cuyo rango incluye la fecha.
func (r *RepositorioConfiguracion) AnioQueCubreFecha(ctx context.Context, fecha time.Time) (AnioEscolar, error) {
	fila, err := r.consultas.AnioEscolarQueCubreFecha(ctx, parametroFecha(fecha))
	if err != nil {
		return AnioEscolar{}, err
	}
	return aAnioEscolar(fila), nil
}

// ListarAnios devuelve todos los años escolares.
func (r *RepositorioConfiguracion) ListarAnios(ctx context.Context) ([]AnioEscolar, error) {
	filas, err := r.consultas.ListarAniosEscolares(ctx)
	if err != nil {
		return nil, err
	}
	anios := make([]AnioEscolar, 0, len(filas))
	for _, fila := range filas {
		anios = append(anios, aAnioEscolar(fila))
	}
	return anios, nil
}

// CerrarAnio cierra el año activo y devuelve falso si ya estaba cerrado.
func (r *RepositorioConfiguracion) CerrarAnio(ctx context.Context, id int64) (bool, error) {
	filas, err := r.consultas.CerrarAnioEscolar(ctx, id)
	if err != nil {
		return false, err
	}
	return filas > 0, nil
}

// CrearGrado registra un grado en el año indicado.
func (r *RepositorioConfiguracion) CrearGrado(ctx context.Context, anioID int64, nombre string) (int64, error) {
	return r.consultas.CrearGrado(ctx, generado.CrearGradoParams{
		AnioEscolarID: anioID,
		Nombre:        nombre,
	})
}

// BuscarGradoPorNombre devuelve el identificador del grado dentro del año.
func (r *RepositorioConfiguracion) BuscarGradoPorNombre(ctx context.Context, anioID int64, nombre string) (int64, error) {
	return r.consultas.BuscarGradoPorNombre(ctx, generado.BuscarGradoPorNombreParams{
		AnioEscolarID: anioID,
		Nombre:        nombre,
	})
}

// BuscarGradoPorID devuelve el grado y el estado de su año escolar.
func (r *RepositorioConfiguracion) BuscarGradoPorID(ctx context.Context, id int64) (Grado, string, error) {
	fila, err := r.consultas.BuscarGradoPorID(ctx, id)
	if err != nil {
		return Grado{}, "", err
	}
	return Grado{
		ID:            fila.ID,
		AnioEscolarID: fila.AnioEscolarID,
		Nombre:        fila.Nombre,
	}, fila.AnioEstado, nil
}

// ListarGradosPorAnio devuelve los grados del año indicado.
func (r *RepositorioConfiguracion) ListarGradosPorAnio(ctx context.Context, anioID int64) ([]Grado, error) {
	filas, err := r.consultas.ListarGradosPorAnio(ctx, anioID)
	if err != nil {
		return nil, err
	}
	grados := make([]Grado, 0, len(filas))
	for _, fila := range filas {
		grados = append(grados, Grado{
			ID:            fila.ID,
			AnioEscolarID: fila.AnioEscolarID,
			Nombre:        fila.Nombre,
		})
	}
	return grados, nil
}

// BuscarSeccionPorID devuelve la sección con ese identificador.
func (r *RepositorioConfiguracion) BuscarSeccionPorID(ctx context.Context, id int64) (Seccion, error) {
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

// BuscarCursoPorNombre devuelve el identificador del curso con ese nombre en el año.
func (r *RepositorioConfiguracion) BuscarCursoPorNombre(ctx context.Context, anioID int64, nombre string) (int64, error) {
	return r.consultas.BuscarCursoPorNombre(ctx, generado.BuscarCursoPorNombreParams{
		AnioEscolarID: anioID,
		Lower:         nombre,
	})
}

// CrearSeccion registra una sección en el grado indicado.
func (r *RepositorioConfiguracion) CrearSeccion(ctx context.Context, gradoID int64, nombre string) (int64, error) {
	return r.consultas.CrearSeccion(ctx, generado.CrearSeccionParams{
		GradoID: gradoID,
		Nombre:  nombre,
	})
}

// BuscarSeccionPorNombre devuelve el identificador de la sección dentro del grado.
func (r *RepositorioConfiguracion) BuscarSeccionPorNombre(ctx context.Context, gradoID int64, nombre string) (int64, error) {
	return r.consultas.BuscarSeccionPorNombre(ctx, generado.BuscarSeccionPorNombreParams{
		GradoID: gradoID,
		Nombre:  nombre,
	})
}

// CrearTurnoSeccion guarda el turno de una sección desde una fecha.
func (r *RepositorioConfiguracion) CrearTurnoSeccion(ctx context.Context, seccionID int64, turno string, vigenteDesde time.Time) (int64, error) {
	return r.consultas.CrearTurnoSeccion(ctx, generado.CrearTurnoSeccionParams{
		SeccionID:    seccionID,
		Turno:        turno,
		VigenteDesde: parametroFecha(vigenteDesde),
	})
}

// TurnoVigenteSeccion devuelve el turno de la sección en la fecha indicada.
func (r *RepositorioConfiguracion) TurnoVigenteSeccion(ctx context.Context, seccionID int64, fecha time.Time) (string, error) {
	return r.consultas.TurnoVigenteSeccion(ctx, generado.TurnoVigenteSeccionParams{
		SeccionID:    seccionID,
		VigenteDesde: parametroFecha(fecha),
	})
}

// ListarTurnosSeccion devuelve la historia de turnos de una sección.
func (r *RepositorioConfiguracion) ListarTurnosSeccion(ctx context.Context, seccionID int64) ([]TurnoSeccion, error) {
	filas, err := r.consultas.ListarTurnosSeccion(ctx, seccionID)
	if err != nil {
		return nil, err
	}
	turnos := make([]TurnoSeccion, 0, len(filas))
	for _, fila := range filas {
		turnos = append(turnos, TurnoSeccion{
			ID:           fila.ID,
			SeccionID:    fila.SeccionID,
			Turno:        fila.Turno,
			VigenteDesde: fila.VigenteDesde.Time,
		})
	}
	return turnos, nil
}

// ListarSeccionesPorAnio devuelve las secciones del año indicado.
func (r *RepositorioConfiguracion) ListarSeccionesPorAnio(ctx context.Context, anioID int64) ([]Seccion, error) {
	filas, err := r.consultas.ListarSeccionesPorAnio(ctx, anioID)
	if err != nil {
		return nil, err
	}
	return aSecciones(filas), nil
}

func aSecciones(filas []generado.ListarSeccionesPorAnioRow) []Seccion {
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
	return secciones
}

func aAnioEscolar(fila generado.AniosEscolare) AnioEscolar {
	anio := AnioEscolar{
		ID:          fila.ID,
		Nombre:      fila.Nombre,
		FechaInicio: fila.FechaInicio.Time,
		FechaFin:    fila.FechaFin.Time,
		Estado:      fila.Estado,
	}
	if fila.CerradoEn.Valid {
		cerrado := fila.CerradoEn.Time
		anio.CerradoEn = &cerrado
	}
	return anio
}

func parametroHora(hora time.Time) pgtype.Time {
	segundos := hora.Hour()*3600 + hora.Minute()*60 + hora.Second()
	return pgtype.Time{Microseconds: int64(segundos) * 1_000_000, Valid: true}
}

func horaDeTexto(texto string) (time.Time, error) {
	hora, err := fechas.ParsearHora(texto)
	if err != nil {
		return time.Time{}, fmt.Errorf("interpretando la hora %q: %w", texto, err)
	}
	return hora, nil
}
