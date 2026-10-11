package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/MrMoneyInTheBank/jobit/internal/model"
	"github.com/MrMoneyInTheBank/jobit/internal/store"
)

func buildReleaseBinary(t *testing.T) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "jobit")
	command := exec.Command("go", "build", "-o", binary, ".")
	command.Env = append(os.Environ(), "CGO_ENABLED=1")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build executable: %v\n%s", err, output)
	}
	return binary
}

func isolatedDatabasePath(home string) string {
	if runtime.GOOS == "darwin" {
		return filepath.Join(home, "Library", "Application Support", "jobit", "jobit.db")
	}
	return filepath.Join(home, "jobit", "jobit.db")
}

func isolatedEnvironment(home string) []string {
	return append(os.Environ(), "HOME="+home, "XDG_DATA_HOME="+home, "TERM=dumb")
}

func fixtureDatabase(t *testing.T, home string) *sql.DB {
	t.Helper()
	path := isolatedDatabasePath(home)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	database, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	return database
}

// runInTerminal exercises the actual executable and Bubble Tea startup. TERM=dumb
// avoids terminal capability queries that require an interactive responder.
func runInTerminal(t *testing.T, binary, home string, keystrokes ...string) (string, error) {
	t.Helper()
	if _, err := exec.LookPath("script"); err != nil {
		t.Fatalf("terminal smoke test requires script: %v", err)
	}
	context, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var command *exec.Cmd
	if runtime.GOOS == "darwin" {
		command = exec.CommandContext(context, "script", "-q", "/dev/null", "/bin/sh", "-c", `stty rows 24 cols 80; exec "$1"`, "jobit", binary)
	} else {
		quotedBinary := "'" + strings.ReplaceAll(binary, "'", `'\''`) + "'"
		command = exec.CommandContext(context, "script", "-q", "-c", "stty rows 24 cols 80; exec "+quotedBinary, "/dev/null")
	}
	command.Env = isolatedEnvironment(home)
	inputReader, inputWriter := io.Pipe()
	command.Stdin = inputReader
	defer inputReader.Close()
	go func() {
		time.Sleep(500 * time.Millisecond)
		if len(keystrokes) == 0 {
			keystrokes = []string{"q"}
		}
		for _, keys := range keystrokes {
			_, _ = io.WriteString(inputWriter, keys)
			time.Sleep(300 * time.Millisecond)
		}
		_ = inputWriter.Close()
	}()
	output, err := command.CombinedOutput()
	if context.Err() != nil {
		t.Fatalf("TUI did not exit after q: %v\n%s", context.Err(), output)
	}
	return string(output), err
}

func runCLI(t *testing.T, binary, home string, args ...string) (string, string, error) {
	t.Helper()
	command := exec.Command(binary, args...)
	command.Env = isolatedEnvironment(home)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	return stdout.String(), stderr.String(), err
}

