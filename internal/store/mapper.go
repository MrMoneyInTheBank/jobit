package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/MrMoneyInTheBank/jobit/internal/model"
)

type boundJobApplication struct {
	id                 int64
	companyName        string
	position           string
	applicationDate    string
	status             string
	referral           bool
	remote             sql.NullString
	location           sql.NullString
	payMin             sql.NullInt64
	payMax             sql.NullInt64
	payCurrency        sql.NullString
	ranking            sql.NullInt64
	notes              sql.NullString
	jobPostingLink     sql.NullString
	companyWebsiteLink sql.NullString
}

func bindJobApplication(application model.JobApplication) boundJobApplication {
	boundedApp := boundJobApplication{
		companyName:     application.CompanyName,
		position:        application.Position,
		applicationDate: application.ApplicationDate.UTC().Format("2006-01-02"),
		status:          string(application.Status),
		referral:        application.Referral,
	}

	if application.WorkMode != "" {
		boundedApp.remote = sql.NullString{String: string(application.WorkMode), Valid: true}
	}

	if application.Location != "" {
		boundedApp.location = sql.NullString{String: application.Location, Valid: true}
	}

	if application.Pay != nil {
		if application.Pay.Min != nil {
			boundedApp.payMin = sql.NullInt64{Int64: int64(*application.Pay.Min), Valid: true}
		}

		if application.Pay.Max != nil {
			boundedApp.payMax = sql.NullInt64{Int64: int64(*application.Pay.Max), Valid: true}
		}

		if application.Pay.Currency != nil {
			boundedApp.payCurrency = sql.NullString{String: *application.Pay.Currency, Valid: true}
		}
	}

	if application.Rank != 0 {
		boundedApp.ranking = sql.NullInt64{Int64: int64(application.Rank), Valid: true}
	}
	if application.Notes != "" {
		boundedApp.notes = sql.NullString{String: application.Notes, Valid: true}
	}
	if application.JobPostingLink != "" {
		boundedApp.jobPostingLink = sql.NullString{String: application.JobPostingLink, Valid: true}
	}
	if application.CompanyWebsiteLink != "" {
		boundedApp.companyWebsiteLink = sql.NullString{String: application.CompanyWebsiteLink, Valid: true}
	}

	return boundedApp
}

func (b boundJobApplication) insertArgs() []any {
	return []any{
		b.companyName,
		b.position,
		b.applicationDate,
		b.status,
		b.referral,
		b.remote,
		b.location,
		b.payMin,
		b.payMax,
		b.payCurrency,
		b.ranking,
		b.notes,
		b.jobPostingLink,
		b.companyWebsiteLink,
	}
}

var ErrInvalidID = errors.New("Invalid ID")

func bindJobApplicationPatch(app model.StoredJobApplication) (boundJobApplication, error) {
	if app.ID <= 0 {
		return boundJobApplication{}, ErrInvalidID
	}
	boundedApp := bindJobApplication(app.JobApplication)
	boundedApp.id = app.ID
	return boundedApp, nil
}

func (b boundJobApplication) updateArgs() []any {
	return []any{
		b.companyName,
		b.position,
		b.applicationDate,
		b.status,
		b.referral,
		b.remote,
		b.location,
		b.payMin,
		b.payMax,
		b.payCurrency,
		b.ranking,
		b.notes,
		b.jobPostingLink,
		b.companyWebsiteLink,
		b.id,
	}
}

func (b boundJobApplication) toModel() (model.StoredJobApplication, error) {
	date, err := time.Parse("2006-01-02", b.applicationDate)
	if err != nil {
		return model.StoredJobApplication{}, fmt.Errorf("application %d: invalid application date %q: %w", b.id, b.applicationDate, err)
	}
	app := model.JobApplication{
		CompanyName:     b.companyName,
		Position:        b.position,
		ApplicationDate: date,
		Status:          model.Status(b.status),
		Referral:        b.referral,
	}

	if b.remote.Valid {
		app.WorkMode = model.WorkMode(b.remote.String)
	}

	if b.location.Valid {
		app.Location = b.location.String
	}

	if b.payMin.Valid || b.payMax.Valid || b.payCurrency.Valid {
		p := &model.Pay{}

		if b.payMin.Valid {
			v := int(b.payMin.Int64)
			p.Min = &v
		}
		if b.payMax.Valid {
			v := int(b.payMax.Int64)
			p.Max = &v
		}
		if b.payCurrency.Valid {
			p.Currency = &b.payCurrency.String
		}

		app.Pay = p
	}

	if b.ranking.Valid {
		app.Rank = model.Ranking(b.ranking.Int64)
	}

	if b.notes.Valid {
		app.Notes = b.notes.String
	}

	if b.jobPostingLink.Valid {
		app.JobPostingLink = b.jobPostingLink.String
	}

	if b.companyWebsiteLink.Valid {
		app.CompanyWebsiteLink = b.companyWebsiteLink.String
	}

	return model.StoredJobApplication{ID: b.id, JobApplication: app}, nil
}
