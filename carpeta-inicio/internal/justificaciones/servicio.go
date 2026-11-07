package justificaciones

import (
	"context"
	"fmt"
	"strings"
	"time"

	"asistencia/internal/alumnos"
	"asistencia/internal/asistencia"
	"asistencia/internal/justificaciones/generado"
	"asistencia/internal/plataforma/auditoria"
	"asistencia/internal/plataforma/fechas"
)

// Servicio aplica las reglas de justificaciones.
type Servicio struct {
	dependencias Dependencias
}

// NuevoServicio crea el servicio.
func NuevoServicio(dependencias Dependencias) *Servicio {
	if dependencias.Reloj == nil {
		dependencias.Reloj = time.Now
	}
	return &Servicio{dependencias: dependencias}
}

func (s *Servicio) hoy() time.Time {
	return fechas.Fecha(s.dependencias.Reloj())
}

// Registrar crea una justificación para un rango de fechas.
func (s *Servicio) Registrar(
	ctx context.Context,
	usuarioID int64,
	rol string,
	alumnoID int64,
	motivo string,
	desde, hasta time.Time,
	adjuntos []EntradaAdjunto,
) (ResultadoRegistro, error) {
	motivo = strings.TrimSpace(motivo)
	if len([]rune(motivo)) < 3 {
		return ResultadoRegistro{}, ErrMotivoRequerido
	}
	if len([]rune(motivo)) > 500 {
		return ResultadoRegistro{}, ErrMotivoLargo
	}
	if hasta.Before(desde) {
		return ResultadoRegistro{}, ErrRangoInvalido
	}
	if err := s.validarAdjuntos(adjuntos); err != nil {
		return ResultadoRegistro{}, err
	}

	alumno, err := s.dependencias.Alumnos.Obtener(ctx, alumnoID)
	if err != nil {
		return ResultadoRegistro{}, err
	}
	if alumno.Estado != "activo" {
		return ResultadoRegistro{}, ErrAlumnoInactivo
	}
	if alumno.SeccionID == 0 {
		return ResultadoRegistro{}, ErrSinAcceso
	}
	if rol == "docente" {
		asignado, err := s.dependencias.Asistencia.DocenteAsignadoASeccionEnFecha(ctx, usuarioID, alumno.SeccionID, desde)
		if err != nil {
			return ResultadoRegistro{}, err
		}
		if !asignado {
			return ResultadoRegistro{}, ErrSinAcceso
		}
	}

	plazo, err := s.dependencias.Asistencia.PlazoJustificacion(ctx)
	if err != nil {
		return ResultadoRegistro{}, err
	}
	hoy := s.hoy()
	limite := hoy.AddDate(0, 0, -plazo)

	resultado := ResultadoRegistro{}
	for fecha := desde; !fecha.After(hasta); fecha = fecha.AddDate(0, 0, 1) {
		tipo, omitir, motivoOmision, err := s.clasificarFecha(ctx, alumno.SeccionID, alumnoID, fecha, hoy)
		if err != nil {
			return ResultadoRegistro{}, err
		}
		if omitir {
			resultado.Omitidas = append(resultado.Omitidas, Omision{Fecha: fecha, Motivo: motivoOmision})
			continue
		}
		if fecha.Before(limite) {
			return ResultadoRegistro{}, ErrFueraDePlazo
		}
		activa, err := s.dependencias.Repo.ExisteActiva(ctx, alumnoID, fecha)
		if err != nil {
			return ResultadoRegistro{}, err
		}
		if activa {
			resultado.Omitidas = append(resultado.Omitidas, Omision{Fecha: fecha, Motivo: "ya justificado o pendiente"})
			continue
		}
		resultado.Cubiertas = append(resultado.Cubiertas, FechaJustificada{
			AlumnoID: alumnoID,
			Fecha:    fecha,
			Tipo:     tipo,
		})
	}
	if len(resultado.Cubiertas) == 0 {
		return resultado, ErrSinFechasAplicables
	}

	adjuntosGuardados, err := s.guardarAdjuntos(ctx, adjuntos)
	if err != nil {
		return ResultadoRegistro{}, err
	}
	justificacion := Justificacion{
		AlumnoID:      alumnoID,
		SeccionID:     alumno.SeccionID,
		Motivo:        motivo,
		RegistranteID: usuarioID,
	}
	var id int64
	err = s.dependencias.Repo.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		creado, err := s.dependencias.Repo.Insertar(ctx, consultas, justificacion, resultado.Cubiertas, adjuntosGuardados)
		if err != nil {
			return err
		}
		id = creado
		return auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID: usuarioID,
			Entidad:   EntidadJustificacion,
			EntidadID: creado,
			Accion:    AccionCrearJustificacion,
			ValorNuevo: map[string]any{
				"alumno_id": alumnoID,
				"motivo":    motivo,
				"fechas":    fechasDe(resultado.Cubiertas),
			},
		})
	})
	if err != nil {
		return ResultadoRegistro{}, err
	}
	resultado.JustificacionID = id
	return resultado, nil
}

