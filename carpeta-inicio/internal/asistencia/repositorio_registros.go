package asistencia

import (
	"context"
	"time"

	"asistencia/internal/asistencia/generado"

	"github.com/jackc/pgx/v5/pgtype"
)

// ListarRegistrosClase devuelve los registros de una clase.
func (r *Repositorio) ListarRegistrosClase(ctx context.Context, fecha time.Time, seccionID, cursoID int64, bloque int) ([]RegistroDocente, error) {
	filas, err := r.consultas.ListarRegistrosClase(ctx, generado.ListarRegistrosClaseParams{
		Fecha:        pgtype.Date{Time: fecha, Valid: true},
		SeccionID:    seccionID,
		CursoID:      cursoID,
		BloqueNumero: int32(bloque),
	})
	if err != nil {
		return nil, err
	}
	registros := make([]RegistroDocente, 0, len(filas))
	for _, fila := range filas {
		registros = append(registros, aRegistro(generado.RegistrosDocente(fila)))
	}
	return registros, nil
}

// GuardarRegistro inserta o actualiza el registro de un alumno en una clase.
func (r *Repositorio) GuardarRegistro(ctx context.Context, registro RegistroDocente) (int64, error) {
	return r.consultas.GuardarRegistro(ctx, generado.GuardarRegistroParams{
		AlumnoID:     registro.AlumnoID,
		Fecha:        pgtype.Date{Time: registro.Fecha, Valid: true},
		CursoID:      registro.CursoID,
		SeccionID:    registro.SeccionID,
		BloqueNumero: int32(registro.BloqueNumero),
		Estado:       registro.Estado,
		Observacion:  textoOpcional(registro.Observacion),
		DocenteID:    registro.DocenteID,
	})
}

// InsertarRegistroSiFalta crea un presente solo si el alumno no tiene registro.
func (r *Repositorio) InsertarRegistroSiFalta(ctx context.Context, registro RegistroDocente) (bool, error) {
	filas, err := r.consultas.InsertarRegistroSiFalta(ctx, generado.InsertarRegistroSiFaltaParams{
		AlumnoID:     registro.AlumnoID,
		Fecha:        pgtype.Date{Time: registro.Fecha, Valid: true},
		CursoID:      registro.CursoID,
		SeccionID:    registro.SeccionID,
		BloqueNumero: int32(registro.BloqueNumero),
		DocenteID:    registro.DocenteID,
	})
	if err != nil {
		return false, err
	}
	return filas > 0, nil
}

// BuscarRegistroPorID devuelve el registro con ese identificador.
func (r *Repositorio) BuscarRegistroPorID(ctx context.Context, id int64) (RegistroDocente, error) {
	fila, err := r.consultas.BuscarRegistroPorID(ctx, id)
	if err != nil {
		return RegistroDocente{}, err
	}
	return aRegistro(generado.RegistrosDocente(fila)), nil
}

// CorregirRegistro actualiza un registro existente.
func (r *Repositorio) CorregirRegistro(ctx context.Context, id int64, estado, observacion string, docenteID int64) error {
	_, err := r.consultas.CorregirRegistro(ctx, generado.CorregirRegistroParams{
		ID:          id,
		Estado:      estado,
		Observacion: textoOpcional(observacion),
		DocenteID:   docenteID,
	})
	return err
}

// ListarRegistrosConsulta devuelve los registros de una sección en una fecha.
func (r *Repositorio) ListarRegistrosConsulta(ctx context.Context, fecha time.Time, seccionID int64) ([]RegistroDocente, error) {
	filas, err := r.consultas.ListarRegistrosConsulta(ctx, generado.ListarRegistrosConsultaParams{
		Fecha:     pgtype.Date{Time: fecha, Valid: true},
		SeccionID: seccionID,
	})
	if err != nil {
		return nil, err
	}
	registros := make([]RegistroDocente, 0, len(filas))
	for _, fila := range filas {
		registros = append(registros, aRegistro(generado.RegistrosDocente(fila)))
	}
	return registros, nil
}

// ExportarPorDocente devuelve los registros del docente en el rango.
func (r *Repositorio) ExportarPorDocente(ctx context.Context, desde, hasta time.Time, docenteID int64) ([]RegistroDocente, error) {
	filas, err := r.consultas.ListarRegistrosExportacionDocente(ctx, generado.ListarRegistrosExportacionDocenteParams{
		Fecha:     pgtype.Date{Time: desde, Valid: true},
		Fecha_2:   pgtype.Date{Time: hasta, Valid: true},
		DocenteID: docenteID,
	})
	if err != nil {
		return nil, err
	}
	registros := make([]RegistroDocente, 0, len(filas))
	for _, fila := range filas {
		registros = append(registros, aRegistro(generado.RegistrosDocente(fila)))
	}
	return registros, nil
}

// ExportarPorSeccion devuelve los registros de una sección en el rango.
func (r *Repositorio) ExportarPorSeccion(ctx context.Context, desde, hasta time.Time, seccionID, cursoID int64) ([]RegistroDocente, error) {
	filas, err := r.consultas.ListarRegistrosExportacionSeccion(ctx, generado.ListarRegistrosExportacionSeccionParams{
		Fecha:     pgtype.Date{Time: desde, Valid: true},
		Fecha_2:   pgtype.Date{Time: hasta, Valid: true},
		SeccionID: seccionID,
		Column4:   cursoID,
	})
	if err != nil {
		return nil, err
	}
	registros := make([]RegistroDocente, 0, len(filas))
	for _, fila := range filas {
		registros = append(registros, aRegistro(generado.RegistrosDocente(fila)))
	}
	return registros, nil
}

func aRegistro(fila generado.RegistrosDocente) RegistroDocente {
	return RegistroDocente{
		ID:            fila.ID,
		AlumnoID:      fila.AlumnoID,
		Fecha:         fila.Fecha.Time,
		CursoID:       fila.CursoID,
		SeccionID:     fila.SeccionID,
		BloqueNumero:  int(fila.BloqueNumero),
		Estado:        fila.Estado,
		Observacion:   fila.Observacion.String,
		DocenteID:     fila.DocenteID,
		CreadoEn:      fila.CreadoEn.Time,
		ActualizadoEn: fila.ActualizadoEn.Time,
	}
}

func textoOpcional(texto string) pgtype.Text {
	return pgtype.Text{String: texto, Valid: texto != ""}
}
