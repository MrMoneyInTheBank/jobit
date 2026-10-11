package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/MrMoneyInTheBank/jobit/internal/model"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

const (
	companyField = iota
	positionField
	locationField
	rankField
	dateField
	statusField
	referralField
	modeField
	minField
	maxField
	currencyField
	notesField
	postingField
	websiteField
	fieldCount
)

type formField struct {
	name, hint string
	choices    []string
}

var formFields = [fieldCount]formField{
	{"Company *", "Required · company name", nil},
	{"Position *", "Required · role or job title", nil},
	{"Location", "Optional · city, region, or remote", nil},
	{"Ranking", "Blank or 0 = unranked · otherwise 1–5", nil},
	{"Application date *", "YYYY-MM-DD · defaults to today in UTC", nil},
	{"Status *", "←/→ choose · submitted, interviewing, offer, rejected, accepted", []string{"submitted", "interviewing", "offer", "rejected", "accepted"}},
	{"Referral *", "←/→ choose · yes or no", []string{"no", "yes"}},
	{"Work mode", "←/→ choose · blank, remote, hybrid, onsite", []string{"", "remote", "hybrid", "onsite"}},
	{"Minimum pay", "Optional · nonnegative whole amount · no commas", nil},
	{"Maximum pay", "Optional · whole amount, at least minimum", nil},
	{"Currency", "Optional · 3-letter code, e.g. USD, EUR, INR", nil},
	{"Notes", "Optional · Enter inserts newline · Tab moves on", nil},
	{"Job posting link", "Optional · paste the full link", nil},
	{"Company website", "Optional · paste the full link", nil},
}

func defaultFields() [fieldCount]string {
	var f [fieldCount]string
	f[dateField] = time.Now().UTC().Format("2006-01-02")
	f[statusField] = string(model.StatusSubmitted)
	f[referralField] = "no"
	return f
}

func applicationFromFields(fields [fieldCount]string) (model.JobApplication, int, error) {
	f := fields
	for i := range f {
		f[i] = strings.TrimSpace(f[i])
	}
	fail := func(i int, message string) (model.JobApplication, int, error) {
		return model.JobApplication{}, i, fmt.Errorf("%s", message)
	}
	if f[companyField] == "" {
		return fail(companyField, "Company is required")
	}
	if f[positionField] == "" {
		return fail(positionField, "Position is required")
	}
	date, err := time.Parse("2006-01-02", f[dateField])
	if err != nil || date.IsZero() {
		return fail(dateField, "Use a valid date: YYYY-MM-DD")
	}
	status := strings.ToLower(f[statusField])
	if !contains(formFields[statusField].choices, status) {
		return fail(statusField, "Choose a valid application status")
	}
	referral := strings.ToLower(f[referralField])
	if !contains(formFields[referralField].choices, referral) {
		return fail(referralField, "Referral must be yes or no")
	}
	mode := strings.ToLower(f[modeField])
	if !contains(formFields[modeField].choices, mode) {
		return fail(modeField, "Work mode must be remote, hybrid, onsite, or blank")
	}
	rank := 0
	if f[rankField] != "" {
		rank, err = strconv.Atoi(f[rankField])
		if err != nil || rank < 0 || rank > 5 {
			return fail(rankField, "Ranking must be 0–5 or blank")
		}
	}
	var pay model.Pay
	for _, i := range []int{minField, maxField} {
		if f[i] == "" {
			continue
		}
		v, e := strconv.Atoi(f[i])
		if e != nil || v < 0 {
			return fail(i, "Pay must be a nonnegative whole number")
		}
		if i == minField {
			pay.Min = &v
		} else {
			pay.Max = &v
		}
	}
	if pay.Min != nil && pay.Max != nil && *pay.Min > *pay.Max {
		return fail(maxField, "Maximum pay must be at least minimum pay")
	}
	if f[currencyField] != "" {
		currency := strings.ToUpper(f[currencyField])
		if len(currency) != 3 {
			return fail(currencyField, "Use a 3-letter currency code")
		}
		for _, r := range currency {
			if r < 'A' || r > 'Z' {
				return fail(currencyField, "Use a 3-letter currency code")
			}
		}
		pay.Currency = &currency
	}
	app := model.JobApplication{CompanyName: f[companyField], Position: f[positionField], Location: f[locationField], Rank: model.Ranking(rank), ApplicationDate: date, Status: model.Status(status), Referral: referral == "yes", WorkMode: model.WorkMode(mode), Notes: f[notesField], JobPostingLink: f[postingField], CompanyWebsiteLink: f[websiteField]}
	if pay.Min != nil || pay.Max != nil || pay.Currency != nil {
		app.Pay = &pay
	}
	return app, 0, nil
}
func contains(values []string, s string) bool {
	for _, v := range values {
		if v == s {
			return true
		}
	}
	return false
}

