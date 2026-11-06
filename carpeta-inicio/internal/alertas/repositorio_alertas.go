package alertas

import (
	"context"
	"errors"
	"time"

	"asistencia/internal/alertas/generado"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// InsertarAlerta intenta crear una alerta y devuelve su identificador.
// El segundo valor es falso si la alerta ya existía.
func (r *Repositorio) InsertarAlerta(ctx context.Context, alerta Alerta) (int64, bool, error) {
	return insertarAlerta(ctx, r.consultas, alerta)
}

// InsertarAlertaCon permite insertar dentro de una transacción abierta.
func (r *Repositorio) InsertarAlertaCon(ctx context.Context, consultas *generado.Queries, alerta Alerta) (int64, bool, error) {
	return insertarAlerta(ctx, consultas, alerta)
}

func insertarAlerta(ctx context.Context, consultas *generado.Queries, alerta Alerta) (int64, bool, error) {
	id, err := consultas.InsertarAlerta(ctx, generado.InsertarAlertaParams{
		Tipo:           alerta.Tipo,
		AlumnoID:       alerta.AlumnoID,
		SeccionID:      alerta.SeccionID,
		CursoID:        enteroOpcional(alerta.CursoID),
		BloqueNumero:   entero32Opcional(alerta.BloqueNumero),
		FechaEvento:    fechaParametro(alerta.FechaEvento),
		Mes:            fechaOpcional(alerta.Mes),
		MarcacionID:    enteroOpcional(alerta.MarcacionID),
		DestinatarioID: alerta.DestinatarioID,
		SinResponsable: alerta.SinResponsable,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return id, true, nil
}

// ListarAlertas devuelve las alertas que cumplen los filtros.
func (r *Repositorio) ListarAlertas(ctx context.Context, filtros FiltrosAlerta) ([]Alerta, error) {
	filas, err := r.consultas.ListarAlertas(ctx, generado.ListarAlertasParams{
		Column1:        filtros.Todos,
		DestinatarioID: filtros.DestinatarioID,
		Column3:        fechaOpcional(filtros.Desde),
		Column4:        fechaOpcional(filtros.Hasta),
		Column5:        filtros.SeccionID,
		Column6:        filtros.Tipo,
		Column7:        filtros.Estado,
	})
	if err != nil {
		return nil, err
	}
	alertas := make([]Alerta, 0, len(filas))
	for _, fila := range filas {
		alertas = append(alertas, aAlerta(fila))
	}
	return alertas, nil
}

// ContarNuevas devuelve la cantidad de alertas nuevas del destinatario.
func (r *Repositorio) ContarNuevas(ctx context.Context, destinatarioID int64, todos bool) (int, error) {
	cantidad, err := r.consultas.ContarAlertasNuevas(ctx, generado.ContarAlertasNuevasParams{
		Column1:        todos,
		DestinatarioID: destinatarioID,
	})
	return int(cantidad), err
}

// MarcarVista marca una alerta como vista si pertenece al usuario.
func (r *Repositorio) MarcarVista(ctx context.Context, id, usuarioID int64, todos bool) (bool, error) {
	filas, err := r.consultas.MarcarAlertaVista(ctx, generado.MarcarAlertaVistaParams{
		ID:       id,
		LeidaPor: pgtype.Int8{Int64: usuarioID, Valid: true},
		Column3:  todos,
	})
	if err != nil {
		return false, err
	}
	return filas > 0, nil
}

// MarcarVistas marca un conjunto de alertas como vistas.
func (r *Repositorio) MarcarVistas(ctx context.Context, ids []int64, usuarioID int64, todos bool) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	filas, err := r.consultas.MarcarAlertasVistas(ctx, generado.MarcarAlertasVistasParams{
		Column1:  ids,
		LeidaPor: pgtype.Int8{Int64: usuarioID, Valid: true},
		Column3:  todos,
	})
	return int(filas), err
}

// AsignarResponsable cambia el destinatario de una alerta.
func (r *Repositorio) AsignarResponsable(ctx context.Context, id, usuarioID int64) (bool, error) {
	filas, err := r.consultas.AsignarResponsableAlerta(ctx, generado.AsignarResponsableAlertaParams{
		ID:             id,
		DestinatarioID: usuarioID,
	})
	if err != nil {
		return false, err
	}
	return filas > 0, nil
}

// MarcarSituacionCambiada marca las alertas de un alumno y fecha.
func (r *Repositorio) MarcarSituacionCambiada(ctx context.Context, alumnoID int64, fecha time.Time, motivo string) (int, error) {
	filas, err := r.consultas.MarcarAlertasSituacionCambiada(ctx, generado.MarcarAlertasSituacionCambiadaParams{
		AlumnoID:     alumnoID,
		MotivoCambio: pgtype.Text{String: motivo, Valid: motivo != ""},
		FechaEvento:  fechaParametro(fecha),
	})
	return int(filas), err
}

// RegistrarCierre intenta registrar el cierre diario y devuelve si fue nuevo.
func (r *Repositorio) RegistrarCierre(ctx context.Context, fecha time.Time, turno string) (bool, error) {
	return r.RegistrarCierreCon(ctx, r.consultas, fecha, turno)
}

// RegistrarCierreCon registra el cierre dentro de una transacción abierta.
func (r *Repositorio) RegistrarCierreCon(ctx context.Context, consultas *generado.Queries, fecha time.Time, turno string) (bool, error) {
	filas, err := consultas.RegistrarCierreAlerta(ctx, generado.RegistrarCierreAlertaParams{
		Fecha: fechaParametro(fecha),
		Turno: turno,
	})
	if err != nil {
		return false, err
	}
	return filas > 0, nil
}

// BloquearProcesosCon toma el bloqueo de asesoría dentro de la transacción.
func (r *Repositorio) BloquearProcesosCon(ctx context.Context, consultas *generado.Queries, clave int64) error {
	return consultas.BloquearProcesos(ctx, clave)
}

func aAlerta(fila generado.ListarAlertasRow) Alerta {
	alerta := Alerta{
		ID:                fila.ID,
		Tipo:              fila.Tipo,
		AlumnoID:          fila.AlumnoID,
		SeccionID:         fila.SeccionID,
		CursoID:           enteroDesdePg(fila.CursoID),
		BloqueNumero:      enteroDesdePg32(fila.BloqueNumero),
		FechaEvento:       fila.FechaEvento.Time,
		DestinatarioID:    fila.DestinatarioID,
		Estado:            fila.Estado,
		SinResponsable:    fila.SinResponsable,
		SituacionCambiada: fila.SituacionCambiada,
		MotivoCambio:      fila.MotivoCambio,
		CreadaEn:          fila.CreadaEn.Time,
	}
	if fila.Mes.Valid {
		mes := fila.Mes.Time
		alerta.Mes = &mes
	}
	if fila.MarcacionID.Valid {
		marcacion := fila.MarcacionID.Int64
		alerta.MarcacionID = &marcacion
	}
	if fila.LeidaEn.Valid {
		leida := fila.LeidaEn.Time
		alerta.LeidaEn = &leida
	}
	if fila.LeidaPor.Valid {
		leidaPor := fila.LeidaPor.Int64
		alerta.LeidaPor = &leidaPor
	}
	return alerta
}

func fechaParametro(fecha time.Time) pgtype.Date {
	return pgtype.Date{Time: fecha, Valid: true}
}

func fechaOpcional(fecha *time.Time) pgtype.Date {
	if fecha == nil {
		return pgtype.Date{}
	}
	return pgtype.Date{Time: *fecha, Valid: true}
}

func enteroOpcional(valor *int64) pgtype.Int8 {
	if valor == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: *valor, Valid: true}
}

func entero32Opcional(valor *int) pgtype.Int4 {
	if valor == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: int32(*valor), Valid: true}
}

func enteroDesdePg(valor pgtype.Int8) *int64 {
	if !valor.Valid {
		return nil
	}
	numero := valor.Int64
	return &numero
}

func enteroDesdePg32(valor pgtype.Int4) *int {
	if !valor.Valid {
		return nil
	}
	numero := int(valor.Int32)
	return &numero
}
