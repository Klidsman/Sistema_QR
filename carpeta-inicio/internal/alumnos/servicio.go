package alumnos

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"asistencia/internal/alumnos/generado"
	"asistencia/internal/plataforma/auditoria"
	"asistencia/internal/plataforma/fechas"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// codigoViolacionUnica es el código de PostgreSQL para un índice único duplicado.
const codigoViolacionUnica = "23505"

// Servicio aplica las reglas de negocio de alumnos y credenciales QR.
type Servicio struct {
	repositorio *Repositorio
}

// NuevoServicio crea el servicio sobre el pool indicado.
func NuevoServicio(pool *pgxpool.Pool) *Servicio {
	return &Servicio{repositorio: NuevoRepositorio(pool)}
}

// Listar devuelve los alumnos que cumplen los filtros.
func (s *Servicio) Listar(ctx context.Context, filtros FiltrosListado) ([]AlumnoListado, error) {
	listado, err := s.repositorio.listar(ctx, filtros)
	if err != nil {
		return nil, fmt.Errorf("listando alumnos: %w", err)
	}
	return listado, nil
}

// ListarCarnets devuelve el lote imprimible y los avisos previos de la selección.
func (s *Servicio) ListarCarnets(ctx context.Context, seleccion SeleccionCarnets) (LoteCarnets, error) {
	if !seleccion.Valida() {
		return LoteCarnets{}, ErrSeleccionCarnetsInvalida
	}
	if seleccion.SeccionID > 0 {
		seleccion.GradoID = 0
	}
	alumnos, err := s.repositorio.listarCarnets(ctx, seleccion)
	if err != nil {
		return LoteCarnets{}, fmt.Errorf("listando alumnos para carnés: %w", err)
	}
	return PrepararLote(alumnos), nil
}

// Obtener devuelve la ficha de un alumno.
func (s *Servicio) Obtener(ctx context.Context, id int64) (Alumno, error) {
	alumno, err := s.repositorio.buscar(ctx, id)
	if err != nil {
		return Alumno{}, err
	}
	return alumno, nil
}

// Crear registra un alumno y devuelve su identificador.
func (s *Servicio) Crear(ctx context.Context, usuarioID int64, datos DatosAlumno) (int64, error) {
	datos = datosNormalizados(datos)
	if err := Validar(datos); err != nil {
		return 0, err
	}

	var alumnoID int64
	err := s.repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		if _, err := consultas.BuscarAlumnoPorIdentificador(ctx, datos.Identificador); err == nil {
			return ErrIdentificadorExistente
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("comprobando el identificador: %w", err)
		}

		id, err := consultas.CrearAlumno(ctx, generado.CrearAlumnoParams{
			Identificador:     datos.Identificador,
			Nombres:           datos.Nombres,
			Apellidos:         datos.Apellidos,
			NombresBusqueda:   Normalizar(datos.Nombres),
			ApellidosBusqueda: Normalizar(datos.Apellidos),
			SeccionID:         datos.SeccionID,
			ApoderadoNombre:   textoOpcional(datos.ApoderadoNombre),
			ApoderadoTelefono: textoOpcional(datos.ApoderadoTelefono),
		})
		if esViolacionUnica(err) {
			return ErrIdentificadorExistente
		}
		if err != nil {
			return fmt.Errorf("creando el alumno: %w", err)
		}
		alumnoID = id

		if _, err := consultas.CrearHistorialSeccion(ctx, generado.CrearHistorialSeccionParams{
			AlumnoID:    id,
			SeccionID:   datos.SeccionID,
			FechaInicio: fechaPG(fechas.Hoy()),
			FechaFin:    pgtype.Date{},
		}); err != nil {
			return fmt.Errorf("abriendo el historial de sección: %w", err)
		}

		return auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID:  usuarioID,
			Entidad:    "alumno",
			EntidadID:  id,
			Accion:     "crear",
			ValorNuevo: datos,
		})
	})
	if err != nil {
		return 0, err
	}
	return alumnoID, nil
}

