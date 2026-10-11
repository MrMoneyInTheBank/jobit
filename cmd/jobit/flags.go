package main

import (
	"flag"
	"fmt"
	"io"

	"github.com/MrMoneyInTheBank/jobit/internal/store"
)

type options struct {
	version   bool
	exportCSV string
}

func parseOptions(args []string, output io.Writer) (options, error) {
	var parsed options
	flags := flag.NewFlagSet("jobit", flag.ContinueOnError)
	flags.SetOutput(output)
	flags.BoolVar(&parsed.version, "version", false, "Print version and exit")
	flags.StringVar(&parsed.exportCSV, "export-csv", "", "Export all applications to a new CSV file, or - for stdout")
	flags.Usage = func() {
		fmt.Fprintln(output, "Jobit — local job application tracker\n\nUsage: jobit [--help] [--version] [--export-csv FILE]\n\n  -h, --help  Show this help without opening the database.")
		flags.PrintDefaults()
		path, err := store.DefaultDatabasePath()
		if err != nil {
			fmt.Fprintf(output, "\nDefault database: unavailable (%v)\n", err)
		} else {
			fmt.Fprintf(output, "\nDefault database: %s\n", path)
		}
		effectivePath, err := store.DatabaseDSN(nil)
		if err != nil {
			fmt.Fprintf(output, "Effective database: unavailable (%v)\n", err)
		} else {
			fmt.Fprintf(output, "Effective database: %s\n", effectivePath)
		}
		fmt.Fprintln(output, "\nOverride with JOBIT_DB_DSN (a SQLite path or DSN).\nRelative overrides resolve from the working directory.\nThe default data directory and database are created on first launch.\nNo repository or SQLite server is required.")
	}

	if err := flags.Parse(args); err != nil {
		return parsed, err
	}
	if flags.NArg() != 0 {
		return parsed, fmt.Errorf("unexpected argument %q; use jobit --help", flags.Arg(0))
	}
	if parsed.version && parsed.exportCSV != "" {
		return parsed, fmt.Errorf("--version and --export-csv cannot be combined")
	}
	return parsed, nil
}

func parseArgs(args []string, output io.Writer) error {
	_, err := parseOptions(args, output)
	return err
}
