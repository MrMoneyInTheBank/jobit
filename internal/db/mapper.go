package db

import (
	"database/sql"
	"time"

	"github.com/MrMoneyInTheBank/jobit/internal/model"
)

func bindJobApplication(application model.JobApplication) []any {
	var (
		location           sql.NullString
		remote             sql.NullString
		payMin             sql.NullInt64
		payMax             sql.NullInt64
		payCurrency        sql.NullString
		ranking            sql.NullInt64
		notes              sql.NullString
		jobPostingLink     sql.NullString
		companyWebsiteLink sql.NullString
	)

	if application.Location != nil {
		location = sql.NullString{String: *application.Location, Valid: true}
	}

	if application.Remote != nil {
		remote = sql.NullString{String: string(*application.Remote), Valid: true}
	}

	if application.Pay != nil {
		if application.Pay.Min != nil {
			payMin = sql.NullInt64{Int64: int64(*application.Pay.Min), Valid: true}
		}

		if application.Pay.Max != nil {
			payMax = sql.NullInt64{Int64: int64(*application.Pay.Max), Valid: true}
		}

		if application.Pay.Currency != nil {
			payCurrency = sql.NullString{String: *application.Pay.Currency, Valid: true}
		}
	}

	if application.Ranking != nil {
		ranking = sql.NullInt64{Int64: int64(*application.Ranking), Valid: true}
	}
	if application.Notes != nil {
		notes = sql.NullString{String: *application.Notes, Valid: true}
	}
	if application.JobPositingLink != nil {
		jobPostingLink = sql.NullString{String: *application.JobPositingLink, Valid: true}
	}
	if application.CompanyWebsiteLink != nil {
		companyWebsiteLink = sql.NullString{String: *application.CompanyWebsiteLink, Valid: true}
	}

	return []any{
		application.CompanyName,
		application.Position,
		application.ApplicationDate.UTC().Format(time.RFC3339),
		string(application.Status),
		application.Referral,
		remote,
		location,
		payMin,
		payMax,
		payCurrency,
		ranking,
		notes,
		jobPostingLink,
		companyWebsiteLink,
	}
}

func scanJobApplication(
	id int64,
	companyName string,
	position string,
	applicationDate string,
	status string,
	referral bool,
	remote sql.NullString,
	location sql.NullString,
	payMin sql.NullInt64,
	payMax sql.NullInt64,
	payCurrency sql.NullString,
	ranking sql.NullInt64,
	notes sql.NullString,
	jobPostingLink sql.NullString,
	companyWebsiteLink sql.NullString,
) (model.JobApplication, error) {
	application := model.JobApplication{
		ID:              id,
		CompanyName:     companyName,
		Position:        position,
		Status:          model.Status(status),
		Referral:        referral,
		ApplicationDate: parseTimeString(applicationDate),
	}

	if remote.Valid {
		r := model.RemoteType(remote.String)
		application.Remote = &r
	}

	if location.Valid {
		application.Location = &location.String
	}

	if payMin.Valid || payMax.Valid || payCurrency.Valid {
		p := &model.Pay{}

		if payMin.Valid {
			v := int(payMin.Int64)
			p.Min = &v

		}

		if payMax.Valid {
			v := int(payMax.Int64)
			p.Max = &v
		}

		if payCurrency.Valid {
			p.Currency = &payCurrency.String
		}

		application.Pay = p
	}

	if ranking.Valid {
		v := int(ranking.Int64)
		application.Ranking = &v
	}
	if notes.Valid {
		application.Notes = &notes.String
	}
	if jobPostingLink.Valid {
		application.JobPositingLink = &jobPostingLink.String
	}
	if companyWebsiteLink.Valid {
		application.CompanyWebsiteLink = &companyWebsiteLink.String
	}

	return application, nil
}

func parseTimeString(applicationDate string) time.Time {
	time, err := time.Parse(time.RFC3339, applicationDate)
	if err != nil {
		panic(err)
	}
	return time
}
