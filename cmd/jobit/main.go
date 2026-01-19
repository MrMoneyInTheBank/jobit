package main

import (
	"fmt"
	"log"

	store "github.com/MrMoneyInTheBank/jobit/internal/store"
)

func main() {
	db, err := store.OpenDB(nil)
	if err != nil {
		log.Fatalf("Could not open database: %v", err)
	}
	defer store.CloseDB(db)
	store.InitDB(db)

	var seedCount int
	fmt.Print("How many applications to seed: ")
	fmt.Scanln(&seedCount)
	err = store.SeedDB(db, seedCount)
	if err != nil {
		log.Fatalf("Could not seed database: %v", err)
	}

	fmt.Printf("Seeded database.\n\n")
	apps, err := store.GetJobApplications(db)
	fmt.Printf("Got %v applications\n", len(apps))

	for _, app := range apps {
		fmt.Printf(
			"id: %v, Company Name: %v, Position: %v\n",
			app.ID,
			app.CompanyName,
			app.Position)
	}
}
