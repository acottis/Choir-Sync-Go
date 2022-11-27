package cloudstorage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/googleapi"
)

type song struct {
	song       string
	part       string
	recordable bool
	secret     bool
}

// Creates a bucket if it does not already exist
func CreateBucket(bucketName string) error {
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
			// Error code 409 is bucket already exists
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

// Parse and return the songs from the bucket
func GetSongsInBucket(bucketName string) ([]song, error) {
	ctx := context.Background()

	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	it := client.Bucket(bucketName).Objects(ctx, nil)

	var songs []song
	for {
		if bucketAttrs, e := it.Next(); e != nil {
			break
		} else {
			tmp := strings.Split(bucketAttrs.Name, "_")
			// If the Name does not include a "_" skip the object
			if len(tmp) == 1 {
				continue
			}
			part := strings.Split(tmp[1], ".")[0]

			var recordable bool
			if bucketAttrs.Metadata["norec"] == "true" {
				recordable = true
			} else {
				recordable = false
			}
			var secret bool
			if bucketAttrs.Metadata["secret"] == "true" {
				secret = true
			} else {
				secret = false
			}
			songs = append(songs, song{
				song:       bucketAttrs.Name,
				part:       part,
				recordable: recordable,
				secret:     secret,
			})
		}
	}

	return songs, nil
}

// Gets the list of bucket name in the current project
func ListBuckets() []string {
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

// Upload a file to a bucket, a same named file uploaded will overwrite an
// exisiting object in the bucket
func UploadFileToGoogle(
	bucketName string,
	srcFileName string,
	destFileName string,
) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	// Open local file.
	f, err := os.Open(srcFileName)
	if err != nil {
		return fmt.Errorf("os.Open: %v", err)
	}
	defer f.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	o := client.Bucket(bucketName).Object(destFileName)

	// Upload an object with storage.Writer.
	writer := o.NewWriter(ctx)
	if _, err = io.Copy(writer, f); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}
	log.Printf("Blob %v uploaded.\n", destFileName)
	return nil
}
