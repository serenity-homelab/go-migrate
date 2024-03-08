package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type MigrationFile struct {
	Number        int
	Name          string
	FileName      string
	FilePath      string
	MigrationType string // Should be 'UP' or 'DOWN'
	Extension     string
}

func (m *MigrationFile) ToString() string {
	return fmt.Sprintf("%d | %s | %s | %s", m.Number, m.Name, m.MigrationType, m.Extension)
}

func processMigrations(dir string, mdb, db *sqlx.DB) {
	mFiles := getMigrationFiles(dir)

	if len(mFiles) == 0 {
		zap.S().Infof("No migration files found in %s", dir)
		return
	}

	migration := getDatabaseMigration(mdb, DATABASE_NAME)

	for _, m := range mFiles {

		if m.Number <= migration.State { // skip migrations already completed
			zap.S().Infof("SKIP    | %v", m.ToString())
			continue
		}

		script := getFile(m.FilePath)

		db.MustExec(script)

		migration.LastUpdated = time.Now()
		migration.State += 1
		updateDatabaseMigration(mdb, migration)
		zap.S().Infof("SUCCESS | %v", m.ToString())

	}

}

func getMigrationFiles(dir string) (migrations []MigrationFile) {
	files, err := os.ReadDir(dir)
	if err != nil {
		zap.S().Panic(err)
	}

	for _, file := range files {
		if !file.IsDir() {
			migration, err := parseFileName(file.Name())
			migration.FilePath = fmt.Sprintf("%v/%v", filepath.Clean(dir), migration.FileName)
			fmt.Printf("%v\n", migration.FilePath)

			if err != nil {
				zap.S().Error(err)
			} else {
				migrations = append(migrations, migration)
			}
		}
	}

	return migrations
}

func parseFileName(filename string) (MigrationFile, error) {
	var migration = MigrationFile{}

	migration.FileName = filename

	filenameParts := strings.Split(filename, ".")

	if len(filenameParts) != 3 {
		errorMsg := fmt.Sprintf("%s does not follow the format {version}_{title}.{up|down}.{extension}", filename)
		return migration, errors.New(errorMsg)
	}

	prefix := filenameParts[0]
	migrationType := filenameParts[1]
	extension := filenameParts[2]

	migration.Extension = extension

	var err error
	migration.Number, migration.Name, err = getNumberFromName(prefix)
	if err != nil {
		return migration, err
	}
	migration.MigrationType, err = getMigrationType(migrationType)
	if err != nil {
		return migration, err
	}

	return migration, err
}

func getMigrationType(name string) (string, error) {
	migrationType := strings.ToLower(name)

	if migrationType != "up" && migrationType != "down" {
		errorMsg := fmt.Sprintf("%s is not a valid type. Needs to be 'up' or 'down'", migrationType)
		return migrationType, errors.New(errorMsg)
	}

	return migrationType, nil
}

func getNumberFromName(name string) (int, string, error) {
	split := strings.SplitN(name, "_", 2)
	if len(split) != 2 {
		errorMsg := fmt.Sprintf("%s is not a valid file name", name)
		return 0, name, errors.New(errorMsg)
	}

	numberStr := split[0]
	filename := split[1]

	num, err := strconv.Atoi(numberStr)

	return num, filename, err
}

func getFile(filepath string) string {
	bytes, err := os.ReadFile(filepath)
	if err != nil {
		zap.S().Panicf("failed to read file %v", filepath)
	}

	return string(bytes)
}
