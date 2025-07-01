package postgres

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Config struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func LoadConfig() *Config {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}

	user := os.Getenv("DB_USER")
	if user == "" {
		user = "tyokeren"
	}

	password := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "sea_catering"
	}

	sslMode := os.Getenv("DB_SSLMODE")
	if sslMode == "" {
		sslMode = "disable"
	}

	maxOpenConns := 25
	if envMaxOpen := os.Getenv("DB_MAX_OPEN_CONNS"); envMaxOpen != "" {
		if parsed, err := strconv.Atoi(envMaxOpen); err == nil {
			maxOpenConns = parsed
		}
	}

	maxIdleConns := 5
	if envMaxIdle := os.Getenv("DB_MAX_IDLE_CONNS"); envMaxIdle != "" {
		if parsed, err := strconv.Atoi(envMaxIdle); err == nil {
			maxIdleConns = parsed
		}
	}

	connMaxLifetime := 5 * time.Minute
	if envLifetime := os.Getenv("DB_CONN_MAX_LIFETIME"); envLifetime != "" {
		if parsed, err := time.ParseDuration(envLifetime); err == nil {
			connMaxLifetime = parsed
		}
	}

	connMaxIdleTime := 5 * time.Minute
	if envIdleTime := os.Getenv("DB_CONN_MAX_IDLE_TIME"); envIdleTime != "" {
		if parsed, err := time.ParseDuration(envIdleTime); err == nil {
			connMaxIdleTime = parsed
		}
	}

	return &Config{
		Host:            host,
		Port:            port,
		User:            user,
		Password:        password,
		DBName:          dbName,
		SSLMode:         sslMode,
		MaxOpenConns:    maxOpenConns,
		MaxIdleConns:    maxIdleConns,
		ConnMaxLifetime: connMaxLifetime,
		ConnMaxIdleTime: connMaxIdleTime,
	}
}

func New() (*sqlx.DB, error) {
	config := LoadConfig()
	return NewWithConfig(config)
}

func NewWithConfig(config *Config) (*sqlx.DB, error) {
	if config == nil {
		config = LoadConfig()
	}

	dsn := FormatDSN(config)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func FormatDSN(config *Config) string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.DBName,
		config.SSLMode,
	)
}

func FormatDSNFromEnv() string {
	config := LoadConfig()
	return FormatDSN(config)
}

func TestConnection() error {
	db, err := New()
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Ping()
}

func CreateDatabase() error {
	config := LoadConfig()

	tempConfig := *config
	tempConfig.DBName = "postgres"

	db, err := NewWithConfig(&tempConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres database: %w", err)
	}
	defer db.Close()

	var exists bool
	query := "SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = $1)"
	err = db.QueryRow(query, config.DBName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	if !exists {

		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", config.DBName))
		if err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
	}

	return nil
}

func DropDatabase() error {
	config := LoadConfig()

	tempConfig := *config
	tempConfig.DBName = "postgres"

	db, err := NewWithConfig(&tempConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres database: %w", err)
	}
	defer db.Close()

	terminateQuery := `
		SELECT pg_terminate_backend(pid)
		FROM pg_stat_activity
		WHERE datname = $1 AND pid <> pg_backend_pid()
	`
	_, err = db.Exec(terminateQuery, config.DBName)
	if err != nil {
		return fmt.Errorf("failed to terminate connections: %w", err)
	}

	_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", config.DBName))
	if err != nil {
		return fmt.Errorf("failed to drop database: %w", err)
	}

	return nil
}

func GetDatabaseInfo() (map[string]interface{}, error) {
	db, err := New()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	info := make(map[string]interface{})

	var version string
	err = db.QueryRow("SELECT version()").Scan(&version)
	if err != nil {
		return nil, fmt.Errorf("failed to get version: %w", err)
	}
	info["version"] = version

	var currentDB string
	err = db.QueryRow("SELECT current_database()").Scan(&currentDB)
	if err != nil {
		return nil, fmt.Errorf("failed to get current database: %w", err)
	}
	info["current_database"] = currentDB

	stats := db.Stats()
	info["max_open_connections"] = stats.MaxOpenConnections
	info["open_connections"] = stats.OpenConnections
	info["in_use"] = stats.InUse
	info["idle"] = stats.Idle
	info["wait_count"] = stats.WaitCount
	info["wait_duration"] = stats.WaitDuration

	var size string
	sizeQuery := "SELECT pg_size_pretty(pg_database_size(current_database()))"
	err = db.QueryRow(sizeQuery).Scan(&size)
	if err != nil {
		return nil, fmt.Errorf("failed to get database size: %w", err)
	}
	info["database_size"] = size

	var tableCount int
	tableCountQuery := `
		SELECT COUNT(*) 
		FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
	`
	err = db.QueryRow(tableCountQuery).Scan(&tableCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get table count: %w", err)
	}
	info["table_count"] = tableCount

	return info, nil
}

func GetTableNames() ([]string, error) {
	db, err := New()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get table names: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, tableName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return tables, nil
}

func CheckHealth() error {
	db, err := New()
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	var result int
	if err := db.QueryRow("SELECT 1").Scan(&result); err != nil {
		return fmt.Errorf("test query failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("unexpected test query result: %d", result)
	}

	return nil
}

func CreateExtension(extensionName string) error {
	db, err := New()
	if err != nil {
		return err
	}
	defer db.Close()

	query := fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS %s", extensionName)
	_, err = db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create extension %s: %w", extensionName, err)
	}

	return nil
}

func GetExtensions() ([]string, error) {
	db, err := New()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := "SELECT extname FROM pg_extension ORDER BY extname"
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get extensions: %w", err)
	}
	defer rows.Close()

	var extensions []string
	for rows.Next() {
		var extName string
		if err := rows.Scan(&extName); err != nil {
			return nil, fmt.Errorf("failed to scan extension name: %w", err)
		}
		extensions = append(extensions, extName)
	}

	return extensions, nil
}

func EnableUUID() error {
	return CreateExtension("uuid-ossp")
}

func EnableCrypto() error {
	return CreateExtension("pgcrypto")
}

func TruncateTable(tableName string) error {
	db, err := New()
	if err != nil {
		return err
	}
	defer db.Close()

	query := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", tableName)
	_, err = db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to truncate table %s: %w", tableName, err)
	}

	return nil
}

func TruncateAllTables() error {
	tables, err := GetTableNames()
	if err != nil {
		return err
	}

	db, err := New()
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec("SET session_replication_role = replica")
	if err != nil {
		return fmt.Errorf("failed to disable foreign key checks: %w", err)
	}

	for _, table := range tables {
		query := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)
		_, err = tx.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", table, err)
		}
	}

	_, err = tx.Exec("SET session_replication_role = DEFAULT")
	if err != nil {
		return fmt.Errorf("failed to re-enable foreign key checks: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func BackupDatabase(backupPath string) error {
	config := LoadConfig()

	return fmt.Errorf("database backup requires pg_dump tool - please use: pg_dump -h %s -p %s -U %s -d %s > %s",
		config.Host, config.Port, config.User, config.DBName, backupPath)
}

func RestoreDatabase(backupPath string) error {
	config := LoadConfig()

	return fmt.Errorf("database restore requires psql tool - please use: psql -h %s -p %s -U %s -d %s < %s",
		config.Host, config.Port, config.User, config.DBName, backupPath)
}

func GetConnectionString() string {
	config := LoadConfig()
	return FormatDSN(config)
}

func Close(db *sqlx.DB) error {
	if db == nil {
		return nil
	}
	return db.Close()
}
