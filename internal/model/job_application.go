package model

import (
	"fmt"
	"strings"
	"time"
)

type Status string

const (
	StatusSubmitted    Status = "submitted"
	StatusInterviewing Status = "interviewing"
	StatusOffer        Status = "offer"
	StatusRejected     Status = "rejected"
	StatusAccepted     Status = "accepted"
)

type RemoteType string

const (
	RemoteRemote RemoteType = "remote"
	RemoteHybrid RemoteType = "hybrid"
	RemoteOnsite RemoteType = "onsite"
)

type Pay struct {
	Min      *int
	Max      *int
	Currency *string
}

func (p Pay) String() string {
	if p.Min == nil && p.Max == nil && p.Currency == nil {
		return "Unknown"
	}

	var b strings.Builder

	if p.Currency != nil {
		b.WriteString(*p.Currency)
		b.WriteByte(' ')
	}

	switch {
	case p.Min != nil && p.Max != nil:
		fmt.Fprintf(&b, "%d - %d", *p.Min, *p.Max)
	case p.Min != nil:
		fmt.Fprintf(&b, "%d+", *p.Min)
	case p.Max != nil:
		fmt.Fprintf(&b, "<= %d", *p.Max)
	default:
		// currency only
		b.WriteString("Unknown")
	}

	return strings.TrimSpace(b.String())
}

type JobApplication struct {
	ID int64

	CompanyName     string
	Position        string
	ApplicationDate time.Time
	Status          Status
	Referral        bool
	Remote          *RemoteType

	Location           *string
	Pay                *Pay
	Ranking            *int
	Notes              *string
	JobPostingLink     *string
	CompanyWebsiteLink *string
}

func (a JobApplication) Compare(b JobApplication) bool {
	compareDate := func(t1, t2 time.Time) bool {
		y1, m1, d1 := t1.Date()
		y2, m2, d2 := t2.Date()
		return y1 == y2 && m1 == m2 && d1 == d2
	}
	return a.CompanyName == b.CompanyName &&
		a.Position == b.Position &&
		compareDate(a.ApplicationDate, b.ApplicationDate) &&
		a.Status == b.Status &&
		a.Referral == b.Referral &&
		a.Remote == b.Remote &&
		a.Location == b.Location &&
		a.Pay == b.Pay &&
		a.Ranking == b.Ranking &&
		a.Notes == b.Notes &&
		a.JobPostingLink == b.JobPostingLink &&
		a.CompanyWebsiteLink == b.CompanyWebsiteLink
}

type JobApplicationPatch struct {
	CompanyName     *string
	Position        *string
	ApplicationDate *time.Time
	Status          *Status
	Referral        *bool
	Remote          *RemoteType

	Location           *string
	Pay                *Pay
	Ranking            *int
	Notes              *string
	JobPostingLink     *string
	CompanyWebsiteLink *string
}

func (a *JobApplication) Apply(p JobApplicationPatch) {
	if p.CompanyName != nil {
		a.CompanyName = *p.CompanyName
	}
	if p.Position != nil {
		a.Position = *p.Position
	}
	if p.ApplicationDate != nil {
		a.ApplicationDate = *p.ApplicationDate
	}
	if p.Status != nil {
		a.Status = *p.Status
	}
	if p.Referral != nil {
		a.Referral = *p.Referral
	}
	if p.Remote != nil {
		a.Remote = p.Remote
	}
	if p.Location != nil {
		a.Location = p.Location
	}
	if p.Pay != nil {
		a.Pay = p.Pay
	}
	if p.Ranking != nil {
		a.Ranking = p.Ranking
	}
	if p.Notes != nil {
		a.Notes = p.Notes
	}
	if p.JobPostingLink != nil {
		a.JobPostingLink = p.JobPostingLink
	}
	if p.CompanyWebsiteLink != nil {
		a.CompanyWebsiteLink = p.CompanyWebsiteLink
	}
}
