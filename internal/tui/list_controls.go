package tui

import (
	"cmp"
	"fmt"
	"sort"
	"strings"

	"github.com/MrMoneyInTheBank/jobit/internal/model"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

type listControl uint8

const (
	noControl listControl = iota
	sortControl
	filterControl
)

type sortChoice struct{ field, direction string }

var sortChoices = []sortChoice{
	{"Added", "oldest first"}, {"Added", "newest first"},
	{"Company", "A–Z"}, {"Company", "Z–A"},
	{"Position", "A–Z"}, {"Position", "Z–A"},
	{"Location", "A–Z"}, {"Location", "Z–A"},
	{"Ranking", "low–high"}, {"Ranking", "high–low"},
}
var statuses = []model.Status{"", model.StatusSubmitted, model.StatusInterviewing, model.StatusOffer, model.StatusRejected, model.StatusAccepted}

func (jl *JobList) selectedID() int64 {
	if s := jl.selected(); s != nil {
		return s.ID
	}
	return 0
}
func (jl *JobList) hasFilter() bool { return jl.query != "" || jl.status != "" }
func (jl *JobList) rebuild(preferred int64) {
	jl.visible = nil
	query := strings.ToLower(strings.TrimSpace(jl.query))
	for i, stored := range jl.jobs {
		j := stored.JobApplication
		if jl.status != "" && j.Status != jl.status {
			continue
		}
		haystack := strings.ToLower(strings.Join([]string{j.CompanyName, j.Position, j.Location, j.Notes}, " "))
		if query != "" && !strings.Contains(haystack, query) {
			continue
		}
		jl.visible = append(jl.visible, i)
	}
	choice := sortChoices[jl.sortIndex]
	sort.SliceStable(jl.visible, func(i, k int) bool {
		a, b := jl.jobs[jl.visible[i]], jl.jobs[jl.visible[k]]
		var comparison int
		var av, bv string
		switch choice.field {
		case "Added":
			comparison = cmp.Compare(a.ID, b.ID)
		case "Company":
			av, bv = a.JobApplication.CompanyName, b.JobApplication.CompanyName
		case "Position":
			av, bv = a.JobApplication.Position, b.JobApplication.Position
		case "Location":
			av, bv = a.JobApplication.Location, b.JobApplication.Location
		case "Ranking":
			ar, br := a.JobApplication.Rank, b.JobApplication.Rank
			if (ar == 0) != (br == 0) {
				return ar != 0
			}
			comparison = cmp.Compare(ar, br)
		}
		if choice.field != "Added" && choice.field != "Ranking" {
			av, bv = strings.ToLower(strings.TrimSpace(av)), strings.ToLower(strings.TrimSpace(bv))
			if (av == "") != (bv == "") {
				return av != ""
			}
			comparison = strings.Compare(av, bv)
		}
		if comparison == 0 {
			return a.ID < b.ID
		}
		if jl.sortIndex%2 == 1 {
			return comparison > 0
		}
		return comparison < 0
	})
	rows := make([]table.Row, len(jl.visible))
	cursor := 0
	for i, index := range jl.visible {
		rows[i] = constructRow(jl.jobs[index])
		if jl.jobs[index].ID == preferred {
			cursor = i
		}
	}
	jl.JobsTable.SetRows(rows)
	jl.JobsTable.SetCursor(cursor)
	jl.markSelection()
}

func (jl *JobList) updateControl(msg tea.KeyMsg) {
	switch msg.String() {
	case "esc":
		jl.control = noControl
		return
	case "enter":
		id := jl.selectedID()
		if jl.control == sortControl {
			jl.sortIndex = jl.sortDraft
		} else {
			jl.query = strings.TrimSpace(jl.queryDraft)
			jl.status = statuses[jl.statusDraft]
		}
		jl.control = noControl
		jl.notice = ""
		jl.rebuild(id)
		return
	}
	if jl.control == sortControl {
		switch msg.String() {
		case "j", "down", "tab":
			jl.sortDraft = (jl.sortDraft + 1) % len(sortChoices)
		case "k", "up", "shift+tab":
			jl.sortDraft = (jl.sortDraft + len(sortChoices) - 1) % len(sortChoices)
		case "g", "home":
			jl.sortDraft = 0
		case "G", "end":
			jl.sortDraft = len(sortChoices) - 1
		}
		return
	}
	switch msg.String() {
	case "tab", "right":
		jl.statusDraft = (jl.statusDraft + 1) % len(statuses)
	case "shift+tab", "left":
		jl.statusDraft = (jl.statusDraft + len(statuses) - 1) % len(statuses)
	case "backspace", "ctrl+h":
		r := []rune(jl.queryDraft)
		if len(r) > 0 {
			jl.queryDraft = string(r[:len(r)-1])
		}
	case "ctrl+u":
		jl.queryDraft = ""
	default:
		if msg.Type == tea.KeyRunes {
			jl.queryDraft += plain(string(msg.Runes))
		}
		if msg.Type == tea.KeySpace {
			jl.queryDraft += " "
		}
	}
}
func (jl *JobList) controlView(height int) string {
	if jl.control == filterControl {
		status := string(statuses[jl.statusDraft])
		if status == "" {
			status = "All statuses"
		}
		query := jl.queryDraft + "▏"
		query = ansi.TruncateLeft(query, max(0, ansi.StringWidth(query)-jl.contentWidth()+3), "…")
		if height < 7 {
			return strings.Join([]string{query, label.Render("Status: " + status), muted.Render("Tab status · Enter apply")}, "\n")
		}
		return strings.Join([]string{accent.Render("Find an application"), "", label.Render("Search company, position, location, or notes"), query, "", label.Render("Status  ‹ " + status + " ›"), muted.Render("Tab cycles status · Ctrl+U clears search")}, "\n")
	}
	lines := []string{}
	start := max(0, jl.sortDraft-max(1, height-2)/2)
	start = min(start, max(0, len(sortChoices)-height))
	for i := start; i < min(len(sortChoices), start+height); i++ {
		c := sortChoices[i]
		marker := "  "
		if i == jl.sortDraft {
			marker = "› "
		}
		value := fmt.Sprintf("%s%-10s %s", marker, c.field, c.direction)
		if i == jl.sortDraft {
			value = accent.Render(value)
		}
		lines = append(lines, value)
	}
	return strings.Join(lines, "\n")
}

// Keep selection identifiable even when terminal colors are disabled.
func (jl *JobList) markSelection() {
	rows := jl.JobsTable.Rows()
	for i, index := range jl.visible {
		prefix := "  "
		if i == jl.JobsTable.Cursor() {
			prefix = "› "
		}
		rows[i][0] = prefix + plain(jl.jobs[index].JobApplication.CompanyName)
	}
	jl.JobsTable.SetRows(rows)
}
