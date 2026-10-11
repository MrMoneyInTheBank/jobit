package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"

	store "github.com/MrMoneyInTheBank/jobit/internal/store"
	"github.com/MrMoneyInTheBank/jobit/internal/tui"
)

func main() {

	db, err := store.OpenDB(nil)
	if err != nil {
		log.Fatalf("Could not open database: %v", err)
	}
	defer store.CloseDB(db)
	if err := store.InitDB(db); err != nil {
		log.Fatalf("Could not initialize database: %v", err)
	}

	job_apps, err := store.GetJobApplications(db)
	if err != nil {
		log.Fatalf("Could not get job applications: %v", err)
	}

	job_list := tui.InitJobList(job_apps)
	if _, err := tea.NewProgram(&job_list, tea.WithAltScreen()).Run(); err != nil {
		log.Fatalf("Could not run TUI: %v", err)
	}
}
