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
	"github.com/MrMoneyInTheBank/jobit/internal/store"
	"github.com/MrMoneyInTheBank/jobit/internal/tui"
)

func main() {
	if err := run(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func run() error {
	parsedOptions, err := parseOptions(os.Args[1:], os.Stdout)
	if errors.Is(err, flag.ErrHelp) {
		return nil
	}
	if err != nil {
		return err
	}
	if parsedOptions.version {
		fmt.Fprintln(os.Stdout, versionString())
		return nil
	}
	if parsedOptions.exportCSV != "" {
		return exportCSV(parsedOptions.exportCSV, os.Stdout)
	}

	instanceLock, err := acquireInstance()
	if err != nil {
		return err
	}
	defer instanceLock.Close()

	database, err := store.OpenDB(nil)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer store.CloseDB(database)
	if err := store.InitDB(database); err != nil {
		return fmt.Errorf("initialize database: %w", err)
	}

	applications, err := store.GetJobApplications(database)
	if err != nil {
		return fmt.Errorf("load applications: %w", err)
	}

	jobList := tui.InitJobList(applications, func(application model.JobApplication) (model.StoredJobApplication, error) {
		id, err := store.AddJobApplication(database, application)
		return model.StoredJobApplication{ID: id, JobApplication: application}, err
	})
	jobList.SetMutationHandlers(func(application model.StoredJobApplication) error {
		return store.PatchJobApplication(database, application, model.JobApplicationPatch{})
	}, func(id int64) error { return store.DeleteJobApplicationByID(database, id) })

	program := tea.NewProgram(&jobList, tea.WithAltScreen(), tea.WithoutSignalHandler(), tea.WithFilter(tui.ProtectPendingWrite))
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
