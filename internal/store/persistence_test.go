package store

import (
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MrMoneyInTheBank/jobit/internal/model"
)

func TestFilePersistenceAndConfiguration(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("JOBIT_DB_DSN", "")
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	defaultPath, err := DefaultDatabasePath()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, env, explicit, file string }{
		{"default", "", "", defaultPath},
		{"environment", "environment.db", "", "environment.db"},
		{"explicit", "unused.db", "explicit.db", "explicit.db"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("JOBIT_DB_DSN", tc.env)
			db, err := OpenDB(&tc.explicit)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if db != nil {
					db.Close()
				}
			})
			mustInitDB(t, db)
			rows, err := GetJobApplications(db)
			if err != nil || len(rows) != 0 {
				t.Fatalf("empty database: %v %v", rows, err)
			}
			apps := []model.JobApplication{testApp, testApp}
			apps[1].Location = "London"
			apps[1].WorkMode = model.WorkModeRemote
			apps[1].Rank = 3
			apps[1].Notes = "Notes"
			apps[1].Referral = true
			apps[1].JobPostingLink = "https://example.com/job"
			apps[1].CompanyWebsiteLink = "https://example.com"
			apps[1].Pay = &model.Pay{Min: ptr(0), Currency: ptr("GBP")}
			ids := make([]int64, len(apps))
			for i, app := range apps {
				ids[i], err = AddJobApplication(db, app)
				if err != nil {
					t.Fatal(err)
				}
			}
			if err := db.Close(); err != nil {
				t.Fatal(err)
			}
			// Reopen the expected path explicitly to prove which file was selected.
			db, err = OpenDB(ptr(tc.file))
			if err != nil {
				t.Fatal(err)
			}
			mustInitDB(t, db)
			// Reverse unspecified SQLite scan ordering to ensure ORDER BY is effective.
			if _, err = db.Exec("PRAGMA reverse_unordered_selects=ON"); err != nil {
				t.Fatal(err)
			}
			rows, err = GetJobApplications(db)
			if err != nil {
				t.Fatal(err)
			}
			if len(rows) != len(apps) {
				t.Fatalf("got %d rows", len(rows))
			}
			for i, app := range apps {
				if rows[i].ID != ids[i] || !rows[i].JobApplication.Equal(app) {
					t.Fatalf("list mismatch: %+v", rows[i])
				}
				got, err := GetJobApplicationByID(db, ids[i])
				if err != nil {
					t.Fatal(err)
				}
				if !got.JobApplication.Equal(app) {
					t.Fatalf("lookup mismatch: %+v", got)
				}
			}
		})
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
	if _, err = OpenDB(ptr(filepath.Join(t.TempDir(), "missing", "db"))); err == nil {
		t.Fatal("opened invalid path")
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
