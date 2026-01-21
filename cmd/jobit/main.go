package main

import (
	"log"

	store "github.com/MrMoneyInTheBank/jobit/internal/store"
)

func ptr[T any](v T) *T {
	return &v
}

func main() {
	db, err := store.OpenDB(ptr("seed.db"))
	if err != nil {
		log.Fatalf("Could not open database: %v", err)
	}
	defer store.CloseDB(db)
	store.InitDB(db)

	job_apps, err := store.GetJobApplications(db)
	if err != nil {
		log.Fatalf("Could not get job applications: %v", err)
	}

	}
}
