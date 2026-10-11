package tui

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/MrMoneyInTheBank/jobit/internal/model"
	"github.com/MrMoneyInTheBank/jobit/internal/store"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func validFields() [fieldCount]string {
	f := defaultFields()
	f[companyField] = " Acme "
	f[positionField] = " Engineer "
	return f
}
func TestFormValidation(t *testing.T) {
	for _, tc := range []struct {
		name  string
		field int
		value string
	}{
		{"company", companyField, " "}, {"position", positionField, ""},
		{"invalid date", dateField, "2026-02-30"}, {"date format", dateField, "11/10/2026"},
		{"missing date", dateField, ""}, {"status", statusField, "pending"}, {"referral", referralField, "maybe"},
		{"work mode", modeField, "office"}, {"rank negative", rankField, "-1"}, {"rank high", rankField, "6"},
		{"rank fractional", rankField, "2.5"}, {"pay negative", minField, "-1"}, {"pay fractional", maxField, "1.5"},
		{"pay comma", minField, "1,000"}, {"pay overflow", maxField, strings.Repeat("9", 100)},
		{"currency length", currencyField, "US"}, {"currency characters", currencyField, "1$!"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := validFields()
			f[tc.field] = tc.value
			_, field, err := applicationFromFields(f)
			if err == nil || field != tc.field {
				t.Fatalf("field=%d error=%v", field, err)
			}
		})
	}
	f := validFields()
	f[minField] = "200"
	f[maxField] = "100"
	if _, field, err := applicationFromFields(f); err == nil || field != maxField {
		t.Fatal("inverted pay range accepted")
	}
}
func TestFormOptionalValues(t *testing.T) {
	f := validFields()
	app, _, err := applicationFromFields(f)
	if err != nil {
		t.Fatal(err)
	}
	if app.Pay != nil || app.Referral || app.Rank != 0 || app.WorkMode != "" {
		t.Fatal("omitted values changed")
	}
	if app.CompanyName != "Acme" || app.Position != "Engineer" {
		t.Fatal("required text not trimmed")
	}
	f[minField] = "0"
	f[currencyField] = " usd "
	app, _, err = applicationFromFields(f)
	if err != nil {
		t.Fatal(err)
	}
	if app.Pay.Min == nil || *app.Pay.Min != 0 || app.Pay.Max != nil || *app.Pay.Currency != "USD" {
		t.Fatal("zero or partial pay lost")
	}
	f[minField] = ""
	f[maxField] = "500"
	app, _, err = applicationFromFields(f)
	if err != nil {
		t.Fatal(err)
	}
	if app.Pay.Min != nil || *app.Pay.Max != 500 {
		t.Fatal("maximum-only pay lost")
	}
	f[maxField] = ""
	app, _, err = applicationFromFields(f)
	if err != nil || app.Pay == nil || app.Pay.Currency == nil {
		t.Fatal("currency-only pay lost")
	}
}
func TestFormKeyboardAndSmallScreens(t *testing.T) {
	j := InitJobList(nil)
	key(&j, "a")
	key(&j, "会社")
	j.Update(tea.KeyMsg{Type: tea.KeyLeft})
	key(&j, "新")
	if j.fields[companyField] != "会新社" {
		t.Fatal("Unicode insertion failed")
	}
	j.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if j.fields[companyField] != "会社" {
		t.Fatal("Unicode deletion failed")
	}
	j.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
	if j.fields[companyField] != "" {
		t.Fatal("clear field failed")
	}
	j.field = statusField
	j.Update(tea.KeyMsg{Type: tea.KeyRight})
	if j.fields[statusField] != "interviewing" {
		t.Fatal("status selection failed")
	}
	j.field = referralField
	j.Update(tea.KeyMsg{Type: tea.KeyRight})
	if j.fields[referralField] != "yes" {
		t.Fatal("referral selection failed")
	}
	j.field = modeField
	j.Update(tea.KeyMsg{Type: tea.KeyRight})
	if j.fields[modeField] != "remote" {
		t.Fatal("mode selection failed")
	}
	j.field = notesField
	j.cursor = 0
	key(&j, "First paragraph")
	key(&j, "enter")
	key(&j, "Second paragraph")
	if j.fields[notesField] != "First paragraph\nSecond paragraph" {
		t.Fatal("notes lost newline")
	}
	for _, size := range [][2]int{{80, 24}, {40, 12}, {24, 9}} {
		j.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		for i := 0; i < fieldCount; i++ {
			j.field = i
			j.cursor = len([]rune(j.fields[i]))
			view := j.View()
			if lipgloss.Width(view) > size[0] || lipgloss.Height(view) > size[1] {
				t.Fatalf("overflow %v field %d", size, i)
			}
			if !strings.Contains(view, "▏") && !strings.Contains(view, "‹") {
				t.Fatalf("cursor hidden %v field %d", size, i)
			}
		}
	}
}
func TestFullFormPersistence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "applications.db")
	db, err := store.OpenDB(&path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { db.Close() }()
	if err = store.InitDB(db); err != nil {
		t.Fatal(err)
	}
	saveCalls := 0
	j := InitJobList(nil, func(app model.JobApplication) (model.StoredJobApplication, error) {
		saveCalls++
		id, err := store.AddJobApplication(db, app)
		return model.StoredJobApplication{ID: id, JobApplication: app}, err
	})
	j.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	key(&j, "a")
	key(&j, "Discard me")
	key(&j, "esc")
	if saveCalls != 0 {
		t.Fatal("cancel wrote a record")
	}
	key(&j, "a")
	input := [fieldCount]string{" Acme ", " Engineer ", " London ", "5", "2026-10-10", "offer", "yes", "hybrid", "0", "200", "gbp", "Notes\nSecond line", "https://example.com/job", "https://example.com"}
	// Type through every field, replacing defaults using Ctrl+U.
	for i, value := range input {
		if j.field != i {
			t.Fatalf("unexpected focus %d", j.field)
		}
		j.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
		if len(formFields[i].choices) > 0 {
			for j.fields[i] != value {
				j.Update(tea.KeyMsg{Type: tea.KeyRight})
			}
		} else if i == notesField {
			key(&j, "Notes")
			key(&j, "enter")
			key(&j, "Second line")
		} else {
			key(&j, value)
		}
		key(&j, "tab")
	}
	expected, _, err := applicationFromFields(input)
	if err != nil {
		t.Fatal(err)
	}
	cmd := key(&j, "ctrl+s")
	if cmd == nil {
		t.Fatalf("save rejected: %s", j.formError)
	}
	if key(&j, "ctrl+s") != nil {
		t.Fatal("duplicate save allowed")
	}
	j.Update(cmd())
	if saveCalls != 1 || j.screen != listScreen || j.selected() == nil {
		t.Fatal("save did not return selected row")
	}
	if !j.selected().JobApplication.Equal(expected) {
		t.Fatal("saved form differs")
	}
	key(&j, "enter")
	if j.screen != detailsScreen {
		t.Fatal("new row details unavailable")
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = store.OpenDB(&path)
	if err != nil {
		t.Fatal(err)
	}
	jobs, err := store.GetJobApplications(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 || !jobs[0].JobApplication.Equal(expected) {
		t.Fatalf("reopened data differs: %+v", jobs)
	}
	reopened := InitJobList(jobs)
	key(&reopened, "enter")
	if reopened.screen != detailsScreen {
		t.Fatal("reopened record details unavailable")
	}
}
