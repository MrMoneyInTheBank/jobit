package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPlatformLocations(t *testing.T) {
	root := t.TempDir()
	home := func() (string, error) { return root, nil }
	for _, tc := range []struct{ platform, env, value, want string }{
		{"darwin", "", "", filepath.Join(root, "Library", "Application Support", "jobit", "jobit.db")},
		{"linux", "", "", filepath.Join(root, ".local", "share", "jobit", "jobit.db")},
		{"linux", "XDG_DATA_HOME", filepath.Join(root, "data"), filepath.Join(root, "data", "jobit", "jobit.db")},
		{"linux", "XDG_DATA_HOME", "relative", filepath.Join(root, ".local", "share", "jobit", "jobit.db")},
	} {
		actual, err := databasePath(tc.platform, func(key string) string {
			if key == tc.env {
				return tc.value
			}
			return ""
		}, home)
		if err != nil || actual != tc.want {
			t.Fatalf("%s: %q %v, want %q", tc.platform, actual, err, tc.want)
		}
	}
}

func TestDefaultDatabaseIndependentOfWorkingDirectory(t *testing.T) {
	root := t.TempDir()
	for _, key := range []string{"HOME", "XDG_DATA_HOME"} {
		t.Setenv(key, root)
	}
	path, err := DefaultDatabasePath()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(filepath.Dir(path)); !os.IsNotExist(err) {
		t.Fatal("resolving location created data")
	}
	db, err := OpenDB()
	if err != nil {
		t.Fatal(err)
	}
	if err = InitDB(db); err != nil {
		t.Fatal(err)
	}
	id, err := AddJobApplication(db, testApp)
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
	t.Chdir(t.TempDir())
	db, err = OpenDB()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err = GetJobApplicationByID(db, id); err != nil {
		t.Fatal("working directory changed database:", err)
	}
}
