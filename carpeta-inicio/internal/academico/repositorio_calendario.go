package academico

import (
	"context"
	"time"

	"asistencia/internal/academico/generado"
)

// GuardarHorarioTurno registra o reemplaza la versión del horario de un turno.
func (r *RepositorioConfiguracion) GuardarHorarioTurno(ctx context.Context, horario HorarioTurno) (int64, error) {
	return r.consultas.GuardarHorarioTurno(ctx, generado.GuardarHorarioTurnoParams{
		Turno:              horario.Turno,
		HoraIngreso:        parametroHora(horario.HoraIngreso),
		ToleranciaMinutos:  int32(horario.ToleranciaMinutos),
		HoraLimiteTardanza: parametroHora(horario.HoraLimiteTardanza),
		HoraCierre:         parametroHora(horario.HoraCierre),
		VigenteDesde:       parametroFecha(horario.VigenteDesde),
	})
}

// HorarioTurnoVigente devuelve el horario del turno vigente en la fecha.
func (r *RepositorioConfiguracion) HorarioTurnoVigente(ctx context.Context, turno string, fecha time.Time) (HorarioTurno, error) {
	fila, err := r.consultas.HorarioTurnoVigente(ctx, generado.HorarioTurnoVigenteParams{
		Turno:        turno,
		VigenteDesde: parametroFecha(fecha),
	})
	if err != nil {
		return HorarioTurno{}, err
	}
	return aHorarioTurno(fila)
}

// ListarHorariosTurno devuelve la historia de horarios de un turno.
func (r *RepositorioConfiguracion) ListarHorariosTurno(ctx context.Context, turno string) ([]HorarioTurno, error) {
	filas, err := r.consultas.ListarHorariosTurno(ctx, turno)
	if err != nil {
		return nil, err
	}
	horarios := make([]HorarioTurno, 0, len(filas))
	for _, fila := range filas {
		horario, err := aHorarioTurno(generado.HorarioTurnoVigenteRow(fila))
		if err != nil {
			return nil, err
		}
		horarios = append(horarios, horario)
	}
	return horarios, nil
}

// CrearExcepcion registra una excepción de calendario y sus grados alcanzados.
func (r *RepositorioConfiguracion) CrearExcepcion(ctx context.Context, excepcion ExcepcionCalendario) (int64, error) {
	var id int64
	err := r.EnTransaccion(ctx, func(consultas *generado.Queries, _ generado.DBTX) error {
		creado, err := consultas.CrearExcepcionCalendario(ctx, generado.CrearExcepcionCalendarioParams{
			AnioEscolarID: excepcion.AnioEscolarID,
			Tipo:          excepcion.Tipo,
			FechaInicio:   parametroFecha(excepcion.FechaInicio),
			FechaFin:      parametroFecha(excepcion.FechaFin),
			Motivo:        excepcion.Motivo,
			CreadoPor:     excepcion.CreadoPor,
		})
		if err != nil {
			return err
		}
		for _, gradoID := range excepcion.Grados {
			if err := consultas.CrearExcepcionGrado(ctx, generado.CrearExcepcionGradoParams{
				ExcepcionID: creado,
				GradoID:     gradoID,
			}); err != nil {
				return err
			}
		}
		id = creado
		return nil
	})
	return id, err
}

// ListarExcepcionesPorAnio devuelve las excepciones del año con sus grados.
func (r *RepositorioConfiguracion) ListarExcepcionesPorAnio(ctx context.Context, anioID int64) ([]ExcepcionCalendario, error) {
	filas, err := r.consultas.ListarExcepcionesPorAnio(ctx, anioID)
	if err != nil {
		return nil, err
	}
	grados, err := r.consultas.ListarGradosDeExcepciones(ctx, anioID)
	if err != nil {
		return nil, err
	}
	porExcepcion := make(map[int64][]int64, len(filas))
	for _, fila := range grados {
		porExcepcion[fila.ExcepcionID] = append(porExcepcion[fila.ExcepcionID], fila.GradoID)
	}

	excepciones := make([]ExcepcionCalendario, 0, len(filas))
	for _, fila := range filas {
		excepciones = append(excepciones, ExcepcionCalendario{
			ID:            fila.ID,
			AnioEscolarID: fila.AnioEscolarID,
			Tipo:          fila.Tipo,
			FechaInicio:   fila.FechaInicio.Time,
			FechaFin:      fila.FechaFin.Time,
			Motivo:        fila.Motivo,
			Grados:        porExcepcion[fila.ID],
			CreadoPor:     fila.CreadoPor,
			CreadoEn:      fila.CreadoEn.Time,
		})
	}
	return excepciones, nil
}

