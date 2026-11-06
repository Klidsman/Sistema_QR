package reportes

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"time"

	"asistencia/internal/academico"
	"asistencia/internal/asistencia"
	"asistencia/internal/plataforma/fechas"
	"asistencia/internal/reportes/generado"
)

// FuenteAcademica aporta horarios y umbrales.
type FuenteAcademica interface {
	HorarioVigente(ctx context.Context, turno string, fecha time.Time) (academico.HorarioTurno, error)
	UmbralVigente(ctx context.Context, tipo string, fecha time.Time) (int, bool, error)
}

// Dependencias reúne las fuentes del módulo.
type Dependencias struct {
	Repo      *Repositorio
	Academico FuenteAcademica
	Reloj     func() time.Time
}

// Servicio calcula los reportes.
type Servicio struct {
	dependencias Dependencias
}

// NuevoServicio crea el servicio de reportes.
func NuevoServicio(dependencias Dependencias) *Servicio {
	if dependencias.Reloj == nil {
		dependencias.Reloj = time.Now
	}
	return &Servicio{dependencias: dependencias}
}

// Diario devuelve el reporte de un día con los filtros aplicados.
func (s *Servicio) Diario(ctx context.Context, filtros Filtros) ([]FilaDiaria, error) {
	fecha := filtros.Desde
	alumnos, err := s.dependencias.Repo.AlumnosEnFecha(ctx, fecha)
	if err != nil {
		return nil, err
	}
	marcaciones, err := s.dependencias.Repo.MarcacionesDelDia(ctx, fecha)
	if err != nil {
		return nil, err
	}
	justificaciones, err := s.dependencias.Repo.JustificacionesAprobadas(ctx, fecha, fecha)
	if err != nil {
		return nil, err
	}
	permitidas, err := s.seccionesPermitidas(ctx, filtros)
	if err != nil {
		return nil, err
	}
	porAlumno := make(map[int64]generado.MarcacionesDelDiaRow, len(marcaciones))
	for _, marcacion := range marcaciones {
		porAlumno[marcacion.AlumnoID] = marcacion
	}
	justificadas := make(map[string]bool, len(justificaciones))
	for _, justificacion := range justificaciones {
		justificadas[claveDia(justificacion.AlumnoID, justificacion.FechaEvento.Time)] = true
	}
	excepciones, err := s.dependencias.Repo.Excepciones(ctx, fecha, fecha)
	if err != nil {
		return nil, err
	}
	noLaborables, porGrado := mapaExcepciones(excepciones)
	hoy := fechas.Fecha(s.dependencias.Reloj())
	cerrado, err := s.estaCerrado(ctx, fecha, hoy, alumnos)
	if err != nil {
		return nil, err
	}

	filas := make([]FilaDiaria, 0, len(alumnos))
	for _, alumno := range alumnos {
		if !permitida(permitidas, alumno.SeccionID) {
			continue
		}
		if fecha.Weekday() == time.Saturday || fecha.Weekday() == time.Sunday {
			continue
		}
		if esNoLaborable(noLaborables, porGrado, fecha, alumno.GradoID) {
			continue
		}
		marcacion, existe := porAlumno[alumno.ID]
		estado := EstadoPendiente
		hora := ""
		if existe {
			hora = fechas.Hora(marcacion.Hora.Time)
			switch marcacion.Estado {
			case asistencia.MarcacionPuntual:
				estado = EstadoPresente
			case asistencia.MarcacionTardanza, asistencia.MarcacionTardanzaGrave:
				if marcacion.Justificada {
					estado = EstadoTardanzaJustificada
				} else {
					estado = EstadoTardanza
				}
			}
		} else if justificadas[claveDia(alumno.ID, fecha)] {
			estado = EstadoFaltaJustificada
		} else if cerrado {
			estado = EstadoFalta
		}

		fila := FilaDiaria{
			AlumnoID:      alumno.ID,
			Identificador: alumno.Identificador,
			Alumno:        alumno.Apellidos + ", " + alumno.Nombres,
			Grado:         alumno.Grado,
			Seccion:       alumno.Seccion,
			Turno:         alumno.Turno,
			Estado:        estado,
			Hora:          hora,
		}
		if coincideFila(fila, filtros) {
			filas = append(filas, fila)
		}
	}
	sort.SliceStable(filas, func(i, j int) bool {
		if filas[i].Turno != filas[j].Turno {
			return filas[i].Turno < filas[j].Turno
		}
		if filas[i].Grado != filas[j].Grado {
			return filas[i].Grado < filas[j].Grado
		}
		if filas[i].Seccion != filas[j].Seccion {
			return filas[i].Seccion < filas[j].Seccion
		}
		return filas[i].Alumno < filas[j].Alumno
	})
	return filas, nil
}

