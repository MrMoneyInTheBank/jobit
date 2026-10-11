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
func (jobList *JobList) SetMutationHandlers(update UpdateApplication, remove DeleteApplication) {
	jobList.update = update
	jobList.remove = remove
}

func fieldsFromApplication(application model.JobApplication) [fieldCount]string {
	var fields [fieldCount]string
	fields[companyField] = application.CompanyName
	fields[positionField] = application.Position
	fields[locationField] = application.Location
	fields[rankField] = strconv.Itoa(int(application.Rank))
	if !application.ApplicationDate.IsZero() {
		fields[dateField] = application.ApplicationDate.UTC().Format("2006-01-02")
	}
	fields[statusField] = string(application.Status)
	fields[referralField] = "no"
	if application.Referral {
		fields[referralField] = "yes"
	}
	fields[modeField] = string(application.WorkMode)
	fields[notesField] = application.Notes
	fields[postingField] = application.JobPostingLink
	fields[websiteField] = application.CompanyWebsiteLink
	if application.Pay != nil {
		if application.Pay.Min != nil {
			fields[minField] = strconv.Itoa(*application.Pay.Min)
		}
		if application.Pay.Max != nil {
			fields[maxField] = strconv.Itoa(*application.Pay.Max)
		}
		if application.Pay.Currency != nil {
			fields[currencyField] = *application.Pay.Currency
		}
	}
	return fields
}

func (jobList *JobList) beginEdit(statusOnly bool) {
	stored := jobList.selected()
	if stored == nil {
		return
	}
	if jobList.update == nil {
		jobList.notice = "Editing is unavailable"
		return
	}
	jobList.returnScreen = jobList.screen
	jobList.editID = stored.ID
	jobList.fields = fieldsFromApplication(stored.JobApplication)
	jobList.field = companyField
	if statusOnly {
		jobList.field = statusField
	}
	jobList.cursor = len([]rune(jobList.fields[jobList.field]))
	jobList.formError = ""
	jobList.screen = addScreen
}

func (jobList *JobList) updateDelete(msg tea.KeyMsg) tea.Cmd {
	if jobList.deleting {
		return nil
	}
	switch msg.String() {
	case "esc", "n":
		jobList.screen = jobList.returnScreen
		jobList.notice = ""
	case "enter":
		id := jobList.selectedID()
		if id == 0 || jobList.remove == nil {
			return nil
		}
		jobList.deleting = true
		jobList.notice = ""
		remove := jobList.remove
		return func() tea.Msg { return deletedMsg{id, remove(id)} }
	}
	return nil
}
