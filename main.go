package main

import (
	"log"
)

func main() {
	// Initialize a new Postgres store.
	store, err := NewPostgresStore()
	if err != nil {
		log.Fatal(err)
	}

	// Initialize the store.
	if err := store.Init(); err != nil {
		log.Fatal(err)
	}

	// Create a new API server with the specified address and store and run it.
	server := NewAPIServer(":8080", store)
	server.Run()
}