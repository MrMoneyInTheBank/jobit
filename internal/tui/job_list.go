package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"
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
)

type SaveApplication func(model.JobApplication) (model.StoredJobApplication, error)
type savedMsg struct {
	application model.StoredJobApplication
	err         error
}

type JobList struct {
	JobsTable            table.Model
	jobs                 []model.StoredJobApplication
	width, height        int
	screen               screen
	details              viewport.Model
	save                 SaveApplication
	fields               [4]string
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

func (jl *JobList) Init() tea.Cmd { return tea.SetWindowTitle("Jobit · Applications") }

// plain prevents stored control characters from affecting terminal rendering.
func plain(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, ansi.Strip(s))
}
func optional(s string) string {
	if s == "" {
		return "—"
	}
	return plain(s)
}
func ranking(r model.Ranking) string {
	if r == 0 {
		return "—"
	}
	return strconv.Itoa(int(r)) + "/5"
}
func constructRow(stored model.StoredJobApplication) table.Row {
	j := stored.JobApplication
	return table.Row{plain(j.CompanyName), plain(j.Position), optional(j.Location), ranking(j.Rank)}
}

// The optional save callback keeps persistence outside the view model.
func InitJobList(jobs []model.StoredJobApplication, save ...SaveApplication) JobList {
	rows := make([]table.Row, len(jobs))
	for i, j := range jobs {
		rows[i] = constructRow(j)
	}
	styles := table.DefaultStyles()
	styles.Header = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Padding(0, 1)
	styles.Cell = lipgloss.NewStyle().Padding(0, 1)
	styles.Selected = lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("24")).Bold(true)
	t := table.New(table.WithColumns(columns(72)), table.WithRows(rows), table.WithFocused(true), table.WithStyles(styles), table.WithHeight(10))
	jl := JobList{JobsTable: t, jobs: append([]model.StoredJobApplication(nil), jobs...), details: viewport.New(72, 10)}
	if len(save) > 0 {
		jl.save = save[0]
	}
	jl.rebuild(0)
	return jl
}

func columns(width int) []table.Column {
	available := max(4, width-8) // Two padding cells per column.
	rank := min(7, max(1, available/5))
	rest := available - rank
	company := max(1, rest*30/100)
	position := max(1, rest*42/100)
	location := max(1, rest-company-position)
	titles := []string{"Company", "Position", "Location", "Ranking"}
	if width < 48 {
		titles = []string{"Co.", "Role", "Place", "Rank"}
	}
	return []table.Column{{Title: titles[0], Width: company}, {Title: titles[1], Width: position}, {Title: titles[2], Width: location}, {Title: titles[3], Width: rank}}
}
func (jl *JobList) contentWidth() int { return max(1, min(112, jl.width-4)) }
func (jl *JobList) resize() {
	w := jl.contentWidth()
	h := max(1, jl.height-8)
	jl.JobsTable.SetColumns(columns(w))
	jl.JobsTable.SetWidth(w)
	jl.JobsTable.SetHeight(h)
	jl.details.Width = w
	jl.details.Height = h
	if jl.screen == detailsScreen {
		jl.setDetails()
	}
}
func (jl *JobList) selected() *model.StoredJobApplication {
	i := jl.JobsTable.Cursor()
	if i < 0 || i >= len(jl.visible) {
		return nil
	}
	return &jl.jobs[jl.visible[i]]
}

