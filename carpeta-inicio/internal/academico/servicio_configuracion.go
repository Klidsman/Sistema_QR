package academico

import (
	"context"
	"errors"
	"time"

	"asistencia/internal/plataforma/fechas"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Reloj devuelve la fecha y hora actuales. Se inyecta en los servicios y en las pruebas.
type Reloj func() time.Time

// RelojDelSistema devuelve un reloj que usa la fecha de hoy en la zona del colegio.
func RelojDelSistema() Reloj {
	return fechas.Hoy
}

// ResumenConfiguracion es el estado que muestra el panel de configuración.
type ResumenConfiguracion struct {
	Anio      *AnioEscolar
	Grados    int
	Secciones int
}

// ServicioConfiguracion aplica las reglas de negocio de la configuración académica.
type ServicioConfiguracion struct {
	repositorio *RepositorioConfiguracion
	reloj       Reloj
}

// NuevoServicioConfiguracion crea el servicio sobre el pool indicado.
func NuevoServicioConfiguracion(pool *pgxpool.Pool) *ServicioConfiguracion {
	return &ServicioConfiguracion{
		repositorio: NuevoRepositorioConfiguracion(pool),
		reloj:       RelojDelSistema(),
	}
}

// Resumen devuelve el año activo y sus cantidades, o un resumen vacío si no hay año activo.
func (s *ServicioConfiguracion) Resumen(ctx context.Context) (ResumenConfiguracion, error) {
	anio, err := s.repositorio.BuscarAnioActivo(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return ResumenConfiguracion{}, nil
	}
	if err != nil {
		return ResumenConfiguracion{}, err
	}

	grados, err := s.repositorio.ListarGradosPorAnio(ctx, anio.ID)
	if err != nil {
		return ResumenConfiguracion{}, err
	}
	secciones, err := s.repositorio.ListarSeccionesPorAnio(ctx, anio.ID)
	if err != nil {
		return ResumenConfiguracion{}, err
	}
	return ResumenConfiguracion{
		Anio:      &anio,
		Grados:    len(grados),
		Secciones: len(secciones),
	}, nil
}