func (jl *JobList) moveField(delta int) {
	jl.field = (jl.field + delta + fieldCount) % fieldCount
	jl.cursor = len([]rune(jl.fields[jl.field]))
}
func (jl *JobList) insertText(text string) {
	if jl.field != notesField {
		text = plain(text)
	} else {
		text = strings.ReplaceAll(text, "\r\n", "\n")
		text = strings.Map(func(r rune) rune {
			if r == '\n' || r == '\t' {
				return r
			}
			if unicode.IsControl(r) {
				return -1
			}
			return r
		}, ansi.Strip(text))
	}
	r := []rune(jl.fields[jl.field])
	jl.cursor = min(jl.cursor, len(r))
	input := []rune(text)
	jl.fields[jl.field] = string(r[:jl.cursor]) + string(input) + string(r[jl.cursor:])
	jl.cursor += len(input)
}
func (jl *JobList) updateForm(msg tea.KeyMsg) tea.Cmd {
	if jl.saving {
		return nil
	}
	r := []rune(jl.fields[jl.field])
	jl.cursor = min(jl.cursor, len(r))
	switch msg.String() {
	case "esc":
		jl.screen = listScreen
		jl.notice = "Draft discarded"
	case "tab", "down":
		jl.moveField(1)
	case "shift+tab", "up":
		jl.moveField(-1)
	case "enter":
		if jl.field == notesField {
			jl.insertText("\n")
		} else {
			jl.moveField(1)
		}
	case "left", "right":
		choices := formFields[jl.field].choices
		delta := 1
		if msg.String() == "left" {
			delta = -1
		}
		if len(choices) > 0 {
			current := 0
			for i, v := range choices {
				if v == jl.fields[jl.field] {
					current = i
				}
			}
			jl.fields[jl.field] = choices[(current+delta+len(choices))%len(choices)]
			jl.cursor = len([]rune(jl.fields[jl.field]))
		} else {
			jl.cursor = max(0, min(len(r), jl.cursor+delta))
		}
	case "home", "ctrl+a":
		jl.cursor = 0
	case "end", "ctrl+e":
		jl.cursor = len(r)
	case "ctrl+u":
		jl.fields[jl.field] = ""
		jl.cursor = 0
	case "backspace", "ctrl+h":
		if jl.cursor > 0 {
			jl.fields[jl.field] = string(r[:jl.cursor-1]) + string(r[jl.cursor:])
			jl.cursor--
		}
	case "delete":
		if jl.cursor < len(r) {
			jl.fields[jl.field] = string(r[:jl.cursor]) + string(r[jl.cursor+1:])
		}
	case "ctrl+s":
		if jl.save == nil {
			jl.formError = "Saving is unavailable in this view"
			return nil
		}
		app, field, err := applicationFromFields(jl.fields)
		if err != nil {
			jl.formError = err.Error()
			jl.field = field
			jl.cursor = len([]rune(jl.fields[field]))
			return nil
		}
		jl.saving = true
		jl.formError = ""
		save := jl.save
		return func() tea.Msg { stored, err := save(app); return savedMsg{stored, err} }
	default:
		if msg.Type == tea.KeyRunes {
			jl.insertText(string(msg.Runes))
		}
		if msg.Type == tea.KeySpace {
			jl.insertText(" ")
		}
	}
	return nil
}

func (jl *JobList) formView() string {
	w := jl.contentWidth()
	h := max(1, jl.height-8)
	var lines []string
	active := 0
	for i, f := range formFields {
		marker := "  "
		if i == jl.field {
			marker = "› "
		}
		lines = append(lines, label.Render(marker+f.name))
		value := jl.fields[i]
		if i == jl.field {
			r := []rune(value)
			cursor := min(jl.cursor, len(r))
			value = string(r[:cursor]) + "▏" + string(r[cursor:])
			// Show the cursor's line even for long notes and narrow terminals.
			before := ansi.Hardwrap(strings.ReplaceAll(string(r[:cursor]), "\t", "    "), max(1, w-2), true)
			active = len(lines) + strings.Count(before, "\n")
		} else if value == "" {
			value = "—"
		}
		wrapped := ansi.Hardwrap(strings.ReplaceAll(value, "\t", "    "), max(1, w-2), true)
		for _, line := range strings.Split(wrapped, "\n") {
			lines = append(lines, "  "+line)
		}
		lines = append(lines, "")
	}
	start := max(0, active-h/2)
	start = min(start, max(0, len(lines)-h))
	return strings.Join(lines[start:min(len(lines), start+h)], "\n")
}