// Panel calcula los totales del día.
func (s *Servicio) Panel(ctx context.Context, filtros Filtros) (PanelDelDia, error) {
	filtros.Estado = ""
	filas, err := s.Diario(ctx, filtros)
	if err != nil {
		return PanelDelDia{}, err
	}
	panel := PanelDelDia{Fecha: filtros.Desde, Total: len(filas)}
	for _, fila := range filas {
		switch fila.Estado {
		case EstadoPresente:
			panel.Presentes++
		case EstadoTardanza, EstadoTardanzaJustificada:
			panel.Tardanzas++
			panel.Presentes++
		case EstadoFalta:
			panel.Ausentes++
		case EstadoPendiente:
			panel.Pendientes++
		}
	}
	return panel, nil
}

// Resumen calcula el acumulado por alumno y sección en el periodo.
func (s *Servicio) Resumen(ctx context.Context, filtros Filtros) ([]ResumenAlumno, error) {
	desde, hasta := filtros.Desde, filtros.Hasta
	alumnos, err := s.dependencias.Repo.AlumnosConHistorial(ctx, desde, hasta)
	if err != nil {
		return nil, err
	}
	marcaciones, err := s.dependencias.Repo.MarcacionesDelRango(ctx, desde, hasta)
	if err != nil {
		return nil, err
	}
	justificaciones, err := s.dependencias.Repo.JustificacionesAprobadas(ctx, desde, hasta)
	if err != nil {
		return nil, err
	}
	excepciones, err := s.dependencias.Repo.Excepciones(ctx, desde, hasta)
	if err != nil {
		return nil, err
	}
	permitidas, err := s.seccionesPermitidas(ctx, filtros)
	if err != nil {
		return nil, err
	}

	porDia := make(map[string]generado.MarcacionesDelRangoRow, len(marcaciones))
	for _, marcacion := range marcaciones {
		porDia[claveDia(marcacion.AlumnoID, marcacion.Fecha.Time)] = marcacion
	}
	justificadas := make(map[string]string, len(justificaciones))
	for _, justificacion := range justificaciones {
		justificadas[claveDia(justificacion.AlumnoID, justificacion.FechaEvento.Time)] = justificacion.Tipo
	}
	noLaborables, globales := mapaExcepciones(excepciones)
	hoy := fechas.Fecha(s.dependencias.Reloj())

	type acumulado struct {
		ResumenAlumno
		racha int
	}
	acumulados := make(map[string]*acumulado)
	for _, alumno := range alumnos {
		if !permitida(permitidas, alumno.SeccionID) {
			continue
		}
		clave := claveSeccion(alumno.ID, alumno.SeccionID)
		registro, existe := acumulados[clave]
		if !existe {
			registro = &acumulado{ResumenAlumno: ResumenAlumno{
				AlumnoID:      alumno.ID,
				Identificador: alumno.Identificador,
				Alumno:        alumno.Apellidos + ", " + alumno.Nombres,
				Grado:         alumno.Grado,
				Seccion:       alumno.Seccion,
				Turno:         alumno.Turno,
			}}
			acumulados[clave] = registro
		}
		inicio := maxFecha(desde, alumno.FechaInicio.Time)
		fin := minFecha(hasta, alumno.FechaFin.Time)
		if fin.After(hoy) {
			fin = hoy
		}
		for fecha := inicio; !fecha.After(fin); fecha = fecha.AddDate(0, 0, 1) {
			if fecha.Weekday() == time.Saturday || fecha.Weekday() == time.Sunday {
				continue
			}
			if esNoLaborable(noLaborables, globales, fecha, alumno.GradoID) {
				continue
			}
			registro.DiasLaborables++
			marcacion, tiene := porDia[claveDia(alumno.ID, fecha)]
			tipoJustificacion := justificadas[claveDia(alumno.ID, fecha)]
			if tiene {
				switch marcacion.Estado {
				case asistencia.MarcacionPuntual:
					registro.Asistencias++
					registro.racha = 0
				case asistencia.MarcacionTardanza, asistencia.MarcacionTardanzaGrave:
					if marcacion.Justificada || tipoJustificacion == "tardanza" {
						registro.TardanzasJustificadas++
					} else {
						registro.Asistencias++
						registro.Tardanzas++
					}
					registro.racha = 0
				}
				continue
			}
			if tipoJustificacion == "falta" {
				registro.FaltasJustificadas++
				continue
			}
			registro.Faltas++
			registro.racha++
			if registro.racha > registro.FaltasSeguidas {
				registro.FaltasSeguidas = registro.racha
			}
		}
	}

	resumen := make([]ResumenAlumno, 0, len(acumulados))
	for _, registro := range acumulados {
		denominador := registro.DiasLaborables - registro.FaltasJustificadas - registro.TardanzasJustificadas
		if denominador > 0 {
			registro.Porcentaje = float64(registro.Asistencias) / float64(denominador) * 100
		}
		if coincideResumen(registro.ResumenAlumno, filtros) {
			resumen = append(resumen, registro.ResumenAlumno)
		}
	}
	sort.SliceStable(resumen, func(i, j int) bool {
		if resumen[i].Grado != resumen[j].Grado {
			return resumen[i].Grado < resumen[j].Grado
		}
		if resumen[i].Seccion != resumen[j].Seccion {
			return resumen[i].Seccion < resumen[j].Seccion
		}
		return resumen[i].Alumno < resumen[j].Alumno
	})
	return resumen, nil
}

