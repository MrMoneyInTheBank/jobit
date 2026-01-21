package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"

	store "github.com/MrMoneyInTheBank/jobit/internal/store"
	"github.com/MrMoneyInTheBank/jobit/internal/tui"
)

func ptr[T any](v T) *T {
	return &v
}

func main() {
	db, err := store.OpenDB(ptr("seed.db"))
	if err != nil {
		log.Fatalf("Could not open database: %v", err)
	}
	defer store.CloseDB(db)
	store.InitDB(db)

	job_apps, err := store.GetJobApplications(db)
	if err != nil {
		log.Fatalf("Could not get job applications: %v", err)
	}

	job_list := tui.InitJobList(job_apps)
	if _, err := tea.NewProgram(&job_list, tea.WithAltScreen()).Run(); err != nil {
		log.Fatalf("Could not run TUI: %v", err)
	}
}
