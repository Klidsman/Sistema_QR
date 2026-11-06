package alumnos

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"asistencia/internal/academico"
	"asistencia/internal/plataforma/almacenamiento"
	"asistencia/internal/plataforma/fechas"
	"asistencia/internal/plataforma/intermediarios"
	vista "asistencia/web/plantillas/alumnos"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
)

// ServicioAlumnos son las operaciones de alumnos que usa el manejador.
type ServicioAlumnos interface {
	Listar(ctx context.Context, filtros FiltrosListado) ([]AlumnoListado, error)
	Obtener(ctx context.Context, id int64) (Alumno, error)
	Crear(ctx context.Context, usuarioID int64, datos DatosAlumno) (int64, error)
	Actualizar(ctx context.Context, usuarioID, alumnoID int64, datos DatosAlumno) error
	ActualizarFoto(ctx context.Context, usuarioID, alumnoID int64, clave string) error
	Desactivar(ctx context.Context, usuarioID, alumnoID int64) error
}

// ServicioQR son las operaciones de credenciales que usa el manejador.
type ServicioQR interface {
	CredencialActiva(ctx context.Context, alumnoID int64) (CredencialQR, error)
	HistorialQR(ctx context.Context, alumnoID int64) ([]CredencialQR, error)
	EmitirQR(ctx context.Context, usuarioID, alumnoID int64) (CredencialQR, error)
	ReemplazarQR(ctx context.Context, usuarioID, alumnoID int64, motivo string) (CredencialQR, error)
}

// Dependencias reúne todo lo que el manejador de alumnos necesita.
type Dependencias struct {
	Alumnos      ServicioAlumnos
	QR           ServicioQR
	Secciones    academico.BuscadorSeccion
	Almacen      almacenamiento.AlmacenFotos
	Sesiones     *scs.SessionManager
	Registro     *slog.Logger
	FotoMaximaMB int
}

// Manejador atiende las pantallas y acciones de alumnos y de sus códigos QR.
type Manejador struct {
	alumnos      ServicioAlumnos
	qr           ServicioQR
	secciones    academico.BuscadorSeccion
	almacen      almacenamiento.AlmacenFotos
	sesiones     *scs.SessionManager
	registro     *slog.Logger
	fotoMaximaMB int
}

// NuevoManejador crea el manejador de alumnos.
func NuevoManejador(dependencias Dependencias) *Manejador {
	return &Manejador{
		alumnos:      dependencias.Alumnos,
		qr:           dependencias.QR,
		secciones:    dependencias.Secciones,
		almacen:      dependencias.Almacen,
		sesiones:     dependencias.Sesiones,
		registro:     dependencias.Registro,
		fotoMaximaMB: dependencias.FotoMaximaMB,
	}
}

// Montar registra las rutas de alumnos agrupadas por rol.
func (m *Manejador) Montar(enrutador chi.Router) {
	enrutador.Group(func(lectura chi.Router) {
		lectura.Use(intermediarios.RequerirRol(m.sesiones, intermediarios.RolAdministrador, intermediarios.RolDirector))
		lectura.Get("/alumnos", m.listar)
		lectura.Get("/alumnos/{id}", m.ficha)
		lectura.Get("/alumnos/{id}/foto", m.foto)
		lectura.Get("/alumnos/{id}/qr/imagen", m.imagenQR)
	})

	enrutador.Group(func(escritura chi.Router) {
		escritura.Use(intermediarios.RequerirRol(m.sesiones, intermediarios.RolAdministrador))
		escritura.Get("/alumnos/nuevo", m.nuevo)
		escritura.Post("/alumnos", m.crear)
		escritura.Get("/alumnos/{id}/editar", m.editar)
		escritura.Post("/alumnos/{id}", m.actualizar)
		escritura.Post("/alumnos/{id}/desactivacion", m.desactivar)
		escritura.Post("/alumnos/{id}/foto", m.subirFoto)
		escritura.Post("/alumnos/{id}/qr", m.emitirQR)
		escritura.Post("/alumnos/{id}/qr/reemplazo", m.reemplazarQR)
	})
}