// Reiteracion lista los alumnos que superan los umbrales del periodo.
func (s *Servicio) Reiteracion(ctx context.Context, filtros Filtros) ([]Reiteracion, error) {
	resumen, err := s.Resumen(ctx, filtros)
	if err != nil {
		return nil, err
	}
	umbralTardanzas, existeTardanzas, err := s.dependencias.Academico.UmbralVigente(ctx, "tardanzas", filtros.Hasta)
	if err != nil {
		return nil, err
	}
	if !existeTardanzas {
		umbralTardanzas = 0
	}
	umbralFaltas, existeFaltas, err := s.dependencias.Academico.UmbralVigente(ctx, "faltas_seguidas", filtros.Hasta)
	if err != nil {
		return nil, err
	}
	if !existeFaltas {
		umbralFaltas = 0
	}

	resultado := make([]Reiteracion, 0)
	for _, fila := range resumen {
		porTardanzas := umbralTardanzas > 0 && fila.Tardanzas >= umbralTardanzas
		porFaltas := umbralFaltas > 0 && fila.FaltasSeguidas >= umbralFaltas
		if !porTardanzas && !porFaltas {
			continue
		}
		resultado = append(resultado, Reiteracion{
			Alumno:          fila.Alumno,
			Seccion:         fila.Grado + " " + fila.Seccion,
			Tardanzas:       fila.Tardanzas,
			FaltasSeguidas:  fila.FaltasSeguidas,
			UmbralTardanzas: umbralTardanzas,
			UmbralFaltas:    umbralFaltas,
		})
	}
	return resultado, nil
}

func (s *Servicio) seccionesPermitidas(ctx context.Context, filtros Filtros) (map[int64]bool, error) {
	if filtros.Rol != "docente" {
		return nil, nil
	}
	ids, err := s.dependencias.Repo.SeccionesDelDocente(ctx, filtros.UsuarioID, filtros.Desde, filtros.Hasta)
	if err != nil {
		return nil, err
	}
	permitidas := make(map[int64]bool, len(ids))
	for _, id := range ids {
		permitidas[id] = true
	}
	return permitidas, nil
}

