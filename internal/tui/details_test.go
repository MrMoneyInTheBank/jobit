package tui

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/MrMoneyInTheBank/jobit/internal/model"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func TestDetailsFullRecord(t *testing.T) {
	min, max, currency := 100, 200, "GBP"
	app := model.JobApplication{
		CompanyName: "Acme", Position: "Engineer", Location: "London", Rank: 5,
		ApplicationDate: time.Date(2026, 10, 11, 1, 0, 0, 0, time.FixedZone("east", 7200)),
		Status:          model.StatusOffer, WorkMode: model.WorkModeHybrid, Referral: true,
		Pay: &model.Pay{Min: &min, Max: &max, Currency: &currency}, Notes: "First paragraph\n\nSecond paragraph",
		JobPostingLink: "https://example.com/jobs/123", CompanyWebsiteLink: "https://example.com",
	}
	content := ansi.Strip(detailsContent(model.StoredJobApplication{ID: 42, JobApplication: app}, 80))
	for _, pair := range [][2]string{
		{"Company", "Acme"}, {"Position", "Engineer"}, {"Location", "London"}, {"Ranking", "5/5"},
		{"Status", "offer"}, {"Applied", "10 Oct 2026"}, {"Work mode", "hybrid"}, {"Pay", "GBP 100 - 200"},
		{"Referral", "Yes"}, {"Notes", "First paragraph\n\nSecond paragraph"},
		{"Job posting", app.JobPostingLink}, {"Company website", app.CompanyWebsiteLink}, {"Record ID", "42"},
	} {
		if !strings.Contains(content, pair[0]+"\n"+pair[1]) {
			t.Errorf("missing field %q with value %q", pair[0], pair[1])
		}
	}
}

func TestDetailsMissingAndPartialValues(t *testing.T) {
	content := ansi.Strip(detailsContent(model.StoredJobApplication{}, 80))
	for _, field := range []string{"Location", "Status", "Applied", "Work mode", "Notes", "Job posting", "Company website"} {
		if !strings.Contains(content, field+"\n—") {
			t.Errorf("missing placeholder for %s", field)
		}
	}
	if !strings.Contains(content, "Referral\nNo") {
		t.Fatal("false referral omitted")
	}
	zero, maximum, currency := 0, 200, "USD"
	for _, tc := range []struct {
		pay  *model.Pay
		want string
	}{
		{nil, "Unknown"}, {&model.Pay{}, "Unknown"}, {&model.Pay{Min: &zero}, "0+"},
		{&model.Pay{Max: &maximum}, "<= 200"}, {&model.Pay{Currency: &currency}, "USD Unknown"},
	} {
		actual := ansi.Strip(detailsContent(model.StoredJobApplication{JobApplication: model.JobApplication{Pay: tc.pay}}, 80))
		if !strings.Contains(actual, "Pay\n"+tc.want) {
			t.Errorf("missing partial pay %q", tc.want)
		}
	}
}

func TestDetailsLongContent(t *testing.T) {
	url := "https://example.com/" + strings.Repeat("path/", 30) + "?first=one&second=two#fragment"
	notes := "alpha beta gamma\r\n\r\n\t日本語 " + strings.Repeat("longword", 30) + "\x1b[31m"
	stored := model.StoredJobApplication{JobApplication: model.JobApplication{Notes: notes, JobPostingLink: url}}
	for _, width := range []int{20, 40, 80} {
		content := ansi.Strip(detailsContent(stored, width))
		if lipgloss.Width(content) > width {
			t.Fatalf("content exceeds %d cells", width)
		}
		if !strings.Contains(content, "alpha beta gamma\n\n") {
			t.Fatal("paragraph breaks lost")
		}
		link := strings.Split(strings.Split(content, "Job posting\n")[1], "\n\nCompany website")[0]
		if strings.ReplaceAll(link, "\n", "") != url {
			t.Fatal("wrapped URL lost characters")
		}
		if strings.Contains(content, "\x1b") {
			t.Fatal("stored terminal escape leaked")
		}
	}
	// At 12 cells, word wrapping should break between beta and gamma.
	content := ansi.Strip(detailsContent(model.StoredJobApplication{JobApplication: model.JobApplication{Notes: "alpha beta gamma"}}, 12))
	if !strings.Contains(content, "alpha beta\ngamma") {
		t.Fatal("notes split a word")
	}
}

func TestDetailsResizeAndListRestoration(t *testing.T) {
	jobs := controlJobs()
	for i := range jobs {
		jobs[i].JobApplication.Notes = strings.Repeat("Paragraph of readable notes. ", 100)
	}
	j := InitJobList(jobs)
	j.query = "acme"
	j.sortIndex = 3
	j.rebuild(3)
	j.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	before := visibleIDs(&j)
	selected := j.selectedID()
	key(&j, "enter")
	key(&j, "G")
	if !j.details.AtBottom() {
		t.Fatal("G did not reach end")
	}
	j.Update(tea.WindowSizeMsg{Width: 30, Height: 12})
	if !j.details.AtBottom() {
		t.Fatal("resize lost end of document")
	}
	key(&j, "g")
	if j.details.YOffset != 0 {
		t.Fatal("g did not reach top")
	}
	for i := 0; i < 15; i++ {
		key(&j, "j")
	}
	progress := j.details.ScrollPercent()
	j.Update(tea.WindowSizeMsg{Width: 60, Height: 20})
	if math.Abs(j.details.ScrollPercent()-progress) > 0.03 {
		t.Fatal("resize lost relative reading position")
	}
	for _, size := range [][2]int{{24, 9}, {8, 3}, {100, 30}} {
		j.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		if lipgloss.Width(j.View()) > size[0] || lipgloss.Height(j.View()) > size[1] {
			t.Fatal("details overflow")
		}
	}
	key(&j, "esc")
	if j.screen != listScreen || j.query != "acme" || j.sortIndex != 3 || j.selectedID() != selected {
		t.Fatal("back lost list settings")
	}
	if len(before) != len(j.visible) {
		t.Fatal("details changed results")
	}
	key(&j, "enter")
	if j.details.YOffset != 0 {
		t.Fatal("opening details did not start at top")
	}
}

func TestDetailsEmptyResultsStayOnList(t *testing.T) {
	for _, jobs := range [][]model.StoredJobApplication{nil, controlJobs()} {
		j := InitJobList(jobs)
		j.query = "does not exist"
		j.rebuild(0)
		key(&j, "enter")
		if j.screen != listScreen {
			t.Fatal("empty results opened details")
		}
	}
}
