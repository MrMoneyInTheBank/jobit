package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/MrMoneyInTheBank/jobit/internal/model"
	store "github.com/MrMoneyInTheBank/jobit/internal/store"
	"github.com/MrMoneyInTheBank/jobit/internal/tui"
)

func main() {
	if err := run(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func run() error {
	opts, err := parseOptions(os.Args[1:], os.Stdout)
	if errors.Is(err, flag.ErrHelp) {
		return nil
	}
	if err != nil {
		return err
	}
	if opts.version {
		fmt.Fprintln(os.Stdout, versionString())
		return nil
	}
	if opts.exportCSV != "" {
		return exportCSV(opts.exportCSV, os.Stdout)
	}
	lock, err := acquireInstance()
	if err != nil {
		return err
	}
	defer lock.Close()

	db, err := store.OpenDB(nil)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer store.CloseDB(db)
	if err := store.InitDB(db); err != nil {
		return fmt.Errorf("initialize database: %w", err)
	}

	job_apps, err := store.GetJobApplications(db)
	if err != nil {
		return fmt.Errorf("load applications: %w", err)
	}

	job_list := tui.InitJobList(job_apps, func(app model.JobApplication) (model.StoredJobApplication, error) {
		id, err := store.AddJobApplication(db, app)
		return model.StoredJobApplication{ID: id, JobApplication: app}, err
	})
	job_list.SetMutationHandlers(func(app model.StoredJobApplication) error {
		return store.PatchJobApplication(db, app, model.JobApplicationPatch{})
	}, func(id int64) error { return store.DeleteJobApplicationByID(db, id) })

	program := tea.NewProgram(&job_list, tea.WithAltScreen(), tea.WithoutSignalHandler(), tea.WithFilter(tui.ProtectPendingWrite))
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signals)
	done := make(chan struct{})
	defer close(done)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-signals:
				program.Send(tea.QuitMsg{})
			}
		}
	}()
	if _, err := program.Run(); err != nil {
		return fmt.Errorf("run TUI: %w", err)
	}
	return nil
}
