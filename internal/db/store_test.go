package db

import (
	"testing"
	"time"

	"github.com/MrMoneyInTheBank/jobit/internal/model"
)

func TestOpenDB(t *testing.T) {
	db, err := OpenDB()
	if err != nil {
		t.Fatalf("Could not open database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.Ping(); err != nil {
		t.Fatalf("Could not ping database: %v", err)
	}
}

func TestInitDB(t *testing.T) {
	db, err := OpenDB()
	if err != nil {
		t.Fatalf("Could not open database: %v", err)
	}

	db, err = InitDB(db)
	if err != nil {
		t.Fatalf("Could not initialize database: %v", err)
	}

	application := model.JobApplication{
		CompanyName:     "OpenAI",
		Position:        "Machine Learning Engineer",
		ApplicationDate: time.Now().UTC(),
		Status:          model.StatusSubmitted,
		Referral:        false,
	}

	boundedApp := bindJobApplication(application)
	if _, err := db.Exec(`
		INSERT INTO job_applications (
			company_name,
			position,
			application_date,
			status,
			referral
		) VALUES (?, ?, ?, ?, ?)
	`, boundedApp.args()[:5]...); err != nil {
		t.Fatalf("insert: %v", err)
	}

	t.Cleanup(func() { db.Close() })
}

func TestCloseDB(t *testing.T) {
	db, err := OpenDB()

	err = CloseDB(db)
	if err != nil {
		t.Fatalf("Could not close database: %v", err)
	}
}

func TestAddJobApplication(t *testing.T) {
	db, _ := OpenDB()
	db, _ = InitDB(db)
	t.Cleanup(func() { db.Close() })

	app := model.JobApplication{
		CompanyName:     "OpenAI",
		Position:        "Machine Learning Engineer",
		ApplicationDate: time.Now().UTC(),
		Status:          model.StatusSubmitted,
		Referral:        false,
	}

	if _, err := AddJobApplication(db, app); err != nil {
		t.Fatalf("Could not add job application: %v", err)
	}
}