// Actualizar modifica los datos editables de un alumno. El identificador no cambia.
func (s *Servicio) Actualizar(ctx context.Context, usuarioID, alumnoID int64, datos DatosAlumno) error {
	datos = datosNormalizados(datos)
	if err := Validar(datos); err != nil {
		return err
	}

	return s.repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		anterior, err := s.repositorio.buscar(ctx, alumnoID)
		if err != nil {
			return err
		}
		datos.Identificador = anterior.Identificador

		err = consultas.ActualizarAlumno(ctx, generado.ActualizarAlumnoParams{
			ID:                alumnoID,
			Nombres:           datos.Nombres,
			Apellidos:         datos.Apellidos,
			NombresBusqueda:   Normalizar(datos.Nombres),
			ApellidosBusqueda: Normalizar(datos.Apellidos),
			SeccionID:         datos.SeccionID,
			ApoderadoNombre:   textoOpcional(datos.ApoderadoNombre),
			ApoderadoTelefono: textoOpcional(datos.ApoderadoTelefono),
		})
		if err != nil {
			return fmt.Errorf("actualizando el alumno: %w", err)
		}

		return auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID:     usuarioID,
			Entidad:       "alumno",
			EntidadID:     alumnoID,
			Accion:        "editar",
			ValorAnterior: anterior,
			ValorNuevo:    datos,
		})
	})
}

// ActualizarFoto guarda la clave de la foto del alumno.
func (s *Servicio) ActualizarFoto(ctx context.Context, usuarioID, alumnoID int64, clave string) error {
	return s.repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		if _, err := s.repositorio.buscar(ctx, alumnoID); err != nil {
			return err
		}
		if err := consultas.ActualizarFotoAlumno(ctx, generado.ActualizarFotoAlumnoParams{
			ID:        alumnoID,
			FotoClave: textoOpcional(clave),
		}); err != nil {
			return fmt.Errorf("actualizando la foto: %w", err)
		}
		return auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID:  usuarioID,
			Entidad:    "alumno",
			EntidadID:  alumnoID,
			Accion:     "editar",
			ValorNuevo: map[string]string{"foto_clave": clave},
		})
	})
}

// Desactivar marca al alumno como desactivado y revoca su QR vigente.
func (s *Servicio) Desactivar(ctx context.Context, usuarioID, alumnoID int64) error {
	return s.repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		anterior, err := s.repositorio.buscar(ctx, alumnoID)
		if err != nil {
			return err
		}
		if anterior.Estado == EstadoDesactivado {
			return nil
		}

		if err := consultas.CambiarEstadoAlumno(ctx, generado.CambiarEstadoAlumnoParams{
			ID:     alumnoID,
			Estado: EstadoDesactivado,
		}); err != nil {
			return fmt.Errorf("desactivando al alumno: %w", err)
		}
		if err := revocarCredencialVigente(ctx, consultas, db, usuarioID, alumnoID, MotivoDesactivacion); err != nil {
			return err
		}

		return auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID:     usuarioID,
			Entidad:       "alumno",
			EntidadID:     alumnoID,
			Accion:        "desactivar",
			ValorAnterior: map[string]string{"estado": anterior.Estado},
			ValorNuevo:    map[string]string{"estado": EstadoDesactivado},
		})
	})
}

// HistorialSecciones devuelve los tramos de sección del alumno, del más reciente al más antiguo.
func (s *Servicio) HistorialSecciones(ctx context.Context, alumnoID int64) ([]HistorialSeccion, error) {
	filas, err := s.repositorio.consultas.ListarHistorialSecciones(ctx, alumnoID)
	if err != nil {
		return nil, fmt.Errorf("listando el historial de secciones: %w", err)
	}

	historial := make([]HistorialSeccion, 0, len(filas))
	for _, fila := range filas {
		tramo := HistorialSeccion{
			ID:          fila.ID,
			SeccionID:   fila.SeccionID,
			Grado:       fila.GradoNombre,
			Seccion:     fila.SeccionNombre,
			AnioEscolar: fila.AnioEscolarNombre,
			FechaInicio: fila.FechaInicio.Time,
		}
		if fila.FechaFin.Valid {
			fin := fila.FechaFin.Time
			tramo.FechaFin = &fin
		}
		historial = append(historial, tramo)
	}
	return historial, nil
}

