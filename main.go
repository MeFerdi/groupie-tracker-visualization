package main

import (
	"log"
	"net/http"
	"os"

	api "groupie/handlers"
)

func main() {
	// Initialize error template
	api.Init()

	if len(os.Args) != 1 {
		log.Println("Usage: go run main.go")
		return
	}

	http.HandleFunc("/", api.HomeHandler)
	http.HandleFunc("/artists/", api.ArtistsHandler)
	http.HandleFunc("/artist/", api.ArtistHandler) // Consolidated artist handler
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Start the server
	log.Println("Server is running on port 3000...")
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
