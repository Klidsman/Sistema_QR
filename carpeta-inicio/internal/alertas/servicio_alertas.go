package alertas

import (
	"context"
	"fmt"
	"time"

	"asistencia/internal/alertas/generado"
	"asistencia/internal/plataforma/auditoria"
)

// FuenteJustificaciones informa si una falta fue justificada.
// La feature 005 reemplazará la implementación inicial, que considera
// injustificada toda ausencia oficial.
type FuenteJustificaciones interface {
	EsFaltaJustificada(ctx context.Context, alumnoID int64, fecha time.Time) (bool, error)
}

// Servicio aplica las reglas de generación y consulta de alertas.
type Servicio struct {
	dependencias    Dependencias
	justificaciones FuenteJustificaciones
}

// NuevoServicio crea el servicio de alertas.
func NuevoServicio(dependencias Dependencias) *Servicio {
	if dependencias.Reloj == nil {
		dependencias.Reloj = time.Now
	}
	return &Servicio{dependencias: dependencias}
}

// ConJustificaciones permite inyectar la fuente de la feature 005.
func (s *Servicio) ConJustificaciones(fuente FuenteJustificaciones) *Servicio {
	s.justificaciones = fuente
	return s
}

func (s *Servicio) ahora() time.Time {
	return s.dependencias.Reloj()
}

func (s *Servicio) administrador(ctx context.Context) (int64, error) {
	return s.dependencias.Usuarios.AdministradorActivo(ctx)
}

func (s *Servicio) esJustificada(ctx context.Context, alumnoID int64, fecha time.Time) (bool, error) {
	if s.justificaciones == nil {
		return false, nil
	}
	return s.justificaciones.EsFaltaJustificada(ctx, alumnoID, fecha)
}

