package database

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/tbauriedel/go-nland-liveticker/internal/model"
)

type Database interface {
	ImportSchema(ctx context.Context) error
	InsertOperation(ctx context.Context, operation *model.Operation) error
	GetLastInsertedOperation(ctx context.Context) (*model.Operation, error)
}

type SQLDatabase struct {
	db         *sqlx.DB
	logger     *slog.Logger
	ConnectErr error
}

func NewSQLDatabase(logger *slog.Logger, driver string, dsn string) SQLDatabase {
	db, err := sqlx.Connect(driver, dsn)

	database := SQLDatabase{
		db:         db,
		logger:     logger,
		ConnectErr: err,
	}

	logger.Info("Connected to database", "driver", driver, "dsn", dsn)

	return database
}

// ImportSchema imports the schema into the database.
//
// If the table already exists, this function returns nil.
func (s *SQLDatabase) ImportSchema(ctx context.Context) error {
	query := s.db.Rebind("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='operations'")

	// Get rows from query
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query: %w", err)
	}
	defer rows.Close()

	// Scan found rows
	var count int
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			log.Fatal(err)
		}
	}

	// If more than 1 row is found (table exists), return nil
	if count > 0 {
		s.logger.Info("Database schema already imported. No need to import schema")
		return nil
	}

	s.logger.Info("Database schema not found. Importing schema")

	// Read schema from file
	schema, err := os.ReadFile("schema/sqlite.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	_, err = s.db.ExecContext(ctx, string(schema))
	if err != nil {
		return fmt.Errorf("failed to import schema: %w", err)
	}

	return nil
}

func (s *SQLDatabase) InsertOperation(ctx context.Context, operation *model.Operation) error {
	query := s.db.Rebind("INSERT INTO operations (time, units, report, district, location) VALUES (?,?,?,?,?)")

	_, err := s.db.ExecContext(ctx,
		query,
		operation.Time.Format("2006-01-02 15:04"), operation.Units, operation.Report, operation.District, operation.Location,
	)

	if err != nil {
		return fmt.Errorf("failed to insert operation: %w", err)
	}

	s.logger.Debug("Inserted operation", "operation", operation.GetIdentifier())

	return nil
}

func (s *SQLDatabase) GetLastInsertedOperation(ctx context.Context) (*model.Operation, error) {
	query := s.db.Rebind("SELECT * FROM operations ORDER BY time DESC LIMIT 1")

	var temp []struct {
		Time     string `db:"time"`
		Units    string `db:"units"`
		Report   string `db:"report"`
		District string `db:"district"`
		Location string `db:"location"`
	}

	if err := s.db.SelectContext(ctx, &temp, query); err != nil {
		return nil, fmt.Errorf("sql select operations failed: %w", err)
	}

	// no operation found. Return nil
	if len(temp) == 0 {
		return nil, nil
	}

	parsedTime, err := time.Parse("2006-01-02 15:04", temp[0].Time)
	if err != nil {
		return nil, fmt.Errorf("failed to parse time: %w (raw value: %s)", err, temp[0].Time)
	}

	operation := model.Operation{
		Time:     parsedTime,
		Units:    temp[0].Units,
		Report:   temp[0].Report,
		District: temp[0].District,
		Location: temp[0].Location,
	}

	return &operation, nil
}
