package main

import (
	"log"

	"github.com/MrMoneyInTheBank/jobit/internal/store"
)

func main() {
	db, err := store.OpenDB(nil)
	if err != nil {
		log.Fatalf("Could not open database: %v", err)
	}
	log.Println("Opened database.")

	defer store.CloseDB(db)
	if err := store.InitDB(db); err != nil {
		log.Fatalf("Could not initialize database: %v", err)
	}

	err = store.SeedDB(db, 50)
	if err != nil {
		log.Fatalf("Could not seed database.")
	}
	log.Println("Seeded database.")
}
