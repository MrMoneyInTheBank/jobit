package store

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/MrMoneyInTheBank/jobit/internal/model"
)

var ErrInvalidID = errors.New("Invalid ID")

// applicationArgs converts optional pay fields to SQL NULLs. Other fields use model values.
func applicationArgs(app model.JobApplication) []any {
	var minimumPay, maximumPay, currency any
	if app.Pay != nil {
		if app.Pay.Min != nil {
			minimumPay = *app.Pay.Min
		}
		if app.Pay.Max != nil {
			maximumPay = *app.Pay.Max
		}
		if app.Pay.Currency != nil {
			currency = *app.Pay.Currency
		}
	}
	return []any{
		app.CompanyName, app.Position, app.ApplicationDate.UTC().Format("2006-01-02"),
		string(app.Status), app.Referral, string(app.WorkMode), app.Location,
		minimumPay, maximumPay, currency, int(app.Rank), app.Notes, app.JobPostingLink, app.CompanyWebsiteLink,
	}
}

func validateApplication(app model.JobApplication) error {
	if strings.TrimSpace(app.CompanyName) == "" || strings.TrimSpace(app.Position) == "" {
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
	if app.Pay != nil {
		if app.Pay.Min != nil && *app.Pay.Min < 0 {
			return fmt.Errorf("minimum pay must be nonnegative")
		}
		if app.Pay.Max != nil && *app.Pay.Max < 0 {
			return fmt.Errorf("maximum pay must be nonnegative")
		}
		if app.Pay.Min != nil && app.Pay.Max != nil && *app.Pay.Min > *app.Pay.Max {
			return fmt.Errorf("maximum pay must be at least minimum pay")
		}
		if app.Pay.Currency != nil {
			currency := *app.Pay.Currency
			if len(currency) != 3 {
				return fmt.Errorf("currency must be three uppercase letters")
			}
			for _, character := range currency {
				if character < 'A' || character > 'Z' {
					return fmt.Errorf("currency must be three uppercase letters")
				}
			}
		}
	}
	for _, link := range []string{app.JobPostingLink, app.CompanyWebsiteLink} {
		if link == "" {
			continue
		}
		parsedURL, err := url.ParseRequestURI(link)
		if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") || parsedURL.Hostname() == "" {
			return fmt.Errorf("links must use a full http:// or https:// URL")
		}
	}
	return nil
}

type rowScanner interface{ Scan(...any) error }

func scanApplication(row rowScanner) (model.StoredJobApplication, error) {
	var storedApplication model.StoredJobApplication
	application := &storedApplication.JobApplication
	var applicationDate, status, workMode string
	var ranking int64
	var minimumPay, maximumPay sql.NullInt64
	var currency sql.NullString

	if err := row.Scan(
		&storedApplication.ID, &application.CompanyName, &application.Position, &applicationDate, &status,
		&application.Referral, &workMode, &application.Location, &minimumPay, &maximumPay, &currency,
		&ranking, &application.Notes, &application.JobPostingLink, &application.CompanyWebsiteLink,
	); err != nil {
		return storedApplication, err
	}

	parsedDate, err := time.Parse("2006-01-02", applicationDate)
	if err != nil {
		return storedApplication, fmt.Errorf("application %d: invalid application date %q: %w", storedApplication.ID, applicationDate, err)
	}
	application.ApplicationDate = parsedDate
	application.Status = model.Status(status)
	application.WorkMode = model.WorkMode(workMode)
	application.Rank = model.Ranking(ranking)

	if minimumPay.Valid || maximumPay.Valid || currency.Valid {
		application.Pay = &model.Pay{}
		if minimumPay.Valid {
			amount := int(minimumPay.Int64)
			application.Pay.Min = &amount
		}
		if maximumPay.Valid {
			amount := int(maximumPay.Int64)
			application.Pay.Max = &amount
		}
		if currency.Valid {
			application.Pay.Currency = &currency.String
		}
	}

	return storedApplication, nil
}
