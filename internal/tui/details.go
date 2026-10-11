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
	j := stored.JobApplication
	referral := "No"
	if j.Referral {
		referral = "Yes"
	}
	date := "—"
	if !j.ApplicationDate.IsZero() {
		date = j.ApplicationDate.UTC().Format("02 Jan 2006")
	}
	fields := [][2]string{
		{"Company", j.CompanyName},
		{"Position", j.Position},
		{"Location", j.Location},
		{"Ranking", ranking(j.Rank)},
		{"Status", string(j.Status)},
		{"Applied", date},
		{"Work mode", string(j.WorkMode)},
		{"Pay", j.Pay.String()},
		{"Referral", referral},
		{"Notes", j.Notes},
		{"Job posting", j.JobPostingLink},
		{"Company website", j.CompanyWebsiteLink},
		{"Record ID", strconv.FormatInt(stored.ID, 10)},
	}
	var parts []string
	for _, f := range fields {
		value := optional(f[1])
		if f[0] == "Notes" && f[1] != "" {
			// Keep paragraphs and indentation, but strip terminal control sequences.
			notes := strings.ReplaceAll(f[1], "\r\n", "\n")
			notes = strings.ReplaceAll(notes, "\t", "    ")
			lines := strings.Split(notes, "\n")
			for i, line := range lines {
				lines[i] = plain(line)
			}
			value = strings.Join(lines, "\n")
		}
		wrapped := ansi.Wrap(value, width, "")
		if f[0] == "Job posting" || f[0] == "Company website" {
			// Hard wrapping keeps every URL character, including long paths and queries.
			wrapped = ansi.Hardwrap(value, width, true)
		}
		parts = append(parts, label.Render(f[0])+"\n"+wrapped)
	}
	return strings.Join(parts, "\n\n")
}

func (jl *JobList) setDetails() {
	if stored := jl.selected(); stored != nil {
		jl.details.SetContent(detailsContent(*stored, jl.details.Width))
	}
}
