// Comando servidor arranca la aplicación web o ejecuta tareas de preparación.
//
// Uso:
//
//	servidor              arranca el servidor HTTP
//	servidor migrar       aplica las migraciones pendientes
//	servidor crear-admin  crea el administrador inicial (variables ADMIN_*)
//	servidor sembrar      crea el administrador y datos académicos de demostración
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"asistencia/internal/academico"
	"asistencia/internal/alumnos"
	"asistencia/internal/asistencia"
	"asistencia/internal/autenticacion"
	"asistencia/internal/plataforma/almacenamiento"
	"asistencia/internal/plataforma/basedatos"
	"asistencia/internal/plataforma/configuracion"
	"asistencia/internal/plataforma/intermediarios"
	"asistencia/internal/plataforma/registro"
	"asistencia/internal/plataforma/rutas"
	"asistencia/internal/plataforma/semilla"
	"asistencia/migraciones"
)

const (
	duracionSesion           = 8 * time.Hour
	maximoIntentos           = 5
	ventanaIntentos          = 15 * time.Minute
	tiempoDeApagado          = 10 * time.Second
	tiempoLecturaEncabezados = 10 * time.Second
)

func main() {
	if err := ejecutar(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func ejecutar(argumentos []string) error {
	cfg, err := configuracion.CargarDesdeEntorno()
	if err != nil {
		return fmt.Errorf("configuración no válida: %w", err)
	}

	comando := ""
	if len(argumentos) > 0 {
		comando = argumentos[0]
	}

	switch comando {
	case "migrar":
		return migraciones.Subir(cfg.BaseDatosURL)
	case "crear-admin":
		return crearAdministrador(cfg)
	case "sembrar":
		return sembrar(cfg)
	case "":
		return servir(cfg)
	default:
		return fmt.Errorf("comando desconocido %q; use migrar, crear-admin o sembrar", comando)
	}
}

func crearAdministrador(cfg configuracion.Configuracion) error {
	opciones, err := opcionesAdministrador()
	if err != nil {
		return err
	}

	ctx := context.Background()
	pool, err := basedatos.Abrir(ctx, cfg.BaseDatosURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	id, creado, err := semilla.CrearAdministrador(ctx, pool, opciones)
	if err != nil {
		return err
	}
	if creado {
		fmt.Printf("Administrador %s creado con el identificador %d.\n", opciones.Correo, id)
	} else {
		fmt.Printf("El administrador %s ya existía; no se modificó.\n", opciones.Correo)
	}
	return nil
}

func sembrar(cfg configuracion.Configuracion) error {
	opciones, err := opcionesAdministrador()
	if err != nil {
		return err
	}

	ctx := context.Background()
	pool, err := basedatos.Abrir(ctx, cfg.BaseDatosURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	if _, _, err := semilla.CrearAdministrador(ctx, pool, opciones); err != nil {
		return err
	}

	anio := time.Now().Year()
	if valor := strings.TrimSpace(os.Getenv("ANIO_NOMBRE")); valor != "" {
		anio = anioDelNombre(valor, anio)
	}
	err = semilla.CrearAcademicoDemo(ctx, pool, semilla.OpcionesAcademicas{
		Nombre:      strconv.Itoa(anio),
		FechaInicio: time.Date(anio, time.March, 1, 0, 0, 0, 0, time.UTC),
		FechaFin:    time.Date(anio, time.December, 31, 0, 0, 0, 0, time.UTC),
		Grados:      []string{"1.°", "2.°", "3.°", "4.°", "5.°"},
		Secciones:   []string{"A", "B"},
		Turno:       "manana",
	})
	if err != nil {
		return err
	}

	fmt.Println("Semilla aplicada: administrador y secciones de demostración listos.")
	fmt.Println("Ingrese en /ingreso con el correo y la contraseña configurados en ADMIN_*.")
	return nil
}

func opcionesAdministrador() (semilla.OpcionesAdministrador, error) {
	correo := strings.ToLower(strings.TrimSpace(os.Getenv("ADMIN_CORREO")))
	contrasenia := os.Getenv("ADMIN_CONTRASENIA")
	nombre := strings.TrimSpace(os.Getenv("ADMIN_NOMBRE"))
	if nombre == "" {
		nombre = "Administrador"
	}

	if correo == "" {
		return semilla.OpcionesAdministrador{}, errors.New("falta la variable ADMIN_CORREO")
	}
	if len(contrasenia) < 8 {
		return semilla.OpcionesAdministrador{}, errors.New("ADMIN_CONTRASENIA debe tener al menos 8 caracteres")
	}
	return semilla.OpcionesAdministrador{Correo: correo, Contrasenia: contrasenia, Nombre: nombre}, nil
}

func servir(cfg configuracion.Configuracion) error {
	registrador := registro.Nuevo(os.Stdout)
	ctx, detener := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer detener()

	pool, err := basedatos.Abrir(ctx, cfg.BaseDatosURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	sesiones := intermediarios.NuevaSesion(intermediarios.NuevoAlmacenSesion(pool), duracionSesion)
	limite := intermediarios.NuevoLimiteIntentos(maximoIntentos, ventanaIntentos, time.Now)
	ingreso := autenticacion.NuevoManejador(autenticacion.NuevoServicio(pool), sesiones, limite, registrador)

	servicioAcademico := academico.NuevoServicio(pool)
	servicioConfiguracion := academico.NuevoServicioConfiguracion(pool)
	manejadorConfiguracion := academico.NuevoManejador(academico.DependenciasConfiguracion{
		Configuracion: servicioConfiguracion,
		Sesiones:      sesiones,
		Registro:      registrador,
	})
	manejadorAsistencia := asistencia.NuevoManejador(asistencia.Dependencias{
		Sesiones: sesiones,
		Registro: registrador,
	})
	servicioAlumnos := alumnos.NuevoServicio(pool)
	servicioImportacion := alumnos.NuevoServicioImportacion(pool, servicioAcademico, cfg.ImportacionMaxFilas)
	manejadorAlumnos := alumnos.NuevoManejador(alumnos.Dependencias{
		Alumnos:             servicioAlumnos,
		QR:                  servicioAlumnos,
		Importacion:         servicioImportacion,
		Secciones:           servicioAcademico,
		Almacen:             abrirAlmacen(cfg.Almacenamiento),
		Sesiones:            sesiones,
		Registro:            registrador,
		FotoMaximaMB:        cfg.FotoTamanioMaximoMB,
		CarnetAnchoMM:       cfg.CarnetAnchoMM,
		CarnetAltoMM:        cfg.CarnetAltoMM,
		ImportacionMaxFilas: cfg.ImportacionMaxFilas,
	})

	servidor := &http.Server{
		Addr: fmt.Sprintf(":%d", cfg.Puerto),
		Handler: rutas.NuevoEnrutador(rutas.Dependencias{
			Sesiones:      sesiones,
			Verificador:   pool,
			Autenticacion: ingreso,
			Alumnos:       manejadorAlumnos,
			Academico:     manejadorConfiguracion,
			Asistencia:    manejadorAsistencia,
		}),
		ReadHeaderTimeout: tiempoLecturaEncabezados,
	}

	errorServidor := make(chan error, 1)
	go func() {
		registrador.Info("servidor iniciado", "puerto", cfg.Puerto)
		errorServidor <- servidor.ListenAndServe()
	}()

	select {
	case err := <-errorServidor:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
	}

	registrador.Info("apagando el servidor")
	ctxApagado, cancelar := context.WithTimeout(context.Background(), tiempoDeApagado)
	defer cancelar()
	return servidor.Shutdown(ctxApagado)
}

func abrirAlmacen(cfg configuracion.Almacenamiento) almacenamiento.AlmacenFotos {
	if cfg.Tipo == "s3" {
		return almacenamiento.NuevoAlmacenS3(cfg)
	}
	return almacenamiento.NuevoAlmacenDisco(cfg.Ruta)
}

func anioDelNombre(nombre string, porDefecto int) int {
	if numero, err := strconv.Atoi(nombre); err == nil && numero > 0 {
		return numero
	}
	return porDefecto
}