// ExcepcionesDeSeccion devuelve las excepciones que alcanzan una sección en la fecha.
func (r *RepositorioConfiguracion) ExcepcionesDeSeccion(ctx context.Context, seccionID int64, fecha time.Time) ([]ExcepcionCalendario, error) {
	filas, err := r.consultas.ExcepcionesQueAlcanzanSeccion(ctx, generado.ExcepcionesQueAlcanzanSeccionParams{
		ID:          seccionID,
		FechaInicio: parametroFecha(fecha),
	})
	if err != nil {
		return nil, err
	}
	excepciones := make([]ExcepcionCalendario, 0, len(filas))
	for _, fila := range filas {
		excepciones = append(excepciones, ExcepcionCalendario{
			ID:            fila.ID,
			AnioEscolarID: fila.AnioEscolarID,
			Tipo:          fila.Tipo,
			FechaInicio:   fila.FechaInicio.Time,
			FechaFin:      fila.FechaFin.Time,
			Motivo:        fila.Motivo,
			CreadoPor:     fila.CreadoPor,
			CreadoEn:      fila.CreadoEn.Time,
		})
	}
	return excepciones, nil
}

// CrearUmbral registra una versión del umbral de alerta.
func (r *RepositorioConfiguracion) CrearUmbral(ctx context.Context, umbral UmbralAlerta) (int64, error) {
	return r.consultas.CrearUmbralAlerta(ctx, generado.CrearUmbralAlertaParams{
		AnioEscolarID: umbral.AnioEscolarID,
		Tipo:          umbral.Tipo,
		Valor:         int32(umbral.Valor),
		VigenteDesde:  parametroFecha(umbral.VigenteDesde),
	})
}

// UmbralVigente devuelve el umbral del tipo vigente en la fecha.
func (r *RepositorioConfiguracion) UmbralVigente(ctx context.Context, anioID int64, tipo string, fecha time.Time) (UmbralAlerta, error) {
	fila, err := r.consultas.UmbralVigente(ctx, generado.UmbralVigenteParams{
		AnioEscolarID: anioID,
		Tipo:          tipo,
		VigenteDesde:  parametroFecha(fecha),
	})
	if err != nil {
		return UmbralAlerta{}, err
	}
	return UmbralAlerta{
		ID:            fila.ID,
		AnioEscolarID: fila.AnioEscolarID,
		Tipo:          fila.Tipo,
		Valor:         int(fila.Valor),
		VigenteDesde:  fila.VigenteDesde.Time,
	}, nil
}

// ListarUmbralesPorAnio devuelve la historia de umbrales del año.
func (r *RepositorioConfiguracion) ListarUmbralesPorAnio(ctx context.Context, anioID int64) ([]UmbralAlerta, error) {
	filas, err := r.consultas.ListarUmbralesPorAnio(ctx, anioID)
	if err != nil {
		return nil, err
	}
	umbrales := make([]UmbralAlerta, 0, len(filas))
	for _, fila := range filas {
		umbrales = append(umbrales, UmbralAlerta{
			ID:            fila.ID,
			AnioEscolarID: fila.AnioEscolarID,
			Tipo:          fila.Tipo,
			Valor:         int(fila.Valor),
			VigenteDesde:  fila.VigenteDesde.Time,
		})
	}
	return umbrales, nil
}

func aHorarioTurno(fila generado.HorarioTurnoVigenteRow) (HorarioTurno, error) {
	ingreso, err := horaDeTexto(fila.HoraIngreso)
	if err != nil {
		return HorarioTurno{}, err
	}
	limite, err := horaDeTexto(fila.HoraLimiteTardanza)
	if err != nil {
		return HorarioTurno{}, err
	}
	cierre, err := horaDeTexto(fila.HoraCierre)
	if err != nil {
		return HorarioTurno{}, err
	}
	return HorarioTurno{
		ID:                 fila.ID,
		Turno:              fila.Turno,
		HoraIngreso:        ingreso,
		ToleranciaMinutos:  int(fila.ToleranciaMinutos),
		HoraLimiteTardanza: limite,
		HoraCierre:         cierre,
		VigenteDesde:       fila.VigenteDesde.Time,
	}, nil
}
