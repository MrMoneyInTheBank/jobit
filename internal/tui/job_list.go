package tui

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/MrMoneyInTheBank/jobit/internal/model"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type screen uint8

const (
	listScreen screen = iota
	detailsScreen
	addScreen
	deleteScreen
)

type SaveApplication func(model.JobApplication) (model.StoredJobApplication, error)
type savedMsg struct {
	application model.StoredJobApplication
	err         error
}

type pendingQuitMsg struct{}

type JobList struct {
	quitBlocked          bool
	update               UpdateApplication
	remove               DeleteApplication
	editID               int64
	returnScreen         screen
	deleting             bool
	JobsTable            table.Model
	jobs                 []model.StoredJobApplication
	width, height        int
	screen               screen
	details              viewport.Model
	save                 SaveApplication
	fields               [fieldCount]string
	cursor               int
	field                int
	formError            string
	saving               bool
	visible              []int
	sortIndex, sortDraft int
	query, queryDraft    string
	status               model.Status
	statusDraft          int
	control              listControl
	notice               string
}

var (
	accent = lipgloss.NewStyle().Foreground(lipgloss.Color("79")).Bold(true)
	muted  = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	label  = lipgloss.NewStyle().Foreground(lipgloss.Color("110"))
)

func (jobList *JobList) Init() tea.Cmd {
	return tea.SetWindowTitle("Jobit · Applications")
}

// plain prevents stored control characters from affecting terminal rendering.
func plain(value string) string {
	return strings.Map(func(character rune) rune {
		if unicode.IsControl(character) {
			return ' '
		}
		return character
	}, ansi.Strip(value))
}

func optional(value string) string {
	if value == "" {
		return "—"
	}
	return plain(value)
}

func ranking(rank model.Ranking) string {
	return strconv.Itoa(int(rank)) + "/5"
}

func constructRow(stored model.StoredJobApplication) table.Row {
	application := stored.JobApplication
	return table.Row{
		plain(application.CompanyName), plain(application.Position),
		optional(application.Location), ranking(application.Rank),
	}
}

// The optional save callback keeps persistence outside the view model.
func InitJobList(jobs []model.StoredJobApplication, save ...SaveApplication) JobList {
	rows := make([]table.Row, len(jobs))
	for index, application := range jobs {
		rows[index] = constructRow(application)
	}
	styles := table.DefaultStyles()
	styles.Header = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Padding(0, 1)
	styles.Cell = lipgloss.NewStyle().Padding(0, 1)
	styles.Selected = lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("24")).Bold(true)
	jobTable := table.New(table.WithColumns(columns(72)), table.WithRows(rows), table.WithFocused(true), table.WithStyles(styles), table.WithHeight(10))
	jobList := JobList{JobsTable: jobTable, jobs: append([]model.StoredJobApplication(nil), jobs...), details: viewport.New(72, 10)}
	if len(save) > 0 {
		jobList.save = save[0]
	}
	jobList.rebuild(0)
	return jobList
}

func columns(width int) []table.Column {
	available := max(4, width-8) // Two padding cells per column.
	rank := min(7, max(1, available/5))
	remaining := available - rank
	company := max(1, remaining*30/100)
	position := max(1, remaining*42/100)
	location := max(1, remaining-company-position)
	titles := []string{"Company", "Position", "Location", "Ranking"}
	if width < 48 {
		titles = []string{"Co.", "Role", "Place", "Rank"}
	}
	return []table.Column{{Title: titles[0], Width: company}, {Title: titles[1], Width: position}, {Title: titles[2], Width: location}, {Title: titles[3], Width: rank}}
}

func (jobList *JobList) contentWidth() int {
	return max(1, min(112, jobList.width-4))
}

func (jobList *JobList) resize() {
	contentWidth := jobList.contentWidth()
	contentHeight := max(1, jobList.height-8)
	jobList.JobsTable.SetColumns(columns(contentWidth))
	jobList.JobsTable.SetWidth(contentWidth)
	jobList.JobsTable.SetHeight(contentHeight)
	progress := jobList.details.ScrollPercent()
	jobList.details.Width = contentWidth
	jobList.details.Height = contentHeight
	if jobList.screen == detailsScreen {
		jobList.setDetails()
		jobList.details.SetYOffset(int(progress * float64(max(0, jobList.details.TotalLineCount()-jobList.details.Height))))
	}
}

func (jobList *JobList) selected() *model.StoredJobApplication {
	cursor := jobList.JobsTable.Cursor()
	if cursor < 0 || cursor >= len(jobList.visible) {
		return nil
	}
	return &jobList.jobs[jobList.visible[cursor]]
}