// Resolver aprueba o rechaza una justificación pendiente.
func (s *Servicio) Resolver(ctx context.Context, usuarioID int64, rol string, id int64, aprobar bool, comentario string) error {
	if rol != "administrador" && rol != "director" {
		return ErrSinAcceso
	}
	comentario = strings.TrimSpace(comentario)
	if !aprobar && comentario == "" {
		return ErrComentarioRequerido
	}
	justificacion, err := s.dependencias.Repo.Obtener(ctx, id)
	if err != nil {
		return err
	}
	if justificacion.Estado != EstadoPendiente {
		return ErrYaResuelta
	}
	if justificacion.RegistranteID == usuarioID && rol != "administrador" {
		return ErrAutoresolucion
	}

	estado := EstadoRechazada
	accion := AccionRechazarJustificacion
	if aprobar {
		estado = EstadoAprobada
		accion = AccionAprobarJustificacion
	}
	err = s.dependencias.Repo.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		cambiada, err := s.dependencias.Repo.Resolver(ctx, consultas, id, estado, usuarioID, comentario)
		if err != nil {
			return err
		}
		if !cambiada {
			return ErrYaResuelta
		}
		if err := s.dependencias.Repo.ActualizarFechas(ctx, consultas, id, estado); err != nil {
			return err
		}
		return auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID:     usuarioID,
			Entidad:       EntidadJustificacion,
			EntidadID:     id,
			Accion:        accion,
			ValorAnterior: map[string]any{"estado": EstadoPendiente},
			ValorNuevo:    map[string]any{"estado": estado, "comentario": comentario},
		})
	})
	if err != nil {
		return err
	}
	if aprobar {
		return s.aplicarAprobacion(ctx, justificacion)
	}
	return nil
}

func (s *Servicio) aplicarAprobacion(ctx context.Context, justificacion Justificacion) error {
	for _, fecha := range justificacion.Fechas {
		if fecha.Tipo == TipoTardanza {
			if err := s.dependencias.Asistencia.MarcarTardanzaJustificada(ctx, justificacion.AlumnoID, fecha.Fecha, true); err != nil {
				return err
			}
		}
		if s.dependencias.Alertas != nil {
			if err := s.dependencias.Alertas.MarcarSituacionCambiada(ctx, justificacion.AlumnoID, fecha.Fecha, "Justificación aprobada"); err != nil {
				return err
			}
		}
	}
	return nil
}

// Anular revierte una justificación aprobada. Solo el administrador puede hacerlo.
func (s *Servicio) Anular(ctx context.Context, usuarioID int64, rol string, id int64, motivo string) error {
	if rol != "administrador" {
		return ErrSoloAdministrador
	}
	motivo = strings.TrimSpace(motivo)
	if motivo == "" {
		return ErrMotivoAnulacion
	}
	justificacion, err := s.dependencias.Repo.Obtener(ctx, id)
	if err != nil {
		return err
	}
	if justificacion.Estado != EstadoAprobada {
		return ErrNoAnulable
	}
	err = s.dependencias.Repo.EnTransaccion(ctx, func(consultas *generado.Queries, db generado.DBTX) error {
		anulada, err := s.dependencias.Repo.Anular(ctx, consultas, id, usuarioID, motivo)
		if err != nil {
			return err
		}
		if !anulada {
			return ErrNoAnulable
		}
		if err := s.dependencias.Repo.ActualizarFechas(ctx, consultas, id, EstadoAnulada); err != nil {
			return err
		}
		return auditoria.Registrar(ctx, db, auditoria.Evento{
			UsuarioID:     usuarioID,
			Entidad:       EntidadJustificacion,
			EntidadID:     id,
			Accion:        AccionAnularJustificacion,
			ValorAnterior: map[string]any{"estado": EstadoAprobada},
			ValorNuevo:    map[string]any{"estado": EstadoAnulada},
			Motivo:        motivo,
		})
	})
	if err != nil {
		return err
	}
	for _, fecha := range justificacion.Fechas {
		if fecha.Tipo == TipoTardanza {
			if err := s.dependencias.Asistencia.MarcarTardanzaJustificada(ctx, justificacion.AlumnoID, fecha.Fecha, false); err != nil {
				return err
			}
		}
	}
	return nil
}

