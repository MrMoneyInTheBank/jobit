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

var testApp model.JobApplication = model.JobApplication{
	CompanyName:     "OpenAI",
	Position:        "Machine Learning Engineer",
	ApplicationDate: time.Now().UTC(),
	Status:          model.StatusSubmitted,
	Referral:        false,
}

func TestGetJobApplications(t *testing.T) {
	db, _ := OpenDB()
	t.Cleanup(func() { db.Close() })
	db, _ = InitDB(db)

	id1, _ := AddJobApplication(db, testApp)
	id2, _ := AddJobApplication(db, testApp)

	got, err := GetJobApplications(db)
	if err != nil {
		t.Fatalf("Could not get job applications: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("Number of applications mismatch")
	}

	if got[0].id != id1 || got[1].id != id2 {
		t.Fatalf("ID mismatch")
	}
}

func TestGetJobApplicationByID(t *testing.T) {
	db, _ := OpenDB()
	t.Cleanup(func() { db.Close() })
	db, _ = InitDB(db)

	id, _ := AddJobApplication(db, testApp)
	got, err := GetJobApplicationByID(db, id)
	if err != nil {
		t.Fatalf("Could not get job application: %v", err)
	}
	if got.id != id {
		t.Fatalf("ID mismatch")
	}
}
