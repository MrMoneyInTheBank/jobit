package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/MrMoneyInTheBank/jobit/internal/model"
)

var ErrInvalidID = errors.New("Invalid ID")

// applicationArgs converts optional pay fields to SQL NULLs. Other fields use model values.
func applicationArgs(app model.JobApplication) []any {
	var min, max, currency any
	if app.Pay != nil {
		if app.Pay.Min != nil {
			min = *app.Pay.Min
		}
		if app.Pay.Max != nil {
			max = *app.Pay.Max
		}
		if app.Pay.Currency != nil {
			currency = *app.Pay.Currency
		}
	}
	return []any{
		app.CompanyName, app.Position, app.ApplicationDate.UTC().Format("2006-01-02"),
		string(app.Status), app.Referral, string(app.WorkMode), app.Location,
		min, max, currency, int(app.Rank), app.Notes, app.JobPostingLink, app.CompanyWebsiteLink,
	}
}

func validateApplication(app model.JobApplication) error {
	if app.CompanyName == "" || app.Position == "" {
		return fmt.Errorf("company and position are required")
	}
	if app.ApplicationDate.IsZero() {
		return fmt.Errorf("application date is required")
	}
	if !app.Rank.Valid() {
		return fmt.Errorf("ranking must be 1–5")
	}
	switch app.Status {
	case model.StatusSubmitted, model.StatusInterviewing, model.StatusOffer, model.StatusRejected, model.StatusAccepted:
	default:
		return fmt.Errorf("invalid status %q", app.Status)
	}
	switch app.WorkMode {
	case "", model.WorkModeRemote, model.WorkModeHybrid, model.WorkModeOnsite:
	default:
		return fmt.Errorf("invalid work mode %q", app.WorkMode)
	}
	return nil
}

type rowScanner interface{ Scan(...any) error }

func scanApplication(row rowScanner) (model.StoredJobApplication, error) {
	var stored model.StoredJobApplication
	app := &stored.JobApplication
	var date, status, mode string
	var rank int64
	var min, max sql.NullInt64
	var currency sql.NullString
	if err := row.Scan(
		&stored.ID, &app.CompanyName, &app.Position, &date, &status,
		&app.Referral, &mode, &app.Location, &min, &max, &currency,
		&rank, &app.Notes, &app.JobPostingLink, &app.CompanyWebsiteLink,
	); err != nil {
		return stored, err
	}
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		return stored, fmt.Errorf("application %d: invalid application date %q: %w", stored.ID, date, err)
	}
	app.ApplicationDate = parsed
	app.Status = model.Status(status)
	app.WorkMode = model.WorkMode(mode)
	app.Rank = model.Ranking(rank)
	if min.Valid || max.Valid || currency.Valid {
		app.Pay = &model.Pay{}
		if min.Valid {
			v := int(min.Int64)
			app.Pay.Min = &v
		}
		if max.Valid {
			v := int(max.Int64)
			app.Pay.Max = &v
		}
		if currency.Valid {
			app.Pay.Currency = &currency.String
		}
	}
	return stored, nil
}
