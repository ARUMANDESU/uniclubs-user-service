package main

import (
	"errors"
	"flag"
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var postgresURL, migrationsPath string

	flag.StringVar(&postgresURL, "postgres-url", "postgresql://username:password@localhost:5432/databaseName", "PostgreSQL database URL")
	flag.StringVar(&migrationsPath, "migration-path", "./migrations", "path to migrations")
	flag.Parse()

	validate(postgresURL, migrationsPath)

	m, err := migrate.New(
		"file://"+migrationsPath,
		postgresURL,
	)
	if err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")
			return
		}
		panic(err)
	}

	fmt.Println("migrations applied successfully")
}

func validate(storagePath, migrationPath string) {
	err := validation.Validate(storagePath, validation.Required)
	if err != nil {
		panic(err)
	}
	err = validation.Validate(migrationPath, validation.Required)
	if err != nil {
		panic(err)
	}

}
