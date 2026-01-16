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
