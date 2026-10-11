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
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" {
				return "jobit dev (" + s.Value + ")"
			}
		}
	}
	return "jobit dev"
}
func exportCSV(path string, stdout io.Writer) (err error) {
	db, err := store.OpenDB(nil)
	if err != nil {
		return err
	}
	defer db.Close()
	if err = store.InitDB(db); err != nil {
		return err
	}
	apps, err := store.GetJobApplications(db)
	if err != nil {
		return err
	}
	var out io.Writer = stdout
	if path != "-" {
		f, e := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if e != nil {
			return fmt.Errorf("create export (existing files are not overwritten): %w", e)
		}
		out = f
		defer func() {
			if e := f.Close(); err == nil {
				err = e
			}
			if err != nil {
				os.Remove(path)
			}
		}()
	}
	writer := csv.NewWriter(out)
	if err = writer.Write([]string{"id", "company", "position", "location", "ranking", "application_date", "status", "referral", "work_mode", "pay_min", "pay_max", "currency", "notes", "job_posting_link", "company_website_link"}); err != nil {
		return err
	}
	for _, stored := range apps {
		a := stored.JobApplication
		minimum, maximum, currency := "", "", ""
		if a.Pay != nil {
			if a.Pay.Min != nil {
				minimum = strconv.Itoa(*a.Pay.Min)
			}
			if a.Pay.Max != nil {
				maximum = strconv.Itoa(*a.Pay.Max)
			}
			if a.Pay.Currency != nil {
				currency = *a.Pay.Currency
			}
		}
		if err = writer.Write([]string{strconv.FormatInt(stored.ID, 10), a.CompanyName, a.Position, a.Location, strconv.Itoa(int(a.Rank)), a.ApplicationDate.UTC().Format("2006-01-02"), string(a.Status), strconv.FormatBool(a.Referral), string(a.WorkMode), minimum, maximum, currency, a.Notes, a.JobPostingLink, a.CompanyWebsiteLink}); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}