// CambiarSeccion cierra el tramo actual del alumno y abre uno nuevo con la sección indicada.
func (s *Servicio) CambiarSeccion(ctx context.Context, usuarioID, alumnoID, seccionID int64, fecha time.Time) error {
	fecha = fechas.Fecha(fecha)

	return s.repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		alumno, err := s.repositorio.buscar(ctx, alumnoID)
		if err != nil {
			return err
		}
		if !alumno.Activo() {
			return ErrAlumnoNoActivo
		}

		actual, err := historialActual(ctx, consultas, alumnoID)
		if err != nil {
			return err
		}
		cambio, err := PrepararCambioSeccion(actual, seccionID, fecha)
		if err != nil {
			return err
		}

		inicio := fechas.Fecha(alumno.CreadoEn)
		if fecha.Before(inicio) {
			return ErrFechaInvalida
		}

		if actual != nil {
			if err := consultas.CerrarHistorialSeccion(ctx, generado.CerrarHistorialSeccionParams{
				AlumnoID: alumnoID,
				FechaFin: fechaPG(cambio.FechaCierre),
			}); err != nil {
				return fmt.Errorf("cerrando el tramo de sección: %w", err)
			}
		} else if _, err := consultas.CrearHistorialSeccion(ctx, generado.CrearHistorialSeccionParams{
			AlumnoID:    alumnoID,
			SeccionID:   alumno.SeccionID,
			FechaInicio: fechaPG(inicio),
			FechaFin:    fechaPG(cambio.FechaCierre),
		}); err != nil {
			return fmt.Errorf("registrando la sección anterior: %w", err)
		}

		if err := consultas.ActualizarSeccionAlumno(ctx, generado.ActualizarSeccionAlumnoParams{
			ID:        alumnoID,
			SeccionID: seccionID,
		}); err != nil {
			return fmt.Errorf("actualizando la sección del alumno: %w", err)
		}

		if _, err := consultas.CrearHistorialSeccion(ctx, generado.CrearHistorialSeccionParams{
			AlumnoID:    alumnoID,
			SeccionID:   cambio.SeccionNueva,
			FechaInicio: fechaPG(cambio.FechaApertura),
			FechaFin:    pgtype.Date{},
		}); err != nil {
			return fmt.Errorf("abriendo el nuevo tramo de sección: %w", err)
		}

		return auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID:     usuarioID,
			Entidad:       "alumno",
			EntidadID:     alumnoID,
			Accion:        "cambiar_seccion",
			ValorAnterior: map[string]any{"seccion_id": cambio.SeccionAnterior},
			ValorNuevo: map[string]any{
				"seccion_id": cambio.SeccionNueva,
				"fecha":      fechas.FormatearISO(cambio.FechaApertura),
			},
		})
	})
}

// RegistrarSalida marca al alumno como retirado o trasladado, cierra su tramo de sección y
// revoca el QR vigente en la misma transacción.
func (s *Servicio) RegistrarSalida(ctx context.Context, usuarioID, alumnoID int64, tipo string, fecha time.Time) error {
	if !TipoSalidaValido(tipo) {
		return ErrTipoSalidaInvalido
	}
	fecha = fechas.Fecha(fecha)
	if fecha.IsZero() {
		return ErrFechaInvalida
	}

	return s.repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		alumno, err := s.repositorio.buscar(ctx, alumnoID)
		if err != nil {
			return err
		}
		if !alumno.Activo() {
			return ErrAlumnoNoActivo
		}

		inicio := fechas.Fecha(alumno.CreadoEn)
		if fecha.Before(inicio) {
			return ErrFechaInvalida
		}

		actual, err := historialActual(ctx, consultas, alumnoID)
		if err != nil {
			return err
		}
		if actual != nil && fecha.Before(actual.FechaInicio) {
			return ErrFechaInvalida
		}

		if err := consultas.RegistrarSalidaAlumno(ctx, generado.RegistrarSalidaAlumnoParams{
			ID:          alumnoID,
			Estado:      tipo,
			FechaSalida: fechaPG(fecha),
		}); err != nil {
			return fmt.Errorf("registrando la salida del alumno: %w", err)
		}

		if actual != nil {
			if err := consultas.CerrarHistorialSeccion(ctx, generado.CerrarHistorialSeccionParams{
				AlumnoID: alumnoID,
				FechaFin: fechaPG(fecha),
			}); err != nil {
				return fmt.Errorf("cerrando el tramo de sección: %w", err)
			}
		} else if _, err := consultas.CrearHistorialSeccion(ctx, generado.CrearHistorialSeccionParams{
			AlumnoID:    alumnoID,
			SeccionID:   alumno.SeccionID,
			FechaInicio: fechaPG(inicio),
			FechaFin:    fechaPG(fecha),
		}); err != nil {
			return fmt.Errorf("registrando el tramo cerrado: %w", err)
		}

		if err := revocarCredencialVigente(ctx, consultas, db, usuarioID, alumnoID, MotivoRetiro); err != nil {
			return err
		}

		accion := "retirar"
		if tipo == EstadoTrasladado {
			accion = "trasladar"
		}
		return auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID:     usuarioID,
			Entidad:       "alumno",
			EntidadID:     alumnoID,
			Accion:        accion,
			ValorAnterior: map[string]any{"estado": alumno.Estado},
			ValorNuevo: map[string]any{
				"estado":       tipo,
				"fecha_salida": fechas.FormatearISO(fecha),
			},
		})
	})
}

