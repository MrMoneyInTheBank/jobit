package db

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/MrMoneyInTheBank/jobit/internal/model"
)

func TestApplySchema(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	defer db.Close()

	if _, err := db.Exec(schemaSQL); err != nil {
		t.Fatalf("Apply scheme: %v", err)
	}

	app := model.JobApplication{
		CompanyName:     "OpenAI",
		Position:        "Machine Learning Engineer",
		ApplicationDate: time.Now().UTC(),
		Status:          model.StatusSubmitted,
		Referral:        false,
	}

	args := bindJobApplication(app)
	res, err := db.Exec(`
		INSERT INTO job_applications (
			company_name,
			position,
			application_date,
			status,
			referral
		) VALUES (?, ?, ?, ?, ?)
	`, args[:5]...)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("last insert id: %v", err)
	}

	row := db.QueryRow(`
		SELECT
			id,
			company_name,
			position,
			application_date,
			status,
			referral
		FROM job_applications
		WHERE id = ?
	`, id)
	var (
		got  model.JobApplication
		date string
	)

	if err := row.Scan(
		&got.ID,
		&got.CompanyName,
		&got.Position,
		&date,
		&got.Status,
		&got.Referral,
	); err != nil {
		t.Fatalf("reading error while scan: %v", err)
	}

	_, err = time.Parse(time.RFC3339, date)
	if err != nil {
		t.Fatal(err)
	}

	if got.CompanyName != app.CompanyName {
		t.Fatalf("Company mismatch")
	}

	if got.Referral != app.Referral {
		t.Fatalf("Referral mismatch")
	}
}