func TestReleaseExecutableStartup(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("Jobit supports macOS and Linux")
	}
	binary := buildReleaseBinary(t)
	t.Run("fresh database", func(t *testing.T) {
		home := t.TempDir()
		output, err := runInTerminal(t, binary, home)
		if err != nil || !strings.Contains(output, "JOBIT") {
			t.Fatalf("fresh TUI failed: %v\n%s", err, output)
		}
		database := fixtureDatabase(t, home)
		if _, err := store.GetJobApplications(database); err != nil {
			t.Fatal("fresh database cannot be read:", err)
		}
	})

	t.Run("current database with applications", func(t *testing.T) {
		home := t.TempDir()
		database := fixtureDatabase(t, home)
		if err := store.InitDB(database); err != nil {
			t.Fatal(err)
		}
		application := model.JobApplication{CompanyName: "Acme", Position: "Engineer", ApplicationDate: time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC), Status: model.StatusSubmitted, Rank: model.RankThree}
		id, err := store.AddJobApplication(database, application)
		if err != nil {
			t.Fatal(err)
		}
		output, err := runInTerminal(t, binary, home)
		if err != nil || !strings.Contains(output, "Acme") {
			t.Fatalf("existing data did not reach TUI: %v\n%s", err, output)
		}
		stored, err := store.GetJobApplicationByID(database, id)
		if err != nil || !stored.JobApplication.Equal(application) {
			t.Fatalf("startup changed application: %+v %v", stored, err)
		}
	})

	t.Run("create application through terminal", func(t *testing.T) {
		home := t.TempDir()
		output, err := runInTerminal(t, binary, home, "a", "Acme", "\tEngineer\t\t3\x13", "q")
		if err != nil || !strings.Contains(output, "Acme") {
			t.Fatalf("form did not save through TUI: %v\n%s", err, output)
		}
		database := fixtureDatabase(t, home)
		applications, err := store.GetJobApplications(database)
		if err != nil || len(applications) != 1 {
			t.Fatalf("terminal save did not persist: %+v %v", applications, err)
		}
		if applications[0].JobApplication.CompanyName != "Acme" || applications[0].JobApplication.Position != "Engineer" || applications[0].JobApplication.Rank != model.RankThree {
			t.Fatalf("terminal saved incorrect fields: %+v", applications[0])
		}
		output, err = runInTerminal(t, binary, home)
		if err != nil || !strings.Contains(output, "Acme") {
			t.Fatalf("saved application did not appear after restart: %v\n%s", err, output)
		}
		output, err = runInTerminal(t, binary, home, "\r", "e", "\t\x15Senior Engineer\x13", "d", "\r", "q")
		if err != nil || !strings.Contains(output, "Senior Engineer") || !strings.Contains(output, "deleted") {
			t.Fatalf("edit/delete through TUI failed: %v\n%s", err, output)
		}
		applications, err = store.GetJobApplications(database)
		if err != nil || len(applications) != 0 {
			t.Fatalf("terminal delete did not persist: %+v %v", applications, err)
		}
		output, err = runInTerminal(t, binary, home)
		if err != nil || !strings.Contains(output, "No applications yet") {
			t.Fatalf("deleted application reappeared after restart: %v\n%s", err, output)
		}
	})

}

func TestReleaseExecutableFlagsAndExport(t *testing.T) {
	binary := buildReleaseBinary(t)
	home := t.TempDir()
	path := isolatedDatabasePath(home)
	for _, arguments := range [][]string{{"--help"}, {"--version"}} {
		stdout, stderr, err := runCLI(t, binary, home, arguments...)
		if err != nil || stderr != "" || !strings.Contains(stdout, "╭") {
			t.Fatalf("%v: stdout=%q stderr=%q err=%v", arguments, stdout, stderr, err)
		}
	}
	stdout, stderr, err := runCLI(t, binary, home, "--export-csv=")
	if err == nil || stdout != "" || !strings.Contains(stderr, "requires a filename") {
		t.Fatalf("empty export flag launched app: stdout=%q stderr=%q err=%v", stdout, stderr, err)
	}
	if _, err := os.Stat(filepath.Dir(path)); !os.IsNotExist(err) {
		t.Fatal("help, version, or invalid flag created data")
	}
	stdout, stderr, err = runCLI(t, binary, home, "--export-csv", "-")
	if err == nil || stdout != "" || !strings.Contains(stderr, "open database") {
		t.Fatalf("missing database export: stdout=%q stderr=%q err=%v", stdout, stderr, err)
	}
	if _, err := os.Stat(filepath.Dir(path)); !os.IsNotExist(err) {
		t.Fatal("missing database export created data")
	}

	database := fixtureDatabase(t, home)
	if err := store.InitDB(database); err != nil {
		t.Fatal(err)
	}
	application := model.JobApplication{CompanyName: "Acme, Inc.", Position: "Engineer", ApplicationDate: time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC), Status: model.StatusSubmitted, Rank: model.RankThree, Notes: "line one\nline two"}
	if _, err := store.AddJobApplication(database, application); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	stdout, stderr, err = runCLI(t, binary, home, "--export-csv", "-")
	if err != nil || stderr != "" {
		t.Fatalf("CSV stdout failed: stderr=%q err=%v", stderr, err)
	}
	records, err := csv.NewReader(strings.NewReader(stdout)).ReadAll()
	if err != nil || len(records) != 2 || records[1][1] != application.CompanyName || records[1][12] != application.Notes {
		t.Fatalf("CSV output: records=%v err=%v", records, err)
	}
	exportPath := filepath.Join(t.TempDir(), "applications.csv")
	stdout, stderr, err = runCLI(t, binary, home, "--export-csv", exportPath)
	if err != nil || stdout != "" || stderr != "" {
		t.Fatalf("CSV file failed: stdout=%q stderr=%q err=%v", stdout, stderr, err)
	}
	if _, _, err := runCLI(t, binary, home, "--export-csv", exportPath); err == nil {
		t.Fatal("export overwrote existing file")
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("export changed source database: %v", err)
	}
}