// CredencialActiva devuelve el QR vigente del alumno.
func (s *Servicio) CredencialActiva(ctx context.Context, alumnoID int64) (CredencialQR, error) {
	fila, err := s.repositorio.consultas.BuscarCredencialActiva(ctx, alumnoID)
	if errors.Is(err, pgx.ErrNoRows) {
		return CredencialQR{}, ErrQRSinActivo
	}
	if err != nil {
		return CredencialQR{}, fmt.Errorf("buscando el QR vigente: %w", err)
	}
	return CredencialQR{
		ID:        fila.ID,
		Token:     fila.Token,
		Estado:    fila.Estado,
		EmitidaEn: fila.EmitidaEn.Time,
	}, nil
}

// HistorialQR devuelve todas las credenciales del alumno, de la más reciente a la más antigua.
func (s *Servicio) HistorialQR(ctx context.Context, alumnoID int64) ([]CredencialQR, error) {
	filas, err := s.repositorio.consultas.ListarCredenciales(ctx, alumnoID)
	if err != nil {
		return nil, fmt.Errorf("listando las credenciales: %w", err)
	}
	credenciales := make([]CredencialQR, 0, len(filas))
	for _, fila := range filas {
		credencial := CredencialQR{
			ID:        fila.ID,
			Token:     fila.Token,
			Estado:    fila.Estado,
			Motivo:    fila.MotivoRevocacion.String,
			EmitidaEn: fila.EmitidaEn.Time,
		}
		if fila.RevocadaEn.Valid {
			revocada := fila.RevocadaEn.Time
			credencial.RevocadaEn = &revocada
		}
		credenciales = append(credenciales, credencial)
	}
	return credenciales, nil
}

// EmitirQR crea el primer QR vigente de un alumno activo.
func (s *Servicio) EmitirQR(ctx context.Context, usuarioID, alumnoID int64) (CredencialQR, error) {
	var credencial CredencialQR
	err := s.repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		alumno, err := s.repositorio.buscar(ctx, alumnoID)
		if err != nil {
			return err
		}
		if !alumno.Activo() {
			return ErrAlumnoNoActivo
		}
		if _, err := consultas.BuscarCredencialActiva(ctx, alumnoID); err == nil {
			return ErrQRYaActivo
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("comprobando el QR vigente: %w", err)
		}

		token, err := GenerarTokenQR()
		if err != nil {
			return err
		}
		id, err := consultas.CrearCredencialQR(ctx, generado.CrearCredencialQRParams{
			AlumnoID: alumnoID,
			Token:    token,
		})
		if esViolacionUnica(err) {
			return ErrQRYaActivo
		}
		if err != nil {
			return fmt.Errorf("creando el QR: %w", err)
		}
		credencial = CredencialQR{ID: id, Token: token, Estado: CredencialActiva}

		return auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID:  usuarioID,
			Entidad:    "credencial_qr",
			EntidadID:  id,
			Accion:     "emitir_qr",
			ValorNuevo: map[string]any{"alumno_id": alumnoID, "estado": CredencialActiva},
		})
	})
	if err != nil {
		return CredencialQR{}, err
	}
	return credencial, nil
}

