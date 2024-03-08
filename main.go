package main

import "go.uber.org/zap"

const DATABASE_NAME = "database_name"
const MIGRATION_DATABASE_NAME = "migrations"

const MigrationsFolder = "scripts"

var host string = getEnv("POSTGRESQL_URL", "localhost")
var port string = getEnv("POSTGRESQL_PORT", "5432")
var dbname string = getEnv("POSTGRESQL_DBNAME", DATABASE_NAME)
var migrationDbName string = getEnv("POSTGRESQL_MIGRATION_DBNAME", MIGRATION_DATABASE_NAME)

func main() {
	logger := configureLogger()
	defer logger.Sync() // flushes buffer, if any

	mdb := openDatabase(host, port, migrationDbName)
	defer mdb.Close()
	zap.S().Info("Migration database connected")

	validateDatabase(mdb)

	db := openDatabase(host, port, dbname)
	defer db.Close()
	zap.S().Info("Postgresql connected")

	processMigrations(MigrationsFolder, mdb, db)

}

func configureLogger() *zap.Logger {
	logger, _ := zap.NewProduction()
	zap.ReplaceGlobals(logger)

	return logger
}
