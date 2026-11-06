package integracion

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"asistencia/internal/autenticacion"
	"asistencia/internal/plataforma/intermediarios"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const claveLoginPrueba = "clave-segura-2026"

func TestLogin_RedirigeAlInicioDeCadaRol(t *testing.T) {
	pool, _ := AbrirBaseDePruebasConURL(t)
	crearCuentasDeLogin(t, pool)
	enrutador := nuevoEnrutadorLogin(pool)

	casos := []struct {
		correo  string
		destino string
	}{
		{"admin@login.pe", "/alumnos"},
		{"director@login.pe", "/alumnos"},
		{"docente@login.pe", "/asistencia"},
		{"porteria@login.pe", "/porteria"},
	}

	for _, caso := range casos {
		cliente := nuevoClienteLogin(t, enrutador)
		respuesta := cliente.ingresar(t, caso.correo, claveLoginPrueba)
		assert.Equal(t, http.StatusSeeOther, respuesta.StatusCode, caso.correo)
		assert.Equal(t, caso.destino, respuesta.Header.Get("Location"), caso.correo)
	}
}

func TestLogin_AdminSalirDocenteLlegaAlInicioDocente(t *testing.T) {
	pool, _ := AbrirBaseDePruebasConURL(t)
	crearCuentasDeLogin(t, pool)
	cliente := nuevoClienteLogin(t, nuevoEnrutadorLogin(pool))

	respuesta := cliente.ingresar(t, "admin@login.pe", claveLoginPrueba)
	require.Equal(t, "/alumnos", respuesta.Header.Get("Location"))

	respuesta = cliente.solicitar(t, http.MethodPost, "/salir", nil)
	require.Equal(t, "/ingreso", respuesta.Header.Get("Location"))

	respuesta = cliente.ingresar(t, "docente@login.pe", claveLoginPrueba)
	assert.Equal(t, "/asistencia", respuesta.Header.Get("Location"))
}

func TestLogin_IgnoraUrlDeRetornoDeOtroRol(t *testing.T) {
	pool, _ := AbrirBaseDePruebasConURL(t)
	crearCuentasDeLogin(t, pool)
	enrutador := nuevoEnrutadorLogin(pool)

	cliente := nuevoClienteLogin(t, enrutador)
	respuesta := cliente.solicitar(t, http.MethodGet, "/alumnos", nil)
	require.Equal(t, http.StatusSeeOther, respuesta.StatusCode)
	require.Equal(t, "/ingreso", respuesta.Header.Get("Location"))

	respuesta = cliente.ingresar(t, "docente@login.pe", claveLoginPrueba)
	assert.Equal(t, "/asistencia", respuesta.Header.Get("Location"),
		"la URL de administrador no debe usarse para una docente")

	cliente = nuevoClienteLogin(t, enrutador)
	respuesta = cliente.solicitar(t, http.MethodGet, "/alumnos", nil)
	require.Equal(t, "/ingreso", respuesta.Header.Get("Location"))
	respuesta = cliente.ingresar(t, "admin@login.pe", claveLoginPrueba)
	assert.Equal(t, "/alumnos", respuesta.Header.Get("Location"))
}

func TestLogin_RolSinPermisoVePaginaConSuInicio(t *testing.T) {
	pool, _ := AbrirBaseDePruebasConURL(t)
	crearCuentasDeLogin(t, pool)
	cliente := nuevoClienteLogin(t, nuevoEnrutadorLogin(pool))

	require.Equal(t, "/asistencia", cliente.ingresar(t, "docente@login.pe", claveLoginPrueba).Header.Get("Location"))
	respuesta := cliente.solicitar(t, http.MethodGet, "/alumnos", nil)

	assert.Equal(t, http.StatusForbidden, respuesta.StatusCode)
	cuerpo, err := io.ReadAll(respuesta.Body)
	require.NoError(t, err)
	assert.Contains(t, string(cuerpo), "Acceso no permitido")
	assert.Contains(t, string(cuerpo), `href="/asistencia"`)
}

