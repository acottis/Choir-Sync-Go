package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/googleapi"
)

func main() {

	if err := createBucket("mytestbucketforchoirsync"); err != nil {
		log.Print(err)
	}

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

	// Start listening on port specified
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
	log.Printf("Listening on port %s", port)
}

func indexHandler(res http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/api" {
		http.NotFound(res, req)
		return
	}
	fmt.Fprint(res, "Hello, World!")
}

func uploadFileHandler(res http.ResponseWriter, req *http.Request) {
	fmt.Fprint(res, "Upload Endpoint")
}

// Creates a bucket if it does not already exist
func createBucket(bucketName string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()
	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	buckattrs := &storage.BucketAttrs{
		Location: "europe-west2",
	}
	bucket := client.Bucket(bucketName)
	if err := bucket.Create(
		ctx,
		os.Getenv("GOOGLE_CLOUD_PROJECT"),
		buckattrs,
	); err != nil {
		var e *googleapi.Error
		if ok := errors.As(err, &e); ok {
			if e.Code == 409 {
				log.Printf("Bucket: %s already exists, skipping...", bucketName)
				return nil
			}
		}
		return fmt.Errorf("Bucket(%q).Create: %v", bucketName, err)
	}
	log.Printf("Created bucket %s in %s with storage class %s\n",
		bucketName, buckattrs.Location, buckattrs.StorageClass,
	)
	return nil
}

// Gets the list of bucket name in the current project
func listBuckets() []string {
	ctx := context.Background()
	// Create a client to interact with google storage
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	// Dont close the client until the function goes out of scope
	defer client.Close()

	// Set a timeout for the ctx
	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	// Initialise an empty array
	var buckets []string

	// Iterator over gcloud buckets
	it := client.Buckets(ctx, os.Getenv("GOOGLE_CLOUD_PROJECT"))
	for {
		buckattr, err := it.Next()
		if err == nil {
			buckets = append(buckets, buckattr.Name)
		} else {
			break
		}
	}
	return buckets
}
