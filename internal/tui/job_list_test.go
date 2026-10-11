package tui

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/MrMoneyInTheBank/jobit/internal/model"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func fixture() []model.StoredJobApplication {
	var jobs []model.StoredJobApplication
	for i := 0; i < 30; i++ {
		jobs = append(jobs, model.StoredJobApplication{ID: int64(i + 1), JobApplication: model.JobApplication{CompanyName: fmt.Sprintf("Company %02d", i), Position: "Software engineer", Location: "London", Rank: 4, Status: model.StatusSubmitted, ApplicationDate: time.Date(2026, 10, 11, 0, 0, 0, 0, time.UTC)}})
	}
	return jobs
}
func key(j *JobList, k string) tea.Cmd {
	special := map[string]tea.KeyType{"enter": tea.KeyEnter, "esc": tea.KeyEsc, "tab": tea.KeyTab, "ctrl+s": tea.KeyCtrlS, "ctrl+c": tea.KeyCtrlC}
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
	if v, ok := special[k]; ok {
		msg = tea.KeyMsg{Type: v}
	}
	_, cmd := j.Update(msg)
	return cmd
}
func TestNavigationDetailsAndResize(t *testing.T) {
	j := InitJobList(fixture())
	j.Update(tea.WindowSizeMsg{Width: 90, Height: 24})
	key(&j, "j")
	if j.JobsTable.Cursor() != 1 {
		t.Fatal("j did not move")
	}
	key(&j, "k")
	if j.JobsTable.Cursor() != 0 {
		t.Fatal("k did not move")
	}
	key(&j, "G")
	if j.JobsTable.Cursor() != 29 {
		t.Fatal("G did not reach end")
	}
	key(&j, "enter")
	if j.screen != detailsScreen {
		t.Fatal("details did not open")
	}
	for _, size := range [][2]int{{40, 14}, {120, 32}, {24, 9}, {8, 3}} {
		j.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		view := j.View()
		if lipgloss.Width(view) > size[0] || lipgloss.Height(view) > size[1] {
			t.Fatalf("overflow at %v: %dx%d", size, lipgloss.Width(view), lipgloss.Height(view))
		}
		if j.JobsTable.Cursor() != 29 {
			t.Fatal("resize lost selection")
		}
	}
	key(&j, "esc")
	if j.screen != listScreen || j.JobsTable.Cursor() != 29 {
		t.Fatal("back lost selection")
	}
	key(&j, "g")
	if j.JobsTable.Cursor() != 0 {
		t.Fatal("g did not reach start")
	}
	if _, ok := key(&j, "q")().(tea.QuitMsg); !ok {
		t.Fatal("q did not quit")
	}
}
func TestDetailsContainEveryField(t *testing.T) {
	jobs := fixture()[:1]
	a := &jobs[0].JobApplication
	a.Notes = strings.Repeat("Long notes with 日本語 text. ", 50)
	a.JobPostingLink = "https://example.com/jobs/123"
	a.CompanyWebsiteLink = "https://example.com"
	a.Referral = true
	a.WorkMode = model.WorkModeHybrid
	j := InitJobList(jobs)
	j.Update(tea.WindowSizeMsg{Width: 80, Height: 20})
	key(&j, "enter")
	var seen strings.Builder
	for i := 0; i < 150; i++ {
		seen.WriteString(j.View())
		key(&j, "j")
	}
	for _, value := range []string{"Company 00", "Software engineer", "London", "4/5", "submitted", "11 Oct 2026", "hybrid", "Unknown", "Yes", "Long notes", "https://example.com/jobs/123", "Record ID"} {
		if !strings.Contains(seen.String(), value) {
			t.Errorf("missing detail %q", value)
		}
	}
}
func TestListColumnsAndBounds(t *testing.T) {
	jobs := fixture()
	jobs[0].JobApplication.CompanyName = strings.Repeat("会社", 100) + "\n\x1b[31m"
	j := InitJobList(jobs)
	for _, size := range [][2]int{{120, 32}, {80, 24}, {40, 12}, {24, 9}} {
		j.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		if len(j.JobsTable.Columns()) != 4 || len(j.JobsTable.Rows()[0]) != 4 {
			t.Fatal("unexpected columns")
		}
		v := j.View()
		if lipgloss.Width(v) > size[0] || lipgloss.Height(v) > size[1] {
			t.Fatalf("overflow %v", size)
		}
		if !strings.Contains(v, "q quit") {
			t.Fatalf("missing quit hint at %v", size)
		}
	}
}
func TestEmptyAndForm(t *testing.T) {
	var saved model.JobApplication
	j := InitJobList(nil, func(a model.JobApplication) (model.StoredJobApplication, error) {
		saved = a
		return model.StoredJobApplication{ID: 1, JobApplication: a}, nil
	})
	j.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	if !strings.Contains(j.View(), "first application") {
		t.Fatal("missing empty action")
	}
	key(&j, "enter")
	if j.screen != listScreen {
		t.Fatal("empty details opened")
	}
	key(&j, "a")
	key(&j, "q")
	if !strings.Contains(j.View(), "q▏") {
		t.Fatal("typed text and cursor must be visible")
	}
	if j.fields[0] != "q" {
		t.Fatal("q must type in forms")
	}
	key(&j, "esc")
	if len(j.jobs) != 0 {
		t.Fatal("cancel wrote data")
	}
	key(&j, "a")
	key(&j, "ctrl+s")
	if j.formError == "" {
		t.Fatal("missing validation")
	}
	copy(j.fields[:4], []string{" Acme ", " Engineer ", " Remote ", "9"})
	key(&j, "ctrl+s")
	if j.saving {
		t.Fatal("invalid rank saved")
	}
	j.fields[3] = "5"
	cmd := key(&j, "ctrl+s")
	if cmd == nil {
		t.Fatal("missing save command")
	}
	j.Update(cmd())
	if j.screen != listScreen || len(j.jobs) != 1 || saved.CompanyName != "Acme" || saved.Status != model.StatusSubmitted {
		t.Fatal("save failed")
	}
	key(&j, "enter")
	if !strings.Contains(j.View(), "Acme") {
		t.Fatal("new row not selected")
	}
}
func TestSaveFailureRetainsInput(t *testing.T) {
	j := InitJobList(fixture(), func(model.JobApplication) (model.StoredJobApplication, error) {
		return model.StoredJobApplication{}, errors.New("database locked")
	})
	key(&j, "j")
	key(&j, "a")
	copy(j.fields[:4], []string{"Acme", "Engineer", "", ""})
	cmd := key(&j, "ctrl+s")
	j.Update(cmd())
	if j.screen != addScreen || j.fields[0] != "Acme" || j.formError != "database locked" {
		t.Fatal("failed save lost input")
	}
	key(&j, "esc")
	if j.JobsTable.Cursor() != 1 {
		t.Fatal("cancel lost selection")
	}
}