func (m *Manejador) listar(respuesta http.ResponseWriter, peticion *http.Request) {
	filtros := FiltrosListado{
		Texto:     peticion.URL.Query().Get("texto"),
		Estado:    peticion.URL.Query().Get("estado"),
		SeccionID: enteroDeTexto(peticion.URL.Query().Get("seccion_id")),
	}
	if !estadoValido(filtros.Estado) {
		filtros.Estado = ""
	}

	listado, err := m.alumnos.Listar(peticion.Context(), filtros)
	if err != nil {
		m.errorInterno(respuesta, "listando alumnos", err)
		return
	}
	secciones, err := m.secciones.ListarSecciones(peticion.Context())
	if err != nil {
		m.errorInterno(respuesta, "listando secciones", err)
		return
	}

	m.renderizar(respuesta, peticion, http.StatusOK, vista.Listado(
		m.nombreUsuario(peticion.Context()),
		intermediarios.RolUsuario(peticion.Context()),
		m.tokenCSRF(peticion.Context()),
		secciones,
		vista.Filtros{Texto: filtros.Texto, Estado: filtros.Estado, SeccionID: filtros.SeccionID},
		aListadoVista(listado),
		mensajeCorrecto(peticion.URL.Query().Get("ok")),
		mensajeError(peticion.URL.Query().Get("error")),
	))
}

func (m *Manejador) nuevo(respuesta http.ResponseWriter, peticion *http.Request) {
	m.renderizarFormulario(respuesta, peticion, http.StatusOK, DatosAlumno{}, nil, "")
}

func (m *Manejador) editar(respuesta http.ResponseWriter, peticion *http.Request) {
	id, ok := idDeRuta(peticion)
	if !ok {
		http.NotFound(respuesta, peticion)
		return
	}
	alumno, err := m.alumnos.Obtener(peticion.Context(), id)
	if errors.Is(err, ErrAlumnoNoEncontrado) {
		http.NotFound(respuesta, peticion)
		return
	}
	if err != nil {
		m.errorInterno(respuesta, "obteniendo el alumno", err)
		return
	}

	datos := DatosAlumno{
		Identificador:     alumno.Identificador,
		Nombres:           alumno.Nombres,
		Apellidos:         alumno.Apellidos,
		SeccionID:         alumno.SeccionID,
		ApoderadoNombre:   alumno.ApoderadoNombre,
		ApoderadoTelefono: alumno.ApoderadoTelefono,
	}
	m.renderizarFormularioConID(respuesta, peticion, http.StatusOK, alumno.ID, datos, nil, "")
}

func (m *Manejador) crear(respuesta http.ResponseWriter, peticion *http.Request) {
	datos, foto, errFoto := m.leerFormulario(respuesta, peticion)
	if validacion := Validar(datos); validacion != nil {
		m.renderizarFormulario(respuesta, peticion, http.StatusUnprocessableEntity, datos, validacion.Campos, mensajeDeErrorFoto(errFoto))
		return
	}

	usuarioID := intermediarios.UsuarioID(peticion.Context())
	id, err := m.alumnos.Crear(peticion.Context(), usuarioID, datos)
	if errors.Is(err, ErrIdentificadorExistente) {
		m.renderizarFormulario(respuesta, peticion, http.StatusUnprocessableEntity, datos,
			map[string]string{"identificador": "Ya existe un alumno con este DNI o código."}, mensajeDeErrorFoto(errFoto))
		return
	}
	if err != nil {
		m.errorInterno(respuesta, "creando el alumno", err)
		return
	}

	destino := fmt.Sprintf("/alumnos/%d?ok=creado", id)
	if errFoto != nil {
		destino += "&error=" + codigoDeErrorFoto(errFoto)
	} else if foto != nil {
		if err := m.guardarFoto(peticion.Context(), usuarioID, id, foto); err != nil {
			m.registro.Error("no se pudo guardar la foto", "alumno_id", id, "error", err)
			destino += "&error=foto_no_guardada"
		}
	}
	http.Redirect(respuesta, peticion, destino, http.StatusSeeOther)
}

