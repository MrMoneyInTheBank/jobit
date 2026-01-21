package main

import (
	"log"

	store "github.com/MrMoneyInTheBank/jobit/internal/store"
)

func main() {
	db, err := store.OpenDB(ptr("seed.db"))
	if err != nil {
		log.Fatalf("Could not open database: %v", err)
	}
	defer store.CloseDB(db)
	store.InitDB(db)

	if err != nil {
	}

	}
}
