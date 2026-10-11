package store

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/MrMoneyInTheBank/jobit/internal/model"
)

func TestModelMappingAndSchema(t *testing.T) {
	db := newTestDB(t)
	app := testApp
	app.Location = ""
	app.WorkMode = ""
	app.Notes = ""
	zero := 0
	app.Pay = &model.Pay{Min: &zero}
	id, err := AddJobApplication(db, app)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := GetJobApplicationByID(db, id)
	if err != nil {
		t.Fatal(err)
	}
	if !actual.JobApplication.Equal(app) || actual.JobApplication.Pay == nil || actual.JobApplication.Pay.Min == nil || *actual.JobApplication.Pay.Min != 0 || actual.JobApplication.Pay.Max != nil {
		t.Fatalf("model round trip: %+v", actual)
	}
	var location, mode, notes string
	var min, max sql.NullInt64
	if err := db.QueryRow("SELECT location, remote, notes, pay_min, pay_max FROM job_applications WHERE id = ?", id).Scan(&location, &mode, &notes, &min, &max); err != nil {
		t.Fatal(err)
	}
	if location != "" || mode != "" || notes != "" || !min.Valid || min.Int64 != 0 || max.Valid {
		t.Fatalf("unexpected SQL values: location=%q mode=%q notes=%q min=%v max=%v", location, mode, notes, min, max)
	}
}

func TestLegacySchemaIsPreserved(t *testing.T) {
	db := mustOpenDB(t)
	if _, err := db.Exec("CREATE TABLE job_applications (id INTEGER PRIMARY KEY, ranking INTEGER); INSERT INTO job_applications (ranking) VALUES (NULL)"); err != nil {
		t.Fatal(err)
	}
	if err := InitDB(db); err == nil {
		t.Fatal("legacy database was silently accepted")
	}
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM job_applications").Scan(&count); err != nil || count != 1 {
		t.Fatalf("legacy rows changed: count=%d err=%v", count, err)
	}
	if err := db.QueryRow("PRAGMA user_version").Scan(&count); err != nil || count != 0 {
		t.Fatalf("legacy version changed: version=%d err=%v", count, err)
	}
}

func TestVersionOneColumnMigrationPreservesRows(t *testing.T) {
	db := mustOpenDB(t)
	legacySchema := strings.Replace(schemaSQL, "job_posting_link", "job_positing_link", 1)
	if _, err := db.Exec(legacySchema); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("PRAGMA user_version = 1"); err != nil {
		t.Fatal(err)
	}
	legacyInsert := strings.Replace(insertQuery, "job_posting_link", "job_positing_link", 1)
	if _, err := db.Exec(legacyInsert, applicationArgs(testApp)...); err != nil {
		t.Fatal(err)
	}
	if err := InitDB(db); err != nil {
		t.Fatal(err)
	}
	rows, err := GetJobApplications(db)
	if err != nil || len(rows) != 1 {
		t.Fatalf("migrated rows lost: %v %v", rows, err)
	}
	if !rows[0].JobApplication.Equal(testApp) {
		t.Fatalf("migrated row changed: %+v", rows[0])
	}
	if err := InitDB(db); err != nil {
		t.Fatalf("migration was not idempotent: %v", err)
	}
}

func TestRequiredDatabaseValues(t *testing.T) {
	db := newTestDB(t)
	for _, rank := range []model.Ranking{0, 6} {
		app := testApp
		app.Rank = rank
		if _, err := AddJobApplication(db, app); err == nil {
			t.Fatalf("accepted rank %d", rank)
		}
	}
	id, err := AddJobApplication(db, testApp)
	if err != nil {
		t.Fatal(err)
	}
	for _, column := range []string{"company_name", "position", "application_date", "status", "referral", "remote", "location", "ranking", "notes", "job_posting_link", "company_website_link"} {
		if _, err := db.Exec("UPDATE job_applications SET "+column+" = NULL WHERE id = ?", id); err == nil {
			t.Fatalf("accepted NULL in %s", column)
		}
	}
}

func TestStoreRejectsInvalidApplicationFields(t *testing.T) {
	database := newTestDB(t)
	id, err := AddJobApplication(database, testApp)
	if err != nil {
		t.Fatal(err)
	}

	negative := -1
	minimum, maximum := 200, 100
	badCurrency := "usd"
	for _, testCase := range []struct {
		name   string
		change func(*model.JobApplication)
	}{
		{"blank company", func(app *model.JobApplication) { app.CompanyName = "  " }},
		{"negative pay", func(app *model.JobApplication) { app.Pay = &model.Pay{Min: &negative} }},
		{"inverted pay", func(app *model.JobApplication) { app.Pay = &model.Pay{Min: &minimum, Max: &maximum} }},
		{"currency", func(app *model.JobApplication) { app.Pay = &model.Pay{Currency: &badCurrency} }},
		{"posting link", func(app *model.JobApplication) { app.JobPostingLink = "javascript:alert(1)" }},
		{"website link", func(app *model.JobApplication) { app.CompanyWebsiteLink = "example.com" }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			application := testApp
			testCase.change(&application)
			if _, err := AddJobApplication(database, application); err == nil {
				t.Fatal("invalid insert accepted")
			}
			stored := model.StoredJobApplication{ID: id, JobApplication: application}
			if err := PatchJobApplication(database, stored, model.JobApplicationPatch{}); err == nil {
				t.Fatal("invalid update accepted")
			}
		})
	}
}