func (m *Manejador) actualizar(respuesta http.ResponseWriter, peticion *http.Request) {
	id, ok := idDeRuta(peticion)
	if !ok {
		http.NotFound(respuesta, peticion)
		return
	}
	datos, foto, errFoto := m.leerFormulario(respuesta, peticion)
	if validacion := Validar(datos); validacion != nil {
		m.renderizarFormularioConID(respuesta, peticion, http.StatusUnprocessableEntity, id, datos, validacion.Campos, mensajeDeErrorFoto(errFoto))
		return
	}

	usuarioID := intermediarios.UsuarioID(peticion.Context())
	err := m.alumnos.Actualizar(peticion.Context(), usuarioID, id, datos)
	if errors.Is(err, ErrAlumnoNoEncontrado) {
		http.NotFound(respuesta, peticion)
		return
	}
	if err != nil {
		m.errorInterno(respuesta, "actualizando el alumno", err)
		return
	}

	destino := fmt.Sprintf("/alumnos/%d?ok=actualizado", id)
	if errFoto != nil {
		destino += "&error=" + codigoDeErrorFoto(errFoto)
	} else if foto != nil {
		if err := m.guardarFoto(peticion.Context(), usuarioID, id, foto); err != nil {
			m.registro.Error("no se pudo guardar la foto", "alumno_id", id, "error", err)
			destino += "&error=foto_no_guardada"
		}
	}
	http.Redirect(respuesta, peticion, destino, http.StatusSeeOther)
}

func (m *Manejador) desactivar(respuesta http.ResponseWriter, peticion *http.Request) {
	id, ok := idDeRuta(peticion)
	if !ok {
		http.NotFound(respuesta, peticion)
		return
	}
	err := m.alumnos.Desactivar(peticion.Context(), intermediarios.UsuarioID(peticion.Context()), id)
	if errors.Is(err, ErrAlumnoNoEncontrado) {
		http.NotFound(respuesta, peticion)
		return
	}
	if err != nil {
		m.errorInterno(respuesta, "desactivando el alumno", err)
		return
	}
	http.Redirect(respuesta, peticion, fmt.Sprintf("/alumnos/%d?ok=desactivado", id), http.StatusSeeOther)
}

func (m *Manejador) subirFoto(respuesta http.ResponseWriter, peticion *http.Request) {
	id, ok := idDeRuta(peticion)
	if !ok {
		http.NotFound(respuesta, peticion)
		return
	}
	archivo, _, err := peticion.FormFile("foto")
	if err != nil {
		http.Redirect(respuesta, peticion, fmt.Sprintf("/alumnos/%d?error=foto_invalida", id), http.StatusSeeOther)
		return
	}
	defer archivo.Close()

	foto, err := LeerFoto(archivo, m.fotoMaximaMB)
	if err != nil {
		http.Redirect(respuesta, peticion, fmt.Sprintf("/alumnos/%d?error=%s", id, codigoDeErrorFoto(err)), http.StatusSeeOther)
		return
	}
	if err := m.guardarFoto(peticion.Context(), intermediarios.UsuarioID(peticion.Context()), id, &foto); err != nil {
		m.errorInterno(respuesta, "guardando la foto", err)
		return
	}
	http.Redirect(respuesta, peticion, fmt.Sprintf("/alumnos/%d?ok=foto_actualizada", id), http.StatusSeeOther)
}

