package tui

import (
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/MrMoneyInTheBank/jobit/internal/model"
)

type JobList struct {
	JobsTable table.Model
	width     int
	height    int
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func (jl *JobList) Init() tea.Cmd {
	return tea.SetWindowTitle("Jobit")
}

// recalculates column widths to fill the available terminal width
func (jl *JobList) resizeColumns() {
	// How much width does your border consume?
	// NormalBorder adds 1 char on left + 1 char on right => 2 total.
	// If you later add padding/margins, subtract those too.
	available := jl.width - 2
	if available < 20 {
		available = 20
	}

	idW := 5
	remaining := available - idW
	if remaining < 10 {
		remaining = 10
	}

	companyW := remaining / 2
	positionW := remaining - companyW

	jl.JobsTable.SetColumns([]table.Column{
		{Title: "ID", Width: idW},
		{Title: "Company Name", Width: companyW},
		{Title: "Position", Width: positionW},
	})
}

func (jl *JobList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		jl.width = msg.Width
		jl.height = msg.Height
		jl.resizeColumns()
		return jl, tea.ClearScreen

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return jl, tea.Quit
		}
	}

	// (Optional but common) forward messages to the table so it can handle arrows, etc.
	var cmd tea.Cmd
	jl.JobsTable, cmd = jl.JobsTable.Update(msg)
	return jl, cmd
}

func (jl *JobList) View() string {
	return baseStyle.Render(jl.JobsTable.View()) + "\n"
}

func InitJobList(jobs []model.JobApplication) JobList {
	// Start with whatever; it'll get resized as soon as WindowSizeMsg arrives.
	columns := []table.Column{
		{Title: "ID", Width: 5},
		{Title: "Company Name", Width: 20},
		{Title: "Position", Width: 20},
	}
	rows := make([]table.Row, len(jobs))

	for idx, job := range jobs {
		rows[idx] = table.Row{
			strconv.Itoa(int(job.ID)),
			job.CompanyName,
			job.Position,
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
	)

	return JobList{JobsTable: t}
}
