package manejadores

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"asistencia/internal/plataforma/intermediarios"

	"github.com/alexedwards/scs/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sesionConTokenCSRF(t *testing.T) (*scs.SessionManager, context.Context) {
	t.Helper()
	sesiones := scs.New()
	ctx, err := sesiones.Load(context.Background(), "token-de-prueba")
	require.NoError(t, err)
	sesiones.Put(ctx, "token_csrf", "token-csrf-de-prueba")
	return sesiones, ctx
}

func formularioMultipartConToken(token string) (*bytes.Buffer, string) {
	var cuerpo bytes.Buffer
	escritor := multipart.NewWriter(&cuerpo)
	_ = escritor.WriteField("token_csrf", token)
	_ = escritor.WriteField("nombre", "valor")
	_ = escritor.Close()
	return &cuerpo, escritor.FormDataContentType()
}

func TestProteccionCSRF_AceptaTokenEnFormularioMultipart(t *testing.T) {
	sesiones, ctx := sesionConTokenCSRF(t)
	llamado := false
	manejador := intermediarios.ProteccionCSRF(sesiones)(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		llamado = true
		rw.WriteHeader(http.StatusOK)
	}))

	cuerpo, tipo := formularioMultipartConToken("token-csrf-de-prueba")
	peticion := httptest.NewRequest(http.MethodPost, "/alumnos", cuerpo).WithContext(ctx)
	peticion.Header.Set("Content-Type", tipo)
	respuesta := httptest.NewRecorder()

	manejador.ServeHTTP(respuesta, peticion)

	assert.True(t, llamado)
	assert.Equal(t, http.StatusOK, respuesta.Code)
}

func TestSesion_CookiesSonSeguras(t *testing.T) {
	sesiones := intermediarios.NuevaSesion(nil, 8*time.Hour)

	assert.Equal(t, "sesion_asistencia", sesiones.Cookie.Name)
	assert.True(t, sesiones.Cookie.HttpOnly)
	assert.True(t, sesiones.Cookie.Secure)
	assert.Equal(t, http.SameSiteStrictMode, sesiones.Cookie.SameSite)
	assert.Equal(t, 8*time.Hour, sesiones.Lifetime)
}

func TestProteccionCSRF_RechazaSolicitudSinToken(t *testing.T) {
	sesiones, ctx := sesionConTokenCSRF(t)
	manejador := intermediarios.ProteccionCSRF(sesiones)(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))

	peticion := httptest.NewRequest(http.MethodPost, "/alumnos", bytes.NewBufferString("campo=valor")).
		WithContext(ctx)
	peticion.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	respuesta := httptest.NewRecorder()

	manejador.ServeHTTP(respuesta, peticion)

	assert.Equal(t, http.StatusForbidden, respuesta.Code)
}

func TestProteccionCSRF_RechazaTokenIncorrecto(t *testing.T) {
	sesiones, ctx := sesionConTokenCSRF(t)
	manejador := intermediarios.ProteccionCSRF(sesiones)(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))

	cuerpo, tipo := formularioMultipartConToken("token-equivocado")
	peticion := httptest.NewRequest(http.MethodPost, "/alumnos", cuerpo).WithContext(ctx)
	peticion.Header.Set("Content-Type", tipo)
	respuesta := httptest.NewRecorder()

	manejador.ServeHTTP(respuesta, peticion)

	assert.Equal(t, http.StatusForbidden, respuesta.Code)
}
