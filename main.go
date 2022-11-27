package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/storage"
)

func main() {
	// Get the port from the env, or default to port 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	// Serve our static files to the root directory
	http.Handle("/", http.FileServer(http.Dir("static")))
	// Our API endpoints
	http.HandleFunc("/api", indexHandler)
	http.HandleFunc("/api/v1/uploadfile", uploadFileHandler)

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func indexHandler(res http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/api" {
		http.NotFound(res, req)
		return
	}
	fmt.Fprint(res, "Hello, World!")
}

func uploadFileHandler(res http.ResponseWriter, req *http.Request) {
	fmt.Fprint(res, "Hello, World!")

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()
}
