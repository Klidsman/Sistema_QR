package alertas

import (
	"context"
	"errors"
	"time"

	"asistencia/internal/alertas/generado"
	"asistencia/internal/plataforma/auditoria"
)

// Errores del servicio de alertas.
var (
	ErrAlertaNoEncontrada = errors.New("la alerta no existe")
	ErrAlertaSinAcceso    = errors.New("no tiene acceso a esta alerta")
)

// ListarAlertas devuelve las alertas visibles para el usuario.
func (s *Servicio) ListarAlertas(ctx context.Context, filtros FiltrosAlerta) ([]Alerta, error) {
	return s.dependencias.Repositorio.ListarAlertas(ctx, filtros)
}

// AlertaDetalle agrega datos legibles a una alerta para el panel.
type AlertaDetalle struct {
	Alerta
	AlumnoNombre    string
	Identificador   string
	SeccionEtiqueta string
}

// Detallar completa las alertas con nombres de alumno y sección.
func (s *Servicio) Detallar(ctx context.Context, alertas []Alerta) ([]AlertaDetalle, error) {
	alumnos := make(map[int64]AlumnoEvento)
	secciones := make(map[int64]string)
	detalles := make([]AlertaDetalle, 0, len(alertas))
	for _, alerta := range alertas {
		alumno, existe := alumnos[alerta.AlumnoID]
		if !existe {
			var err error
			alumno, err = s.dependencias.Alumnos.AlumnoPorID(ctx, alerta.AlumnoID)
			if err != nil {
				return nil, err
			}
			alumnos[alerta.AlumnoID] = alumno
		}
		etiqueta, existe := secciones[alerta.SeccionID]
		if !existe {
			var err error
			etiqueta, err = s.dependencias.Academico.EtiquetaSeccion(ctx, alerta.SeccionID)
			if err != nil {
				return nil, err
			}
			secciones[alerta.SeccionID] = etiqueta
		}
		detalles = append(detalles, AlertaDetalle{
			Alerta:          alerta,
			AlumnoNombre:    alumno.Apellidos + ", " + alumno.Nombres,
			Identificador:   alumno.Identificador,
			SeccionEtiqueta: etiqueta,
		})
	}
	return detalles, nil
}

// EtiquetaTipo devuelve el texto visible de un tipo de alerta.
func EtiquetaTipo(tipo string) string {
	switch tipo {
	case TipoTardanza:
		return "Tardanza"
	case TipoReiteracionTardanza:
		return "Reiteración de tardanzas"
	case TipoReiteracionFalta:
		return "Faltas seguidas"
	default:
		return tipo
	}
}

// ContarNuevas devuelve la cantidad de alertas nuevas del usuario.
func (s *Servicio) ContarNuevas(ctx context.Context, destinatarioID int64, todos bool) (int, error) {
	return s.dependencias.Repositorio.ContarNuevas(ctx, destinatarioID, todos)
}

// MarcarVista marca una alerta como vista.
func (s *Servicio) MarcarVista(ctx context.Context, id, usuarioID int64, todos bool) error {
	return s.dependencias.Repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		marcada, err := s.dependencias.Repositorio.MarcarVista(ctx, id, usuarioID, todos)
		if err != nil {
			return err
		}
		if !marcada {
			return ErrAlertaSinAcceso
		}
		return auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID:  usuarioID,
			Entidad:    EntidadAlerta,
			EntidadID:  id,
			Accion:     AccionMarcarVista,
			ValorNuevo: map[string]any{"estado": EstadoVista},
		})
	})
}

// MarcarVistas marca como vistas las alertas indicadas.
func (s *Servicio) MarcarVistas(ctx context.Context, ids []int64, usuarioID int64, todos bool) (int, error) {
	var total int
	err := s.dependencias.Repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		marcadas, err := s.dependencias.Repositorio.MarcarVistas(ctx, ids, usuarioID, todos)
		if err != nil {
			return err
		}
		total = marcadas
		if marcadas == 0 {
			return nil
		}
		return auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID:  usuarioID,
			Entidad:    EntidadAlerta,
			EntidadID:  0,
			Accion:     AccionMarcarVistas,
			ValorNuevo: map[string]any{"ids": ids, "marcadas": marcadas},
		})
	})
	return total, err
}

// AsignarResponsable cambia el destinatario de una alerta sin responsable.
func (s *Servicio) AsignarResponsable(ctx context.Context, id, usuarioID, destinatarioID int64) error {
	return s.dependencias.Repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		asignada, err := s.dependencias.Repositorio.AsignarResponsable(ctx, id, destinatarioID)
		if err != nil {
			return err
		}
		if !asignada {
			return ErrAlertaNoEncontrada
		}
		return auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID:  usuarioID,
			Entidad:    EntidadAlerta,
			EntidadID:  id,
			Accion:     AccionAsignarResponsable,
			ValorNuevo: map[string]any{"destinatario_id": destinatarioID},
		})
	})
}

// MarcarSituacionCambiada marca las alertas de un alumno y fecha como corregidas.
func (s *Servicio) MarcarSituacionCambiada(ctx context.Context, alumnoID int64, fecha time.Time, motivo string) error {
	return s.dependencias.Repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		cambiadas, err := s.dependencias.Repositorio.MarcarSituacionCambiada(ctx, alumnoID, fecha, motivo)
		if err != nil {
			return err
		}
		if cambiadas == 0 {
			return nil
		}
		return auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID:  s.administradorID(ctx),
			Entidad:    EntidadAlerta,
			EntidadID:  0,
			Accion:     AccionMarcarSituacionCambiada,
			ValorNuevo: map[string]any{"alumno_id": alumnoID, "fecha": fecha.Format("2006-01-02"), "motivo": motivo},
		})
	})
}

func (s *Servicio) administradorID(ctx context.Context) int64 {
	id, err := s.dependencias.Usuarios.AdministradorActivo(ctx)
	if err != nil {
		return 0
	}
	return id
}