// ReemplazarQR revoca el QR vigente y emite uno nuevo en la misma transacción.
func (s *Servicio) ReemplazarQR(ctx context.Context, usuarioID, alumnoID int64, motivo string) (CredencialQR, error) {
	if !MotivoRevocacionValido(motivo) {
		return CredencialQR{}, ErrMotivoRequerido
	}

	var credencial CredencialQR
	err := s.repositorio.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		alumno, err := s.repositorio.buscar(ctx, alumnoID)
		if err != nil {
			return err
		}
		if !alumno.Activo() {
			return ErrAlumnoNoActivo
		}
		anterior, err := consultas.BuscarCredencialActiva(ctx, alumnoID)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrQRSinActivo
		}
		if err != nil {
			return fmt.Errorf("buscando el QR vigente: %w", err)
		}

		if err := consultas.RevocarCredencialActiva(ctx, generado.RevocarCredencialActivaParams{
			AlumnoID:         alumnoID,
			MotivoRevocacion: textoOpcional(motivo),
		}); err != nil {
			return fmt.Errorf("revocando el QR vigente: %w", err)
		}
		if err := auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID:     usuarioID,
			Entidad:       "credencial_qr",
			EntidadID:     anterior.ID,
			Accion:        "revocar_qr",
			ValorAnterior: map[string]any{"estado": CredencialActiva},
			ValorNuevo:    map[string]any{"estado": CredencialRevocada},
			Motivo:        motivo,
		}); err != nil {
			return err
		}

		token, err := GenerarTokenQR()
		if err != nil {
			return err
		}
		id, err := consultas.CrearCredencialQR(ctx, generado.CrearCredencialQRParams{
			AlumnoID: alumnoID,
			Token:    token,
		})
		if err != nil {
			return fmt.Errorf("creando el QR nuevo: %w", err)
		}
		credencial = CredencialQR{ID: id, Token: token, Estado: CredencialActiva}

		return auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID:  usuarioID,
			Entidad:    "credencial_qr",
			EntidadID:  id,
			Accion:     "emitir_qr",
			ValorNuevo: map[string]any{"alumno_id": alumnoID, "estado": CredencialActiva},
		})
	})
	if err != nil {
		return CredencialQR{}, err
	}
	return credencial, nil
}

func revocarCredencialVigente(ctx context.Context, consultas *generado.Queries, db generado.DBTX, usuarioID, alumnoID int64, motivo string) error {
	anterior, err := consultas.BuscarCredencialActiva(ctx, alumnoID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("buscando el QR vigente: %w", err)
	}
	if err := consultas.RevocarCredencialActiva(ctx, generado.RevocarCredencialActivaParams{
		AlumnoID:         alumnoID,
		MotivoRevocacion: textoOpcional(motivo),
	}); err != nil {
		return fmt.Errorf("revocando el QR: %w", err)
	}
	return auditoria.Registrar(ctx, db, auditoria.Evento{
		UsuarioID:     usuarioID,
		Entidad:       "credencial_qr",
		EntidadID:     anterior.ID,
		Accion:        "revocar_qr",
		ValorAnterior: map[string]any{"estado": CredencialActiva},
		ValorNuevo:    map[string]any{"estado": CredencialRevocada},
		Motivo:        motivo,
	})
}

func datosNormalizados(datos DatosAlumno) DatosAlumno {
	datos.Identificador = LimpiarIdentificador(datos.Identificador)
	datos.Nombres = strings.TrimSpace(datos.Nombres)
	datos.Apellidos = strings.TrimSpace(datos.Apellidos)
	datos.ApoderadoNombre = strings.TrimSpace(datos.ApoderadoNombre)
	datos.ApoderadoTelefono = strings.TrimSpace(datos.ApoderadoTelefono)
	return datos
}

func esViolacionUnica(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == codigoViolacionUnica
}

// historialActual devuelve el tramo vigente del alumno o nil si todavía no tiene historial.
func historialActual(ctx context.Context, consultas *generado.Queries, alumnoID int64) (*HistorialSeccion, error) {
	fila, err := consultas.BuscarHistorialActual(ctx, alumnoID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("buscando el tramo de sección vigente: %w", err)
	}
	tramo := HistorialSeccion{
		ID:          fila.ID,
		SeccionID:   fila.SeccionID,
		Grado:       fila.GradoNombre,
		Seccion:     fila.SeccionNombre,
		AnioEscolar: fila.AnioEscolarNombre,
		FechaInicio: fila.FechaInicio.Time,
	}
	if fila.FechaFin.Valid {
		fin := fila.FechaFin.Time
		tramo.FechaFin = &fin
	}
	return &tramo, nil
}

// fechaPG convierte una fecha del colegio al tipo date de PostgreSQL.
func fechaPG(fecha time.Time) pgtype.Date {
	return pgtype.Date{Time: fecha, Valid: !fecha.IsZero()}
}
