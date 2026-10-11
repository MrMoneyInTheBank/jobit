package store

import (
	"database/sql"
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
	got, err := GetJobApplicationByID(db, id)
	if err != nil {
		t.Fatal(err)
	}
	if !got.JobApplication.Equal(app) || got.JobApplication.Pay == nil || got.JobApplication.Pay.Min == nil || *got.JobApplication.Pay.Min != 0 || got.JobApplication.Pay.Max != nil {
		t.Fatalf("model round trip: %+v", got)
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

func TestLegacySchemaIsResetOnce(t *testing.T) {
	db := mustOpenDB(t)
	if _, err := db.Exec("CREATE TABLE job_applications (id INTEGER PRIMARY KEY, ranking INTEGER); INSERT INTO job_applications (ranking) VALUES (NULL)"); err != nil {
		t.Fatal(err)
	}
	if err := InitDB(db); err != nil {
		t.Fatal(err)
	}
	rows, err := GetJobApplications(db)
	if err != nil || len(rows) != 0 {
		t.Fatalf("old rows retained: %v %v", rows, err)
	}
	if _, err := AddJobApplication(db, testApp); err != nil {
		t.Fatal(err)
	}
	if err := InitDB(db); err != nil {
		t.Fatal(err)
	}
	rows, err = GetJobApplications(db)
	if err != nil || len(rows) != 1 {
		t.Fatalf("current rows lost: %v %v", rows, err)
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
	for _, column := range []string{"company_name", "position", "application_date", "status", "referral", "remote", "location", "ranking", "notes", "job_positing_link", "company_website_link"} {
		if _, err := db.Exec("UPDATE job_applications SET "+column+" = NULL WHERE id = ?", id); err == nil {
			t.Fatalf("accepted NULL in %s", column)
		}
	}
}
