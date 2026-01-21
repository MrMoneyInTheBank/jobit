package tui

import (
	"log"
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
	return tea.SetWindowTitle("\U0001F4BC Jobit")
}

func (jl *JobList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return jl, tea.Quit
		}
	}
	jl.JobsTable, cmd = jl.JobsTable.Update(msg)
	return jl, cmd
}

func (jl *JobList) View() string {
	content := baseStyle.Render(jl.JobsTable.View())
	return lipgloss.Place(jl.width, jl.height, lipgloss.Center, lipgloss.Center, content)
}

func InitJobList(jobs []model.JobApplication) JobList {
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
