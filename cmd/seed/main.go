package main

import (
	"log"

	"github.com/MrMoneyInTheBank/jobit/internal/store"
)

func ptr[T any](v T) *T {
	return &v
}

func main() {
	db, err := store.OpenDB(ptr("seed.db"))
	if err != nil {
		log.Fatalf("Could not open database: %v", err)
	}
	log.Printf("Opened database at %v\n", "seed.db")

	defer store.CloseDB(db)
	store.InitDB(db)

	err = store.SeedDB(db, 50)
	if err != nil {
		log.Fatalf("Could not seed database.")
	}
	log.Println("Seeded database.")
}
