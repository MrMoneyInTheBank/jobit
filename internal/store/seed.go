package store

import (
	"context"
	"database/sql"
	"math/rand"
	"time"

	"github.com/MrMoneyInTheBank/jobit/internal/model"
)

func randomJobApplication(randomGenerator *rand.Rand) model.JobApplication {
	now := time.Now()

	companyNames := []string{
		"OpenAI", "Google", "Meta", "Amazon", "Microsoft",
		"Jane Street", "Citadel", "Optiver", "Stripe", "Uber",
	}

	positions := []string{
		"Software Engineer",
		"Backend Engineer",
		"Quantitative Trader",
		"Quantitative Researcher",
		"Machine Learning Engineer",
		"Systems Engineer",
	}

	locations := []string{
		"New York, NY",
		"San Francisco, CA",
		"London, UK",
		"Amsterdam, NL",
		"Remote",
	}

	currencies := []string{"USD", "EUR", "GBP"}

	statuses := []model.Status{
		model.StatusSubmitted,
		model.StatusInterviewing,
		model.StatusOffer,
		model.StatusRejected,
		model.StatusAccepted,
	}

	remoteTypes := []model.WorkMode{
		model.WorkModeRemote,
		model.WorkModeHybrid,
		model.WorkModeOnsite,
	}

	application := model.JobApplication{
		CompanyName:     companyNames[randomGenerator.Intn(len(companyNames))],
		Position:        positions[randomGenerator.Intn(len(positions))],
		ApplicationDate: now.AddDate(0, 0, -randomGenerator.Intn(180)), // last 6 months
		Status:          statuses[randomGenerator.Intn(len(statuses))],
		Referral:        randomGenerator.Float64() < 0.3, // 30% referral
		Rank:            model.Ranking(1 + randomGenerator.Intn(5)),
	}

	if randomGenerator.Float64() < 0.8 {
		workMode := remoteTypes[randomGenerator.Intn(len(remoteTypes))]
		application.WorkMode = workMode
	}

	if randomGenerator.Float64() < 0.7 {
		location := locations[randomGenerator.Intn(len(locations))]
		application.Location = location
	}

	if randomGenerator.Float64() < 0.6 {
		minimumPay := 80_000 + randomGenerator.Intn(70_000)
		maximumPay := minimumPay + randomGenerator.Intn(50_000)
		currency := currencies[randomGenerator.Intn(len(currencies))]

		application.Pay = &model.Pay{
			Min:      &minimumPay,
			Max:      &maximumPay,
			Currency: &currency,
		}
	}

	if randomGenerator.Float64() < 0.4 {
		note := "Reached out to recruiter on LinkedIn"
		application.Notes = note
	}

	if randomGenerator.Float64() < 0.5 {
		postingLink := "https://jobs.example.com/posting"
		application.JobPostingLink = postingLink
	}
	if randomGenerator.Float64() < 0.5 {
		websiteLink := "https://www." + application.CompanyName + ".com"
		application.CompanyWebsiteLink = websiteLink
	}

	return application
}

func SeedDB(db *sql.DB, count int) error {
	ctx := context.Background()
	transaction, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer transaction.Rollback()

	randomGenerator := rand.New(rand.NewSource(time.Now().UnixNano()))
	for range count {
		application := randomJobApplication(randomGenerator)
		if _, err2 := transaction.ExecContext(ctx, insertQuery, applicationArgs(application)...); err2 != nil {
			return err2
		}
	}
	return transaction.Commit()
}
