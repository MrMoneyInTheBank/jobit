package store

import (
	"database/sql"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/MrMoneyInTheBank/jobit/internal/model"
)

func TestFilePersistenceAndConfiguration(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("XDG_DATA_HOME", root)
	path, err := DefaultDatabasePath()
	if err != nil {
		t.Fatal(err)
	}
	db, err := OpenDB()
	if err != nil {
		t.Fatal(err)
	}
	mustInitDB(t, db)
	id, err := AddJobApplication(db, testApp)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	db, err = OpenDB()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	actual, err := GetJobApplicationByID(db, id)
	if err != nil || !actual.JobApplication.Equal(testApp) {
		t.Fatalf("reopened application: %+v %v", actual, err)
	}
}

func TestPersistenceErrors(t *testing.T) {
	db := newTestDB(t)
	for _, id := range []int64{-1, 0, 999} {
		want := ErrInvalidID
		if id > 0 {
			want = sql.ErrNoRows
		}
		_, err := GetJobApplicationByID(db, id)
		if !errors.Is(err, want) {
			t.Fatalf("lookup %d: %v", id, err)
		}
		if err := DeleteJobApplicationByID(db, id); !errors.Is(err, want) {
			t.Fatalf("delete %d: %v", id, err)
		}
		app := model.StoredJobApplication{ID: id, JobApplication: testApp}
		if err := PatchJobApplication(db, app, model.JobApplicationPatch{}); !errors.Is(err, want) {
			t.Fatalf("patch %d: %v", id, err)
		}
	}
	invalid := testApp
	invalid.Status = "invalid"
	if _, err := AddJobApplication(db, invalid); err == nil {
		t.Fatal("invalid insert accepted")
	}
	id, err := AddJobApplication(db, testApp)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("UPDATE job_applications SET application_date = ? WHERE id = ?", "bad-date", id); err != nil {
		t.Fatal(err)
	}
	if _, err = GetJobApplicationByID(db, id); err == nil || !strings.Contains(err.Error(), "invalid application date") {
		t.Fatalf("lookup corrupt date: %v", err)
	}
	if _, err = GetJobApplications(db); err == nil || !strings.Contains(err.Error(), "invalid application date") {
		t.Fatalf("list corrupt date: %v", err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	if err = InitDB(db); err == nil {
		t.Fatal("initialized closed database")
	}
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	if _, err = OpenReadDB(); err == nil {
		t.Fatal("read-only open created missing database")
	}
}

func TestSeedRollback(t *testing.T) {
	db := newTestDB(t)
	// Fail after one successful insert inside the transaction.
	_, err := db.Exec(`CREATE TRIGGER fail_seed BEFORE INSERT ON job_applications
 WHEN (SELECT COUNT(*) FROM job_applications) >= 1
 BEGIN SELECT RAISE(ABORT, 'forced insert failure'); END`)
	if err != nil {
		t.Fatal(err)
	}
	if err = SeedDB(db, 3); err == nil {
		t.Fatal("seed failure swallowed")
	}
	rows, err := GetJobApplications(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("partial seed persisted: %d", len(rows))
	}
}

func TestSeedCommitFailure(t *testing.T) {
	db := newTestDB(t)
	// A deferred foreign key lets inserts succeed and forces Commit to fail.
	for _, q := range []string{
		"PRAGMA foreign_keys=ON",
		"CREATE TABLE parent (id INTEGER PRIMARY KEY)",
		"CREATE TABLE child (id INTEGER REFERENCES parent(id) DEFERRABLE INITIALLY DEFERRED)",
		"CREATE TRIGGER fail_commit AFTER INSERT ON job_applications BEGIN INSERT INTO child VALUES (999); END",
	} {
		if _, err := db.Exec(q); err != nil {
			t.Fatal(err)
		}
	}
	if err := SeedDB(db, 1); err == nil {
		t.Fatal("commit failure swallowed")
	}
	rows, err := GetJobApplications(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatal("failed commit persisted rows")
	}
}

func TestSchemaInitializationFailure(t *testing.T) {
	db := mustOpenDB(t)
	if _, err := db.Exec("PRAGMA query_only=ON"); err != nil {
		t.Fatal(err)
	}
	if err := InitDB(db); err == nil {
		t.Fatal("schema write failure was ignored")
	}
}
