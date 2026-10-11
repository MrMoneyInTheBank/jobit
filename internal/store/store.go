package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	"github.com/MrMoneyInTheBank/jobit/internal/model"
)

func OpenDB(dsn *string) (*sql.DB, error) {
	databaseDSN, err := DatabaseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("resolve database location: %w", err)
	}
	if (dsn == nil || *dsn == "") && os.Getenv("JOBIT_DB_DSN") == "" {
		if err := os.MkdirAll(filepath.Dir(databaseDSN), 0o700); err != nil {
			return nil, fmt.Errorf("create database directory: %w", err)
		}
	}

	database, err := sql.Open("sqlite3", databaseDSN)
	if err != nil {
		return nil, err
	}

	if err := database.Ping(); err != nil {
		database.Close()
		return nil, fmt.Errorf("open database: %w", err)
	}
	return database, nil
}

func InitDB(db *sql.DB) error {
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	var version int
	if err := db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return err
	}
	if version >= 1 {
		_, err := db.Exec(schemaSQL)
		return err
	}

	// Version zero is the old nullable schema (or a new database). Discard its rows.
	transaction, err := db.Begin()
	if err != nil {
		return err
	}
	defer transaction.Rollback()
	if _, err := transaction.Exec("DROP TABLE IF EXISTS job_applications"); err != nil {
		return err
	}
	if _, err := transaction.Exec(schemaSQL); err != nil {
		return err
	}
	if _, err := transaction.Exec("PRAGMA user_version = 1"); err != nil {
		return err
	}
	if err := transaction.Commit(); err != nil {
		return err
	}
	return nil
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
