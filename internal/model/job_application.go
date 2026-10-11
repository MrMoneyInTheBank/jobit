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

type WorkMode string

const (
	WorkModeRemote WorkMode = "remote"
	WorkModeHybrid WorkMode = "hybrid"
	WorkModeOnsite WorkMode = "onsite"
)

type Pay struct {
	Min      *int
	Max      *int
	Currency *string
}

func (p *Pay) isEmpty() bool {
	return p == nil || (p.Min == nil && p.Max == nil && p.Currency == nil)
}

func (p *Pay) String() string {
	if p.isEmpty() {
		return "Unknown"
	}

	var display strings.Builder

	if p.Currency != nil {
		display.WriteString(*p.Currency)
		display.WriteByte(' ')
	}

	switch {
	case p.Min != nil && p.Max != nil:
		fmt.Fprintf(&display, "%d - %d", *p.Min, *p.Max)
	case p.Min != nil:
		fmt.Fprintf(&display, "%d+", *p.Min)
	case p.Max != nil:
		fmt.Fprintf(&display, "<= %d", *p.Max)
	default:
		display.WriteString("Unknown")
	}

	return strings.TrimSpace(display.String())
}

type Ranking uint8

const (
	RankOne Ranking = iota + 1
	RankTwo
	RankThree
	RankFour
	RankFive
)

func (rank Ranking) Valid() bool {
	return rank >= RankOne && rank <= RankFive
}

type JobApplication struct {
	CompanyName        string
	Position           string
	Location           string
	Pay                *Pay
	WorkMode           WorkMode
	ApplicationDate    time.Time
	Status             Status
	Rank               Ranking
	Referral           bool
	Notes              string
	JobPostingLink     string
	CompanyWebsiteLink string
}

func ptrEq[T comparable](left, right *T) bool {
	if left == right {
		return true
	}
	if left == nil || right == nil {
		return false
	}
	return *left == *right
}

func payEq(left, right *Pay) bool {
	if left.isEmpty() || right.isEmpty() {
		return left.isEmpty() && right.isEmpty()
	}

	return ptrEq(left.Min, right.Min) &&
		ptrEq(left.Max, right.Max) &&
		ptrEq(left.Currency, right.Currency)
}

func sameDay(left, right time.Time) bool {
	leftYear, leftMonth, leftDay := left.UTC().Date()
	rightYear, rightMonth, rightDay := right.UTC().Date()
	return leftYear == rightYear && leftMonth == rightMonth && leftDay == rightDay
}

// Equal compares application data, treating dates as UTC calendar days.
func (application JobApplication) Equal(other JobApplication) bool {
	return application.CompanyName == other.CompanyName &&
		application.Position == other.Position &&
		application.Location == other.Location &&
		payEq(application.Pay, other.Pay) &&
		application.WorkMode == other.WorkMode &&
		sameDay(application.ApplicationDate, other.ApplicationDate) &&
		application.Status == other.Status &&
		application.Rank == other.Rank &&
		application.Referral == other.Referral &&
		application.Notes == other.Notes &&
		application.JobPostingLink == other.JobPostingLink &&
		application.CompanyWebsiteLink == other.CompanyWebsiteLink
}

// JobApplicationPatch uses nil for unchanged fields. Pay points to nil to clear pay.
type JobApplicationPatch struct {
	CompanyName        *string
	Position           *string
	Location           *string
	Pay                **Pay
	WorkMode           *WorkMode
	ApplicationDate    *time.Time
	Status             *Status
	Rank               *Ranking
	Referral           *bool
	Notes              *string
	JobPostingLink     *string
	CompanyWebsiteLink *string
}

func (application *JobApplication) Apply(patch JobApplicationPatch) {
	if patch.CompanyName != nil {
		application.CompanyName = *patch.CompanyName
	}
	if patch.Position != nil {
		application.Position = *patch.Position
	}
	if patch.Location != nil {
		application.Location = *patch.Location
	}
	if patch.Pay != nil {
		application.Pay = *patch.Pay
	}
	if patch.WorkMode != nil {
		application.WorkMode = *patch.WorkMode
	}
	if patch.ApplicationDate != nil {
		application.ApplicationDate = *patch.ApplicationDate
	}
	if patch.Status != nil {
		application.Status = *patch.Status
	}
	if patch.Rank != nil {
		application.Rank = *patch.Rank
	}
	if patch.Referral != nil {
		application.Referral = *patch.Referral
	}
	if patch.Notes != nil {
		application.Notes = *patch.Notes
	}
	if patch.JobPostingLink != nil {
		application.JobPostingLink = *patch.JobPostingLink
	}
	if patch.CompanyWebsiteLink != nil {
		application.CompanyWebsiteLink = *patch.CompanyWebsiteLink
	}
}
