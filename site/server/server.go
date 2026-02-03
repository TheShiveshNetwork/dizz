package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

// @ignore-unused
func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create file server for public directory
	fs := http.FileServer(http.Dir("../public"))

	// Handle all requests
	http.Handle("/", fs)

	fmt.Printf("Server starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