func (jl *JobList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		jl.width = msg.Width
		jl.height = msg.Height
		jl.resize()
		return jl, nil
	case savedMsg:
		jl.saving = false
		if msg.err != nil {
			jl.formError = msg.err.Error()
			return jl, nil
		}
		jl.jobs = append(jl.jobs, msg.application)
		jl.rebuild(msg.application.ID)
		jl.notice = "Application saved"
		if selected := jl.selected(); selected == nil || selected.ID != msg.application.ID {
			jl.notice = "Saved · hidden by filter; Esc clears it"
		}
		jl.screen = listScreen
		return jl, nil
	case tea.KeyMsg:
		key := msg.String()
		if key == "ctrl+c" {
			return jl, tea.Quit
		}
		if jl.screen == addScreen {
			return jl, jl.updateForm(msg)
		}
		if jl.screen == listScreen && jl.control != noControl {
			jl.updateControl(msg)
			return jl, nil
		}
		if key == "q" {
			return jl, tea.Quit
		}
		if key == "esc" {
			if jl.screen == listScreen {
				id := jl.selectedID()
				jl.query = ""
				jl.status = ""
				jl.notice = ""
				jl.rebuild(id)
			}
			jl.screen = listScreen
			return jl, nil
		}
		if jl.screen == listScreen {
			if key == "s" {
				jl.control = sortControl
				jl.sortDraft = jl.sortIndex
				return jl, nil
			}
			if key == "/" {
				jl.control = filterControl
				jl.queryDraft = jl.query
				jl.statusDraft = 0
				for i, status := range statuses {
					if status == jl.status {
						jl.statusDraft = i
					}
				}
				return jl, nil
			}
			if key == "enter" && jl.selected() != nil {
				jl.screen = detailsScreen
				jl.setDetails()
				jl.details.GotoTop()
				return jl, nil
			}
			if key == "a" {
				jl.screen = addScreen
				jl.fields = [4]string{}
				jl.field = 0
				jl.formError = ""
				return jl, nil
			}
		}
	}
	var cmd tea.Cmd
	if jl.screen == listScreen {
		jl.JobsTable, cmd = jl.JobsTable.Update(msg)
		jl.markSelection()
	}
	if jl.screen == detailsScreen {
		jl.details, cmd = jl.details.Update(msg)
	}
	return jl, cmd
}

func (jl *JobList) updateForm(msg tea.KeyMsg) tea.Cmd {
	if jl.saving {
		return nil
	}
	switch msg.String() {
	case "esc":
		jl.screen = listScreen
	case "tab", "enter", "down":
		jl.field = (jl.field + 1) % len(jl.fields)
	case "shift+tab", "up":
		jl.field = (jl.field + len(jl.fields) - 1) % len(jl.fields)
	case "backspace", "ctrl+h":
		r := []rune(jl.fields[jl.field])
		if len(r) > 0 {
			jl.fields[jl.field] = string(r[:len(r)-1])
		}
	case "ctrl+s":
		if jl.save == nil {
			jl.formError = "Saving is unavailable in this view."
			return nil
		}
		company, position := strings.TrimSpace(jl.fields[0]), strings.TrimSpace(jl.fields[1])
		if company == "" || position == "" {
			jl.formError = "Company and position are required."
			return nil
		}
		rank := 0
		var err error
		if strings.TrimSpace(jl.fields[3]) != "" {
			rank, err = strconv.Atoi(strings.TrimSpace(jl.fields[3]))
		}
		if err != nil || rank < 0 || rank > 5 {
			jl.formError = "Ranking must be 1–5, or blank for unranked."
			return nil
		}
		app := model.JobApplication{CompanyName: company, Position: position, Location: strings.TrimSpace(jl.fields[2]), Rank: model.Ranking(rank), Status: model.StatusSubmitted, ApplicationDate: time.Now().UTC()}
		jl.saving = true
		jl.formError = ""
		save := jl.save
		return func() tea.Msg { stored, err := save(app); return savedMsg{stored, err} }
	default:
		if msg.Type == tea.KeyRunes {
			jl.fields[jl.field] += plain(string(msg.Runes))
		}
		if msg.Type == tea.KeySpace {
			jl.fields[jl.field] += " "
		}
	}
	return nil
}

func (jl *JobList) formView() string {
	names := []string{"Company *", "Position *", "Location", "Ranking · 1–5"}
	var lines []string
	lines = append(lines, muted.Render("Defaults: submitted · today’s date"), "")
	for i, name := range names {
		marker := "  "
		if i == jl.field {
			marker = "› "
		}
		value := jl.fields[i]
		if i == jl.field {
			value += "▏"
		}
		lines = append(lines, label.Render(marker+name), "  "+ansi.TruncateLeft(value, max(0, lipgloss.Width(value)-max(1, jl.contentWidth()-3)), "…"), "")
	}
	if jl.saving {
		lines = append(lines, accent.Render("Saving…"))
	}
	if jl.formError != "" {
		lines = append(lines, ansi.Hardwrap(plain(jl.formError), jl.contentWidth(), true))
	}
	// Keep the active field visible when a short terminal cannot fit the form.
	start := 0
	h := max(1, jl.height-8)
	if len(lines) > h {
		start = min(max(0, 2+jl.field*3-h+3), len(lines)-h)
		if jl.formError != "" {
			start = len(lines) - h
		}
	}
	return strings.Join(lines[start:min(len(lines), start+h)], "\n")
}

