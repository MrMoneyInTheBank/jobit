package tui

import (
	"strconv"

	"github.com/MrMoneyInTheBank/jobit/internal/model"
	tea "github.com/charmbracelet/bubbletea"
)

type UpdateApplication func(model.StoredJobApplication) error
type DeleteApplication func(int64) error
type deletedMsg struct {
	id  int64
	err error
}

// SetMutationHandlers enables editing and deletion using the caller's repository.
func (jl *JobList) SetMutationHandlers(update UpdateApplication, remove DeleteApplication) {
	jl.update = update
	jl.remove = remove
}

func fieldsFromApplication(a model.JobApplication) [fieldCount]string {
	var f [fieldCount]string
	f[companyField] = a.CompanyName
	f[positionField] = a.Position
	f[locationField] = a.Location
	if a.Rank != 0 {
		f[rankField] = strconv.Itoa(int(a.Rank))
	}
	if !a.ApplicationDate.IsZero() {
		f[dateField] = a.ApplicationDate.UTC().Format("2006-01-02")
	}
	f[statusField] = string(a.Status)
	f[referralField] = "no"
	if a.Referral {
		f[referralField] = "yes"
	}
	f[modeField] = string(a.WorkMode)
	f[notesField] = a.Notes
	f[postingField] = a.JobPostingLink
	f[websiteField] = a.CompanyWebsiteLink
	if a.Pay != nil {
		if a.Pay.Min != nil {
			f[minField] = strconv.Itoa(*a.Pay.Min)
		}
		if a.Pay.Max != nil {
			f[maxField] = strconv.Itoa(*a.Pay.Max)
		}
		if a.Pay.Currency != nil {
			f[currencyField] = *a.Pay.Currency
		}
	}
	return f
}
func (jl *JobList) beginEdit(statusOnly bool) {
	stored := jl.selected()
	if stored == nil {
		return
	}
	if jl.update == nil {
		jl.notice = "Editing is unavailable"
		return
	}
	jl.returnScreen = jl.screen
	jl.editID = stored.ID
	jl.fields = fieldsFromApplication(stored.JobApplication)
	jl.field = companyField
	if statusOnly {
		jl.field = statusField
	}
	jl.cursor = len([]rune(jl.fields[jl.field]))
	jl.formError = ""
	jl.screen = addScreen
}
func (jl *JobList) updateDelete(msg tea.KeyMsg) tea.Cmd {
	if jl.deleting {
		return nil
	}
	switch msg.String() {
	case "esc", "n":
		jl.screen = jl.returnScreen
		jl.notice = ""
	case "enter":
		id := jl.selectedID()
		if id == 0 || jl.remove == nil {
			return nil
		}
		jl.deleting = true
		jl.notice = ""
		remove := jl.remove
		return func() tea.Msg { return deletedMsg{id, remove(id)} }
	}
	return nil
}
