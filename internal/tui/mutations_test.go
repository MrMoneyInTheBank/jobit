package tui

import (
	"errors"
	"testing"

	"github.com/MrMoneyInTheBank/jobit/internal/model"
	"github.com/MrMoneyInTheBank/jobit/internal/store"
	tea "github.com/charmbracelet/bubbletea"
)

func TestEditDeletePersistence(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("XDG_DATA_HOME", root)
	db, err := store.OpenDB()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { db.Close() }()
	if err = store.InitDB(db); err != nil {
		t.Fatal(err)
	}
	f := validFields()
	f[minField] = "0"
	f[maxField] = "100"
	f[currencyField] = "USD"
	f[notesField] = "old notes"
	original, _, err := applicationFromFields(f)
	if err != nil {
		t.Fatal(err)
	}
	id, err := store.AddJobApplication(db, original)
	if err != nil {
		t.Fatal(err)
	}
	jobs, err := store.GetJobApplications(db)
	if err != nil {
		t.Fatal(err)
	}
	j := InitJobList(jobs)
	j.SetMutationHandlers(func(app model.StoredJobApplication) error {
		return store.PatchJobApplication(db, app, model.JobApplicationPatch{})
	}, func(id int64) error { return store.DeleteJobApplicationByID(db, id) })
	key(&j, "enter")
	key(&j, "e")
	if j.editID != id || j.fields[minField] != "0" || j.fields[notesField] != "old notes" {
		t.Fatal("edit not prefilled")
	}
	j.field = notesField
	j.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
	key(&j, "cancelled")
	key(&j, "esc")
	if j.screen != detailsScreen || j.jobs[0].JobApplication.Notes != "old notes" {
		t.Fatal("cancel changed record or screen")
	}
	key(&j, "u")
	if j.field != statusField {
		t.Fatal("status shortcut did not focus status")
	}
	j.Update(tea.KeyMsg{Type: tea.KeyRight})
	// Edit other fields and clear all pay info.
	j.fields[positionField] = "Senior engineer"
	j.fields[notesField] = ""
	j.fields[minField] = ""
	j.fields[maxField] = ""
	j.fields[currencyField] = ""
	cmd := key(&j, "ctrl+s")
	if cmd == nil {
		t.Fatalf("update rejected: %s", j.formError)
	}
	j.Update(cmd())
	if len(j.jobs) != 1 || j.jobs[0].ID != id {
		t.Fatal("update created duplicate")
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = store.OpenDB()
	if err != nil {
		t.Fatal(err)
	}
	actual, err := store.GetJobApplicationByID(db, id)
	if err != nil {
		t.Fatal(err)
	}
	if actual.JobApplication.Status != model.StatusInterviewing || actual.JobApplication.Position != "Senior engineer" || actual.JobApplication.Pay != nil || actual.JobApplication.Notes != "" {
		t.Fatalf("update not persisted: %+v", actual)
	}
	key(&j, "d")
	if j.screen != deleteScreen {
		t.Fatal("missing delete confirmation")
	}
	key(&j, "esc")
	if _, err = store.GetJobApplicationByID(db, id); err != nil {
		t.Fatal("cancel deleted record")
	}
	key(&j, "d")
	cmd = key(&j, "enter")
	if cmd == nil {
		t.Fatal("no delete command")
	}
	if key(&j, "enter") != nil {
		t.Fatal("duplicate delete allowed")
	}
	j.Update(cmd())
	if len(j.jobs) != 0 || j.screen != listScreen {
		t.Fatal("deleted record still displayed")
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = store.OpenDB()
	if err != nil {
		t.Fatal(err)
	}
	remaining, err := store.GetJobApplications(db)
	if err != nil || len(remaining) != 0 {
		t.Fatal("delete not persisted")
	}
	for _, k := range []string{"e", "u", "d"} {
		key(&j, k)
		if j.screen != listScreen {
			t.Fatal("empty list mutation opened")
		}
	}
}

func TestMutationErrorsAndFilters(t *testing.T) {
	jobs := controlJobs()
	j := InitJobList(jobs)
	fail := true
	j.SetMutationHandlers(func(model.StoredJobApplication) error {
		if fail {
			return errors.New("update failed")
		}
		return nil
	}, func(int64) error { return errors.New("delete failed") })
	j.query = "acme"
	j.status = model.StatusSubmitted
	j.sortIndex = 3
	j.rebuild(2)
	key(&j, "u")
	j.Update(tea.KeyMsg{Type: tea.KeyRight})
	// Existing fixture has no date; enter a valid date to exercise repository errors.
	j.fields[dateField] = "2026-10-11"
	cmd := key(&j, "ctrl+s")
	if cmd == nil {
		t.Fatalf("unexpected validation: %s", j.formError)
	}
	j.Update(cmd())
	if j.screen != addScreen || j.jobs[1].JobApplication.Status != model.StatusSubmitted {
		t.Fatal("failed update changed model")
	}
	fail = false
	cmd = key(&j, "ctrl+s")
	j.Update(cmd())
	if len(j.visible) != 0 || !j.hasFilter() || j.sortIndex != 3 {
		t.Fatal("update did not preserve active filter")
	}
	key(&j, "esc")
	key(&j, "d")
	cmd = key(&j, "enter")
	j.Update(cmd())
	if j.screen != deleteScreen || len(j.jobs) != len(jobs) || j.notice == "" {
		t.Fatal("delete failure lost record")
	}
	key(&j, "esc")
	if j.screen != listScreen {
		t.Fatal("delete failure could not cancel")
	}
}