// Listar devuelve las justificaciones visibles para el usuario.
func (s *Servicio) Listar(ctx context.Context, usuarioID int64, rol string) ([]Justificacion, error) {
	switch rol {
	case "administrador", "director":
		return s.dependencias.Repo.ListarPendientes(ctx)
	default:
		return s.dependencias.Repo.ListarDelRegistrante(ctx, usuarioID)
	}
}

// HistorialDeAlumno devuelve las justificaciones de un alumno si el usuario tiene acceso.
func (s *Servicio) HistorialDeAlumno(ctx context.Context, usuarioID int64, rol string, alumnoID int64) ([]Justificacion, error) {
	justificaciones, err := s.dependencias.Repo.ListarDeAlumno(ctx, alumnoID)
	if err != nil {
		return nil, err
	}
	if rol == "administrador" || rol == "director" {
		return justificaciones, nil
	}
	filtradas := make([]Justificacion, 0, len(justificaciones))
	for _, justificacion := range justificaciones {
		if justificacion.RegistranteID == usuarioID {
			filtradas = append(filtradas, justificacion)
			continue
		}
		if len(justificacion.Fechas) > 0 {
			asignado, err := s.dependencias.Asistencia.DocenteAsignadoASeccionEnFecha(
				ctx, usuarioID, justificacion.SeccionID, justificacion.Fechas[0].Fecha)
			if err != nil {
				return nil, err
			}
			if asignado {
				filtradas = append(filtradas, justificacion)
			}
		}
	}
	return filtradas, nil
}

// ObtenerConAcceso devuelve una justificación si el usuario puede verla.
func (s *Servicio) ObtenerConAcceso(ctx context.Context, usuarioID int64, rol string, id int64) (Justificacion, error) {
	justificacion, err := s.dependencias.Repo.Obtener(ctx, id)
	if err != nil {
		return Justificacion{}, err
	}
	if rol == "administrador" || rol == "director" || justificacion.RegistranteID == usuarioID {
		return justificacion, nil
	}
	if len(justificacion.Fechas) > 0 {
		asignado, err := s.dependencias.Asistencia.DocenteAsignadoASeccionEnFecha(
			ctx, usuarioID, justificacion.SeccionID, justificacion.Fechas[0].Fecha)
		if err != nil {
			return Justificacion{}, err
		}
		if asignado {
			return justificacion, nil
		}
	}
	return Justificacion{}, ErrSinAcceso
}

// AbrirAdjunto devuelve el documento si el usuario puede verlo.
func (s *Servicio) AbrirAdjunto(ctx context.Context, usuarioID int64, rol string, adjuntoID int64) (Adjunto, []byte, error) {
	adjunto, err := s.dependencias.Repo.ObtenerAdjunto(ctx, adjuntoID)
	if err != nil {
		return Adjunto{}, nil, err
	}
	if _, err := s.ObtenerConAcceso(ctx, usuarioID, rol, adjunto.JustificacionID); err != nil {
		return Adjunto{}, nil, err
	}
	lector, err := s.dependencias.Almacen.Obtener(ctx, adjunto.Clave)
	if err != nil {
		return Adjunto{}, nil, err
	}
	defer func() { _ = lector.Close() }()
	contenido := make([]byte, 0, adjunto.Tamano)
	buffer := make([]byte, 32*1024)
	for {
		leidos, err := lector.Read(buffer)
		if leidos > 0 {
			contenido = append(contenido, buffer[:leidos]...)
		}
		if err != nil {
			break
		}
	}
	return adjunto, contenido, nil
}

// EsFaltaJustificada responde a la feature 004.
func (s *Servicio) EsFaltaJustificada(ctx context.Context, alumnoID int64, fecha time.Time) (bool, error) {
	return s.dependencias.Repo.EsFaltaJustificada(ctx, alumnoID, fecha)
}

