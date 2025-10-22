package database

import "log/slog"

type SQLiteDatabase struct {
	SQLDatabase
}

// NewSQLiteDatabase creates a new SQLiteDatabase instance.
func NewSQLiteDatabase(dsn string, logger *slog.Logger) *SQLiteDatabase {
	return &SQLiteDatabase{
		SQLDatabase: NewSQLDatabase(logger, "sqlite3", dsn),
	}
}
