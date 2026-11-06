package academico

import (
	"context"
	"errors"
	"fmt"
	"time"

	"asistencia/internal/academico/generado"
	"asistencia/internal/plataforma/auditoria"
	"asistencia/internal/plataforma/fechas"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// codigoViolacionUnica es el código de PostgreSQL para un índice único duplicado.
const codigoViolacionUnica = "23505"

// Errores de negocio de la configuración académica.
var (
	ErrAnioActivoExistente = errors.New("ya existe un año escolar activo")
	ErrAnioNombreExistente = errors.New("ya existe un año escolar con ese nombre")
	ErrAnioNoEncontrado    = errors.New("el año escolar no existe")
	ErrAnioCerrado         = errors.New("el año escolar está cerrado y no admite cambios")
	ErrGradoNoEncontrado   = errors.New("el grado no existe")
	ErrGradoExistente      = errors.New("ya existe un grado con ese nombre en el año")
	ErrSeccionExistente    = errors.New("ya existe una sección con ese nombre en el grado")
	ErrHorarioNoEncontrado = errors.New("no hay un horario vigente para ese turno en la fecha")
	ErrBloqueNoEncontrado  = errors.New("el bloque de clase no existe")
)

// Reloj devuelve la fecha y hora actuales. Se inyecta en los servicios y en las pruebas.
type Reloj func() time.Time

// RelojDelSistema devuelve un reloj que usa la fecha de hoy en la zona del colegio.
func RelojDelSistema() Reloj {
	return fechas.Hoy
}

// DatosAnio son los datos para crear un año escolar.
type DatosAnio struct {
	Nombre      string
	FechaInicio time.Time
	FechaFin    time.Time
}

// DatosGrado son los datos para crear un grado.
type DatosGrado struct {
	AnioEscolarID int64
	Nombre        string
}

// DatosSeccion son los datos para crear una sección con su turno inicial.
type DatosSeccion struct {
	GradoID      int64
	Nombre       string
	Turno        string
	VigenteDesde time.Time
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
	anio, err := s.AnioActivo(ctx)
	if err != nil {
		return ResumenConfiguracion{}, err
	}
	if anio == nil {
		return ResumenConfiguracion{}, nil
	}
	grados, err := s.repositorio.ListarGradosPorAnio(ctx, anio.ID)
	if err != nil {
		return ResumenConfiguracion{}, fmt.Errorf("listando los grados: %w", err)
	}
	secciones, err := s.repositorio.ListarSeccionesPorAnio(ctx, anio.ID)
	if err != nil {
		return ResumenConfiguracion{}, fmt.Errorf("listando las secciones: %w", err)
	}
	return ResumenConfiguracion{
		Anio:      anio,
		Grados:    len(grados),
		Secciones: len(secciones),
	}, nil
}

// AnioActivo devuelve el año escolar activo, o nil si no existe.
func (s *ServicioConfiguracion) AnioActivo(ctx context.Context) (*AnioEscolar, error) {
	anio, err := s.repositorio.BuscarAnioActivo(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("buscando el año escolar activo: %w", err)
	}
	return &anio, nil
}

// ListarAnios devuelve todos los años escolares.
func (s *ServicioConfiguracion) ListarAnios(ctx context.Context) ([]AnioEscolar, error) {
	anios, err := s.repositorio.ListarAnios(ctx)
	if err != nil {
		return nil, fmt.Errorf("listando los años escolares: %w", err)
	}
	return anios, nil
}

// ListarGrados devuelve los grados del año indicado.
func (s *ServicioConfiguracion) ListarGrados(ctx context.Context, anioID int64) ([]Grado, error) {
	grados, err := s.repositorio.ListarGradosPorAnio(ctx, anioID)
	if err != nil {
		return nil, fmt.Errorf("listando los grados: %w", err)
	}
	return grados, nil
}

// ListarSecciones devuelve las secciones del año indicado.
func (s *ServicioConfiguracion) ListarSecciones(ctx context.Context, anioID int64) ([]Seccion, error) {
	secciones, err := s.repositorio.ListarSeccionesPorAnio(ctx, anioID)
	if err != nil {
		return nil, fmt.Errorf("listando las secciones: %w", err)
	}
	return secciones, nil
}

// ListarTurnosSeccion devuelve la historia de turnos de una sección.
func (s *ServicioConfiguracion) ListarTurnosSeccion(ctx context.Context, seccionID int64) ([]TurnoSeccion, error) {
	turnos, err := s.repositorio.ListarTurnosSeccion(ctx, seccionID)
	if err != nil {
		return nil, fmt.Errorf("listando los turnos de la sección: %w", err)
	}
	return turnos, nil
}

// CrearAnio registra un año escolar activo y lo audita.
func (s *ServicioConfiguracion) CrearAnio(ctx context.Context, usuarioID int64, datos DatosAnio) (int64, error) {
	datos.Nombre = NormalizarNombre(datos.Nombre)
	if err := ValidarNombre(datos.Nombre); err != nil {
		return 0, err
	}
	if err := ValidarFechasAnio(datos.FechaInicio, datos.FechaFin); err != nil {
		return 0, err
	}

	var id int64
	err := s.repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		if _, err := consultas.BuscarAnioActivo(ctx); err == nil {
			return ErrAnioActivoExistente
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("comprobando el año activo: %w", err)
		}

		creado, err := consultas.CrearAnioEscolar(ctx, generado.CrearAnioEscolarParams{
			Nombre:      datos.Nombre,
			FechaInicio: parametroFecha(datos.FechaInicio),
			FechaFin:    parametroFecha(datos.FechaFin),
		})
		if err != nil {
			if esViolacionUnica(err) {
				return ErrAnioNombreExistente
			}
			return fmt.Errorf("creando el año escolar: %w", err)
		}
		if err := auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID: usuarioID,
			Entidad:   EntidadAnioEscolar,
			EntidadID: creado,
			Accion:    AccionCrear,
			ValorNuevo: map[string]any{
				"nombre":       datos.Nombre,
				"fecha_inicio": fechas.FormatearISO(datos.FechaInicio),
				"fecha_fin":    fechas.FormatearISO(datos.FechaFin),
				"estado":       EstadoAnioActivo,
			},
		}); err != nil {
			return err
		}
		id = creado
		return nil
	})
	return id, err
}

