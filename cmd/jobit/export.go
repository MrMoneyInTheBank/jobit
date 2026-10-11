package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strconv"

	"github.com/MrMoneyInTheBank/jobit/internal/store"
)

var version = "dev"

func versionString() string {
	if version != "dev" {
		return "jobit " + version
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return "jobit " + info.Main.Version
		}
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return "jobit dev (" + setting.Value + ")"
			}
		}
	}
	return "jobit dev"
}

func exportCSV(path string, stdout io.Writer) (err error) {
	database, err := store.OpenReadDB()
	if err != nil {
		return err
	}
	defer database.Close()

	if err = store.CheckSchema(database); err != nil {
		return err
	}
	applications, err := store.GetJobApplications(database)
	if err != nil {
		return err
	}

	var output io.Writer = stdout
	if path != "-" {
		exportFile, openErr := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if openErr != nil {
			return fmt.Errorf("create export (existing files are not overwritten): %w", openErr)
		}
		output = exportFile
		defer func() {
			if closeErr := exportFile.Close(); err == nil {
				err = closeErr
			}
			if err != nil {
				os.Remove(path)
			}
		}()
	}

	writer := csv.NewWriter(output)
	if err = writer.Write([]string{"id", "company", "position", "location", "ranking", "application_date", "status", "referral", "work_mode", "pay_min", "pay_max", "currency", "notes", "job_posting_link", "company_website_link"}); err != nil {
		return err
	}
	for _, stored := range applications {
		application := stored.JobApplication
		minimum, maximum, currency := "", "", ""
		if application.Pay != nil {
			if application.Pay.Min != nil {
				minimum = strconv.Itoa(*application.Pay.Min)
			}
			if application.Pay.Max != nil {
				maximum = strconv.Itoa(*application.Pay.Max)
			}
			if application.Pay.Currency != nil {
				currency = *application.Pay.Currency
			}
		}
		record := []string{
			strconv.FormatInt(stored.ID, 10), application.CompanyName, application.Position,
			application.Location, strconv.Itoa(int(application.Rank)),
			application.ApplicationDate.UTC().Format("2006-01-02"), string(application.Status),
			strconv.FormatBool(application.Referral), string(application.WorkMode),
			minimum, maximum, currency, application.Notes,
			application.JobPostingLink, application.CompanyWebsiteLink,
		}
		if err = writer.Write(record); err != nil {
			return err
		}
	}

	writer.Flush()
	return writer.Error()
}