func (m *Manejador) ficha(respuesta http.ResponseWriter, peticion *http.Request) {
	id, ok := idDeRuta(peticion)
	if !ok {
		http.NotFound(respuesta, peticion)
		return
	}
	alumno, err := m.alumnos.Obtener(peticion.Context(), id)
	if errors.Is(err, ErrAlumnoNoEncontrado) {
		http.NotFound(respuesta, peticion)
		return
	}
	if err != nil {
		m.errorInterno(respuesta, "obteniendo el alumno", err)
		return
	}

	var credencial *CredencialQR
	activa, err := m.qr.CredencialActiva(peticion.Context(), id)
	switch {
	case err == nil:
		credencial = &activa
	case !errors.Is(err, ErrQRSinActivo):
		m.errorInterno(respuesta, "buscando el QR vigente", err)
		return
	}

	historial, err := m.qr.HistorialQR(peticion.Context(), id)
	if err != nil {
		m.registro.Error("no se pudo leer el historial de QR", "alumno_id", id, "error", err)
		historial = nil
	}

	m.renderizar(respuesta, peticion, http.StatusOK, vista.Ficha(
		m.nombreUsuario(peticion.Context()),
		intermediarios.RolUsuario(peticion.Context()),
		m.tokenCSRF(peticion.Context()),
		aAlumnoVista(alumno),
		aCredencialVista(credencial),
		aHistorialVista(historial),
		mensajeCorrecto(peticion.URL.Query().Get("ok")),
		mensajeError(peticion.URL.Query().Get("error")),
	))
}

func (m *Manejador) foto(respuesta http.ResponseWriter, peticion *http.Request) {
	id, ok := idDeRuta(peticion)
	if !ok {
		http.NotFound(respuesta, peticion)
		return
	}
	alumno, err := m.alumnos.Obtener(peticion.Context(), id)
	if errors.Is(err, ErrAlumnoNoEncontrado) {
		http.NotFound(respuesta, peticion)
		return
	}
	if err != nil {
		m.errorInterno(respuesta, "obteniendo el alumno", err)
		return
	}
	if alumno.FotoClave == "" {
		http.NotFound(respuesta, peticion)
		return
	}

	lectura, err := m.almacen.Obtener(peticion.Context(), alumno.FotoClave)
	if errors.Is(err, almacenamiento.ErrNoEncontrado) {
		http.NotFound(respuesta, peticion)
		return
	}
	if err != nil {
		m.errorInterno(respuesta, "obteniendo la foto", err)
		return
	}
	defer lectura.Close()

	tipo := "image/jpeg"
	if ExtensionDesdeClave(alumno.FotoClave) == ".png" {
		tipo = "image/png"
	}
	respuesta.Header().Set("Content-Type", tipo)
	respuesta.Header().Set("Cache-Control", "private, max-age=300")
	if _, err := io.Copy(respuesta, lectura); err != nil {
		m.registro.Error("no se pudo enviar la foto", "alumno_id", id, "error", err)
	}
}

func (m *Manejador) emitirQR(respuesta http.ResponseWriter, peticion *http.Request) {
	id, ok := idDeRuta(peticion)
	if !ok {
		http.NotFound(respuesta, peticion)
		return
	}
	_, err := m.qr.EmitirQR(peticion.Context(), intermediarios.UsuarioID(peticion.Context()), id)
	switch {
	case errors.Is(err, ErrAlumnoNoEncontrado):
		http.NotFound(respuesta, peticion)
		return
	case errors.Is(err, ErrAlumnoNoActivo):
		m.redirigirConError(respuesta, peticion, id, "alumno_no_activo")
		return
	case errors.Is(err, ErrQRYaActivo):
		m.redirigirConError(respuesta, peticion, id, "qr_ya_activo")
		return
	case err != nil:
		m.errorInterno(respuesta, "emitiendo el QR", err)
		return
	}
	http.Redirect(respuesta, peticion, fmt.Sprintf("/alumnos/%d?ok=qr_emitido", id), http.StatusSeeOther)
}

