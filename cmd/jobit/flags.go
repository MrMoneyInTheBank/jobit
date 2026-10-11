package main

import (
	"flag"
	"fmt"
	"io"

	"github.com/MrMoneyInTheBank/jobit/internal/store"
)

func parseArgs(args []string, output io.Writer) error {
	flags := flag.NewFlagSet("jobit", flag.ContinueOnError)
	flags.SetOutput(output)
	flags.Usage = func() {
		fmt.Fprintln(output, "Jobit — local job application tracker\n\nUsage: jobit [--help]\n\n  -h, --help  Show this help without opening the database.")
		path, err := store.DefaultDatabasePath()
		if err != nil {
			fmt.Fprintf(output, "\nDefault database: unavailable (%v)\n", err)
		} else {
			fmt.Fprintf(output, "\nDefault database: %s\n", path)
		}
		effective, err := store.DatabaseDSN(nil)
		if err != nil {
			fmt.Fprintf(output, "Effective database: unavailable (%v)\n", err)
		} else {
			fmt.Fprintf(output, "Effective database: %s\n", effective)
		}
		fmt.Fprintln(output, "\nOverride with JOBIT_DB_DSN (a SQLite path or DSN).\nRelative overrides resolve from the working directory.\nThe default data directory and database are created on first launch.\nNo repository or SQLite server is required.")
	}
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected argument %q; use jobit --help", flags.Arg(0))
	}
	return nil
}
