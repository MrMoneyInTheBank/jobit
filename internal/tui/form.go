package tui

import (
	"fmt"
	"net/url"
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
	{"Ranking *", "Required · 1–5", nil},
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
	var values [fieldCount]string
	values[dateField] = time.Now().UTC().Format("2006-01-02")
	values[statusField] = string(model.StatusSubmitted)
	values[referralField] = "no"
	return values
}

func applicationFromFields(fields [fieldCount]string) (model.JobApplication, int, error) {
	values := fields
	for fieldIndex := range values {
		values[fieldIndex] = strings.TrimSpace(values[fieldIndex])
	}

	fail := func(fieldIndex int, message string) (model.JobApplication, int, error) {
		return model.JobApplication{}, fieldIndex, fmt.Errorf("%s", message)
	}
	if values[companyField] == "" {
		return fail(companyField, "Company is required")
	}
	if values[positionField] == "" {
		return fail(positionField, "Position is required")
	}
	applicationDate, err := time.Parse("2006-01-02", values[dateField])
	if err != nil || applicationDate.IsZero() {
		return fail(dateField, "Use a valid date: YYYY-MM-DD")
	}
	status := strings.ToLower(values[statusField])
	if !contains(formFields[statusField].choices, status) {
		return fail(statusField, "Choose a valid application status")
	}
	referral := strings.ToLower(values[referralField])
	if !contains(formFields[referralField].choices, referral) {
		return fail(referralField, "Referral must be yes or no")
	}
	workMode := strings.ToLower(values[modeField])
	if !contains(formFields[modeField].choices, workMode) {
		return fail(modeField, "Work mode must be remote, hybrid, onsite, or blank")
	}
	rank, err := strconv.Atoi(values[rankField])
	if err != nil || !model.Ranking(rank).Valid() {
		return fail(rankField, "Ranking must be 1–5")
	}
	var pay model.Pay
	for _, fieldIndex := range []int{minField, maxField} {
		if values[fieldIndex] == "" {
			continue
		}
		amount, parseErr := strconv.Atoi(values[fieldIndex])
		if parseErr != nil || amount < 0 {
			return fail(fieldIndex, "Pay must be a nonnegative whole number")
		}
		if fieldIndex == minField {
			pay.Min = &amount
		} else {
			pay.Max = &amount
		}
	}
	if pay.Min != nil && pay.Max != nil && *pay.Min > *pay.Max {
		return fail(maxField, "Maximum pay must be at least minimum pay")
	}
	if values[currencyField] != "" {
		currency := strings.ToUpper(values[currencyField])
		if len(currency) != 3 {
			return fail(currencyField, "Use a 3-letter currency code")
		}
		for _, character := range currency {
			if character < 'A' || character > 'Z' {
				return fail(currencyField, "Use a 3-letter currency code")
			}
		}
		pay.Currency = &currency
	}
	for _, fieldIndex := range []int{postingField, websiteField} {
		if values[fieldIndex] == "" {
			continue
		}
		parsedURL, err := url.ParseRequestURI(values[fieldIndex])
		if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") || parsedURL.Hostname() == "" {
			return fail(fieldIndex, "Use a full http:// or https:// URL")
		}
	}
	application := model.JobApplication{
		CompanyName:        values[companyField],
		Position:           values[positionField],
		Location:           values[locationField],
		Rank:               model.Ranking(rank),
		ApplicationDate:    applicationDate,
		Status:             model.Status(status),
		Referral:           referral == "yes",
		WorkMode:           model.WorkMode(workMode),
		Notes:              values[notesField],
		JobPostingLink:     values[postingField],
		CompanyWebsiteLink: values[websiteField],
	}
	if pay.Min != nil || pay.Max != nil || pay.Currency != nil {
		application.Pay = &pay
	}
	return application, 0, nil
}

func contains(values []string, target string) bool {
	for _, choice := range values {
		if choice == target {
			return true
		}
	}
	return false
}

func (jobList *JobList) moveField(delta int) {
	jobList.field = (jobList.field + delta + fieldCount) % fieldCount
	jobList.cursor = len([]rune(jobList.fields[jobList.field]))
}

func (jobList *JobList) insertText(text string) {
	if jobList.field != notesField {
		text = plain(text)
	} else {
		text = strings.ReplaceAll(text, "\r\n", "\n")
		text = strings.Map(func(character rune) rune {
			if character == '\n' || character == '\t' {
				return character
			}
			if unicode.IsControl(character) {
				return -1
			}
			return character
		}, ansi.Strip(text))
	}
	currentValue := []rune(jobList.fields[jobList.field])
	jobList.cursor = min(jobList.cursor, len(currentValue))
	input := []rune(text)
	candidate := string(currentValue[:jobList.cursor]) + string(input) + string(currentValue[jobList.cursor:])
	if !validTypedValue(jobList.field, candidate) {
		jobList.formError = "Input rejected · " + formFields[jobList.field].hint
		return
	}
	jobList.fields[jobList.field] = candidate
	jobList.formError = ""
	jobList.cursor += len(input)
}

