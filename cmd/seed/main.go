package main

import (
	"log"
	"os"

	"github.com/MrMoneyInTheBank/jobit/internal/store"
)

func main() {
	var dsn *string
	if os.Getenv("JOBIT_DB_DSN") == "" {
		path := "./dev.db"
		dsn = &path
	}
	db, err := store.OpenDB(dsn)
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
		log.Fatalf("Could not seed database: %v", err)
	}
	log.Println("Seeded database.")
}
