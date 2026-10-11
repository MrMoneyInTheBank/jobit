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

func (jobList *JobList) selectedID() int64 {
	if selected := jobList.selected(); selected != nil {
		return selected.ID
	}
	return 0
}

func (jobList *JobList) hasFilter() bool {
	return jobList.query != "" || jobList.status != ""
}

func (jobList *JobList) rebuild(preferredID int64) {
	jobList.visible = nil
	query := strings.ToLower(strings.TrimSpace(jobList.query))
	for index, stored := range jobList.jobs {
		application := stored.JobApplication
		if jobList.status != "" && application.Status != jobList.status {
			continue
		}
		haystack := strings.ToLower(strings.Join([]string{
			application.CompanyName, application.Position, application.Location, application.Notes,
		}, " "))
		if query != "" && !strings.Contains(haystack, query) {
			continue
		}
		jobList.visible = append(jobList.visible, index)
	}

	choice := sortChoices[jobList.sortIndex]
	sort.SliceStable(jobList.visible, func(leftIndex, rightIndex int) bool {
		left := jobList.jobs[jobList.visible[leftIndex]]
		right := jobList.jobs[jobList.visible[rightIndex]]
		var comparison int
		var leftValue, rightValue string
		switch choice.field {
		case "Added":
			comparison = cmp.Compare(left.ID, right.ID)
		case "Company":
			leftValue, rightValue = left.JobApplication.CompanyName, right.JobApplication.CompanyName
		case "Position":
			leftValue, rightValue = left.JobApplication.Position, right.JobApplication.Position
		case "Location":
			leftValue, rightValue = left.JobApplication.Location, right.JobApplication.Location
		case "Ranking":
			comparison = cmp.Compare(left.JobApplication.Rank, right.JobApplication.Rank)
		}
		if choice.field != "Added" && choice.field != "Ranking" {
			leftValue = strings.ToLower(strings.TrimSpace(leftValue))
			rightValue = strings.ToLower(strings.TrimSpace(rightValue))
			if (leftValue == "") != (rightValue == "") {
				return leftValue != ""
			}
			comparison = strings.Compare(leftValue, rightValue)
		}
		if comparison == 0 {
			return left.ID < right.ID
		}
		if jobList.sortIndex%2 == 1 {
			return comparison > 0
		}
		return comparison < 0
	})

	rows := make([]table.Row, len(jobList.visible))
	cursor := 0
	for index, jobIndex := range jobList.visible {
		rows[index] = constructRow(jobList.jobs[jobIndex])
		if jobList.jobs[jobIndex].ID == preferredID {
			cursor = index
		}
	}
	jobList.JobsTable.SetRows(rows)
	jobList.JobsTable.SetCursor(cursor)
	jobList.markSelection()
}

func (jobList *JobList) updateControl(msg tea.KeyMsg) {
	switch msg.String() {
	case "esc":
		jobList.control = noControl
		return
	case "enter":
		id := jobList.selectedID()
		if jobList.control == sortControl {
			jobList.sortIndex = jobList.sortDraft
		} else {
			jobList.query = strings.TrimSpace(jobList.queryDraft)
			jobList.status = statuses[jobList.statusDraft]
		}
		jobList.control = noControl
		jobList.notice = ""
		jobList.rebuild(id)
		return
	}
	if jobList.control == sortControl {
		switch msg.String() {
		case "j", "down", "tab":
			jobList.sortDraft = (jobList.sortDraft + 1) % len(sortChoices)
		case "k", "up", "shift+tab":
			jobList.sortDraft = (jobList.sortDraft + len(sortChoices) - 1) % len(sortChoices)
		case "g", "home":
			jobList.sortDraft = 0
		case "G", "end":
			jobList.sortDraft = len(sortChoices) - 1
		}
		return
	}
	switch msg.String() {
	case "tab", "right":
		jobList.statusDraft = (jobList.statusDraft + 1) % len(statuses)
	case "shift+tab", "left":
		jobList.statusDraft = (jobList.statusDraft + len(statuses) - 1) % len(statuses)
	case "backspace", "ctrl+h":
		r := []rune(jobList.queryDraft)
		if len(r) > 0 {
			jobList.queryDraft = string(r[:len(r)-1])
		}
	case "ctrl+u":
		jobList.queryDraft = ""
	default:
		if msg.Type == tea.KeyRunes {
			jobList.queryDraft += plain(string(msg.Runes))
		}
		if msg.Type == tea.KeySpace {
			jobList.queryDraft += " "
		}
	}
}

func (jobList *JobList) controlView(height int) string {
	if jobList.control == filterControl {
		status := string(statuses[jobList.statusDraft])
		if status == "" {
			status = "All statuses"
		}
		query := jobList.queryDraft + "▏"
		query = ansi.TruncateLeft(query, max(0, ansi.StringWidth(query)-jobList.contentWidth()+3), "…")
		if height < 7 {
			return strings.Join([]string{query, label.Render("Status: " + status), muted.Render("Tab status · Enter apply")}, "\n")
		}
		return strings.Join([]string{accent.Render("Find an application"), "", label.Render("Search company, position, location, or notes"), query, "", label.Render("Status  ‹ " + status + " ›"), muted.Render("Tab cycles status · Ctrl+U clears search")}, "\n")
	}
	lines := []string{}
	start := max(0, jobList.sortDraft-max(1, height-2)/2)
	start = min(start, max(0, len(sortChoices)-height))
	for i := start; i < min(len(sortChoices), start+height); i++ {
		c := sortChoices[i]
		marker := "  "
		if i == jobList.sortDraft {
			marker = "› "
		}
		value := fmt.Sprintf("%s%-10s %s", marker, c.field, c.direction)
		if i == jobList.sortDraft {
			value = accent.Render(value)
		}
		lines = append(lines, value)
	}
	return strings.Join(lines, "\n")
}

// Keep selection identifiable even when terminal colors are disabled.
func (jobList *JobList) markSelection() {
	rows := jobList.JobsTable.Rows()
	for i, index := range jobList.visible {
		prefix := "  "
		if i == jobList.JobsTable.Cursor() {
			prefix = "› "
		}
		rows[i][0] = prefix + plain(jobList.jobs[index].JobApplication.CompanyName)
	}
	jobList.JobsTable.SetRows(rows)
}
