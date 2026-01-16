package model

import "time"

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
	JobPositingLink    *string
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
		a.JobPositingLink == b.JobPositingLink &&
		a.CompanyWebsiteLink == b.CompanyWebsiteLink
}
