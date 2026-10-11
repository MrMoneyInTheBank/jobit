package main

import (
	"errors"
	"flag"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/MrMoneyInTheBank/jobit/internal/model"
	store "github.com/MrMoneyInTheBank/jobit/internal/store"
	"github.com/MrMoneyInTheBank/jobit/internal/tui"
)

func main() {
	if err := parseArgs(os.Args[1:], os.Stdout); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		log.Print(err)
		os.Exit(2)
	}

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

	job_list := tui.InitJobList(job_apps, func(app model.JobApplication) (model.StoredJobApplication, error) {
		id, err := store.AddJobApplication(db, app)
		return model.StoredJobApplication{ID: id, JobApplication: app}, err
	})
	job_list.SetMutationHandlers(func(app model.StoredJobApplication) error {
		return store.PatchJobApplication(db, app, model.JobApplicationPatch{})
	}, func(id int64) error { return store.DeleteJobApplicationByID(db, id) })
	if _, err := tea.NewProgram(&job_list, tea.WithAltScreen()).Run(); err != nil {
		log.Fatalf("Could not run TUI: %v", err)
	}
}
