package rdbms

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type postgresWrapper struct {
	*rdbms
}

func NewPostgres(cfg *Config) (RDBMS, error) {
	connString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database,
	)

	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("Error openning connection:\n%v", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("Error ping database:\n%v", err)
	}

	return &postgresWrapper{&rdbms{db: db}}, nil
}

func (db *postgresWrapper) MigrateUp(source string) error {
	migrator := func(m *migrate.Migrate) error { return m.Up() }
	return db.migrate(source, migrator)
}

func (db *postgresWrapper) MigrateDown(source string) error {
	migrator := func(m *migrate.Migrate) error { return m.Down() }
	return db.migrate(source, migrator)
}

func (db *postgresWrapper) migrate(source string, migrator func(*migrate.Migrate) error) error {
	instance, err := postgres.WithInstance(db.db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("Error creating migrate instance\n%v", err)
	}

	migration, err := migrate.NewWithDatabaseInstance(source, "postgres", instance)
	if err != nil {
		return fmt.Errorf("Error loading migration files\n%v", err)
	}

	if err := migrator(migration); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("Error doing migrations\n%v", err)
	}

	return nil
}