func (jl *JobList) View() string {
	if jl.width == 0 || jl.height == 0 {
		return ""
	}
	if jl.width < 24 || jl.height < 9 {
		return ansi.Truncate("Resize to 24 × 9 · Ctrl+C quits", max(1, jl.width), "…")
	}
	w := jl.contentWidth()
	h := max(1, jl.height-8)
	title := "Applications"
	help := "↑↓ j/k move   ↵ details   a add   s sort   / filter   q quit"
	subtitle := fmt.Sprintf("%d of %d applications", len(jl.visible), len(jl.jobs))
	if len(jl.visible) > 0 {
		subtitle += fmt.Sprintf("  ·  row %d", jl.JobsTable.Cursor()+1)
	}
	choice := sortChoices[jl.sortIndex]
	toolbar := choice.field + " · " + choice.direction
	if jl.hasFilter() {
		toolbar += "  |  "
		if jl.query != "" {
			toolbar += "“" + jl.query + "” "
		}
		if jl.status != "" {
			toolbar += string(jl.status)
		}
	}
	body := jl.JobsTable.View()
	if len(jl.jobs) == 0 {
		body = accent.Render("A fresh start.") + "\nNo applications yet.\nPress a to add your first application."
	} else if len(jl.visible) == 0 {
		body = accent.Render("No matches.") + "\nTry a different search or status.\nPress Esc to clear filters."
	}
	if jl.screen == detailsScreen {
		title = "Application"
		subtitle = plain(jl.selected().JobApplication.CompanyName) + " · " + plain(jl.selected().JobApplication.Position)
		toolbar = fmt.Sprintf("DETAILS  ·  %.0f%% scrolled", jl.details.ScrollPercent()*100)
		body = jl.details.View()
		help = "↑↓ j/k scroll   esc back   q quit"
	}
	if jl.screen == addScreen {
		title = "New application"
		subtitle = "Start with the essentials"
		toolbar = "COMPANY → POSITION → LOCATION → RANKING"
		body = jl.formView()
		help = "tab next   ctrl+s save   esc cancel"
	}
	if jl.control != noControl && jl.screen == listScreen {
		body = jl.controlView(h)
		help = "j/k choose   enter apply   esc cancel"
		if jl.control == filterControl {
			title = "Filter applications"
			help = "tab status   enter apply   esc cancel"
		} else {
			title = "Sort applications"
		}
	}
	if w < 60 && jl.screen == listScreen && jl.control == noControl {
		help = "j/k ↵view a+ s / q quit"
	}
	if w < 36 && jl.screen == listScreen && jl.control == noControl {
		help = "↵ a+ s / q quit"
	}
	if w < 36 && jl.screen == detailsScreen {
		help = "j/k scroll esc back q"
	}
	if w < 36 && jl.screen == addScreen {
		help = "tab next ^s save esc"
	}
	if w < 36 && jl.control != noControl {
		help = "↵ apply  esc cancel"
	}
	header := accent.Render("◈ JOBIT") + "  " + title
	rule := muted.Render(strings.Repeat("─", w))
	bodyLines := strings.Split(body, "\n")
	if len(bodyLines) > h {
		bodyLines = bodyLines[:h]
	}
	for len(bodyLines) < h {
		bodyLines = append(bodyLines, "")
	}
	statusLine := toolbar
	if jl.notice != "" && jl.screen == listScreen && jl.control == noControl {
		statusLine = jl.notice + "  ·  " + toolbar
	}
	lines := []string{header, muted.Render(subtitle), rule, label.Render(statusLine), strings.Join(bodyLines, "\n"), rule, muted.Render(help)}
	rows := strings.Split(strings.Join(lines, "\n"), "\n")
	for i := range rows {
		rows[i] = ansi.Truncate(rows[i], w, "…")
	}
	return lipgloss.NewStyle().Padding(1, 2).Render(strings.Join(rows, "\n"))
}
