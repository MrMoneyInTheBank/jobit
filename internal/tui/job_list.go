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
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func (jl *JobList) Init() tea.Cmd {
	return nil
}

func (jl *JobList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return jl, tea.Quit
		}
	}
	return jl, nil
}

func (jl *JobList) View() string {
	return baseStyle.Render(jl.JobsTable.View()) + "\n"
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

	return JobList{t}
}