// CerrarAnio cierra el año escolar y lo deja de solo lectura.
func (s *ServicioConfiguracion) CerrarAnio(ctx context.Context, usuarioID, anioID int64) error {
	return s.repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		anio, err := consultas.BuscarAnioEscolarPorID(ctx, anioID)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrAnioNoEncontrado
		}
		if err != nil {
			return fmt.Errorf("buscando el año escolar: %w", err)
		}
		if anio.Estado == EstadoAnioCerrado {
			return ErrAnioCerrado
		}
		filas, err := consultas.CerrarAnioEscolar(ctx, anioID)
		if err != nil {
			return fmt.Errorf("cerrando el año escolar: %w", err)
		}
		if filas == 0 {
			return ErrAnioCerrado
		}
		return auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID:     usuarioID,
			Entidad:       EntidadAnioEscolar,
			EntidadID:     anioID,
			Accion:        AccionCerrar,
			ValorAnterior: map[string]any{"estado": anio.Estado},
			ValorNuevo:    map[string]any{"estado": EstadoAnioCerrado},
		})
	})
}

// CrearGrado registra un grado en el año indicado.
func (s *ServicioConfiguracion) CrearGrado(ctx context.Context, usuarioID int64, datos DatosGrado) (int64, error) {
	datos.Nombre = NormalizarNombre(datos.Nombre)
	if err := ValidarNombre(datos.Nombre); err != nil {
		return 0, err
	}

	var id int64
	err := s.repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		anio, err := consultas.BuscarAnioEscolarPorID(ctx, datos.AnioEscolarID)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrAnioNoEncontrado
		}
		if err != nil {
			return fmt.Errorf("buscando el año escolar: %w", err)
		}
		if anio.Estado == EstadoAnioCerrado {
			return ErrAnioCerrado
		}

		if _, err := consultas.BuscarGradoPorNombre(ctx, generado.BuscarGradoPorNombreParams{
			AnioEscolarID: datos.AnioEscolarID,
			Nombre:        datos.Nombre,
		}); err == nil {
			return ErrGradoExistente
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("comprobando el grado: %w", err)
		}

		creado, err := consultas.CrearGrado(ctx, generado.CrearGradoParams{
			AnioEscolarID: datos.AnioEscolarID,
			Nombre:        datos.Nombre,
		})
		if err != nil {
			if esViolacionUnica(err) {
				return ErrGradoExistente
			}
			return fmt.Errorf("creando el grado: %w", err)
		}
		if err := auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID:  usuarioID,
			Entidad:    EntidadGrado,
			EntidadID:  creado,
			Accion:     AccionCrear,
			ValorNuevo: map[string]any{"anio_escolar_id": datos.AnioEscolarID, "nombre": datos.Nombre},
		}); err != nil {
			return err
		}
		id = creado
		return nil
	})
	return id, err
}

