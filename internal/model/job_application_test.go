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
