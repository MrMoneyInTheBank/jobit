package db

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
	db, err := OpenDB()
	if err != nil {
		t.Fatalf("Could not open database: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func mustInitDB(t *testing.T, db *sql.DB) *sql.DB {
	t.Helper()
	db, err := InitDB(db)
	if err != nil {
		t.Fatalf("Could not initialize database: %v", err)
	}
	return db
}

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	return mustInitDB(t, mustOpenDB(t))
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
	db, err := OpenDB()
	t.Cleanup(func() { db.Close() })

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
	db := newTestDB(t)

	id, _ := AddJobApplication(db, testApp)
	got, err := GetJobApplicationByID(db, id)
	if err != nil {
		t.Fatalf("Could not get job application: %v", err)
	}
	if got.id != id {
		t.Fatalf("ID mismatch")
	}
}

func ptr[T any](t T) *T {
	return &t
}

func TestPatchJobApplicationByID(t *testing.T) {
	db := newTestDB(t)

	id, _ := AddJobApplication(db, testApp)
	got, _ := GetJobApplicationByID(db, id)
	app := got.toModel()

	patch := model.JobApplicationPatch{
		CompanyName: ptr("Anthropic"),
	}

	err := PatchJobApplication(db, app, patch)
	if err != nil {
		t.Fatalf("Could not patch job application: %v", err)
	}
	got, _ = GetJobApplicationByID(db, id)
	app = got.toModel()
	if app.CompanyName != *patch.CompanyName {
		fmt.Printf("app.CompanyName: %s\n", app.CompanyName)
		fmt.Printf("patch.CompanyName: %s\n", *patch.CompanyName)
		t.Fatalf("Company name mismatch")
	}
}

func TestDeleteJobApplicationByID(t *testing.T) {
	db := newTestDB(t)
	id, _ := AddJobApplication(db, testApp)
	boundedApps, _ := GetJobApplications(db)
	cnt := len(boundedApps)

	err := DeleteJobApplicationByID(db, id)
	if err != nil {
		t.Fatalf("Could not delete job application: %v", err)
	}
	boundedApps, _ = GetJobApplications(db)
	if len(boundedApps) != cnt-1 {
		t.Fatalf("Number of applications mismatch post deletion.")
	}
}
