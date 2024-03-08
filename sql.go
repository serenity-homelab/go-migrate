package main

import (
	"fmt"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/serenity-homelab/vault-postgres-driver"
	"go.uber.org/zap"
)

const driverName = "vault-postgres-driver"

const createMigrationTableQuery = `
CREATE TABLE IF NOT EXISTS migrations (
	database_name TEXT NOT NULL,
	state INT NOT NULL DEFAULT 0,
	last_updated TIMESTAMP WITH TIME ZONE NOT NULL
);
`
const createDatabaseMigrationQuery = `INSERT INTO migrations (database_name, state, last_updated) VALUES ($1, 0, NOW());`

const getDatabaseMigrationQuery = `SELECT database_name, state, last_updated FROM migrations WHERE database_name = $1;`

const updateDatabaseMigrationQuery = `
	UPDATE migrations SET
	state = :state,
	last_updated = :last_updated
	WHERE database_name = :database_name;
`

type Migration struct {
	DatabaseName string    `db:"database_name"`
	State        int       `db:"state"`
	LastUpdated  time.Time `db:"last_updated"`
}

func (m *Migration) ToString() string {
	return fmt.Sprintf("%v | %d", m.DatabaseName, m.State)
}

func openDatabase(host, port, dbname string) *sqlx.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=$1 password=$2 dbname=%s sslmode=disable",
		host, port, dbname)

	sqlx.BindDriver(driverName, sqlx.DOLLAR)
	var err error
	db, err := sqlx.Connect(driverName, psqlInfo)

	if err != nil {
		zap.S().Panic(err)
	}

	return db
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func createIfNotExistsMigrationTable(db *sqlx.DB) {
	db.MustExec(createMigrationTableQuery)
}

func getDatabaseMigration(db *sqlx.DB, databaseName string) Migration {
	migration := Migration{}
	err := db.Get(&migration, getDatabaseMigrationQuery, databaseName)

	if err != nil { // create if not exists
		createDatabaseMigration(db, databaseName)
		migration.DatabaseName = databaseName
		migration.State = 0
		migration.LastUpdated = time.Now()
	}

	return migration
}

func updateDatabaseMigration(db *sqlx.DB, migration Migration) {
	db.NamedExec(updateDatabaseMigrationQuery, migration)
}

func createDatabaseMigration(db *sqlx.DB, databaseName string) {
	db.MustExec(createDatabaseMigrationQuery, databaseName)
}

func validateDatabase(db *sqlx.DB) {
	createIfNotExistsMigrationTable(db)
	zap.S().Info("validated migration database")
}
