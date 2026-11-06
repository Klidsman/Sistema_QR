package asistencia

import "time"

// DependenciasServicio reúne los componentes que necesita el servicio de asistencia.
type DependenciasServicio struct {
	Repositorio *Repositorio
	Academico   ResolutorClase
	Alumnos     ListadorAlumnos
	Docentes    BuscadorDocente
	Reloj       Reloj
}

// Servicio aplica las reglas de negocio de asignaciones y registros del docente.
type Servicio struct {
	repositorio *Repositorio
	academico   ResolutorClase
	alumnos     ListadorAlumnos
	docentes    BuscadorDocente
	reloj       Reloj
}

// NuevoServicio crea el servicio de asistencia.
func NuevoServicio(dependencias DependenciasServicio) *Servicio {
	reloj := dependencias.Reloj
	if reloj == nil {
		reloj = time.Now
	}
	return &Servicio{
		repositorio: dependencias.Repositorio,
		academico:   dependencias.Academico,
		alumnos:     dependencias.Alumnos,
		docentes:    dependencias.Docentes,
		reloj:       reloj,
	}
}

func (s *Servicio) hoy() time.Time {
	return s.reloj()
}
