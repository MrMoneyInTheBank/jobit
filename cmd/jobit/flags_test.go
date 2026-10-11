package main

import (
	"bytes"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MrMoneyInTheBank/jobit/internal/store"
)

func TestHelpWithoutDatabaseSideEffects(t *testing.T) {
	root := t.TempDir()
	for _, key := range []string{"HOME", "USERPROFILE", "LOCALAPPDATA", "XDG_DATA_HOME"} {
		t.Setenv(key, root)
	}
	t.Setenv("JOBIT_DB_DSN", "")
	path, err := store.DefaultDatabasePath()
	if err != nil {
		t.Fatal(err)
	}
	for _, arg := range []string{"--help", "-h"} {
		var out bytes.Buffer
		if err := parseArgs([]string{arg}, &out); !errors.Is(err, flag.ErrHelp) {
			t.Fatal(err)
		}
		if !strings.Contains(out.String(), path) {
			t.Fatal("help missing location")
		}
	}
	if _, err := os.Stat(filepath.Dir(path)); !os.IsNotExist(err) {
		t.Fatal("help created data")
	}
	t.Setenv("JOBIT_DB_DSN", "./custom.db")
	var out bytes.Buffer
	_ = parseArgs([]string{"--help"}, &out)
	if !strings.Contains(out.String(), "Effective database: ./custom.db") {
		t.Fatal("override not reported")
	}
	for _, args := range [][]string{{"--unknown"}, {"unexpected"}} {
		if err := parseArgs(args, &out); err == nil {
			t.Fatal("invalid arguments accepted")
		}
	}
}