// CrearSeccion registra una sección con su turno inicial.
func (s *ServicioConfiguracion) CrearSeccion(ctx context.Context, usuarioID int64, datos DatosSeccion) (int64, error) {
	datos.Nombre = NormalizarNombre(datos.Nombre)
	if err := ValidarNombreSeccion(datos.Nombre); err != nil {
		return 0, err
	}
	if err := ValidarTurno(datos.Turno); err != nil {
		return 0, err
	}
	if err := ValidarVigencia(datos.VigenteDesde, s.reloj()); err != nil {
		return 0, err
	}

	var id int64
	err := s.repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		grado, err := consultas.BuscarGradoPorID(ctx, datos.GradoID)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrGradoNoEncontrado
		}
		if err != nil {
			return fmt.Errorf("buscando el grado: %w", err)
		}
		if grado.AnioEstado == EstadoAnioCerrado {
			return ErrAnioCerrado
		}

		if _, err := consultas.BuscarSeccionPorNombre(ctx, generado.BuscarSeccionPorNombreParams{
			GradoID: datos.GradoID,
			Nombre:  datos.Nombre,
		}); err == nil {
			return ErrSeccionExistente
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("comprobando la sección: %w", err)
		}

		creado, err := consultas.CrearSeccion(ctx, generado.CrearSeccionParams{
			GradoID: datos.GradoID,
			Nombre:  datos.Nombre,
		})
		if err != nil {
			if esViolacionUnica(err) {
				return ErrSeccionExistente
			}
			return fmt.Errorf("creando la sección: %w", err)
		}
		if _, err := consultas.CrearTurnoSeccion(ctx, generado.CrearTurnoSeccionParams{
			SeccionID:    creado,
			Turno:        datos.Turno,
			VigenteDesde: parametroFecha(datos.VigenteDesde),
		}); err != nil {
			return fmt.Errorf("asignando el turno de la sección: %w", err)
		}
		if err := auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID: usuarioID,
			Entidad:   EntidadSeccion,
			EntidadID: creado,
			Accion:    AccionCrear,
			ValorNuevo: map[string]any{
				"grado_id":      datos.GradoID,
				"nombre":        datos.Nombre,
				"turno":         datos.Turno,
				"vigente_desde": fechas.FormatearISO(datos.VigenteDesde),
			},
		}); err != nil {
			return err
		}
		id = creado
		return nil
	})
	return id, err
}

