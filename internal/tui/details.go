package tui

import (
	"strconv"
	"strings"

	"github.com/MrMoneyInTheBank/jobit/internal/model"
	"github.com/charmbracelet/x/ansi"
)

// detailsContent keeps full values independent of the abbreviated list cells.
func detailsContent(stored model.StoredJobApplication, width int) string {
	width = max(1, width)
	application := stored.JobApplication
	referral := "No"
	if application.Referral {
		referral = "Yes"
	}
	date := "—"
	if !application.ApplicationDate.IsZero() {
		date = application.ApplicationDate.UTC().Format("02 Jan 2006")
	}
	fields := [][2]string{
		{"Company", application.CompanyName},
		{"Position", application.Position},
		{"Location", application.Location},
		{"Ranking", ranking(application.Rank)},
		{"Status", string(application.Status)},
		{"Applied", date},
		{"Work mode", string(application.WorkMode)},
		{"Pay", application.Pay.String()},
		{"Referral", referral},
		{"Notes", application.Notes},
		{"Job posting", application.JobPostingLink},
		{"Company website", application.CompanyWebsiteLink},
		{"Record ID", strconv.FormatInt(stored.ID, 10)},
	}
	var sections []string
	for _, field := range fields {
		value := optional(field[1])
		if field[0] == "Notes" && field[1] != "" {
			// Keep paragraphs and indentation, but strip terminal control sequences.
			notes := strings.ReplaceAll(field[1], "\r\n", "\n")
			notes = strings.ReplaceAll(notes, "\t", "    ")
			lines := strings.Split(notes, "\n")
			for i, line := range lines {
				lines[i] = plain(line)
			}
			value = strings.Join(lines, "\n")
		}
		wrapped := ansi.Wrap(value, width, "")
		if field[0] == "Job posting" || field[0] == "Company website" {
			// Hard wrapping keeps every URL character, including long paths and queries.
			wrapped = ansi.Hardwrap(value, width, true)
		}
		sections = append(sections, label.Render(field[0])+"\n"+wrapped)
	}
	return strings.Join(sections, "\n\n")
}

func (jobList *JobList) setDetails() {
	if stored := jobList.selected(); stored != nil {
		jobList.details.SetContent(detailsContent(*stored, jobList.details.Width))
	}
}