// ProcesarTardanzas convierte las marcaciones de tardanza en alertas.
// Es idempotente: los índices únicos impiden duplicados.
func (s *Servicio) ProcesarTardanzas(ctx context.Context) error {
	hoy := s.ahora()
	desde := hoy.AddDate(0, 0, -2)
	return s.dependencias.Repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		if err := s.dependencias.Repositorio.BloquearProcesosCon(ctx, consultas, claveBloqueoTardanzas); err != nil {
			return err
		}

		marcaciones, err := s.dependencias.Asistencia.MarcacionesDelRango(ctx, desde, hoy)
		if err != nil {
			return fmt.Errorf("leyendo las marcaciones del rango: %w", err)
		}

		for _, marcacion := range marcaciones {
			if marcacion.Estado != "tardanza" && marcacion.Estado != "tardanza_grave" {
				continue
			}
			seccionID, err := s.dependencias.Alumnos.SeccionDeAlumnoEnFecha(ctx, marcacion.AlumnoID, marcacion.Fecha)
			if err != nil {
				return err
			}
			if seccionID == 0 {
				continue
			}

			destinatarioID, sinResponsable, err := s.destinatarioTardanza(ctx, marcacion.RegistradaPor)
			if err != nil {
				return err
			}
			marcacionID := marcacion.MarcacionID
			id, creada, err := s.dependencias.Repositorio.InsertarAlertaCon(ctx, consultas, Alerta{
				Tipo:           TipoTardanza,
				AlumnoID:       marcacion.AlumnoID,
				SeccionID:      seccionID,
				FechaEvento:    marcacion.Fecha,
				MarcacionID:    &marcacionID,
				DestinatarioID: destinatarioID,
				SinResponsable: sinResponsable,
			})
			if err != nil {
				return fmt.Errorf("creando la alerta de tardanza: %w", err)
			}
			if !creada {
				continue
			}
			if err := s.auditarAlerta(ctx, db, id, destinatarioID, TipoTardanza, map[string]any{
				"alumno_id":       marcacion.AlumnoID,
				"seccion_id":      seccionID,
				"marcacion_id":    marcacionID,
				"destinatario":    destinatarioID,
				"sin_responsable": sinResponsable,
			}); err != nil {
				return err
			}

			if err := s.evaluarReiteracionTardanzas(ctx, consultas, db, marcacion.AlumnoID, seccionID, marcacion.Fecha); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Servicio) destinatarioTardanza(ctx context.Context, registradaPor *int64) (int64, bool, error) {
	administradorID, err := s.administrador(ctx)
	if err != nil {
		return 0, false, err
	}
	destinatario := DestinatarioTardanza(registradaPor, administradorID)
	return destinatario, destinatario == administradorID, nil
}

func (s *Servicio) evaluarReiteracionTardanzas(
	ctx context.Context,
	consultas *generado.Queries,
	db generado.DBTX,
	alumnoID, seccionID int64,
	fecha time.Time,
) error {
	umbral, existe, err := s.dependencias.Academico.UmbralVigente(ctx, "tardanzas", fecha)
	if err != nil {
		return err
	}
	if !existe {
		return nil
	}

	inicioMes := time.Date(fecha.Year(), fecha.Month(), 1, 0, 0, 0, 0, fecha.Location())
	finMes := inicioMes.AddDate(0, 1, -1)
	conteo, err := s.dependencias.Asistencia.ContarTardanzas(ctx, alumnoID, inicioMes, finMes)
	if err != nil {
		return err
	}
	if !AlcanzaUmbral(conteo, umbral) {
		return nil
	}

	destinatarios, sinResponsable, err := s.destinatariosDeSeccion(ctx, seccionID, fecha)
	if err != nil {
		return err
	}
	for _, destinatarioID := range destinatarios {
		id, creada, err := s.dependencias.Repositorio.InsertarAlertaCon(ctx, consultas, Alerta{
			Tipo:           TipoReiteracionTardanza,
			AlumnoID:       alumnoID,
			SeccionID:      seccionID,
			FechaEvento:    fecha,
			Mes:            &inicioMes,
			DestinatarioID: destinatarioID,
			SinResponsable: sinResponsable,
		})
		if err != nil {
			return fmt.Errorf("creando la reiteración de tardanzas: %w", err)
		}
		if !creada {
			continue
		}
		if err := s.auditarAlerta(ctx, db, id, destinatarioID, TipoReiteracionTardanza, map[string]any{
			"alumno_id":       alumnoID,
			"seccion_id":      seccionID,
			"conteo":          conteo,
			"umbral":          umbral,
			"sin_responsable": sinResponsable,
		}); err != nil {
			return err
		}
	}
	return nil
}

// ProcesarCierre evalúa las faltas seguidas de un turno y registra el cierre.
func (s *Servicio) ProcesarCierre(ctx context.Context, fecha time.Time, turno string) error {
	return s.dependencias.Repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		if err := s.dependencias.Repositorio.BloquearProcesosCon(ctx, consultas, claveBloqueoCierre); err != nil {
			return err
		}
		nuevo, err := s.dependencias.Repositorio.RegistrarCierreCon(ctx, consultas, fecha, turno)
		if err != nil {
			return err
		}
		if !nuevo {
			return nil
		}

		administradorID, err := s.administrador(ctx)
		if err != nil {
			return err
		}
		if err := auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID: administradorID,
			Entidad:   EntidadCierreAlerta,
			EntidadID: 0,
			Accion:    AccionEjecutarCierre,
			ValorNuevo: map[string]any{
				"fecha": fecha.Format("2006-01-02"),
				"turno": turno,
			},
		}); err != nil {
			return err
		}

		umbral, existe, err := s.dependencias.Academico.UmbralVigente(ctx, "faltas_seguidas", fecha)
		if err != nil {
			return err
		}
		if !existe {
			return nil
		}

		secciones, err := s.dependencias.Academico.SeccionesActivas(ctx)
		if err != nil {
			return err
		}
		for _, seccion := range secciones {
			if seccion.Turno != turno {
				continue
			}
			if err := s.evaluarSeccion(ctx, consultas, db, seccion.ID, fecha, umbral); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Servicio) evaluarSeccion(
	ctx context.Context,
	consultas *generado.Queries,
	db generado.DBTX,
	seccionID int64,
	fecha time.Time,
	umbral int,
) error {
	alumnos, err := s.dependencias.Alumnos.AlumnosActivosDeSeccionEnFecha(ctx, seccionID, fecha)
	if err != nil {
		return err
	}
	for _, alumno := range alumnos {
		racha, err := s.rachaFaltas(ctx, seccionID, alumno.ID, fecha, umbral)
		if err != nil {
			return err
		}
		if !AlcanzaUmbral(racha, umbral) {
			continue
		}
		destinatarios, sinResponsable, err := s.destinatariosDeSeccion(ctx, seccionID, fecha)
		if err != nil {
			return err
		}
		for _, destinatarioID := range destinatarios {
			fechaEvento := fecha
			id, creada, err := s.dependencias.Repositorio.InsertarAlertaCon(ctx, consultas, Alerta{
				Tipo:           TipoReiteracionFalta,
				AlumnoID:       alumno.ID,
				SeccionID:      seccionID,
				FechaEvento:    fechaEvento,
				DestinatarioID: destinatarioID,
				SinResponsable: sinResponsable,
			})
			if err != nil {
				return fmt.Errorf("creando la reiteración de faltas: %w", err)
			}
			if !creada {
				continue
			}
			if err := s.auditarAlerta(ctx, db, id, destinatarioID, TipoReiteracionFalta, map[string]any{
				"alumno_id":       alumno.ID,
				"seccion_id":      seccionID,
				"racha":           racha,
				"umbral":          umbral,
				"sin_responsable": sinResponsable,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Servicio) destinatariosDeSeccion(ctx context.Context, seccionID int64, fecha time.Time) ([]int64, bool, error) {
	docentes, err := s.dependencias.Asistencia.DocentesDeSeccionEnFecha(ctx, seccionID, fecha)
	if err != nil {
		return nil, false, err
	}
	if len(docentes) > 0 {
		return docentes, false, nil
	}
	administradorID, err := s.administrador(ctx)
	if err != nil {
		return nil, false, err
	}
	return []int64{administradorID}, true, nil
}

// rachaFaltas cuenta los días laborables consecutivos sin marcación oficial.
// Deja de contar al superar el umbral: así solo se alerta el día exacto en que
// se cruza la racha y no se repite la alerta cada día siguiente.
func (s *Servicio) rachaFaltas(ctx context.Context, seccionID, alumnoID int64, fecha time.Time, umbral int) (int, error) {
	if umbral <= 0 {
		return 0, nil
	}
	cursor := fecha
	racha := 0
	for dias := 0; dias < 366; dias++ {
		if cursor.Weekday() == time.Saturday || cursor.Weekday() == time.Sunday {
			cursor = cursor.AddDate(0, 0, -1)
			continue
		}
		noLaborable, err := s.dependencias.Academico.EsDiaNoLaborable(ctx, seccionID, cursor)
		if err != nil {
			return 0, err
		}
		if noLaborable {
			cursor = cursor.AddDate(0, 0, -1)
			continue
		}
		tiene, err := s.dependencias.Asistencia.TieneMarcacion(ctx, alumnoID, cursor)
		if err != nil {
			return 0, err
		}
		if tiene {
			break
		}
		justificada, err := s.esJustificada(ctx, alumnoID, cursor)
		if err != nil {
			return 0, err
		}
		if justificada {
			break
		}
		racha++
		if racha > umbral {
			return racha, nil
		}
		cursor = cursor.AddDate(0, 0, -1)
	}
	return racha, nil
}

func (s *Servicio) auditarAlerta(ctx context.Context, db generado.DBTX, id, usuarioID int64, tipo string, datos map[string]any) error {
	datos["tipo"] = tipo
	return auditoria.Registrar(ctx, db, auditoria.Evento{
		UsuarioID:  usuarioID,
		Entidad:    EntidadAlerta,
		EntidadID:  id,
		Accion:     AccionCrearAlerta,
		ValorNuevo: datos,
	})
}

const (
	claveBloqueoTardanzas int64 = 404001
	claveBloqueoCierre    int64 = 404002
)
