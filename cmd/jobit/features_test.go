package main

import (
	"bytes"
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/MrMoneyInTheBank/jobit/internal/model"
	"github.com/MrMoneyInTheBank/jobit/internal/store"
)

func TestInstanceLock(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("XDG_DATA_HOME", root)
	first, err := acquireInstance()
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	if second, err := acquireInstance(); err == nil {
		second.Close()
		t.Fatal("second instance admitted")
	} else if !strings.Contains(err.Error(), "terminal") {
		t.Fatal(err)
	}
	first.Close()
	third, err := acquireInstance()
	if err != nil {
		t.Fatal("stale file blocks launch:", err)
	}
	third.Close()
}

func TestCSVExport(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("XDG_DATA_HOME", root)
	path, err := store.DefaultDatabasePath()
	if err != nil {
		t.Fatal(err)
	}
	db, err := store.OpenDB()
	if err != nil {
		t.Fatal(err)
	}
	if err = store.InitDB(db); err != nil {
		t.Fatal(err)
	}
	zero := 0
	app := model.JobApplication{CompanyName: "Acme, Inc.", Position: "Engineer", ApplicationDate: time.Now(), Status: model.StatusSubmitted, Rank: model.RankThree, Notes: "line one\n\"quoted\" line two", Pay: &model.Pay{Min: &zero}}
	if _, err = store.AddJobApplication(db, app); err != nil {
		t.Fatal(err)
	}
	db.Close()
	sourceBefore, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sourceInfo, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err = exportCSV("-", &out); err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(&out).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[1][1] != app.CompanyName || rows[1][12] != app.Notes || rows[1][9] != "0" || rows[1][10] != "" {
		t.Fatalf("incorrect CSV: %v", rows)
	}
	dest := filepath.Join(t.TempDir(), "export.csv")
	if err = exportCSV(dest, &out); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(dest)
	if err = exportCSV(dest, &out); err == nil {
		t.Fatal("overwrote existing file")
	}
	after, _ := os.ReadFile(dest)
	if !bytes.Equal(before, after) {
		t.Fatal("existing file changed")
	}
	if err = exportCSV(path, &out); err == nil {
		t.Fatal("database overwritten")
	}
	sourceAfter, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	afterInfo, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(sourceBefore, sourceAfter) || !sourceInfo.ModTime().Equal(afterInfo.ModTime()) {
		t.Fatal("export modified source database")
	}
}

func TestVersionAndExportFlags(t *testing.T) {
	var out bytes.Buffer
	for _, args := range [][]string{{"--version"}, {"--export-csv", "-"}, {"--export-csv", "applications.csv"}} {
		if _, err := parseOptions(args, &out); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := parseOptions([]string{"--export-csv="}, &out); err == nil || !strings.Contains(err.Error(), "filename") {
		t.Fatal("empty export filename accepted")
	}
	if _, err := parseOptions([]string{"--version", "--export-csv", "-"}, &out); err == nil {
		t.Fatal("conflicting flags accepted")
	}
	if !strings.HasPrefix(versionString(), "jobit ") {
		t.Fatal("missing version")
	}
}

func TestExportDoesNotResetLegacyDatabase(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("XDG_DATA_HOME", root)
	database, err := store.OpenDB()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec("CREATE TABLE job_applications (id INTEGER PRIMARY KEY); INSERT INTO job_applications DEFAULT VALUES"); err != nil {
		t.Fatal(err)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	if err := exportCSV("-", &output); err == nil {
		t.Fatal("export silently reset a legacy database")
	}
	database, err = store.OpenDB()
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	var count int
	if err := database.QueryRow("SELECT COUNT(*) FROM job_applications").Scan(&count); err != nil || count != 1 {
		t.Fatalf("export changed legacy rows: count=%d err=%v", count, err)
	}
}

func TestExportMissingDatabaseDoesNotCreateOne(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("XDG_DATA_HOME", root)
	path, err := store.DefaultDatabasePath()
	if err != nil {
		t.Fatal(err)
	}
	if err := exportCSV("-", &bytes.Buffer{}); err == nil {
		t.Fatal("export accepted missing database")
	}
	if _, err := os.Stat(filepath.Dir(path)); !os.IsNotExist(err) {
		t.Fatal("export created database directory")
	}
}
