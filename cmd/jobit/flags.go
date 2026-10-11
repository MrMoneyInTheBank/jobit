package main

import (
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/MrMoneyInTheBank/jobit/internal/store"
	"github.com/charmbracelet/lipgloss"
)

type options struct {
	version         bool
	exportRequested bool
	exportCSV       string
}

var titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
var cardStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63")).Padding(0, 1)

func styledVersion(value string) string { return cardStyle.Render(titleStyle.Render(value)) }
func styledError(err error) string {
	return cardStyle.BorderForeground(lipgloss.Color("196")).Render(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196")).Render("Error") + "  " + err.Error())
}

func parseOptions(args []string, output io.Writer) (options, error) {
	var parsed options
	flags := flag.NewFlagSet("jobit", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.BoolVar(&parsed.version, "version", false, "Print version and exit")
	flags.StringVar(&parsed.exportCSV, "export-csv", "", "Export to a new CSV file, or - for stdout")
	flags.Usage = func() {
		path, err := store.DefaultDatabasePath()
		if err != nil {
			path = "unavailable (" + err.Error() + ")"
		}
		body := strings.Join([]string{
			titleStyle.Render("JOBIT"),
			"Local job application tracker",
			"",
			titleStyle.Render("USAGE"),
			"  jobit [--help] [--version] [--export-csv FILE]",
			"",
			titleStyle.Render("OPTIONS"),
			"  -h, --help           Show this help",
			"  --version           Show the installed version",
			"  --export-csv FILE   Export to a new CSV file; use - for stdout",
			"",
			titleStyle.Render("DATA"),
			"  " + path,
			"  Created on first interactive launch.",
		}, "\n")
		fmt.Fprintln(output, cardStyle.Render(body))
	}
	if err := flags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return parsed, err
		}
		return parsed, err
	}
	flags.Visit(func(f *flag.Flag) {
		if f.Name == "export-csv" {
			parsed.exportRequested = true
		}
	})
	if flags.NArg() != 0 {
		return parsed, fmt.Errorf("unexpected argument %q; use jobit --help", flags.Arg(0))
	}
	if parsed.exportRequested && parsed.exportCSV == "" {
		return parsed, fmt.Errorf("--export-csv requires a filename (or - for stdout)")
	}
	if parsed.version && parsed.exportRequested {
		return parsed, fmt.Errorf("--version and --export-csv cannot be combined")
	}
	return parsed, nil
}

func parseArgs(args []string, output io.Writer) error {
	_, err := parseOptions(args, output)
	return err
}
