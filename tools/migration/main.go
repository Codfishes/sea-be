package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	var (
		command     = flag.String("command", "", "Migration command: up, down, force, version, create, drop")
		version     = flag.Int("version", -1, "Migration version")
		name        = flag.String("name", "", "Migration name for create command")
		steps       = flag.Int("steps", 1, "Number of migration steps")
		migrateDir  = flag.String("path", "./database/migrations", "Migration files directory")
		databaseURL = flag.String("database", "", "Database URL")
	)
	flag.Parse()

	if *databaseURL == "" {
		*databaseURL = os.Getenv("DATABASE_URL")
		if *databaseURL == "" {
			log.Fatal("Database URL is required. Set DATABASE_URL environment variable or use -database flag")
		}
	}

	switch *command {
	case "up":
		migrateUp(*databaseURL, *migrateDir)
	case "down":
		migrateDown(*databaseURL, *migrateDir, *steps)
	case "force":
		if *version == -1 {
			log.Fatal("Version is required for force command")
		}
		migrateForce(*databaseURL, *migrateDir, *version)
	case "version":
		showVersion(*databaseURL, *migrateDir)
	case "create":
		if *name == "" {
			log.Fatal("Name is required for create command")
		}
		createMigration(*migrateDir, *name)
	case "drop":
		migrateDrop(*databaseURL, *migrateDir)
	default:
		fmt.Println("Available commands: up, down, force, version, create, drop")
		flag.Usage()
		os.Exit(1)
	}
}

func getMigrator(databaseURL, migrateDir string) (*migrate.Migrate, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrateDir),
		"postgres",
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}

	return m, nil
}

func migrateUp(databaseURL, migrateDir string) {
	m, err := getMigrator(databaseURL, migrateDir)
	if err != nil {
		log.Fatal(err)
	}
	defer m.Close()

	fmt.Println("Running migrations up...")
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}
	fmt.Println("✓ Migrations applied successfully")
}

func migrateDown(databaseURL, migrateDir string, steps int) {
	m, err := getMigrator(databaseURL, migrateDir)
	if err != nil {
		log.Fatal(err)
	}
	defer m.Close()

	fmt.Printf("Running %d migration(s) down...\n", steps)
	if err := m.Steps(-steps); err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}
	fmt.Printf("✓ %d migration(s) rolled back successfully\n", steps)
}

func migrateForce(databaseURL, migrateDir string, version int) {
	m, err := getMigrator(databaseURL, migrateDir)
	if err != nil {
		log.Fatal(err)
	}
	defer m.Close()

	fmt.Printf("Forcing migration to version %d...\n", version)
	if err := m.Force(version); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ Migration forced to version %d\n", version)
}

func showVersion(databaseURL, migrateDir string) {
	m, err := getMigrator(databaseURL, migrateDir)
	if err != nil {
		log.Fatal(err)
	}
	defer m.Close()

	version, dirty, err := m.Version()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Current migration version: %d\n", version)
	if dirty {
		fmt.Println("⚠️  Database is in dirty state")
	} else {
		fmt.Println("✓ Database is clean")
	}
}

func createMigration(migrateDir, name string) {

	if err := os.MkdirAll(migrateDir, 0755); err != nil {
		log.Fatal(err)
	}

	version := getNextVersion(migrateDir)

	upFile := filepath.Join(migrateDir, fmt.Sprintf("%06d_%s.up.sql", version, name))
	downFile := filepath.Join(migrateDir, fmt.Sprintf("%06d_%s.down.sql", version, name))

	upContent := fmt.Sprintf(`-- Migration: %s
-- Created: %s
-- Description: TODO: Add description

-- Add your up migration here
`, name, time.Now().Format("2006-01-02 15:04:05"))

	if err := os.WriteFile(upFile, []byte(upContent), 0644); err != nil {
		log.Fatal(err)
	}

	downContent := fmt.Sprintf(`-- Migration: %s
-- Created: %s
-- Description: TODO: Add description

-- Add your down migration here
`, name, time.Now().Format("2006-01-02 15:04:05"))

	if err := os.WriteFile(downFile, []byte(downContent), 0644); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("✓ Migration files created:\n")
	fmt.Printf("  - %s\n", upFile)
	fmt.Printf("  - %s\n", downFile)
}

func getNextVersion(migrateDir string) int {
	files, err := os.ReadDir(migrateDir)
	if err != nil {
		return 1
	}

	maxVersion := 0
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		if len(name) >= 6 {
			if version, err := strconv.Atoi(name[:6]); err == nil {
				if version > maxVersion {
					maxVersion = version
				}
			}
		}
	}

	return maxVersion + 1
}

func migrateDrop(databaseURL, migrateDir string) {
	m, err := getMigrator(databaseURL, migrateDir)
	if err != nil {
		log.Fatal(err)
	}
	defer m.Close()

	fmt.Println("⚠️  WARNING: This will drop all tables and data!")
	fmt.Print("Are you sure? [y/N]: ")

	var response string
	fmt.Scanln(&response)

	if response != "y" && response != "Y" {
		fmt.Println("Cancelled.")
		return
	}

	fmt.Println("Dropping database...")
	if err := m.Drop(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("✓ Database dropped successfully")
}
