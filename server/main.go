package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"choir-sync/cloudstorage"
)

// The bucket we are using for storage for this app
const bucketName = "mytestbucketforchoirsync"

// Our json respone schema
type response struct {
	Message string
}

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
	http.HandleFunc("/api/v1/test", testHandler)
	http.HandleFunc("/api/v1/uploadfile", uploadFileHandler)

	// Start listening on port specified
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
	log.Printf("Listening on port %s", port)
}

// A test function to check if the API is up
func testHandler(resW http.ResponseWriter, req *http.Request) {
	res := response{Message: "Hello, world!"}
	bytes, err := json.Marshal(res)
	if err != nil {
		log.Print(err)
	} else {
		resW.Header().Set("Content-Type", "application/json")
		fmt.Fprint(resW, string(bytes))
	}
}

// Http Handler for uploading a file
func uploadFileHandler(resW http.ResponseWriter, req *http.Request) {
	// Create bucket if it does not exist
	if err := cloudstorage.CreateBucket(bucketName); err != nil {
		log.Print(err)
	}
	if err := cloudstorage.UploadFileToGoogle(bucketName, "tmp/test", "HelloWorld"); err != nil {
		log.Print(err)
	}
	fmt.Fprint(resW, "Upload Endpoint")
}
