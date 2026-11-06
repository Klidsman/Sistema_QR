package justificaciones

import (
	"context"
	"errors"
	"fmt"
	"time"

	"asistencia/internal/justificaciones/generado"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repositorio accede a las tablas de justificaciones.
type Repositorio struct {
	pool      *pgxpool.Pool
	consultas *generado.Queries
}

// NuevoRepositorio crea el repositorio.
func NuevoRepositorio(pool *pgxpool.Pool) *Repositorio {
	return &Repositorio{pool: pool, consultas: generado.New(pool)}
}

// EnTransaccion ejecuta la función dentro de una transacción.
func (r *Repositorio) EnTransaccion(ctx context.Context, funcion func(*generado.Queries, generado.DBTX) error) error {
	transaccion, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("iniciando la transacción: %w", err)
	}
	defer func() { _ = transaccion.Rollback(ctx) }()
	if err := funcion(generado.New(transaccion), transaccion); err != nil {
		return err
	}
	if err := transaccion.Commit(ctx); err != nil {
		return fmt.Errorf("confirmando la transacción: %w", err)
	}
	return nil
}

// Insertar crea la cabecera y sus fechas dentro de una transacción.
func (r *Repositorio) Insertar(ctx context.Context, consultas *generado.Queries, justificacion Justificacion, fechas []FechaJustificada, adjuntos []Adjunto) (int64, error) {
	id, err := consultas.InsertarJustificacion(ctx, generado.InsertarJustificacionParams{
		AlumnoID:      justificacion.AlumnoID,
		SeccionID:     justificacion.SeccionID,
		Motivo:        justificacion.Motivo,
		RegistranteID: justificacion.RegistranteID,
	})
	if err != nil {
		return 0, err
	}
	for _, fecha := range fechas {
		if err := consultas.InsertarJustificacionFecha(ctx, generado.InsertarJustificacionFechaParams{
			JustificacionID: id,
			AlumnoID:        fecha.AlumnoID,
			FechaEvento:     fechaParametro(fecha.Fecha),
			Tipo:            fecha.Tipo,
		}); err != nil {
			return 0, err
		}
	}
	for _, adjunto := range adjuntos {
		if err := consultas.InsertarAdjuntoJustificacion(ctx, generado.InsertarAdjuntoJustificacionParams{
			JustificacionID: id,
			Nombre:          adjunto.Nombre,
			Tipo:            adjunto.Tipo,
			Tamano:          adjunto.Tamano,
			Clave:           adjunto.Clave,
		}); err != nil {
			return 0, err
		}
	}
	return id, nil
}

// Obtener devuelve una justificación con fechas y adjuntos.
func (r *Repositorio) Obtener(ctx context.Context, id int64) (Justificacion, error) {
	fila, err := r.consultas.ObtenerJustificacion(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Justificacion{}, ErrJustificacionNoExiste
	}
	if err != nil {
		return Justificacion{}, err
	}
	justificacion := aJustificacion(fila)
	if err := r.cargarDetalle(ctx, &justificacion); err != nil {
		return Justificacion{}, err
	}
	return justificacion, nil
}

// ListarDeAlumno devuelve el historial de un alumno.
func (r *Repositorio) ListarDeAlumno(ctx context.Context, alumnoID int64) ([]Justificacion, error) {
	filas, err := r.consultas.ListarJustificacionesDeAlumno(ctx, alumnoID)
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(filas))
	for _, fila := range filas {
		ids = append(ids, fila.ID)
	}
	return r.cargarVarias(ctx, ids)
}

// ListarPendientes devuelve las justificaciones por resolver.
func (r *Repositorio) ListarPendientes(ctx context.Context) ([]Justificacion, error) {
	filas, err := r.consultas.ListarJustificacionesPendientes(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(filas))
	for _, fila := range filas {
		ids = append(ids, fila.ID)
	}
	return r.cargarVarias(ctx, ids)
}

// ListarDelRegistrante devuelve las justificaciones creadas por un docente.
func (r *Repositorio) ListarDelRegistrante(ctx context.Context, usuarioID int64) ([]Justificacion, error) {
	filas, err := r.consultas.ListarJustificacionesDelRegistrante(ctx, usuarioID)
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(filas))
	for _, fila := range filas {
		ids = append(ids, fila.ID)
	}
	return r.cargarVarias(ctx, ids)
}

// Resolver cambia el estado de una justificación pendiente.
func (r *Repositorio) Resolver(ctx context.Context, consultas *generado.Queries, id int64, estado string, usuarioID int64, comentario string) (bool, error) {
	filas, err := consultas.ResolverJustificacion(ctx, generado.ResolverJustificacionParams{
		ID:          id,
		Estado:      estado,
		ResueltoPor: pgtype.Int8{Int64: usuarioID, Valid: true},
		Comentario:  textoParametro(comentario),
	})
	return filas > 0, err
}

// ActualizarFechas cambia el estado de todas las fechas de una justificación.
func (r *Repositorio) ActualizarFechas(ctx context.Context, consultas *generado.Queries, id int64, estado string) error {
	return consultas.ActualizarEstadoFechas(ctx, generado.ActualizarEstadoFechasParams{
		JustificacionID: id,
		Estado:          estado,
	})
}

// Anular marca una justificación aprobada como anulada.
func (r *Repositorio) Anular(ctx context.Context, consultas *generado.Queries, id, usuarioID int64, motivo string) (bool, error) {
	filas, err := consultas.AnularJustificacion(ctx, generado.AnularJustificacionParams{
		ID:              id,
		AnuladoPor:      pgtype.Int8{Int64: usuarioID, Valid: true},
		MotivoAnulacion: textoParametro(motivo),
	})
	return filas > 0, err
}