func (jobList *JobList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		jobList.width = msg.Width
		jobList.height = msg.Height
		jobList.resize()
		return jobList, nil
	case pendingQuitMsg:
		jobList.quitBlocked = jobList.Busy()
		return jobList, nil
	case savedMsg:
		jobList.quitBlocked = false
		jobList.saving = false
		if msg.err != nil {
			jobList.formError = msg.err.Error()
			return jobList, nil
		}
		updated := false
		for index := range jobList.jobs {
			if jobList.jobs[index].ID == msg.application.ID {
				jobList.jobs[index] = msg.application
				updated = true
				break
			}
		}
		if !updated {
			jobList.jobs = append(jobList.jobs, msg.application)
		}
		jobList.rebuild(msg.application.ID)
		jobList.notice = "Application saved"
		if updated {
			jobList.notice = "Application updated"
		}
		if selected := jobList.selected(); selected == nil || selected.ID != msg.application.ID {
			jobList.notice = "Saved · hidden by filter; Esc clears it"
		}
		jobList.screen = listScreen
		return jobList, nil
	case deletedMsg:
		jobList.quitBlocked = false
		jobList.deleting = false
		if msg.err != nil {
			jobList.notice = "Delete failed: " + msg.err.Error()
			return jobList, nil
		}
		cursor := jobList.JobsTable.Cursor()
		for index, application := range jobList.jobs {
			if application.ID == msg.id {
				jobList.jobs = append(jobList.jobs[:index], jobList.jobs[index+1:]...)
				break
			}
		}
		jobList.rebuild(0)
		jobList.JobsTable.SetCursor(min(cursor, max(0, len(jobList.visible)-1)))
		jobList.markSelection()
		jobList.screen = listScreen
		jobList.notice = "Application deleted"
		return jobList, nil
	case tea.KeyMsg:
		key := msg.String()
		if key == "ctrl+c" {
			if jobList.Busy() {
				jobList.quitBlocked = true
				return jobList, nil
			}
			return jobList, tea.Quit
		}
		if jobList.screen == deleteScreen {
			return jobList, jobList.updateDelete(msg)
		}
		if jobList.screen == addScreen {
			return jobList, jobList.updateForm(msg)
		}
		if jobList.screen == listScreen && jobList.control != noControl {
			jobList.updateControl(msg)
			return jobList, nil
		}
		if key == "e" || key == "u" {
			jobList.beginEdit(key == "u")
			return jobList, nil
		}
		if key == "d" && jobList.selected() != nil {
			if jobList.remove == nil {
				jobList.notice = "Deletion is unavailable"
				return jobList, nil
			}
			jobList.returnScreen = jobList.screen
			jobList.screen = deleteScreen
			jobList.notice = ""
			return jobList, nil
		}
		if key == "q" {
			if jobList.Busy() {
				jobList.quitBlocked = true
				return jobList, nil
			}
			return jobList, tea.Quit
		}
		if key == "esc" {
			if jobList.screen == listScreen {
				id := jobList.selectedID()
				jobList.query = ""
				jobList.status = ""
				jobList.notice = ""
				jobList.rebuild(id)
			}
			jobList.screen = listScreen
			return jobList, nil
		}
		if jobList.screen == detailsScreen {
			switch key {
			case "g", "home":
				jobList.details.GotoTop()
				return jobList, nil
			case "G", "end":
				jobList.details.GotoBottom()
				return jobList, nil
			}
		}
		if jobList.screen == listScreen {
			if key == "s" {
				jobList.control = sortControl
				jobList.sortDraft = jobList.sortIndex
				return jobList, nil
			}
			if key == "/" {
				jobList.control = filterControl
				jobList.queryDraft = jobList.query
				jobList.statusDraft = 0
				for i, status := range statuses {
					if status == jobList.status {
						jobList.statusDraft = i
					}
				}
				return jobList, nil
			}
			if key == "enter" && jobList.selected() != nil {
				jobList.screen = detailsScreen
				jobList.setDetails()
				jobList.details.GotoTop()
				return jobList, nil
			}
			if key == "a" {
				jobList.screen = addScreen
				jobList.editID = 0
				jobList.returnScreen = listScreen
				jobList.fields = defaultFields()
				jobList.cursor = 0
				jobList.field = 0
				jobList.formError = ""
				return jobList, nil
			}
		}
	}
	var cmd tea.Cmd
	if jobList.screen == listScreen {
		jobList.JobsTable, cmd = jobList.JobsTable.Update(msg)
		jobList.markSelection()
	}
	if jobList.screen == detailsScreen {
		jobList.details, cmd = jobList.details.Update(msg)
	}
	return jobList, cmd
}

