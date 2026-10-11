package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"testing"
)

func TestRejectInvalidTyping(t *testing.T) {
	j := InitJobList(nil)
	key(&j, "a")
	for _, field := range []int{statusField, referralField, modeField} {
		j.field = field
		before := j.fields[field]
		key(&j, "garbage")
		if j.fields[field] != before {
			t.Fatal("selector accepted arbitrary text")
		}
		j.Update(tea.KeyMsg{Type: tea.KeyBackspace})
		j.Update(tea.KeyMsg{Type: tea.KeyDelete})
		if j.fields[field] != before {
			t.Fatal("selector allowed partial deletion")
		}
		j.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
		if !contains(formFields[field].choices, j.fields[field]) {
			t.Fatal("clear created invalid choice")
		}
	}
	for _, tc := range []struct {
		field int
		value string
	}{
		{rankField, "6"}, {rankField, "0"}, {rankField, "-1"}, {minField, "1.5"}, {maxField, "-3"},
		{minField, "1,000"}, {currencyField, "12!"}, {currencyField, "DOLLARS"},
		{dateField, "2026-02-30"}, {dateField, "2026-13-01"}, {dateField, "tomorrow"},
	} {
		j.field = tc.field
		j.fields[tc.field] = ""
		j.cursor = 0
		key(&j, tc.value)
		if j.fields[tc.field] != "" {
			t.Fatalf("accepted invalid field %d: %s", tc.field, tc.value)
		}
	}
	f := validFields()
	for _, field := range []int{postingField, websiteField} {
		f[field] = "javascript:alert(1)"
		if _, _, err := applicationFromFields(f); err == nil {
			t.Fatal("accepted invalid URL")
		}
		f[field] = "https://example.com/job"
	}
	if _, _, err := applicationFromFields(f); err != nil {
		t.Fatal(err)
	}
}
