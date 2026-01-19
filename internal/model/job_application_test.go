package model

import (
	"testing"
	"time"
)

func TestJobApplicationCompare(t *testing.T) {
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

	if !a.Compare(b) {
		t.Fatalf("Job application mismatch")
	}

	if a.Compare(c) {
		t.Fatalf("Job application mismatch")
	}
}