func (jobList *JobList) View() string {
	if jobList.width == 0 || jobList.height == 0 {
		return ""
	}
	if jobList.quitBlocked && jobList.Busy() {
		return jobList.pendingPopup()
	}
	if jobList.width < 24 || jobList.height < 9 {
		return ansi.Truncate("Resize to 24 × 9 · Ctrl+C quits", max(1, jobList.width), "…")
	}
	contentWidth := jobList.contentWidth()
	contentHeight := max(1, jobList.height-8)
	title := "Applications"
	help := "j/k ↵ details  a add  e edit  u status  d delete  s sort  / filter  q quit"
	subtitle := fmt.Sprintf("%d of %d applications", len(jobList.visible), len(jobList.jobs))
	if len(jobList.visible) > 0 {
		subtitle += fmt.Sprintf("  ·  row %d", jobList.JobsTable.Cursor()+1)
	}
	choice := sortChoices[jobList.sortIndex]
	toolbar := choice.field + " · " + choice.direction
	if jobList.hasFilter() {
		toolbar += "  |  "
		if jobList.query != "" {
			toolbar += "“" + jobList.query + "” "
		}
		if jobList.status != "" {
			toolbar += string(jobList.status)
		}
	}
	body := jobList.JobsTable.View()
	if len(jobList.jobs) == 0 {
		body = accent.Render("A fresh start.") + "\nNo applications yet.\nPress a to add your first application."
	} else if len(jobList.visible) == 0 {
		body = accent.Render("No matches.") + "\nTry a different search or status.\nPress Esc to clear filters."
	}
	if jobList.screen == detailsScreen {
		title = "Application"
		subtitle = plain(jobList.selected().JobApplication.CompanyName) + " · " + plain(jobList.selected().JobApplication.Position)
		toolbar = fmt.Sprintf("DETAILS  ·  %.0f%% scrolled", jobList.details.ScrollPercent()*100)
		body = jobList.details.View()
		help = "j/k scroll  g/G top/end  e edit  u status  d delete  esc back  q quit"
	}
	if jobList.screen == addScreen {
		title = "New application"
		if jobList.editID != 0 {
			title = "Edit application"
		}
		subtitle = fmt.Sprintf("Field %d of %d · * required", jobList.field+1, fieldCount)
		toolbar = formFields[jobList.field].hint
		if jobList.formError != "" {
			toolbar = "Error: " + jobList.formError
		}
		if jobList.saving {
			toolbar = "Saving…"
		}
		body = jobList.formView()
		help = "tab next   ctrl+s save   esc discard"
	}
	if jobList.screen == deleteScreen {
		title = "Delete application?"
		subtitle = "This permanently removes the selected record."
		toolbar = "Enter confirms · Esc cancels"
		body = ansi.Hardwrap(plain(jobList.selected().JobApplication.CompanyName)+"\n"+plain(jobList.selected().JobApplication.Position), contentWidth, true)
		help = "enter delete   esc cancel"
		if jobList.deleting {
			toolbar = "Deleting…"
		} else if jobList.notice != "" {
			toolbar = jobList.notice
		}
	}
	if jobList.control != noControl && jobList.screen == listScreen {
		body = jobList.controlView(contentHeight)
		help = "j/k choose   enter apply   esc cancel"
		if jobList.control == filterControl {
			title = "Filter applications"
			help = "tab status   enter apply   esc cancel"
		} else {
			title = "Sort applications"
		}
	}
	if contentWidth < 60 && jobList.screen == listScreen && jobList.control == noControl {
		help = "↵ a+ e u d s / q quit"
	}
	if contentWidth < 36 && jobList.screen == listScreen && jobList.control == noControl {
		help = "↵ a e u d s / q quit"
	}
	if contentWidth < 36 && jobList.screen == detailsScreen {
		help = "j/k scroll esc back q"
	}
	if contentWidth < 36 && jobList.screen == addScreen {
		help = "tab ^s save esc discard"
	}
	if contentWidth < 36 && jobList.control != noControl {
		help = "↵ apply  esc cancel"
	}
	header := accent.Render("◈ JOBIT") + "  " + title
	rule := muted.Render(strings.Repeat("─", contentWidth))
	bodyLines := strings.Split(body, "\n")
	if len(bodyLines) > contentHeight {
		bodyLines = bodyLines[:contentHeight]
	}
	for len(bodyLines) < contentHeight {
		bodyLines = append(bodyLines, "")
	}
	statusLine := toolbar
	if jobList.notice != "" && jobList.screen == listScreen && jobList.control == noControl {
		statusLine = jobList.notice + "  ·  " + toolbar
	}
	lines := []string{header, muted.Render(subtitle), rule, label.Render(statusLine), strings.Join(bodyLines, "\n"), rule, muted.Render(help)}
	rows := strings.Split(strings.Join(lines, "\n"), "\n")
	for i := range rows {
		rows[i] = ansi.Truncate(rows[i], contentWidth, "…")
	}
	return lipgloss.NewStyle().Padding(1, 2).Render(strings.Join(rows, "\n"))
}

// Busy reports a pending database write.
func (jobList *JobList) Busy() bool {
	return jobList.saving || jobList.deleting
}

func ProtectPendingWrite(model tea.Model, msg tea.Msg) tea.Msg {
	if jobList, ok := model.(*JobList); ok && jobList.Busy() {
		switch msg.(type) {
		case tea.QuitMsg, tea.InterruptMsg:
			return pendingQuitMsg{}
		}
	}
	return msg
}

func (jobList *JobList) pendingPopup() string {
	if jobList.width < 24 || jobList.height < 9 {
		return ansi.Truncate("Busy — please wait", max(1, jobList.width), "")
	}
	title := "Saving changes…"
	if jobList.deleting {
		title = "Deleting record…"
	}
	width := min(42, jobList.width-6)
	content := accent.Render(title) + "\n\n" + ansi.Wrap("Please wait, then quit again.", width, "")
	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("79")).Padding(1, 2).Width(width + 4).Render(content)
	return lipgloss.Place(jobList.width, jobList.height, lipgloss.Center, lipgloss.Center, box)
}