func (s *Servicio) estaCerrado(ctx context.Context, fecha, hoy time.Time, alumnos []generado.AlumnosEnFechaRow) (bool, error) {
	if fecha.Before(hoy) {
		return true, nil
	}
	if fecha.After(hoy) {
		return false, nil
	}
	ahora, err := fechas.ParsearHora(fechas.Hora(s.dependencias.Reloj()))
	if err != nil {
		return false, err
	}
	cache := make(map[string]academico.HorarioTurno)
	for _, alumno := range alumnos {
		horario, existe := cache[alumno.Turno]
		if !existe {
			var err error
			horario, err = s.dependencias.Academico.HorarioVigente(ctx, alumno.Turno, fecha)
			if err != nil {
				continue
			}
			cache[alumno.Turno] = horario
		}
		if fechas.CompararHoras(ahora, horario.HoraCierre) <= 0 {
			return false, nil
		}
	}
	return true, nil
}

func coincideFila(fila FilaDiaria, filtros Filtros) bool {
	if filtros.Turno != "" && fila.Turno != filtros.Turno {
		return false
	}
	if filtros.Grado != "" && fila.Grado != filtros.Grado {
		return false
	}
	if filtros.Seccion != "" && fila.Seccion != filtros.Seccion {
		return false
	}
	if filtros.Estado != "" && fila.Estado != filtros.Estado {
		return false
	}
	if filtros.Alumno != "" {
		texto := strings.ToLower(filtros.Alumno)
		if !strings.Contains(strings.ToLower(fila.Alumno), texto) &&
			!strings.Contains(strings.ToLower(fila.Identificador), texto) {
			return false
		}
	}
	return true
}

func coincideResumen(fila ResumenAlumno, filtros Filtros) bool {
	if filtros.Turno != "" && fila.Turno != filtros.Turno {
		return false
	}
	if filtros.Grado != "" && fila.Grado != filtros.Grado {
		return false
	}
	if filtros.Seccion != "" && fila.Seccion != filtros.Seccion {
		return false
	}
	if filtros.Alumno != "" {
		texto := strings.ToLower(filtros.Alumno)
		if !strings.Contains(strings.ToLower(fila.Alumno), texto) &&
			!strings.Contains(strings.ToLower(fila.Identificador), texto) {
			return false
		}
	}
	return true
}

func mapaExcepciones(filas []generado.ExcepcionesDelRangoRow) (map[string]bool, map[string]bool) {
	porDia := make(map[string]bool)
	porGrado := make(map[string]bool)
	for _, fila := range filas {
		for fecha := fila.FechaInicio.Time; !fecha.After(fila.FechaFin.Time); fecha = fecha.AddDate(0, 0, 1) {
			clave := fechas.FormatearISO(fecha)
			if fila.GradoID.Valid {
				porGrado[clave+"|"+strconv.FormatInt(fila.GradoID.Int64, 10)] = true
			} else {
				porDia[clave] = true
			}
		}
	}
	return porDia, porGrado
}

func esNoLaborable(globales, porGrado map[string]bool, fecha time.Time, gradoID int64) bool {
	clave := fechas.FormatearISO(fecha)
	return globales[clave] || porGrado[clave+"|"+strconv.FormatInt(gradoID, 10)]
}

func claveDia(alumnoID int64, fecha time.Time) string {
	return strconv.FormatInt(alumnoID, 10) + "|" + fechas.FormatearISO(fecha)
}

func claveSeccion(alumnoID, seccionID int64) string {
	return strconv.FormatInt(alumnoID, 10) + "|" + strconv.FormatInt(seccionID, 10)
}

func maxFecha(primera, segunda time.Time) time.Time {
	if primera.After(segunda) {
		return primera
	}
	return segunda
}

func minFecha(primera, segunda time.Time) time.Time {
	if primera.Before(segunda) {
		return primera
	}
	return segunda
}

func permitida(permitidas map[int64]bool, seccionID int64) bool {
	if permitidas == nil {
		return true
	}
	return permitidas[seccionID]
}
