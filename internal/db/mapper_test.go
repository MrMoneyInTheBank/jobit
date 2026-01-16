package db

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/MrMoneyInTheBank/jobit/internal/model"
)

func TestSchemaAndMappingSmoke(t *testing.T) {
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

	boundedApp := bindJobApplication(app)
	res, err := db.Exec(`
		INSERT INTO job_applications (
			company_name,
			position,
			application_date,
			status,
			referral
		) VALUES (?, ?, ?, ?, ?)
	`, boundedApp.args()[:5]...)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("last insert id: %v", err)
	}

	row := db.QueryRow(`
		SELECT *
		FROM job_applications
		WHERE id = ?
	`, id)

	var got boundJobApplication

	if err := row.Scan(
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
	); err != nil {
		t.Fatalf("reading error while scan: %v", err)
	}

	if !got.toModel().Compare(app) {
		t.Fatalf("Application mismatch")
	}
}
