package db

import (
	"database/sql"

	"github.com/MrMoneyInTheBank/jobit/internal/model"
)

func OpenDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	return db, nil
}

func InitDB(db *sql.DB) (*sql.DB, error) {
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if _, err := db.Exec(schemaSQL); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func CloseDB(db *sql.DB) error {
	if err := db.Close(); err != nil {
		return err
	}

	return nil
}

func AddJobApplication(db *sql.DB, app model.JobApplication) (int64, error) {
	boundedApp := bindJobApplication(app)
	res, err := db.Exec(`
		INSERT INTO job_applications (
			company_name,
			position,
			application_date,
			status,
			referral,
			remote,
			location,
			pay_min,
			pay_max,
			pay_currency,
			ranking,
			notes,
			job_positing_link,
			company_website_link
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, boundedApp.args()...)
	if err != nil {
		return -1, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}

	return id, nil
}