func (m *Manejador) reemplazarQR(respuesta http.ResponseWriter, peticion *http.Request) {
	id, ok := idDeRuta(peticion)
	if !ok {
		http.NotFound(respuesta, peticion)
		return
	}
	motivo := peticion.PostFormValue("motivo")
	_, err := m.qr.ReemplazarQR(peticion.Context(), intermediarios.UsuarioID(peticion.Context()), id, motivo)
	switch {
	case errors.Is(err, ErrAlumnoNoEncontrado):
		http.NotFound(respuesta, peticion)
		return
	case errors.Is(err, ErrMotivoRequerido):
		m.redirigirConError(respuesta, peticion, id, "motivo_invalido")
		return
	case errors.Is(err, ErrQRSinActivo):
		m.redirigirConError(respuesta, peticion, id, "qr_sin_activo")
		return
	case errors.Is(err, ErrAlumnoNoActivo):
		m.redirigirConError(respuesta, peticion, id, "alumno_no_activo")
		return
	case err != nil:
		m.errorInterno(respuesta, "reemplazando el QR", err)
		return
	}
	http.Redirect(respuesta, peticion, fmt.Sprintf("/alumnos/%d?ok=qr_reemplazado", id), http.StatusSeeOther)
}

func (m *Manejador) imagenQR(respuesta http.ResponseWriter, peticion *http.Request) {
	id, ok := idDeRuta(peticion)
	if !ok {
		http.NotFound(respuesta, peticion)
		return
	}
	credencial, err := m.qr.CredencialActiva(peticion.Context(), id)
	if errors.Is(err, ErrQRSinActivo) {
		http.NotFound(respuesta, peticion)
		return
	}
	if err != nil {
		m.errorInterno(respuesta, "buscando el QR vigente", err)
		return
	}

	imagen, err := ImagenPNG(credencial.Token)
	if err != nil {
		m.errorInterno(respuesta, "generando la imagen del QR", err)
		return
	}
	respuesta.Header().Set("Content-Type", "image/png")
	respuesta.Header().Set("Cache-Control", "private, no-store")
	_, _ = respuesta.Write(imagen)
}

func (m *Manejador) leerFormulario(respuesta http.ResponseWriter, peticion *http.Request) (DatosAlumno, *Foto, error) {
	limite := int64(m.fotoMaximaMB+10) * 1024 * 1024
	peticion.Body = http.MaxBytesReader(respuesta, peticion.Body, limite)
	if err := peticion.ParseMultipartForm(limite); err != nil {
		return DatosAlumno{}, nil, ErrFotoMuyGrande
	}

	datos := DatosAlumno{
		Identificador:     peticion.PostFormValue("identificador"),
		Nombres:           peticion.PostFormValue("nombres"),
		Apellidos:         peticion.PostFormValue("apellidos"),
		SeccionID:         enteroDeTexto(peticion.PostFormValue("seccion_id")),
		ApoderadoNombre:   peticion.PostFormValue("apoderado_nombre"),
		ApoderadoTelefono: peticion.PostFormValue("apoderado_telefono"),
	}

	var foto *Foto
	archivo, _, err := peticion.FormFile("foto")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			return datos, nil, nil
		}
		return datos, nil, err
	}
	defer archivo.Close()

	leida, err := LeerFoto(archivo, m.fotoMaximaMB)
	if err != nil {
		return datos, nil, err
	}
	foto = &leida
	return datos, foto, nil
}

func (m *Manejador) guardarFoto(ctx context.Context, usuarioID, alumnoID int64, foto *Foto) error {
	variante := strconv.FormatInt(time.Now().UnixNano(), 10)
	clave := ClaveFoto(alumnoID, variante, foto.Extension)
	if err := m.almacen.Guardar(ctx, clave, bytes.NewReader(foto.Contenido)); err != nil {
		return err
	}
	if err := m.alumnos.ActualizarFoto(ctx, usuarioID, alumnoID, clave); err != nil {
		_ = m.almacen.Eliminar(ctx, clave)
		return err
	}
	return nil
}