func (s *Servicio) clasificarFecha(ctx context.Context, seccionID, alumnoID int64, fecha, hoy time.Time) (string, bool, string, error) {
	if fecha.After(hoy) {
		return "", true, "fecha futura", nil
	}
	if fecha.Weekday() == time.Saturday || fecha.Weekday() == time.Sunday {
		return "", true, "día no laborable", nil
	}
	noLaborable, err := s.dependencias.Academico.EsDiaNoLaborable(ctx, seccionID, fecha)
	if err != nil {
		return "", false, "", err
	}
	if noLaborable {
		return "", true, "día no laborable", nil
	}
	marcacion, existe, err := s.dependencias.Asistencia.MarcacionDelDia(ctx, alumnoID, fecha)
	if err != nil {
		return "", false, "", err
	}
	if !existe {
		return TipoFalta, false, "", nil
	}
	if marcacion.Estado == asistencia.MarcacionPuntual {
		return "", true, "asistencia presente", nil
	}
	if marcacion.Justificada {
		return "", true, "ya justificado", nil
	}
	return TipoTardanza, false, "", nil
}

func (s *Servicio) validarAdjuntos(adjuntos []EntradaAdjunto) error {
	if len(adjuntos) > 3 {
		return ErrAdjuntosExcedidos
	}
	for _, adjunto := range adjuntos {
		if adjunto.Tamano <= 0 || adjunto.Tamano > 5*1024*1024 {
			return ErrAdjuntoInvalido
		}
		switch adjunto.Tipo {
		case "application/pdf", "image/jpeg", "image/png":
		default:
			return ErrAdjuntoInvalido
		}
	}
	return nil
}

func (s *Servicio) guardarAdjuntos(ctx context.Context, adjuntos []EntradaAdjunto) ([]Adjunto, error) {
	if len(adjuntos) == 0 {
		return nil, nil
	}
	guardados := make([]Adjunto, 0, len(adjuntos))
	for indice, adjunto := range adjuntos {
		nombre := sanitizarNombre(adjunto.Nombre)
		clave := fmt.Sprintf("justificaciones/%d-%d-%s", time.Now().UnixNano(), indice, nombre)
		if err := s.dependencias.Almacen.Guardar(ctx, clave, strings.NewReader(string(adjunto.Contenido))); err != nil {
			return nil, err
		}
		guardados = append(guardados, Adjunto{
			Nombre: nombre,
			Tipo:   adjunto.Tipo,
			Tamano: adjunto.Tamano,
			Clave:  clave,
		})
	}
	return guardados, nil
}

func sanitizarNombre(nombre string) string {
	nombre = strings.TrimSpace(nombre)
	if nombre == "" {
		return "documento"
	}
	reemplazos := strings.NewReplacer("/", "-", "\\", "-", "..", "-")
	return reemplazos.Replace(nombre)
}

func fechasDe(fechas []FechaJustificada) []string {
	resultado := make([]string, 0, len(fechas))
	for _, fecha := range fechas {
		resultado = append(resultado, fecha.Fecha.Format("2006-01-02"))
	}
	return resultado
}

// JustificacionDetalle agrega datos legibles para la interfaz.
type JustificacionDetalle struct {
	Justificacion
	Alumno  string
	Seccion string
}

// Describir completa las justificaciones con el nombre del alumno y su sección.
func (s *Servicio) Describir(ctx context.Context, lista []Justificacion) ([]JustificacionDetalle, error) {
	resultado := make([]JustificacionDetalle, 0, len(lista))
	for _, justificacion := range lista {
		alumno, err := s.dependencias.Alumnos.Obtener(ctx, justificacion.AlumnoID)
		if err != nil {
			return nil, err
		}
		seccion, err := s.dependencias.Academico.EtiquetaSeccion(ctx, justificacion.SeccionID)
		if err != nil {
			seccion = ""
		}
		resultado = append(resultado, JustificacionDetalle{
			Justificacion: justificacion,
			Alumno:        alumno.NombreCompleto(),
			Seccion:       seccion,
		})
	}
	return resultado, nil
}

// Alumno devuelve los datos del alumno para la pantalla de registro.
func (s *Servicio) Alumno(ctx context.Context, alumnoID int64) (alumnos.Alumno, error) {
	return s.dependencias.Alumnos.Obtener(ctx, alumnoID)
}
