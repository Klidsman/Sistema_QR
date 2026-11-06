package alumnos

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"asistencia/internal/alumnos/generado"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repositorio accede a las tablas de alumnos y credenciales QR mediante consultas sqlc.
type Repositorio struct {
	pool      *pgxpool.Pool
	consultas *generado.Queries
}

// NuevoRepositorio crea el repositorio sobre el pool indicado.
func NuevoRepositorio(pool *pgxpool.Pool) *Repositorio {
	return &Repositorio{pool: pool, consultas: generado.New(pool)}
}

// EnTransaccion ejecuta la función dentro de una transacción y confirma si no hay error.
// Entrega las consultas tipadas y la conexión transaccional, que también usa la auditoría.
func (r *Repositorio) EnTransaccion(ctx context.Context, funcion func(*generado.Queries, generado.DBTX) error) error {
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

func (r *Repositorio) listar(ctx context.Context, filtros FiltrosListado) ([]AlumnoListado, error) {
	filas, err := r.consultas.ListarAlumnos(ctx, generado.ListarAlumnosParams{
		Estado:        textoOpcional(filtros.Estado),
		SeccionID:     enteroOpcional(filtros.SeccionID),
		AnioEscolarID: enteroOpcional(filtros.AnioEscolarID),
		Texto:         textoOpcional(Normalizar(filtros.Texto)),
	})
	if err != nil {
		return nil, err
	}

	listado := make([]AlumnoListado, 0, len(filas))
	for _, fila := range filas {
		listado = append(listado, AlumnoListado{
			ID:            fila.ID,
			Identificador: fila.Identificador,
			Nombres:       fila.Nombres,
			Apellidos:     fila.Apellidos,
			Estado:        fila.Estado,
			SeccionID:     fila.SeccionID,
			Grado:         fila.GradoNombre,
			Seccion:       fila.SeccionNombre,
			AnioEscolar:   fila.AnioEscolarNombre,
			TieneFoto:     fila.FotoClave.Valid,
		})
	}
	return listado, nil
}

func (r *Repositorio) listarCarnets(ctx context.Context, seleccion SeleccionCarnets) ([]AlumnoCarnet, error) {
	filas, err := r.consultas.ListarAlumnosParaCarnets(ctx, generado.ListarAlumnosParaCarnetsParams{
		SeccionID: enteroOpcional(seleccion.SeccionID),
		GradoID:   enteroOpcional(seleccion.GradoID),
	})
	if err != nil {
		return nil, err
	}

	carnets := make([]AlumnoCarnet, 0, len(filas))
	for _, fila := range filas {
		carnets = append(carnets, AlumnoCarnet{
			ID:            fila.ID,
			Identificador: fila.Identificador,
			Nombres:       fila.Nombres,
			Apellidos:     fila.Apellidos,
			SeccionID:     fila.SeccionID,
			GradoID:       fila.GradoID,
			Grado:         fila.GradoNombre,
			Seccion:       fila.SeccionNombre,
			AnioEscolar:   fila.AnioEscolarNombre,
			FotoClave:     fila.FotoClave.String,
			TokenQR:       fila.QrToken.String,
		})
	}
	return carnets, nil
}

func (r *Repositorio) buscar(ctx context.Context, id int64) (Alumno, error) {
	fila, err := r.consultas.BuscarAlumnoPorID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Alumno{}, ErrAlumnoNoEncontrado
	}
	if err != nil {
		return Alumno{}, err
	}
	return Alumno{
		ID:                fila.ID,
		Identificador:     fila.Identificador,
		Nombres:           fila.Nombres,
		Apellidos:         fila.Apellidos,
		Estado:            fila.Estado,
		SeccionID:         fila.SeccionID,
		Grado:             fila.GradoNombre,
		Seccion:           fila.SeccionNombre,
		Turno:             fila.Turno,
		AnioEscolar:       fila.AnioEscolarNombre,
		FotoClave:         fila.FotoClave.String,
		ApoderadoNombre:   fila.ApoderadoNombre.String,
		ApoderadoTelefono: fila.ApoderadoTelefono.String,
		FechaSalida:       fechaOpcional(fila.FechaSalida),
		CreadoEn:          fila.CreadoEn.Time,
		ActualizadoEn:     fila.ActualizadoEn.Time,
	}, nil
}

func (r *Repositorio) buscarIdentificador(ctx context.Context, consultas *generado.Queries, identificador string) (generado.BuscarAlumnoPorIdentificadorRow, error) {
	return consultas.BuscarAlumnoPorIdentificador(ctx, identificador)
}

func textoOpcional(valor string) pgtype.Text {
	limpio := strings.TrimSpace(valor)
	if limpio == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: limpio, Valid: true}
}

func enteroOpcional(valor int64) pgtype.Int8 {
	if valor <= 0 {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: valor, Valid: true}
}

func fechaOpcional(valor pgtype.Date) *time.Time {
	if !valor.Valid {
		return nil
	}
	fecha := valor.Time
	return &fecha
}
