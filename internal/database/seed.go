package store

import (
	"database/sql"
	"math/rand"
	"time"

	"github.com/MrMoneyInTheBank/jobit/internal/model"
)

func randomJobApplication(r *rand.Rand) model.JobApplication {
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

	remoteTypes := []model.RemoteType{
		model.RemoteRemote,
		model.RemoteHybrid,
		model.RemoteOnsite,
	}

	app := model.JobApplication{
		ID:              r.Int63(),
		CompanyName:     companyNames[r.Intn(len(companyNames))],
		Position:        positions[r.Intn(len(positions))],
		ApplicationDate: now.AddDate(0, 0, -r.Intn(180)), // last 6 months
		Status:          statuses[r.Intn(len(statuses))],
		Referral:        r.Float64() < 0.3, // 30% referral
	}

	// Optional: Remote
	if r.Float64() < 0.8 {
		rt := remoteTypes[r.Intn(len(remoteTypes))]
		app.Remote = &rt
	}

	// Optional: Location
	if r.Float64() < 0.7 {
		loc := locations[r.Intn(len(locations))]
		app.Location = &loc
	}

	// Optional: Pay
	if r.Float64() < 0.6 {
		min := 80_000 + r.Intn(70_000)
		max := min + r.Intn(50_000)
		cur := currencies[r.Intn(len(currencies))]

		app.Pay = &model.Pay{
			Min:      &min,
			Max:      &max,
			Currency: &cur,
		}
	}

	// Optional: Ranking (1–10)
	if r.Float64() < 0.5 {
		rank := 1 + r.Intn(10)
		app.Ranking = &rank
	}

	// Optional: Notes
	if r.Float64() < 0.4 {
		note := "Reached out to recruiter on LinkedIn"
		app.Notes = &note
	}

	// Optional: Links
	if r.Float64() < 0.5 {
		jp := "https://jobs.example.com/posting"
		app.JobPostingLink = &jp
	}
	if r.Float64() < 0.5 {
		ws := "https://www." + app.CompanyName + ".com"
		app.CompanyWebsiteLink = &ws
	}

	return app
}

func seedDB(db *sql.DB) error {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for range 100 {
		app := randomJobApplication(r)
		if _, err := AddJobApplication(db, app); err != nil {
			return err
		}
	}
	return nil
}
