package store

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	"github.com/MrMoneyInTheBank/jobit/internal/model"
)

func OpenDB() (*sql.DB, error) {
	databasePath, err := DefaultDatabasePath()
	if err != nil {
		return nil, fmt.Errorf("resolve database location: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(databasePath), 0o700); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}
	return openSQLite(databasePath)
}

// OpenReadDB opens the fixed database location without creating or changing it.
func OpenReadDB() (*sql.DB, error) {
	databasePath, err := DefaultDatabasePath()
	if err != nil {
		return nil, fmt.Errorf("resolve database location: %w", err)
	}
	readOnlyDSN := (&url.URL{Scheme: "file", Path: databasePath, RawQuery: "mode=ro"}).String()
	return openSQLite(readOnlyDSN)
}

func openSQLite(dsn string) (*sql.DB, error) {
	database, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	database.SetMaxOpenConns(1)
	database.SetMaxIdleConns(1)

	if err := database.Ping(); err != nil {
		database.Close()
		return nil, fmt.Errorf("open database: %w", err)
	}
	return database, nil
}

// InitDB applies the single schema used by this build.
func InitDB(db *sql.DB) error {
	_, err := db.Exec(schemaSQL)
	return err
}

func CloseDB(database *sql.DB) error {
	if err := database.Close(); err != nil {
		return err
	}

	return nil
}

func AddJobApplication(db *sql.DB, app model.JobApplication) (int64, error) {
	if err := validateApplication(app); err != nil {
		return -1, err
	}
	result, err := db.Exec(insertQuery, applicationArgs(app)...)
	if err != nil {
		return -1, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return -1, err
	}

	return id, nil
}

func GetJobApplications(db *sql.DB) ([]model.StoredJobApplication, error) {
	rows, err := db.Query(getAllQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var applications []model.StoredJobApplication
	for rows.Next() {
		application, err := scanApplication(rows)
		if err != nil {
			return nil, err
		}
		applications = append(applications, application)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return applications, nil
}

func GetJobApplicationByID(db *sql.DB, id int64) (*model.StoredJobApplication, error) {
	if id <= 0 {
		return nil, ErrInvalidID
	}
	application, err := scanApplication(db.QueryRow(getByIDQuery, id))
	if err != nil {
		return nil, err
	}

	return &application, nil
}

func PatchJobApplication(
	db *sql.DB,
	app model.StoredJobApplication,
	patch model.JobApplicationPatch,
) error {
	if app.ID <= 0 {
		return ErrInvalidID
	}
	app.JobApplication.Apply(patch)
	if err := validateApplication(app.JobApplication); err != nil {
		return err
	}
	arguments := append(applicationArgs(app.JobApplication), app.ID)
	result, err := db.Exec(updateQuery, arguments...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return sql.ErrNoRows
	}

	return nil
}

func DeleteJobApplicationByID(db *sql.DB, id int64) error {
	if id <= 0 {
		return ErrInvalidID
	}
	result, err := db.Exec(deleteByIDQuery, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows != 1 {
		return sql.ErrNoRows
	}

	return nil
}
