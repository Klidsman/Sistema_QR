package asistencia

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"time"

	"asistencia/internal/alumnos"
	"asistencia/internal/plataforma/fechas"
)

// RegistroExportacion es una fila del reporte CSV del docente.
type RegistroExportacion struct {
	Fecha         time.Time
	Curso         string
	Seccion       string
	Bloque        int
	Identificador string
	Alumno        string
	Estado        string
	Observacion   string
	HoraQR        string
	Docente       string
}

// ExportarDocente devuelve los registros del docente en el rango de fechas.
func (s *Servicio) ExportarDocente(ctx context.Context, docenteID int64, desde, hasta time.Time) ([]RegistroExportacion, error) {
	if err := ValidarRangoExportacion(desde, hasta); err != nil {
		return nil, err
	}
	registros, err := s.repositorio.ExportarPorDocente(ctx, desde, hasta, docenteID)
	if err != nil {
		return nil, fmt.Errorf("exportando los registros del docente: %w", err)
	}
	return s.enriquecerExportacion(ctx, registros)
}

// ExportarSeccion devuelve los registros de una sección en el rango de fechas.
func (s *Servicio) ExportarSeccion(ctx context.Context, seccionID, cursoID int64, desde, hasta time.Time) ([]RegistroExportacion, error) {
	if err := ValidarRangoExportacion(desde, hasta); err != nil {
		return nil, err
	}
	if seccionID <= 0 {
		return nil, ErrClaseNoProgramada
	}
	registros, err := s.repositorio.ExportarPorSeccion(ctx, desde, hasta, seccionID, cursoID)
	if err != nil {
		return nil, fmt.Errorf("exportando los registros de la sección: %w", err)
	}
	return s.enriquecerExportacion(ctx, registros)
}

// GenerarCSV convierte los registros en un archivo CSV con BOM para Excel.
func GenerarCSV(registros []RegistroExportacion) ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString("\xEF\xBB\xBF")
	escritor := csv.NewWriter(&buffer)
	if err := escritor.Write([]string{
		"fecha", "curso", "seccion", "bloque", "identificador",
		"alumno", "estado", "observacion", "hora_qr", "docente",
	}); err != nil {
		return nil, err
	}
	for _, registro := range registros {
		if err := escritor.Write([]string{
			fechas.FormatearDia(registro.Fecha),
			registro.Curso,
			registro.Seccion,
			fmt.Sprintf("%d", registro.Bloque),
			registro.Identificador,
			registro.Alumno,
			registro.Estado,
			registro.Observacion,
			registro.HoraQR,
			registro.Docente,
		}); err != nil {
			return nil, err
		}
	}
	escritor.Flush()
	if err := escritor.Error(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (s *Servicio) enriquecerExportacion(ctx context.Context, registros []RegistroDocente) ([]RegistroExportacion, error) {
	if len(registros) == 0 {
		return nil, nil
	}
	ids := make([]int64, 0, len(registros))
	for _, registro := range registros {
		ids = append(ids, registro.AlumnoID)
	}
	alumnos, err := s.alumnos.AlumnosPorIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	alumnosPorID := make(map[int64]string, len(alumnos))
	for _, alumno := range alumnos {
		alumnosPorID[alumno.ID] = alumno.NombreCompleto()
	}
	desde := registros[0].Fecha
	hasta := registros[0].Fecha
	for _, registro := range registros {
		if fechas.CompararDia(registro.Fecha, desde) < 0 {
			desde = registro.Fecha
		}
		if fechas.CompararDia(registro.Fecha, hasta) > 0 {
			hasta = registro.Fecha
		}
	}
	marcaciones, err := s.repositorio.ListarMarcacionesPorRangoYAlumnos(ctx, desde, hasta, ids)
	if err != nil {
		return nil, fmt.Errorf("listando las marcaciones: %w", err)
	}
	marcacionPorAlumno := make(map[string]*Marcacion, len(marcaciones))
	for i := range marcaciones {
		marcacionPorAlumno[claveMarcacion(marcaciones[i].AlumnoID, marcaciones[i].Fecha)] = &marcaciones[i]
	}

	cursos := make(map[int64]string)
	secciones := make(map[int64]string)
	docentes := make(map[int64]string)
	resultado := make([]RegistroExportacion, 0, len(registros))
	for _, registro := range registros {
		curso, err := s.nombreCursoCache(ctx, cursos, registro.CursoID)
		if err != nil {
			return nil, err
		}
		seccion, err := s.etiquetaSeccionCache(ctx, secciones, registro.SeccionID)
		if err != nil {
			return nil, err
		}
		docente, err := s.nombreDocenteCache(ctx, docentes, registro.DocenteID)
		if err != nil {
			return nil, err
		}
		horaQR := ""
		if marcacion := marcacionPorAlumno[claveMarcacion(registro.AlumnoID, registro.Fecha)]; marcacion != nil {
			horaQR = fechas.Hora(marcacion.Hora)
		}
		resultado = append(resultado, RegistroExportacion{
			Fecha:         registro.Fecha,
			Curso:         curso,
			Seccion:       seccion,
			Bloque:        registro.BloqueNumero,
			Identificador: identificadorDe(alumnos, registro.AlumnoID),
			Alumno:        alumnosPorID[registro.AlumnoID],
			Estado:        registro.Estado,
			Observacion:   registro.Observacion,
			HoraQR:        horaQR,
			Docente:       docente,
		})
	}
	return resultado, nil
}

func identificadorDe(lista []alumnos.AlumnoEnClase, alumnoID int64) string {
	for _, alumno := range lista {
		if alumno.ID == alumnoID {
			return alumno.Identificador
		}
	}
	return ""
}

func claveMarcacion(alumnoID int64, fecha time.Time) string {
	return fmt.Sprintf("%d|%s", alumnoID, fechas.FormatearISO(fecha))
}

func (s *Servicio) nombreCursoCache(ctx context.Context, cache map[int64]string, cursoID int64) (string, error) {
	if nombre, existe := cache[cursoID]; existe {
		return nombre, nil
	}
	nombre, err := s.academico.NombreCurso(ctx, cursoID)
	if err != nil {
		return "", err
	}
	cache[cursoID] = nombre
	return nombre, nil
}

func (s *Servicio) etiquetaSeccionCache(ctx context.Context, cache map[int64]string, seccionID int64) (string, error) {
	if etiqueta, existe := cache[seccionID]; existe {
		return etiqueta, nil
	}
	etiqueta, err := s.academico.EtiquetaSeccion(ctx, seccionID)
	if err != nil {
		return "", err
	}
	cache[seccionID] = etiqueta
	return etiqueta, nil
}

func (s *Servicio) nombreDocenteCache(ctx context.Context, cache map[int64]string, docenteID int64) (string, error) {
	if nombre, existe := cache[docenteID]; existe {
		return nombre, nil
	}
	nombre, err := s.docentes.NombreDocente(ctx, docenteID)
	if err != nil {
		return "", err
	}
	cache[docenteID] = nombre
	return nombre, nil
}
