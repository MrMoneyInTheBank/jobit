package store

import (
	"database/sql"
	"errors"
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

func InitDB(db *sql.DB) error {
	if _, err := db.Exec(schemaSQL); err != nil {
		return err
	}
	if err := CheckSchema(db); err != nil {
		if !errors.Is(err, errSchemaMismatch) {
			return err
		}
		return replaceEmptySchema(db)
	}
	return nil
}

var errSchemaMismatch = errors.New("database schema does not match this Jobit build; back up and remove the old database to start fresh")

// replaceEmptySchema refreshes a pre-release database only when it has no rows.
func replaceEmptySchema(db *sql.DB) error {
	transaction, err := db.Begin()
	if err != nil {
		return err
	}
	defer transaction.Rollback()
	var applicationCount int
	if err := transaction.QueryRow("SELECT COUNT(*) FROM job_applications").Scan(&applicationCount); err != nil {
		return err
	}
	if applicationCount != 0 {
		return errSchemaMismatch
	}
	if _, err := transaction.Exec("DROP TABLE job_applications"); err != nil {
		return err
	}
	if _, err := transaction.Exec(schemaSQL); err != nil {
		return err
	}
	if err := checkSchema(transaction); err != nil {
		return err
	}
	return transaction.Commit()
}

// CheckSchema verifies that a database uses the current layout without writing to it.
func CheckSchema(db *sql.DB) error {
	return checkSchema(db)
}

type schemaQuerier interface {
	Query(string, ...any) (*sql.Rows, error)
}

func checkSchema(db schemaQuerier) error {
	rows, err := db.Query("PRAGMA table_info(job_applications)")
	if err != nil {
		return err
	}
	defer rows.Close()

	expected := []string{"id", "company_name", "position", "application_date", "status", "referral", "remote", "location", "pay_min", "pay_max", "pay_currency", "ranking", "notes", "job_posting_link", "company_website_link"}
	var actual []string
	for rows.Next() {
		var index, notNull, primaryKey int
		var name, columnType string
		var defaultValue sql.NullString
		if err := rows.Scan(&index, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			return err
		}
		actual = append(actual, name)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(actual) != len(expected) {
		return errSchemaMismatch
	}
	for index, name := range expected {
		if actual[index] != name {
			return errSchemaMismatch
		}
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
