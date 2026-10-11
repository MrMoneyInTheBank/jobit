package model

import (
	"testing"
	"time"
)

func TestJobApplicationEqual(t *testing.T) {
	a := JobApplication{
		CompanyName:     "OpenAI",
		Position:        "Machine Learning Engineer",
		ApplicationDate: time.Now().UTC(),
		Status:          StatusSubmitted,
		Referral:        false,
	}
	b := JobApplication{
		CompanyName:     "OpenAI",
		Position:        "Machine Learning Engineer",
		ApplicationDate: time.Now().UTC(),
		Status:          StatusSubmitted,
		Referral:        false,
	}

	c := JobApplication{
		CompanyName:     "OpenAI",
		Position:        "Machine Learning Engineer",
		ApplicationDate: time.Now().UTC(),
		Status:          StatusSubmitted,
		Referral:        true,
	}

	if !a.Equal(b) {
		t.Fatalf("Job application mismatch")
	}

	if a.Equal(c) {
		t.Fatalf("Job application mismatch")
	}
}

func TestEqualPayAndDates(t *testing.T) {
	min1, min2 := 100, 100
	a := JobApplication{Pay: &Pay{Min: &min1}, ApplicationDate: time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC)}
	b := a
	b.Pay = &Pay{Min: &min2}
	b.ApplicationDate = a.ApplicationDate.Add(time.Hour)
	if !a.Equal(b) {
		t.Fatal("equal values and UTC day should match")
	}
	min2++
	if a.Equal(b) {
		t.Fatal("different pay should differ")
	}
	a.Pay = nil
	b.Pay = &Pay{}
	if !a.Equal(b) {
		t.Fatal("nil and empty pay should match")
	}
	b.ApplicationDate = a.ApplicationDate.Add(24 * time.Hour)
	if a.Equal(b) {
		t.Fatal("different UTC days should differ")
	}
}

func testPointer[T any](value T) *T { return &value }

func TestPayDisplay(t *testing.T) {
	for _, testCase := range []struct {
		name string
		pay  *Pay
		want string
	}{
		{"missing", nil, "Unknown"},
		{"empty", &Pay{}, "Unknown"},
		{"minimum and maximum", &Pay{Min: testPointer(0), Max: testPointer(100), Currency: testPointer("USD")}, "USD 0 - 100"},
		{"minimum only", &Pay{Min: testPointer(50)}, "50+"},
		{"maximum only", &Pay{Max: testPointer(100), Currency: testPointer("EUR")}, "EUR <= 100"},
		{"currency only", &Pay{Currency: testPointer("GBP")}, "GBP Unknown"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if actual := testCase.pay.String(); actual != testCase.want {
				t.Fatalf("pay display = %q, want %q", actual, testCase.want)
			}
		})
	}
}

func TestRankingBounds(t *testing.T) {
	for rank := Ranking(0); rank <= 6; rank++ {
		if rank.Valid() != (rank >= RankOne && rank <= RankFive) {
			t.Fatalf("rank %d validity is incorrect", rank)
		}
	}
}

func TestPatchUpdatesEveryFieldAndCanClearPay(t *testing.T) {
	application := JobApplication{
		CompanyName: "Acme", Position: "Engineer", Location: "London",
		Pay: &Pay{Min: testPointer(100)}, WorkMode: WorkModeRemote,
		ApplicationDate: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
		Status:          StatusSubmitted, Rank: RankTwo, Referral: false,
		Notes: "Old", JobPostingLink: "https://example.com/old", CompanyWebsiteLink: "https://example.com",
	}
	original := application
	application.Apply(JobApplicationPatch{})
	if !application.Equal(original) {
		t.Fatal("empty patch changed the application")
	}
	var noPay *Pay
	patch := JobApplicationPatch{
		CompanyName: testPointer("Beta"), Position: testPointer("Designer"),
		Location: testPointer(""), Pay: &noPay, WorkMode: testPointer(WorkModeHybrid),
		ApplicationDate: testPointer(time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)),
		Status:          testPointer(StatusInterviewing), Rank: testPointer(RankFive),
		Referral: testPointer(true), Notes: testPointer(""),
		JobPostingLink:     testPointer("https://example.com/new"),
		CompanyWebsiteLink: testPointer(""),
	}
	application.Apply(patch)
	want := JobApplication{
		CompanyName: "Beta", Position: "Designer", Location: "", WorkMode: WorkModeHybrid,
		ApplicationDate: *patch.ApplicationDate, Status: StatusInterviewing, Rank: RankFive,
		Referral: true, Notes: "", JobPostingLink: "https://example.com/new",
	}
	if !application.Equal(want) || application.Pay != nil {
		t.Fatalf("patch result = %+v, want %+v", application, want)
	}
}

func TestApplicationEqualityChecksEachField(t *testing.T) {
	base := JobApplication{
		CompanyName: "Acme", Position: "Engineer", Location: "London",
		Pay:             &Pay{Min: testPointer(100), Max: testPointer(200), Currency: testPointer("USD")},
		WorkMode:        WorkModeRemote,
		ApplicationDate: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
		Status:          StatusSubmitted, Rank: RankThree, Referral: true,
		Notes: "Note", JobPostingLink: "https://example.com/job", CompanyWebsiteLink: "https://example.com",
	}
	for _, testCase := range []struct {
		name   string
		change func(*JobApplication)
	}{
		{"company", func(a *JobApplication) { a.CompanyName = "Beta" }},
		{"position", func(a *JobApplication) { a.Position = "Designer" }},
		{"location", func(a *JobApplication) { a.Location = "Paris" }},
		{"pay minimum", func(a *JobApplication) {
			a.Pay = &Pay{Min: testPointer(101), Max: testPointer(200), Currency: testPointer("USD")}
		}},
		{"pay maximum", func(a *JobApplication) {
			a.Pay = &Pay{Min: testPointer(100), Max: testPointer(201), Currency: testPointer("USD")}
		}},
		{"pay currency", func(a *JobApplication) {
			a.Pay = &Pay{Min: testPointer(100), Max: testPointer(200), Currency: testPointer("EUR")}
		}},
		{"work mode", func(a *JobApplication) { a.WorkMode = WorkModeHybrid }},
		{"date", func(a *JobApplication) { a.ApplicationDate = a.ApplicationDate.AddDate(0, 0, 1) }},
		{"status", func(a *JobApplication) { a.Status = StatusOffer }},
		{"rank", func(a *JobApplication) { a.Rank = RankFour }},
		{"referral", func(a *JobApplication) { a.Referral = false }},
		{"notes", func(a *JobApplication) { a.Notes = "Changed" }},
		{"posting", func(a *JobApplication) { a.JobPostingLink = "https://example.com/new" }},
		{"website", func(a *JobApplication) { a.CompanyWebsiteLink = "https://example.org" }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			changed := base
			testCase.change(&changed)
			if base.Equal(changed) {
				t.Fatal("changed field was ignored")
			}
		})
	}
}
