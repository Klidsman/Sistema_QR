// Package migraciones aplica las migraciones SQL versionadas con goose.
package migraciones

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"

	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed *.sql
var archivos embed.FS

// Subir aplica todas las migraciones pendientes sobre la base de datos indicada.
func Subir(urlBaseDatos string) error {
	db, err := sql.Open("pgx", urlBaseDatos)
	if err != nil {
		return fmt.Errorf("abriendo la base de datos para migrar: %w", err)
	}
	defer func() { _ = db.Close() }()

	goose.SetBaseFS(archivos)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("configurando el dialecto de migraciones: %w", err)
	}
	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("aplicando migraciones: %w", err)
	}
	return nil
}

// Bajar revierte la última migración aplicada. Se usa en pruebas y en reversiones controladas.
func Bajar(urlBaseDatos string) error {
	db, err := sql.Open("pgx", urlBaseDatos)
	if err != nil {
		return fmt.Errorf("abriendo la base de datos para revertir: %w", err)
	}
	defer func() { _ = db.Close() }()

	goose.SetBaseFS(archivos)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("configurando el dialecto de migraciones: %w", err)
	}
	if err := goose.Down(db, "."); err != nil {
		return fmt.Errorf("revirtiendo migraciones: %w", err)
	}
	return nil
}