func (m *Manejador) renderizarFormulario(respuesta http.ResponseWriter, peticion *http.Request, estado int, datos DatosAlumno, errores map[string]string, mensaje string) {
	m.renderizarFormularioConID(respuesta, peticion, estado, 0, datos, errores, mensaje)
}

func (m *Manejador) renderizarFormularioConID(respuesta http.ResponseWriter, peticion *http.Request, estado int, id int64, datos DatosAlumno, errores map[string]string, mensaje string) {
	secciones, err := m.secciones.ListarSecciones(peticion.Context())
	if err != nil {
		m.errorInterno(respuesta, "listando secciones", err)
		return
	}
	formulario := vista.DatosFormulario{
		EsEdicion:         id > 0,
		ID:                id,
		Identificador:     datos.Identificador,
		Nombres:           datos.Nombres,
		Apellidos:         datos.Apellidos,
		SeccionID:         datos.SeccionID,
		ApoderadoNombre:   datos.ApoderadoNombre,
		ApoderadoTelefono: datos.ApoderadoTelefono,
		Errores:           errores,
	}
	m.renderizar(respuesta, peticion, estado, vista.Formulario(
		m.nombreUsuario(peticion.Context()),
		intermediarios.RolUsuario(peticion.Context()),
		m.tokenCSRF(peticion.Context()),
		secciones,
		formulario,
		mensaje,
	))
}

func (m *Manejador) renderizar(respuesta http.ResponseWriter, peticion *http.Request, estado int, componente interface {
	Render(context.Context, io.Writer) error
}) {
	respuesta.Header().Set("Content-Type", "text/html; charset=utf-8")
	respuesta.WriteHeader(estado)
	if err := componente.Render(peticion.Context(), respuesta); err != nil {
		m.registro.Error("no se pudo renderizar la plantilla", "error", err)
	}
}

func (m *Manejador) tokenCSRF(ctx context.Context) string {
	token, err := intermediarios.TokenCSRF(ctx, m.sesiones)
	if err != nil {
		m.registro.Error("no se pudo preparar el token CSRF", "error", err)
	}
	return token
}

func (m *Manejador) nombreUsuario(ctx context.Context) string {
	nombre := intermediarios.NombreUsuario(ctx, m.sesiones)
	if nombre == "" {
		return "Usuario"
	}
	return nombre
}

func (m *Manejador) redirigirConError(respuesta http.ResponseWriter, peticion *http.Request, id int64, codigo string) {
	http.Redirect(respuesta, peticion, fmt.Sprintf("/alumnos/%d?error=%s", id, codigo), http.StatusSeeOther)
}

func (m *Manejador) errorInterno(respuesta http.ResponseWriter, accion string, err error) {
	m.registro.Error(accion, "error", err)
	http.Error(respuesta, "Ocurrió un error inesperado. Intente de nuevo.", http.StatusInternalServerError)
}

func aAlumnoVista(alumno Alumno) vista.Alumno {
	fechaSalida := ""
	if alumno.FechaSalida != nil {
		fechaSalida = fechas.FormatearDia(*alumno.FechaSalida)
	}
	return vista.Alumno{
		ID:                alumno.ID,
		Identificador:     alumno.Identificador,
		Nombres:           alumno.Nombres,
		Apellidos:         alumno.Apellidos,
		Estado:            alumno.Estado,
		Grado:             alumno.Grado,
		Seccion:           alumno.Seccion,
		Turno:             alumno.Turno,
		AnioEscolar:       alumno.AnioEscolar,
		FotoClave:         alumno.FotoClave,
		ApoderadoNombre:   alumno.ApoderadoNombre,
		ApoderadoTelefono: alumno.ApoderadoTelefono,
		FechaSalida:       fechaSalida,
	}
}

