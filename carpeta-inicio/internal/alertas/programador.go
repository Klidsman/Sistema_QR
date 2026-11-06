package alertas

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

// Procesador son los procesos periódicos que el programador ejecuta.
type Procesador interface {
	ProcesarTardanzas(ctx context.Context) error
	ProcesarCierre(ctx context.Context, fecha time.Time, turno string) error
}

// OpcionesProgramador configuran los horarios de cierre y la zona horaria.
type OpcionesProgramador struct {
	Zona             *time.Location
	HoraCierreManana string
	HoraCierreTarde  string
}

// Programador ejecuta los procesos de alertas con robfig/cron.
type Programador struct {
	procesador Procesador
	registro   *slog.Logger
	opciones   OpcionesProgramador
}

// NuevoProgramador crea el programador con valores por defecto razonables.
func NuevoProgramador(procesador Procesador, registro *slog.Logger, opciones OpcionesProgramador) *Programador {
	if opciones.Zona == nil {
		opciones.Zona = time.Local
	}
	if opciones.HoraCierreManana == "" {
		opciones.HoraCierreManana = "13:00"
	}
	if opciones.HoraCierreTarde == "" {
		opciones.HoraCierreTarde = "19:00"
	}
	if registro == nil {
		registro = slog.Default()
	}
	return &Programador{procesador: procesador, registro: registro, opciones: opciones}
}

// Iniciar arranca los procesos y los detiene al cancelar el contexto.
func (p *Programador) Iniciar(ctx context.Context) error {
	manana, err := expresionCierre(p.opciones.HoraCierreManana)
	if err != nil {
		return fmt.Errorf("hora de cierre de la mañana no válida: %w", err)
	}
	tarde, err := expresionCierre(p.opciones.HoraCierreTarde)
	if err != nil {
		return fmt.Errorf("hora de cierre de la tarde no válida: %w", err)
	}

	programador := cron.New(cron.WithLocation(p.opciones.Zona))
	if _, err := programador.AddFunc("@every 1m", func() {
		if err := p.procesador.ProcesarTardanzas(ctx); err != nil {
			p.registro.Error("procesando tardanzas", "error", err)
		}
	}); err != nil {
		return err
	}
	if _, err := programador.AddFunc(manana.especificacion, func() {
		p.ejecutarCierre(ctx, "manana")
	}); err != nil {
		return err
	}
	if _, err := programador.AddFunc(tarde.especificacion, func() {
		p.ejecutarCierre(ctx, "tarde")
	}); err != nil {
		return err
	}

	programador.Start()
	go func() {
		<-ctx.Done()
		programador.Stop()
	}()
	p.registro.Info("programador de alertas iniciado",
		"tardanzas", "cada minuto",
		"cierre_manana", p.opciones.HoraCierreManana,
		"cierre_tarde", p.opciones.HoraCierreTarde,
	)
	return nil
}

func (p *Programador) ejecutarCierre(ctx context.Context, turno string) {
	fecha := time.Now().In(p.opciones.Zona)
	if err := p.procesador.ProcesarCierre(ctx, fecha, turno); err != nil {
		p.registro.Error("procesando el cierre de alertas", "turno", turno, "error", err)
	}
}

type horarioCierre struct {
	especificacion string
}

func expresionCierre(hora string) (horarioCierre, error) {
	partes := strings.Split(strings.TrimSpace(hora), ":")
	if len(partes) != 2 {
		return horarioCierre{}, fmt.Errorf("se esperaba HH:MM")
	}
	horaDelDia, err := strconv.Atoi(partes[0])
	if err != nil || horaDelDia < 0 || horaDelDia > 23 {
		return horarioCierre{}, fmt.Errorf("hora fuera de rango")
	}
	minuto, err := strconv.Atoi(partes[1])
	if err != nil || minuto < 0 || minuto > 59 {
		return horarioCierre{}, fmt.Errorf("minuto fuera de rango")
	}
	return horarioCierre{especificacion: fmt.Sprintf("%d %d * * *", minuto, horaDelDia)}, nil
}
