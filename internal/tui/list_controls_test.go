package tui

import (
	"reflect"
	"strings"
	"testing"

	"github.com/MrMoneyInTheBank/jobit/internal/model"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func controlJobs() []model.StoredJobApplication {
	names := []string{"beta", "Acme", "acme", "Gamma"}
	var result []model.StoredJobApplication
	for i, name := range names {
		result = append(result, model.StoredJobApplication{ID: int64(i + 1), JobApplication: model.JobApplication{CompanyName: name, Position: "Engineer", Status: model.StatusSubmitted, Rank: model.RankOne}})
	}
	result[0].JobApplication.Location = "London"
	result[0].JobApplication.Rank = 3
	result[1].JobApplication.Location = "Paris"
	result[1].JobApplication.Rank = 5
	result[2].JobApplication.Status = model.StatusInterviewing
	return result
}

func visibleIDs(j *JobList) []int64 {
	ids := []int64{}
	for _, i := range j.visible {
		ids = append(ids, j.jobs[i].ID)
	}
	return ids
}

func TestSortDeterminismAndSelection(t *testing.T) {
	original := controlJobs()
	j := InitJobList(original)
	key(&j, "j")
	for _, tc := range []struct {
		sort int
		want []int64
	}{
		{2, []int64{2, 3, 1, 4}}, {3, []int64{4, 1, 2, 3}},
		{6, []int64{1, 2, 3, 4}}, {7, []int64{2, 1, 3, 4}},
		{8, []int64{3, 4, 1, 2}}, {9, []int64{2, 1, 3, 4}},
		{1, []int64{4, 3, 2, 1}},
	} {
		j.sortIndex = tc.sort
		j.rebuild(2)
		if !reflect.DeepEqual(visibleIDs(&j), tc.want) {
			t.Fatalf("sort %d: %v", tc.sort, visibleIDs(&j))
		}
		if j.selected().ID != 2 {
			t.Fatal("sort lost selected ID")
		}
	}
	if !reflect.DeepEqual(original, j.jobs) {
		t.Fatal("sort mutated source records")
	}
}

func TestFilterApplyCancelAndCombinedSort(t *testing.T) {
	j := InitJobList(controlJobs())
	j.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	key(&j, "/")
	key(&j, "AcMe")
	key(&j, "enter")
	if !reflect.DeepEqual(visibleIDs(&j), []int64{2, 3}) {
		t.Fatal("case insensitive match failed")
	}
	key(&j, "/")
	key(&j, "zz")
	key(&j, "esc")
	if j.query != "AcMe" || len(j.visible) != 2 {
		t.Fatal("cancel applied draft")
	}
	key(&j, "s")
	j.sortDraft = 3
	key(&j, "enter")
	if !reflect.DeepEqual(visibleIDs(&j), []int64{2, 3}) {
		t.Fatal("tie ordering changed")
	}
	key(&j, "/")
	key(&j, "tab")
	key(&j, "tab")
	key(&j, "enter")
	if !reflect.DeepEqual(visibleIDs(&j), []int64{3}) {
		t.Fatal("status and text must intersect")
	}
	key(&j, "enter")
	key(&j, "esc")
	if !j.hasFilter() || j.sortIndex != 3 || j.selected().ID != 3 {
		t.Fatal("details lost settings")
	}
	key(&j, "esc")
	if j.hasFilter() || len(j.visible) != 4 || j.selected().ID != 3 {
		t.Fatal("clear lost selection")
	}
}

func TestNoMatchesAndEmptyControls(t *testing.T) {
	for _, jobs := range [][]model.StoredJobApplication{nil, controlJobs()} {
		j := InitJobList(jobs)
		j.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		key(&j, "/")
		key(&j, "no-match")
		key(&j, "enter")
		key(&j, "enter")
		if j.screen != listScreen || j.selected() != nil {
			t.Fatal("empty result opened details")
		}
		if len(jobs) > 0 && !strings.Contains(j.View(), "No matches") {
			t.Fatal("missing no matches state")
		}
		key(&j, "s")
		key(&j, "G")
		key(&j, "enter")
		if len(j.visible) != 0 {
			t.Fatal("sort removed filter")
		}
	}
}

func TestSavedRowRespectsFilters(t *testing.T) {
	j := InitJobList(controlJobs())
	j.query = "Acme"
	j.rebuild(0)
	j.sortIndex = 3
	j.Update(savedMsg{application: model.StoredJobApplication{ID: 99, JobApplication: model.JobApplication{CompanyName: "Elsewhere"}}})
	if len(j.visible) != 2 || !strings.Contains(j.notice, "hidden by filter") {
		t.Fatal("hidden save not explained")
	}
	j.Update(savedMsg{application: model.StoredJobApplication{ID: 100, JobApplication: model.JobApplication{CompanyName: "Acme"}}})
	if j.selected().ID != 100 || j.query != "Acme" {
		t.Fatal("matching save did not select new record")
	}
}

func TestControlLayoutAndTextFields(t *testing.T) {
	jobs := controlJobs()
	jobs[3].JobApplication.Notes = "Recruiter referral"
	j := InitJobList(jobs)
	for _, query := range []string{"ENGINEER", "london", "recruiter"} {
		j.query = query
		j.rebuild(0)
		if len(j.visible) == 0 {
			t.Fatalf("no match for %q", query)
		}
	}
	for _, size := range [][2]int{{80, 24}, {40, 14}, {24, 9}} {
		j.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		for _, control := range []listControl{sortControl, filterControl} {
			j.control = control
			j.queryDraft = "query"
			if control == sortControl {
				j.sortDraft = 9
			}
			view := j.View()
			if lipgloss.Width(view) > size[0] || lipgloss.Height(view) > size[1] {
				t.Fatalf("control overflow at %v", size)
			}
			if control == filterControl && !strings.Contains(view, "query") {
				t.Fatalf("query hidden at %v", size)
			}
		}
	}
}
