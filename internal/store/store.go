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
	finalDSN, err := DatabaseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("resolve database location: %w", err)
	}
	if (dsn == nil || *dsn == "") && os.Getenv("JOBIT_DB_DSN") == "" {
		if err := os.MkdirAll(filepath.Dir(finalDSN), 0o700); err != nil {
			return nil, fmt.Errorf("create database directory: %w", err)
		}
	}

	db, err := sql.Open("sqlite3", finalDSN)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("open database: %w", err)
	}
	return db, nil
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
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec("DROP TABLE IF EXISTS job_applications"); err != nil {
		return err
	}
	if _, err := tx.Exec(schemaSQL); err != nil {
		return err
	}
	if _, err := tx.Exec("PRAGMA user_version = 1"); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func CloseDB(db *sql.DB) error {
	if err := db.Close(); err != nil {
		return err
	}

	return nil
}

func AddJobApplication(db *sql.DB, app model.JobApplication) (int64, error) {
	if err := validateApplication(app); err != nil {
		return -1, err
	}
	res, err := db.Exec(insertQuery, applicationArgs(app)...)
	if err != nil {
		return -1, err
	}

	id, err := res.LastInsertId()
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
	var got []model.StoredJobApplication
	for rows.Next() {
		mapped, err := scanApplication(rows)
		if err != nil {
			return nil, err
		}
		got = append(got, mapped)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return got, nil
}

func GetJobApplicationByID(db *sql.DB, id int64) (*model.StoredJobApplication, error) {
	if id <= 0 {
		return nil, ErrInvalidID
	}
	res, err := scanApplication(db.QueryRow(getByIDQuery, id))
	if err != nil {
		return nil, err
	}

	return &res, nil
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
	args := append(applicationArgs(app.JobApplication), app.ID)
	res, err := db.Exec(updateQuery, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
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
	res, err := db.Exec(deleteByIDQuery, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows != 1 {
		return sql.ErrNoRows
	}

	return nil
}