// CambiarTurno registra un turno nuevo para una sección desde una fecha.
func (s *ServicioConfiguracion) CambiarTurno(ctx context.Context, usuarioID, seccionID int64, turno string, vigenteDesde time.Time) error {
	if err := ValidarTurno(turno); err != nil {
		return err
	}
	if err := ValidarVigencia(vigenteDesde, s.reloj()); err != nil {
		return err
	}

	return s.repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		gradoID, err := consultas.GradoDeSeccion(ctx, seccionID)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrGradoNoEncontrado
		}
		if err != nil {
			return fmt.Errorf("buscando el grado de la sección: %w", err)
		}
		grado, err := consultas.BuscarGradoPorID(ctx, gradoID)
		if err != nil {
			return fmt.Errorf("buscando el grado: %w", err)
		}
		if grado.AnioEstado == EstadoAnioCerrado {
			return ErrAnioCerrado
		}

		anterior := ""
		turnoAnterior, err := consultas.TurnoVigenteSeccion(ctx, generado.TurnoVigenteSeccionParams{
			SeccionID:    seccionID,
			VigenteDesde: parametroFecha(s.reloj()),
		})
		if err == nil {
			anterior = turnoAnterior
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("buscando el turno anterior: %w", err)
		}

		if _, err := consultas.CrearTurnoSeccion(ctx, generado.CrearTurnoSeccionParams{
			SeccionID:    seccionID,
			Turno:        turno,
			VigenteDesde: parametroFecha(vigenteDesde),
		}); err != nil {
			return fmt.Errorf("cambiando el turno de la sección: %w", err)
		}
		return auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID:     usuarioID,
			Entidad:       EntidadTurno,
			EntidadID:     seccionID,
			Accion:        AccionCambiarTurno,
			ValorAnterior: map[string]any{"turno": anterior},
			ValorNuevo: map[string]any{
				"turno":         turno,
				"vigente_desde": fechas.FormatearISO(vigenteDesde),
			},
		})
	})
}

// ConfigurarHorario registra un horario de turno con su vigencia y lo audita.
func (s *ServicioConfiguracion) ConfigurarHorario(ctx context.Context, usuarioID int64, horario HorarioTurno) error {
	if err := ValidarHorarioConVigencia(horario, s.reloj()); err != nil {
		return err
	}

	return s.repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		if _, err := db.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtext('horarios_turno'))"); err != nil {
			return fmt.Errorf("bloqueando la configuración de horarios: %w", err)
		}

		anio, err := consultas.BuscarAnioActivo(ctx)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrAnioNoEncontrado
		}
		if err != nil {
			return fmt.Errorf("buscando el año escolar activo: %w", err)
		}
		if anio.Estado == EstadoAnioCerrado {
			return ErrAnioCerrado
		}

		anterior, err := consultas.HorarioTurnoVigente(ctx, generado.HorarioTurnoVigenteParams{
			Turno:        horario.Turno,
			VigenteDesde: parametroFecha(horario.VigenteDesde),
		})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("buscando el horario anterior: %w", err)
		}
		valorAnterior := any(nil)
		if err == nil {
			anteriorConvertido, err := aHorarioTurno(anterior)
			if err != nil {
				return err
			}
			valorAnterior = mapaHorario(anteriorConvertido)
		}

		otroTurno := TurnoTarde
		if horario.Turno == TurnoTarde {
			otroTurno = TurnoManana
		}
		otroHorario, err := consultas.HorarioTurnoVigente(ctx, generado.HorarioTurnoVigenteParams{
			Turno:        otroTurno,
			VigenteDesde: parametroFecha(horario.VigenteDesde),
		})
		if err == nil {
			otroConvertido, err := aHorarioTurno(otroHorario)
			if err != nil {
				return err
			}
			if err := ValidarSuperposicionHorarios(horario, otroConvertido); err != nil {
				return err
			}
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("buscando el horario del otro turno: %w", err)
		}

		if _, err := consultas.GuardarHorarioTurno(ctx, generado.GuardarHorarioTurnoParams{
			Turno:              horario.Turno,
			HoraIngreso:        parametroHora(horario.HoraIngreso),
			ToleranciaMinutos:  int32(horario.ToleranciaMinutos),
			HoraLimiteTardanza: parametroHora(horario.HoraLimiteTardanza),
			HoraCierre:         parametroHora(horario.HoraCierre),
			VigenteDesde:       parametroFecha(horario.VigenteDesde),
		}); err != nil {
			return fmt.Errorf("guardando el horario del turno: %w", err)
		}

		return auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID:     usuarioID,
			Entidad:       EntidadTurno,
			EntidadID:     anio.ID,
			Accion:        AccionConfigurarHorario,
			ValorAnterior: valorAnterior,
			ValorNuevo:    mapaHorario(horario),
		})
	})
}

