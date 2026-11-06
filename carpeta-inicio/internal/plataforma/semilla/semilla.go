// Package semilla crea los datos mínimos para probar la aplicación en desarrollo.
// La configuración académica definitiva pertenece a la feature 002.
package semilla

import (
	"context"
	"errors"
	"fmt"
	"time"

	"asistencia/internal/autenticacion"
	"asistencia/internal/autenticacion/generado"

	academicogenerado "asistencia/internal/academico/generado"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// OpcionesAdministrador son los datos del administrador inicial.
type OpcionesAdministrador struct {
	Correo      string
	Contrasenia string
	Nombre      string
}

// CrearAdministrador registra el primer administrador si el correo no existe.
func CrearAdministrador(ctx context.Context, pool *pgxpool.Pool, opciones OpcionesAdministrador) (int64, bool, error) {
	repositorio := autenticacion.NuevoRepositorio(pool)
	existente, err := repositorio.BuscarPorCorreo(ctx, opciones.Correo)
	if err == nil {
		return existente.ID, false, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, false, fmt.Errorf("buscando el administrador: %w", err)
	}

	hash, err := autenticacion.HashearContrasenia(opciones.Contrasenia)
	if err != nil {
		return 0, false, err
	}
	id, err := repositorio.Crear(ctx, generado.CrearUsuarioParams{
		Correo:          opciones.Correo,
		ContraseniaHash: hash,
		NombreCompleto:  opciones.Nombre,
		Rol:             "administrador",
	})
	if err != nil {
		return 0, false, fmt.Errorf("creando el administrador: %w", err)
	}
	return id, true, nil
}

// OpcionesAcademicas son los datos del año escolar de demostración.
type OpcionesAcademicas struct {
	Nombre      string
	FechaInicio time.Time
	FechaFin    time.Time
	Grados      []string
	Secciones   []string
	Turno       string
}

// CrearAcademicoDemo crea el año escolar, los grados y las secciones si no existen.
func CrearAcademicoDemo(ctx context.Context, pool *pgxpool.Pool, opciones OpcionesAcademicas) error {
	consultas := academicogenerado.New(pool)

	anioID, err := consultas.BuscarAnioEscolarPorNombre(ctx, opciones.Nombre)
	if errors.Is(err, pgx.ErrNoRows) {
		anioID, err = consultas.CrearAnioEscolar(ctx, academicogenerado.CrearAnioEscolarParams{
			Nombre:      opciones.Nombre,
			FechaInicio: pgtype.Date{Time: opciones.FechaInicio, Valid: true},
			FechaFin:    pgtype.Date{Time: opciones.FechaFin, Valid: true},
		})
	}
	if err != nil {
		return fmt.Errorf("preparando el año escolar: %w", err)
	}

	for _, nombreGrado := range opciones.Grados {
		gradoID, err := consultas.BuscarGradoPorNombre(ctx, academicogenerado.BuscarGradoPorNombreParams{
			AnioEscolarID: anioID,
			Nombre:        nombreGrado,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			gradoID, err = consultas.CrearGrado(ctx, academicogenerado.CrearGradoParams{
				AnioEscolarID: anioID,
				Nombre:        nombreGrado,
			})
		}
		if err != nil {
			return fmt.Errorf("preparando el grado %s: %w", nombreGrado, err)
		}

		for _, nombreSeccion := range opciones.Secciones {
			_, err := consultas.BuscarSeccionPorNombre(ctx, academicogenerado.BuscarSeccionPorNombreParams{
				GradoID: gradoID,
				Nombre:  nombreSeccion,
			})
			if err == nil {
				continue
			}
			if !errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("buscando la sección %s %s: %w", nombreGrado, nombreSeccion, err)
			}
			if _, err := consultas.CrearSeccion(ctx, academicogenerado.CrearSeccionParams{
				GradoID: gradoID,
				Nombre:  nombreSeccion,
				Turno:   opciones.Turno,
			}); err != nil {
				return fmt.Errorf("creando la sección %s %s: %w", nombreGrado, nombreSeccion, err)
			}
		}
	}
	return nil
}
