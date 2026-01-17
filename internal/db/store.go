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
	res, err := db.Exec(insertQuery, boundedApp.args()...)
	if err != nil {
		return -1, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}

	return id, nil
}

func GetJobApplications(db *sql.DB) ([]boundJobApplication, error) {
	rows, err := db.Query(getAllQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var got []boundJobApplication
	for rows.Next() {
		var app boundJobApplication
		err := rows.Scan(
			&app.id,
			&app.companyName,
			&app.position,
			&app.applicationDate,
			&app.status,
			&app.referral,
			&app.remote,
			&app.location,
			&app.payMin,
			&app.payMax,
			&app.payCurrency,
			&app.ranking,
			&app.notes,
			&app.jobPostingLink,
			&app.companyWebsiteLink,
		)
		if err != nil {
			return nil, err
		}
		got = append(got, app)
	}

	return got, nil
}

func GetJobApplicationByID(db *sql.DB, id int64) (*boundJobApplication, error) {
	var got boundJobApplication

	err := db.QueryRow(getByIDQuery, id).Scan(
		&got.id,
		&got.companyName,
		&got.position,
		&got.applicationDate,
		&got.status,
		&got.referral,
		&got.remote,
		&got.location,
		&got.payMin,
		&got.payMax,
		&got.payCurrency,
		&got.ranking,
		&got.notes,
		&got.jobPostingLink,
		&got.companyWebsiteLink,
	)
	if err != nil {
		return nil, err
	}

	return &got, nil
}