// HorarioVigente devuelve el horario de un turno en la fecha indicada.
func (s *ServicioConfiguracion) HorarioVigente(ctx context.Context, turno string, fecha time.Time) (HorarioTurno, error) {
	if err := ValidarTurno(turno); err != nil {
		return HorarioTurno{}, err
	}
	horario, err := s.repositorio.HorarioTurnoVigente(ctx, turno, fecha)
	if errors.Is(err, pgx.ErrNoRows) {
		return HorarioTurno{}, ErrHorarioNoEncontrado
	}
	if err != nil {
		return HorarioTurno{}, fmt.Errorf("buscando el horario del turno: %w", err)
	}
	return horario, nil
}

// ListarHorarios devuelve la historia de horarios de un turno.
func (s *ServicioConfiguracion) ListarHorarios(ctx context.Context, turno string) ([]HorarioTurno, error) {
	if err := ValidarTurno(turno); err != nil {
		return nil, err
	}
	horarios, err := s.repositorio.ListarHorariosTurno(ctx, turno)
	if err != nil {
		return nil, fmt.Errorf("listando los horarios del turno: %w", err)
	}
	return horarios, nil
}

// CrearBloque registra un bloque de hora de clase y lo audita.
func (s *ServicioConfiguracion) CrearBloque(ctx context.Context, usuarioID int64, bloque BloqueClase) (int64, error) {
	if err := ValidarTurno(bloque.Turno); err != nil {
		return 0, err
	}
	if err := ValidarVigencia(bloque.VigenteDesde, s.reloj()); err != nil {
		return 0, err
	}

	var id int64
	err := s.repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		if _, err := db.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtext($1))", "bloques_"+bloque.Turno); err != nil {
			return fmt.Errorf("bloqueando la configuración de bloques: %w", err)
		}
		if err := verificarAnioAbierto(ctx, consultas); err != nil {
			return err
		}
		horario, otros, err := horarioYBloquesVigentes(ctx, consultas, bloque.Turno, bloque.VigenteDesde)
		if err != nil {
			return err
		}
		if err := ValidarBloque(bloque, horario, otros); err != nil {
			return err
		}
		creado, err := consultas.GuardarBloqueClase(ctx, generado.GuardarBloqueClaseParams{
			Turno:        bloque.Turno,
			Numero:       int32(bloque.Numero),
			HoraInicio:   parametroHora(bloque.HoraInicio),
			HoraFin:      parametroHora(bloque.HoraFin),
			VigenteDesde: parametroFecha(bloque.VigenteDesde),
		})
		if err != nil {
			return fmt.Errorf("guardando el bloque de clase: %w", err)
		}
		if err := auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID:  usuarioID,
			Entidad:    EntidadBloqueClase,
			EntidadID:  creado,
			Accion:     AccionCrearBloque,
			ValorNuevo: mapaBloque(bloque),
		}); err != nil {
			return err
		}
		id = creado
		return nil
	})
	return id, err
}

// AjustarHorasBloque registra una versión nueva de un bloque existente con otras horas.
func (s *ServicioConfiguracion) AjustarHorasBloque(ctx context.Context, usuarioID, bloqueID int64, inicio, fin, vigencia time.Time) error {
	if err := ValidarVigencia(vigencia, s.reloj()); err != nil {
		return err
	}

	anterior, err := s.repositorio.BuscarBloquePorID(ctx, bloqueID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrBloqueNoEncontrado
	}
	if err != nil {
		return fmt.Errorf("buscando el bloque de clase: %w", err)
	}

	nuevo := BloqueClase{
		Turno:        anterior.Turno,
		Numero:       anterior.Numero,
		HoraInicio:   inicio,
		HoraFin:      fin,
		VigenteDesde: vigencia,
	}

	return s.repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		if _, err := db.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtext($1))", "bloques_"+nuevo.Turno); err != nil {
			return fmt.Errorf("bloqueando la configuración de bloques: %w", err)
		}
		if err := verificarAnioAbierto(ctx, consultas); err != nil {
			return err
		}
		horario, otros, err := horarioYBloquesVigentes(ctx, consultas, nuevo.Turno, vigencia)
		if err != nil {
			return err
		}
		if err := ValidarBloque(nuevo, horario, otros); err != nil {
			return err
		}
		if _, err := consultas.GuardarBloqueClase(ctx, generado.GuardarBloqueClaseParams{
			Turno:        nuevo.Turno,
			Numero:       int32(nuevo.Numero),
			HoraInicio:   parametroHora(nuevo.HoraInicio),
			HoraFin:      parametroHora(nuevo.HoraFin),
			VigenteDesde: parametroFecha(nuevo.VigenteDesde),
		}); err != nil {
			return fmt.Errorf("ajustando las horas del bloque: %w", err)
		}
		return auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID:     usuarioID,
			Entidad:       EntidadBloqueClase,
			EntidadID:     bloqueID,
			Accion:        AccionAjustarBloque,
			ValorAnterior: mapaBloque(anterior),
			ValorNuevo:    mapaBloque(nuevo),
		})
	})
}

