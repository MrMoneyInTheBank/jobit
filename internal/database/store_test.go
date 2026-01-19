package database

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/MrMoneyInTheBank/jobit/internal/model"
)

var testApp model.JobApplication = model.JobApplication{
	CompanyName:     "OpenAI",
	Position:        "Machine Learning Engineer",
	ApplicationDate: time.Now().UTC(),
	Status:          model.StatusSubmitted,
	Referral:        false,
}

func mustOpenDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := OpenDB(nil)
	if err != nil {
		t.Fatalf("Could not open database: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func mustInitDB(t *testing.T, db *sql.DB) {
	t.Helper()
	err := InitDB(db)
	if err != nil {
		t.Fatalf("Could not initialize database: %v", err)
	}
}

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db := mustOpenDB(t)
	mustInitDB(t, db)
	return db
}

func TestOpenDB(t *testing.T) {
	db := mustOpenDB(t)
	if err := db.Ping(); err != nil {
		t.Fatalf("Could not ping database: %v", err)
	}
}

func TestInitDB(t *testing.T) {
	db := newTestDB(t)
	boundedApp := bindJobApplication(testApp)

	if _, err := db.Exec(insertQuery, boundedApp.insertArgs()...); err != nil {
		t.Fatalf("insert: %v", err)
	}
}

func TestCloseDB(t *testing.T) {
	db, err := OpenDB(nil)
	t.Cleanup(func() { db.Close() })
	if err != nil {
		t.Fatalf("Could not open database: %v", err)
	}

	err = CloseDB(db)
	if err != nil {
		t.Fatalf("Could not close database: %v", err)
	}
}

func TestAddJobApplication(t *testing.T) {
	db := newTestDB(t)

	if _, err := AddJobApplication(db, testApp); err != nil {
		t.Fatalf("Could not add job application: %v", err)
	}
}

func TestGetJobApplications(t *testing.T) {
	db := newTestDB(t)

	id1, err1 := AddJobApplication(db, testApp)
	id2, err2 := AddJobApplication(db, testApp)
	if err1 != nil || err2 != nil {
		t.Logf("Could not add job application 1: %v", err1)
		t.Fatalf("Could not add job application 2: %v", err2)
	}

	got, err := GetJobApplications(db)
	if err != nil {
		t.Fatalf("Could not get job applications: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("Number of applications mismatch")
	}

	if got[0].ID != id1 || got[1].ID != id2 {
		t.Fatalf("ID mismatch")
	}
}

func TestGetJobApplicationByID(t *testing.T) {
	db := newTestDB(t)

	id, err := AddJobApplication(db, testApp)
	if err != nil {
		t.Fatalf("Could not add job application: %v", err)
	}
	got, err := GetJobApplicationByID(db, id)
	if err != nil {
		t.Fatalf("Could not get job application: %v", err)
	}
	if got.ID != id {
		t.Fatalf("ID mismatch")
	}
}

func ptr[T any](t T) *T {
	return &t
}

func TestPatchJobApplicationByID(t *testing.T) {
	db := newTestDB(t)

	id, err := AddJobApplication(db, testApp)
	if err != nil {
		t.Fatalf("Could not add job application: %v", err)
	}
	app, err := GetJobApplicationByID(db, id)
	if err != nil {
		t.Fatalf("Could not get job application: %v", err)
	}
	if app == nil {
		t.Fatalf("Job application is nil")
	}
	patch := model.JobApplicationPatch{
		CompanyName: ptr("Anthropic"),
	}

	err = PatchJobApplication(db, *app, patch)
	if err != nil {
		t.Fatalf("Could not patch job application: %v", err)
	}
	app, err = GetJobApplicationByID(db, id)
	if err != nil {
		t.Fatalf("Could not get job application: %v", err)
	}
	if app.CompanyName != *patch.CompanyName {
		fmt.Printf("app.CompanyName: %s\n", app.CompanyName)
		fmt.Printf("patch.CompanyName: %s\n", *patch.CompanyName)
		t.Fatalf("Company name mismatch")
	}
}

func TestDeleteJobApplicationByID(t *testing.T) {
	db := newTestDB(t)
	id, err := AddJobApplication(db, testApp)
	if err != nil {
		t.Fatalf("Could not add job application: %v", err)
	}
	boundedApps, err := GetJobApplications(db)
	if err != nil {
		t.Fatalf("Could not get job applications: %v", err)
	}
	cnt := len(boundedApps)

	err = DeleteJobApplicationByID(db, id)
	if err != nil {
		t.Fatalf("Could not delete job application: %v", err)
	}
	boundedApps, _ = GetJobApplications(db)
	if len(boundedApps) != cnt-1 {
		t.Fatalf("Number of applications mismatch post deletion.")
	}
}
