package academico

import (
	"context"
	"time"

	"asistencia/internal/academico/generado"

	"github.com/jackc/pgx/v5/pgtype"
)

// CrearBloque registra una versión de un bloque de hora de clase.
func (r *RepositorioConfiguracion) CrearBloque(ctx context.Context, bloque BloqueClase) (int64, error) {
	return r.consultas.CrearBloqueClase(ctx, generado.CrearBloqueClaseParams{
		Turno:        bloque.Turno,
		Numero:       int32(bloque.Numero),
		HoraInicio:   parametroHora(bloque.HoraInicio),
		HoraFin:      parametroHora(bloque.HoraFin),
		VigenteDesde: parametroFecha(bloque.VigenteDesde),
	})
}

// BloquesVigentesTurno devuelve los bloques del turno vigentes en la fecha.
func (r *RepositorioConfiguracion) BloquesVigentesTurno(ctx context.Context, turno string, fecha time.Time) ([]BloqueClase, error) {
	filas, err := r.consultas.BloquesVigentesTurno(ctx, generado.BloquesVigentesTurnoParams{
		Turno:        turno,
		VigenteDesde: parametroFecha(fecha),
	})
	if err != nil {
		return nil, err
	}
	bloques := make([]BloqueClase, 0, len(filas))
	for _, fila := range filas {
		bloque, err := aBloqueClase(generado.BloquesVigentesTurnoRow(fila))
		if err != nil {
			return nil, err
		}
		bloques = append(bloques, bloque)
	}
	return bloques, nil
}

// ListarBloquesTurno devuelve la historia de bloques de un turno.
func (r *RepositorioConfiguracion) ListarBloquesTurno(ctx context.Context, turno string) ([]BloqueClase, error) {
	filas, err := r.consultas.ListarBloquesTurno(ctx, turno)
	if err != nil {
		return nil, err
	}
	bloques := make([]BloqueClase, 0, len(filas))
	for _, fila := range filas {
		bloque, err := aBloqueClase(generado.BloquesVigentesTurnoRow(fila))
		if err != nil {
			return nil, err
		}
		bloques = append(bloques, bloque)
	}
	return bloques, nil
}

// BuscarBloquePorID devuelve el bloque con ese identificador.
func (r *RepositorioConfiguracion) BuscarBloquePorID(ctx context.Context, id int64) (BloqueClase, error) {
	fila, err := r.consultas.BuscarBloquePorID(ctx, id)
	if err != nil {
		return BloqueClase{}, err
	}
	return aBloqueClase(generado.BloquesVigentesTurnoRow(fila))
}

// BloqueVigentePorNumero devuelve el bloque del número vigente en la fecha.
func (r *RepositorioConfiguracion) BloqueVigentePorNumero(ctx context.Context, turno string, numero int, fecha time.Time) (BloqueClase, error) {
	fila, err := r.consultas.BloqueVigentePorNumero(ctx, generado.BloqueVigentePorNumeroParams{
		Turno:        turno,
		Numero:       int32(numero),
		VigenteDesde: parametroFecha(fecha),
	})
	if err != nil {
		return BloqueClase{}, err
	}
	return aBloqueClase(generado.BloquesVigentesTurnoRow(fila))
}

// GuardarCeldaHorario registra o reemplaza una celda del horario semanal.
func (r *RepositorioConfiguracion) GuardarCeldaHorario(ctx context.Context, celda HorarioSeccion) error {
	var cursoID pgtype.Int8
	if celda.CursoID != nil {
		cursoID = pgtype.Int8{Int64: *celda.CursoID, Valid: true}
	}
	return r.consultas.UpsertHorarioSeccion(ctx, generado.UpsertHorarioSeccionParams{
		SeccionID:    celda.SeccionID,
		DiaSemana:    int16(celda.DiaSemana),
		BloqueNumero: int32(celda.BloqueNumero),
		CursoID:      cursoID,
		VigenteDesde: parametroFecha(celda.VigenteDesde),
	})
}

// HorarioSeccionVigente devuelve las celdas vigentes de una sección en la fecha.
func (r *RepositorioConfiguracion) HorarioSeccionVigente(ctx context.Context, seccionID int64, fecha time.Time) ([]HorarioSeccion, error) {
	filas, err := r.consultas.HorarioSeccionVigente(ctx, generado.HorarioSeccionVigenteParams{
		SeccionID:    seccionID,
		VigenteDesde: parametroFecha(fecha),
	})
	if err != nil {
		return nil, err
	}
	return aHorarioSeccion(filas), nil
}

// HorarioSeccionCelda devuelve la celda vigente de un día y bloque.
func (r *RepositorioConfiguracion) HorarioSeccionCelda(ctx context.Context, seccionID int64, dia, bloque int, fecha time.Time) (HorarioSeccion, error) {
	fila, err := r.consultas.HorarioSeccionCelda(ctx, generado.HorarioSeccionCeldaParams{
		SeccionID:    seccionID,
		DiaSemana:    int16(dia),
		BloqueNumero: int32(bloque),
		VigenteDesde: parametroFecha(fecha),
	})
	if err != nil {
		return HorarioSeccion{}, err
	}
	return HorarioSeccion{
		ID:           fila.ID,
		SeccionID:    fila.SeccionID,
		DiaSemana:    int(fila.DiaSemana),
		BloqueNumero: int(fila.BloqueNumero),
		CursoID:      enteroOpcional(fila.CursoID),
		VigenteDesde: fila.VigenteDesde.Time,
	}, nil
}

func aBloqueClase(fila generado.BloquesVigentesTurnoRow) (BloqueClase, error) {
	inicio, err := horaDeTexto(fila.HoraInicio)
	if err != nil {
		return BloqueClase{}, err
	}
	fin, err := horaDeTexto(fila.HoraFin)
	if err != nil {
		return BloqueClase{}, err
	}
	return BloqueClase{
		ID:           fila.ID,
		Turno:        fila.Turno,
		Numero:       int(fila.Numero),
		HoraInicio:   inicio,
		HoraFin:      fin,
		VigenteDesde: fila.VigenteDesde.Time,
	}, nil
}

func aHorarioSeccion(filas []generado.HorarioSeccionVigenteRow) []HorarioSeccion {
	celdas := make([]HorarioSeccion, 0, len(filas))
	for _, fila := range filas {
		celdas = append(celdas, HorarioSeccion{
			ID:           fila.ID,
			SeccionID:    fila.SeccionID,
			DiaSemana:    int(fila.DiaSemana),
			BloqueNumero: int(fila.BloqueNumero),
			CursoID:      enteroOpcional(fila.CursoID),
			VigenteDesde: fila.VigenteDesde.Time,
		})
	}
	return celdas
}

func enteroOpcional(valor pgtype.Int8) *int64 {
	if !valor.Valid {
		return nil
	}
	numero := valor.Int64
	return &numero
}