func (jobList *JobList) updateForm(msg tea.KeyMsg) tea.Cmd {
	if jobList.saving {
		return nil
	}
	currentValue := []rune(jobList.fields[jobList.field])
	jobList.cursor = min(jobList.cursor, len(currentValue))
	switch msg.String() {
	case "esc":
		jobList.screen = jobList.returnScreen
		jobList.notice = "Draft discarded"
	case "tab", "down":
		jobList.moveField(1)
	case "shift+tab", "up":
		jobList.moveField(-1)
	case "enter":
		if jobList.field == notesField {
			jobList.insertText("\n")
		} else {
			jobList.moveField(1)
		}
	case "left", "right":
		choices := formFields[jobList.field].choices
		delta := 1
		if msg.String() == "left" {
			delta = -1
		}
		if len(choices) > 0 {
			current := 0
			for index, choice := range choices {
				if choice == jobList.fields[jobList.field] {
					current = index
				}
			}
			jobList.fields[jobList.field] = choices[(current+delta+len(choices))%len(choices)]
			jobList.cursor = len([]rune(jobList.fields[jobList.field]))
			jobList.formError = ""
		} else {
			jobList.cursor = max(0, min(len(currentValue), jobList.cursor+delta))
		}
	case "home", "ctrl+a":
		jobList.cursor = 0
	case "end", "ctrl+e":
		jobList.cursor = len(currentValue)
	case "ctrl+u":
		if choices := formFields[jobList.field].choices; len(choices) > 0 {
			jobList.fields[jobList.field] = choices[0]
			jobList.cursor = len([]rune(choices[0]))
		} else {
			jobList.fields[jobList.field] = ""
			jobList.cursor = 0
		}
	case "backspace", "ctrl+h":
		if len(formFields[jobList.field].choices) > 0 {
			return nil
		}
		if jobList.cursor > 0 {
			jobList.fields[jobList.field] = string(currentValue[:jobList.cursor-1]) + string(currentValue[jobList.cursor:])
			jobList.cursor--
		}
	case "delete":
		if len(formFields[jobList.field].choices) > 0 {
			return nil
		}
		if jobList.cursor < len(currentValue) {
			jobList.fields[jobList.field] = string(currentValue[:jobList.cursor]) + string(currentValue[jobList.cursor+1:])
		}
	case "ctrl+s":
		if (jobList.editID == 0 && jobList.save == nil) || (jobList.editID != 0 && jobList.update == nil) {
			jobList.formError = "Saving is unavailable in this view"
			return nil
		}
		application, field, err := applicationFromFields(jobList.fields)
		if err != nil {
			jobList.formError = err.Error()
			jobList.field = field
			jobList.cursor = len([]rune(jobList.fields[field]))
			return nil
		}
		jobList.saving = true
		jobList.formError = ""
		if jobList.editID != 0 {
			stored := model.StoredJobApplication{ID: jobList.editID, JobApplication: application}
			update := jobList.update
			return func() tea.Msg {
				return savedMsg{stored, update(stored)}
			}
		}
		save := jobList.save
		return func() tea.Msg {
			stored, err := save(application)
			return savedMsg{stored, err}
		}
	default:
		if msg.Type == tea.KeyRunes {
			jobList.insertText(string(msg.Runes))
		}
		if msg.Type == tea.KeySpace {
			jobList.insertText(" ")
		}
	}
	return nil
}

func (jobList *JobList) formView() string {
	contentWidth := jobList.contentWidth()
	contentHeight := max(1, jobList.height-8)
	var lines []string
	activeLine := 0
	for fieldIndex, field := range formFields {
		marker := "  "
		if fieldIndex == jobList.field {
			marker = "› "
		}
		lines = append(lines, label.Render(marker+field.name))
		value := jobList.fields[fieldIndex]
		if fieldIndex == jobList.field {
			characters := []rune(value)
			cursor := min(jobList.cursor, len(characters))
			value = string(characters[:cursor]) + "▏" + string(characters[cursor:])
			// Show the cursor's line even for long notes and narrow terminals.
			before := ansi.Hardwrap(strings.ReplaceAll(string(characters[:cursor]), "\t", "    "), max(1, contentWidth-2), true)
			activeLine = len(lines) + strings.Count(before, "\n")
		} else if value == "" {
			value = "—"
		}
		if len(field.choices) > 0 {
			value = jobList.fields[fieldIndex]
			if value == "" {
				value = "Unspecified"
			}
			value = "‹ " + value + " ›"
		}
		wrapped := ansi.Hardwrap(strings.ReplaceAll(value, "\t", "    "), max(1, contentWidth-2), true)
		for _, line := range strings.Split(wrapped, "\n") {
			lines = append(lines, "  "+line)
		}
		lines = append(lines, "")
	}
	start := max(0, activeLine-contentHeight/2)
	start = min(start, max(0, len(lines)-contentHeight))
	return strings.Join(lines[start:min(len(lines), start+contentHeight)], "\n")
}

// Partial dates and currency codes may be incomplete while editing; saving
// still validates the complete value and cross-field constraints.
func validTypedValue(field int, value string) bool {
	if len(formFields[field].choices) > 0 {
		return false
	}
	switch field {
	case rankField:
		return value == "" || (len(value) == 1 && value[0] >= '1' && value[0] <= '5')
	case minField, maxField:
		if value == "" {
			return true
		}
		for _, r := range value {
			if r < '0' || r > '9' {
				return false
			}
		}
		_, err := strconv.Atoi(value)
		return err == nil
	case dateField:
		if len(value) > 10 {
			return false
		}
		for i, r := range value {
			if i == 4 || i == 7 {
				if r != '-' {
					return false
				}
			} else if r < '0' || r > '9' {
				return false
			}
		}
		if len(value) >= 7 {
			month, _ := strconv.Atoi(value[5:7])
			if month < 1 || month > 12 {
				return false
			}
		}
		if len(value) == 10 {
			date, err := time.Parse("2006-01-02", value)
			return err == nil && !date.IsZero()
		}
		return true
	case currencyField:
		if len(value) > 3 {
			return false
		}
		for _, r := range value {
			if !(r >= 'a' && r <= 'z') && !(r >= 'A' && r <= 'Z') {
				return false
			}
		}
	}
	return true
}
