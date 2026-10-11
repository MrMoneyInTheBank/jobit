package store

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
	Rank:            model.RankThree,
}

func mustOpenDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := OpenDB(ptr(":memory:"))
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
	if _, err := db.Exec(insertQuery, applicationArgs(testApp)...); err != nil {
		t.Fatalf("insert: %v", err)
	}
}

func TestCloseDB(t *testing.T) {
	db, err := OpenDB(ptr(":memory:"))
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

	actual, err := GetJobApplications(db)
	if err != nil {
		t.Fatalf("Could not get job applications: %v", err)
	}

	if len(actual) != 2 {
		t.Fatalf("Number of applications mismatch")
	}

	if actual[0].ID != id1 || actual[1].ID != id2 {
		t.Fatalf("ID mismatch")
	}
}

func TestGetJobApplicationByID(t *testing.T) {
	db := newTestDB(t)

	id, err := AddJobApplication(db, testApp)
	if err != nil {
		t.Fatalf("Could not add job application: %v", err)
	}
	actual, err := GetJobApplicationByID(db, id)
	if err != nil {
		t.Fatalf("Could not get job application: %v", err)
	}
	if actual.ID != id {
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
	if app.JobApplication.CompanyName != *patch.CompanyName {
		fmt.Printf("app.JobApplication.CompanyName: %s\n", app.JobApplication.CompanyName)
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

func TestRoundTripAndClear(t *testing.T) {
	db := newTestDB(t)
	original := testApp
	original.WorkMode = model.WorkModeHybrid
	original.Location = "London"
	original.Rank = 5
	original.Notes = "Follow up"
	original.JobPostingLink = "https://example.com/job"
	original.CompanyWebsiteLink = "https://example.com"
	original.Pay = &model.Pay{Min: ptr(100), Max: ptr(200), Currency: ptr("GBP")}
	id, err := AddJobApplication(db, original)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := GetJobApplicationByID(db, id)
	if err != nil {
		t.Fatal(err)
	}
	if !actual.JobApplication.Equal(original) {
		t.Fatalf("round trip mismatch: %+v", actual)
	}
	var noPay *model.Pay
	patch := model.JobApplicationPatch{Location: ptr(""), Notes: ptr(""), Pay: &noPay, Rank: ptr(model.RankOne)}
	if err := PatchJobApplication(db, *actual, patch); err != nil {
		t.Fatal(err)
	}
	original.Apply(patch)
	actual, err = GetJobApplicationByID(db, id)
	if err != nil {
		t.Fatal(err)
	}
	if !actual.JobApplication.Equal(original) {
		t.Fatalf("clearing mismatch: %+v", actual)
	}
}

func TestSeedDB(t *testing.T) {
	db := newTestDB(t)
	if err := SeedDB(db, 50); err != nil {
		t.Fatal(err)
	}
	actual, err := GetJobApplications(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(actual) != 50 {
		t.Fatalf("actual %d rows", len(actual))
	}
}