// ExisteActiva indica si la fecha ya tiene una justificación pendiente o aprobada.
func (r *Repositorio) ExisteActiva(ctx context.Context, alumnoID int64, fecha time.Time) (bool, error) {
	return r.consultas.ExisteJustificacionActiva(ctx, generado.ExisteJustificacionActivaParams{
		AlumnoID:    alumnoID,
		FechaEvento: fechaParametro(fecha),
	})
}

// EsFaltaJustificada responde a la feature 004.
func (r *Repositorio) EsFaltaJustificada(ctx context.Context, alumnoID int64, fecha time.Time) (bool, error) {
	return r.consultas.EsFaltaJustificada(ctx, generado.EsFaltaJustificadaParams{
		AlumnoID:    alumnoID,
		FechaEvento: fechaParametro(fecha),
	})
}

// ObtenerAdjunto devuelve los metadatos del documento.
func (r *Repositorio) ObtenerAdjunto(ctx context.Context, id int64) (Adjunto, error) {
	fila, err := r.consultas.ObtenerAdjunto(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Adjunto{}, ErrJustificacionNoExiste
	}
	if err != nil {
		return Adjunto{}, err
	}
	return aAdjunto(fila.ID, fila.JustificacionID, fila.Nombre, fila.Tipo, fila.Tamano, fila.Clave, fila.CreadoEn), nil
}

func (r *Repositorio) cargarVarias(ctx context.Context, ids []int64) ([]Justificacion, error) {
	resultado := make([]Justificacion, 0, len(ids))
	for _, id := range ids {
		justificacion, err := r.Obtener(ctx, id)
		if err != nil {
			return nil, err
		}
		resultado = append(resultado, justificacion)
	}
	return resultado, nil
}

func (r *Repositorio) cargarDetalle(ctx context.Context, justificacion *Justificacion) error {
	fechas, err := r.consultas.ListarFechasDeJustificacion(ctx, justificacion.ID)
	if err != nil {
		return err
	}
	justificacion.Fechas = make([]FechaJustificada, 0, len(fechas))
	for _, fecha := range fechas {
		justificacion.Fechas = append(justificacion.Fechas, aFecha(fecha))
	}
	adjuntos, err := r.consultas.ListarAdjuntosDeJustificacion(ctx, justificacion.ID)
	if err != nil {
		return err
	}
	justificacion.Adjuntos = make([]Adjunto, 0, len(adjuntos))
	for _, adjunto := range adjuntos {
		justificacion.Adjuntos = append(justificacion.Adjuntos, aAdjunto(adjunto.ID, adjunto.JustificacionID, adjunto.Nombre, adjunto.Tipo, adjunto.Tamano, adjunto.Clave, adjunto.CreadoEn))
	}
	return nil
}

func aJustificacion(fila generado.ObtenerJustificacionRow) Justificacion {
	return aJustificacionFila(fila.ID, fila.AlumnoID, fila.SeccionID, fila.Motivo, fila.Estado,
		fila.RegistranteID, fila.ResueltoPor, fila.ResueltoEn, fila.Comentario,
		fila.AnuladoPor, fila.AnuladoEn, fila.MotivoAnulacion, fila.CreadaEn)
}

func aJustificacionFila(id, alumnoID, seccionID int64, motivo, estado string, registranteID int64,
	resueltoPor pgtype.Int8, resueltoEn pgtype.Timestamptz, comentario string,
	anuladoPor pgtype.Int8, anuladoEn pgtype.Timestamptz, motivoAnulacion string,
	creadaEn pgtype.Timestamptz) Justificacion {
	justificacion := Justificacion{
		ID:              id,
		AlumnoID:        alumnoID,
		SeccionID:       seccionID,
		Motivo:          motivo,
		Estado:          estado,
		RegistranteID:   registranteID,
		Comentario:      comentario,
		MotivoAnulacion: motivoAnulacion,
		CreadaEn:        creadaEn.Time,
	}
	if resueltoPor.Valid {
		valor := resueltoPor.Int64
		justificacion.ResueltoPor = &valor
	}
	if resueltoEn.Valid {
		valor := resueltoEn.Time
		justificacion.ResueltoEn = &valor
	}
	if anuladoPor.Valid {
		valor := anuladoPor.Int64
		justificacion.AnuladoPor = &valor
	}
	if anuladoEn.Valid {
		valor := anuladoEn.Time
		justificacion.AnuladoEn = &valor
	}
	return justificacion
}

func aFecha(fila generado.JustificacionFecha) FechaJustificada {
	return FechaJustificada{
		ID:              fila.ID,
		JustificacionID: fila.JustificacionID,
		AlumnoID:        fila.AlumnoID,
		Fecha:           fila.FechaEvento.Time,
		Tipo:            fila.Tipo,
		Estado:          fila.Estado,
	}
}

func aAdjunto(id, justificacionID int64, nombre, tipo string, tamano int64, clave string, creadoEn pgtype.Timestamptz) Adjunto {
	return Adjunto{
		ID:              id,
		JustificacionID: justificacionID,
		Nombre:          nombre,
		Tipo:            tipo,
		Tamano:          tamano,
		Clave:           clave,
		CreadoEn:        creadoEn.Time,
	}
}
func fechaParametro(fecha time.Time) pgtype.Date {
	return pgtype.Date{Time: fecha, Valid: true}
}

func textoParametro(texto string) pgtype.Text {
	return pgtype.Text{String: texto, Valid: texto != ""}
}