// BloquesVigentes devuelve los bloques de un turno vigentes en la fecha.
func (s *ServicioConfiguracion) BloquesVigentes(ctx context.Context, turno string, fecha time.Time) ([]BloqueClase, error) {
	if err := ValidarTurno(turno); err != nil {
		return nil, err
	}
	bloques, err := s.repositorio.BloquesVigentesTurno(ctx, turno, fecha)
	if err != nil {
		return nil, fmt.Errorf("listando los bloques del turno: %w", err)
	}
	return bloques, nil
}

func verificarAnioAbierto(ctx context.Context, consultas *generado.Queries) error {
	anio, err := consultas.BuscarAnioActivo(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrAnioNoEncontrado
	}
	if err != nil {
		return fmt.Errorf("buscando el año escolar activo: %w", err)
	}
	if anio.Estado == EstadoAnioCerrado {
		return ErrAnioCerrado
	}
	return nil
}

func horarioYBloquesVigentes(ctx context.Context, consultas *generado.Queries, turno string, fecha time.Time) (HorarioTurno, []BloqueClase, error) {
	filaHorario, err := consultas.HorarioTurnoVigente(ctx, generado.HorarioTurnoVigenteParams{
		Turno:        turno,
		VigenteDesde: parametroFecha(fecha),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return HorarioTurno{}, nil, ErrHorarioNoEncontrado
	}
	if err != nil {
		return HorarioTurno{}, nil, fmt.Errorf("buscando el horario del turno: %w", err)
	}
	horario, err := aHorarioTurno(filaHorario)
	if err != nil {
		return HorarioTurno{}, nil, err
	}

	filas, err := consultas.BloquesVigentesTurno(ctx, generado.BloquesVigentesTurnoParams{
		Turno:        turno,
		VigenteDesde: parametroFecha(fecha),
	})
	if err != nil {
		return HorarioTurno{}, nil, fmt.Errorf("listando los bloques vigentes: %w", err)
	}
	otros := make([]BloqueClase, 0, len(filas))
	for _, fila := range filas {
		bloque, err := aBloqueClase(fila)
		if err != nil {
			return HorarioTurno{}, nil, err
		}
		otros = append(otros, bloque)
	}
	return horario, otros, nil
}

func mapaBloque(bloque BloqueClase) map[string]any {
	return map[string]any{
		"turno":         bloque.Turno,
		"numero":        bloque.Numero,
		"hora_inicio":   fechas.FormatearHora(bloque.HoraInicio),
		"hora_fin":      fechas.FormatearHora(bloque.HoraFin),
		"vigente_desde": fechas.FormatearISO(bloque.VigenteDesde),
	}
}

func mapaHorario(horario HorarioTurno) map[string]any {
	return map[string]any{
		"turno":                horario.Turno,
		"hora_ingreso":         fechas.FormatearHora(horario.HoraIngreso),
		"tolerancia_minutos":   horario.ToleranciaMinutos,
		"hora_limite_tardanza": fechas.FormatearHora(horario.HoraLimiteTardanza),
		"hora_cierre":          fechas.FormatearHora(horario.HoraCierre),
		"vigente_desde":        fechas.FormatearISO(horario.VigenteDesde),
	}
}

func esViolacionUnica(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == codigoViolacionUnica
}

// parametroFecha convierte una fecha a su parámetro de base de datos.
func parametroFecha(fecha time.Time) pgtype.Date {
	return pgtype.Date{Time: fecha, Valid: true}
}