func TestLogin_RegeneraLaSesion(t *testing.T) {
	pool, _ := AbrirBaseDePruebasConURL(t)
	crearCuentasDeLogin(t, pool)
	cliente := nuevoClienteLogin(t, nuevoEnrutadorLogin(pool))

	respuesta := cliente.solicitar(t, http.MethodGet, "/ingreso", nil)
	require.Equal(t, http.StatusOK, respuesta.StatusCode)
	antes := cliente.cookieSesion()
	require.NotEmpty(t, antes)

	respuesta = cliente.ingresar(t, "docente@login.pe", claveLoginPrueba)
	require.Equal(t, http.StatusSeeOther, respuesta.StatusCode)
	despues := cliente.cookieSesion()
	assert.NotEqual(t, antes, despues, "el token de sesión debe regenerarse al ingresar")
}

type clienteLogin struct {
	t        *testing.T
	servidor *httptest.Server
	cliente  *http.Client
}

func nuevoClienteLogin(t *testing.T, enrutador http.Handler) *clienteLogin {
	t.Helper()
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	servidor := httptest.NewServer(enrutador)
	t.Cleanup(servidor.Close)
	return &clienteLogin{t: t, servidor: servidor, cliente: &http.Client{Jar: jar}}
}

func (c *clienteLogin) ingresar(t *testing.T, correo, clave string) *http.Response {
	t.Helper()
	formulario := url.Values{"correo": {correo}, "contrasenia": {clave}}
	return c.solicitar(t, http.MethodPost, "/ingreso", strings.NewReader(formulario.Encode()))
}

func (c *clienteLogin) solicitar(t *testing.T, metodo, ruta string, cuerpo io.Reader) *http.Response {
	t.Helper()
	peticion, err := http.NewRequest(metodo, c.servidor.URL+ruta, cuerpo)
	require.NoError(t, err)
	if cuerpo != nil {
		peticion.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	cliente := &http.Client{
		Jar: c.cliente.Jar,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	respuesta, err := cliente.Do(peticion)
	require.NoError(t, err)
	c.cliente = cliente
	return respuesta
}

func (c *clienteLogin) cookieSesion() string {
	base, _ := url.Parse(c.servidor.URL)
	for _, cookie := range c.cliente.Jar.Cookies(base) {
		if cookie.Name == "sesion_asistencia" {
			return cookie.Value
		}
	}
	return ""
}

func nuevoEnrutadorLogin(pool *pgxpool.Pool) http.Handler {
	sesiones := scs.New()
	sesiones.Lifetime = time.Hour
	sesiones.Cookie.Name = "sesion_asistencia"
	servicio := autenticacion.NuevoServicio(pool)
	limite := intermediarios.NuevoLimiteIntentos(5, 15*time.Minute, time.Now)
	registro := slog.New(slog.NewTextHandler(io.Discard, nil))
	manejador := autenticacion.NuevoManejador(servicio, sesiones, limite, registro)

	enrutador := chi.NewRouter()
	enrutador.Use(sesiones.LoadAndSave)
	enrutador.Group(func(protegidas chi.Router) {
		protegidas.Use(intermediarios.RequerirRol(sesiones,
			intermediarios.RolAdministrador, intermediarios.RolDirector))
		protegidas.Get("/alumnos", func(respuesta http.ResponseWriter, _ *http.Request) {
			respuesta.WriteHeader(http.StatusOK)
		})
	})
	manejador.Montar(enrutador)
	return enrutador
}

func crearCuentasDeLogin(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	hash, err := autenticacion.HashearContrasenia(claveLoginPrueba)
	require.NoError(t, err)
	cuentas := []struct{ correo, rol string }{
		{"admin@login.pe", "administrador"},
		{"director@login.pe", "director"},
		{"docente@login.pe", "docente"},
		{"porteria@login.pe", "porteria"},
	}
	for _, cuenta := range cuentas {
		_, err := pool.Exec(t.Context(),
			`INSERT INTO usuarios (correo, contrasenia_hash, nombre_completo, rol)
			 VALUES ($1, $2, $3, $4)`,
			cuenta.correo, hash, cuenta.correo, cuenta.rol)
		require.NoError(t, err)
	}
}
