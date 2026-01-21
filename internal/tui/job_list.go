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
	log.Println("Init called.")
	return tea.SetWindowTitle("this is the title just please show up")
}

func (jl *JobList) resizeColumns() {
	idW := 5 // index width

	available := max(jl.width-2, 20)    // total remaining width
	remaining := max(available-idW, 10) // remaining width after accounting for index

	companyW := remaining / 2 // split remaining width in half
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
		log.Printf("Window size changed to %d x %d\n", msg.Width, msg.Height)
		jl.width, jl.height = msg.Width, msg.Height
		jl.resizeColumns()
		return jl, nil

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

	return JobList{JobsTable: t}
}
