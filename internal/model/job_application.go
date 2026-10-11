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
		b.WriteString("Unknown")
	}

	return strings.TrimSpace(b.String())
}

type Ranking uint8

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

func ptrEq[T comparable](a, b *T) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func payEq(a, b *Pay) bool {
	if a.isEmpty() || b.isEmpty() {
		return a.isEmpty() && b.isEmpty()
	}

	return ptrEq(a.Min, b.Min) &&
		ptrEq(a.Max, b.Max) &&
		ptrEq(a.Currency, b.Currency)
}

func sameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.UTC().Date()
	y2, m2, d2 := t2.UTC().Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
