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
	path := filepath.Join(t.TempDir(), "data.db")
	t.Setenv("JOBIT_DB_DSN", path)
	db, err := store.OpenDB(nil)
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
}

func TestVersionAndExportFlags(t *testing.T) {
	var out bytes.Buffer
	for _, args := range [][]string{{"--version"}, {"--export-csv", "-"}, {"--export-csv", "applications.csv"}} {
		if _, err := parseOptions(args, &out); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := parseOptions([]string{"--version", "--export-csv", "-"}, &out); err == nil {
		t.Fatal("conflicting flags accepted")
	}
	if !strings.HasPrefix(versionString(), "jobit ") {
		t.Fatal("missing version")
	}
}