func aListadoVista(listado []AlumnoListado) []vista.AlumnoListado {
	filas := make([]vista.AlumnoListado, 0, len(listado))
	for _, alumno := range listado {
		filas = append(filas, vista.AlumnoListado{
			ID:            alumno.ID,
			Identificador: alumno.Identificador,
			Nombres:       alumno.Nombres,
			Apellidos:     alumno.Apellidos,
			Estado:        alumno.Estado,
			Grado:         alumno.Grado,
			Seccion:       alumno.Seccion,
			AnioEscolar:   alumno.AnioEscolar,
			TieneFoto:     alumno.TieneFoto,
		})
	}
	return filas
}

func aCredencialVista(credencial *CredencialQR) *vista.Credencial {
	if credencial == nil {
		return nil
	}
	return &vista.Credencial{
		Estado:    credencial.Estado,
		EmitidaEn: fechas.Formatear(credencial.EmitidaEn),
	}
}

func aHistorialVista(historial []CredencialQR) []vista.Credencial {
	filas := make([]vista.Credencial, 0, len(historial))
	for _, credencial := range historial {
		revocada := ""
		if credencial.RevocadaEn != nil {
			revocada = fechas.Formatear(*credencial.RevocadaEn)
		}
		filas = append(filas, vista.Credencial{
			Estado:     credencial.Estado,
			EmitidaEn:  fechas.Formatear(credencial.EmitidaEn),
			RevocadaEn: revocada,
			Motivo:     credencial.Motivo,
		})
	}
	return filas
}

func idDeRuta(peticion *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(peticion, "id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func enteroDeTexto(texto string) int64 {
	numero, err := strconv.ParseInt(texto, 10, 64)
	if err != nil || numero <= 0 {
		return 0
	}
	return numero
}

func estadoValido(estado string) bool {
	switch estado {
	case "", EstadoActivo, EstadoDesactivado, EstadoRetirado, EstadoTrasladado:
		return true
	default:
		return false
	}
}

func mensajeDeErrorFoto(err error) string {
	switch {
	case err == nil:
		return ""
	case errors.Is(err, ErrFotoMuyGrande):
		return "La foto supera el tamaño máximo permitido. El alumno se guardó sin la foto."
	case errors.Is(err, ErrFotoInvalida):
		return "La foto debe ser un archivo JPG o PNG. El alumno se guardó sin la foto."
	default:
		return "No se pudo leer la foto. El alumno se guardó sin la foto."
	}
}

func codigoDeErrorFoto(err error) string {
	switch {
	case errors.Is(err, ErrFotoMuyGrande):
		return "foto_muy_grande"
	case errors.Is(err, ErrFotoInvalida):
		return "foto_invalida"
	default:
		return "foto_invalida"
	}
}

func mensajeCorrecto(codigo string) string {
	switch codigo {
	case "creado":
		return "El alumno se registró correctamente."
	case "actualizado":
		return "Los datos del alumno se actualizaron."
	case "desactivado":
		return "El alumno quedó desactivado y su QR dejó de funcionar."
	case "foto_actualizada":
		return "La foto se actualizó."
	case "qr_emitido":
		return "El QR se generó correctamente."
	case "qr_reemplazado":
		return "El QR anterior quedó invalidado y se generó uno nuevo."
	default:
		return ""
	}
}

func mensajeError(codigo string) string {
	switch codigo {
	case "foto_invalida":
		return "La foto debe ser un archivo JPG o PNG válido."
	case "foto_muy_grande":
		return "La foto supera el tamaño máximo permitido."
	case "foto_no_guardada":
		return "El alumno se guardó, pero la foto no se pudo almacenar."
	case "qr_ya_activo":
		return "El alumno ya tiene un QR vigente."
	case "qr_sin_activo":
		return "El alumno no tiene un QR vigente que reemplazar."
	case "motivo_invalido":
		return "Debe indicar un motivo válido para reemplazar el QR."
	case "alumno_no_activo":
		return "El alumno no está activo, así que no puede tener un QR."
	default:
		return ""
	}
}
