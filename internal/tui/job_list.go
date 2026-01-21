package tui
import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
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

