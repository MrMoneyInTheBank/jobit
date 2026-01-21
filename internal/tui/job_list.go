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
	JobsTable           table.Model
	width               int
	height              int
	minimumColumnWidths []int
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
	return content
}

func InitJobList(jobs []model.JobApplication) JobList {
	columns := []table.Column{
		{Title: "ID"},
		{Title: "Company Name"},
		{Title: "Position"},
		{Title: "Application Date"},
		{Title: "Status"},
		{Title: "Referral"},
		{Title: "Pay"},
	}
	rows := make([]table.Row, len(jobs))

	for idx, job := range jobs {
		rows[idx] = table.Row{
			strconv.Itoa(int(job.ID)),
			job.CompanyName,
			job.Position,
			job.ApplicationDate.Format("2006-01-02"),
			string(job.Status),
			strconv.FormatBool(job.Referral),
			func() string {
				if job.Pay == nil {
					return "Unknown"
				} else {
					return job.Pay.String()
				}
			}(),
		}
	}
	minimumColumnWidths := computeMinColWidths(columns, rows)
	for idx := range columns {
		columns[idx].Width = minimumColumnWidths[idx]
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
	)

	return JobList{JobsTable: t, minimumColumnWidths: minimumColumnWidths}
}

func computeMinColWidths(columns []table.Column, rows []table.Row) []int {
	widths := make([]int, len(columns))

	for idx, column := range columns {
		widths[idx] = lipgloss.Width(column.Title)
	}
	log.Printf("(b rows)Column widths: %v\n", widths)

	for _, row := range rows {
		for idx, cell := range row {
			if len(cell) > widths[idx] {
				widths[idx] = lipgloss.Width(cell)
			}
		}
	}

	log.Printf("(a rows)Column widths: %v\n", widths)
	return widths
}
