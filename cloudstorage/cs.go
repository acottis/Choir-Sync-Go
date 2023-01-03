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
	Song       string `json:"song"`
	Part       string `json:"part"`
	Recordable bool   `json:"recordable"`
	Secret     bool   `json:"secret"`
	Url        string `json:"url"`
}

// Creates a bucket if it does not already exist
func CreateBucket(bucketName string, projectName string) error {
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
		projectName,
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
func GetSongsInBucket(bucketName string, groupName string) ([]song, error) {
	ctx := context.Background()

	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	bucket := client.Bucket(bucketName)

	// Check if bucket exists
	_, err = bucket.Attrs(ctx)
	if err != nil {
		return nil, err
	}

	var songs []song
	it := bucket.Objects(ctx, nil)
	for {
		if objectAttrs, e := it.Next(); e != nil {
			break
		} else {
			// Skip items that are not mpeg's
			if objectAttrs.ContentType != "audio/mpeg" {
				continue
			}
			var tmp []string
			// Parse out just the last part in the full file path
			tmp = strings.Split(objectAttrs.Name, "/")
			rawFileName := tmp[len(tmp)-1]

			// Split the song from the Parts
			tmp = strings.Split(rawFileName, "_")

			// If the Name does not include a "_" skip the object
			if len(tmp) == 1 {
				continue
			}
			songName := tmp[0]
			part := strings.Split(tmp[1], ".")[0]
			recordable := (objectAttrs.Metadata["recordable"] == "true")
			//to implement: groups rather than boolean secret
			//group := (objectAttrs.Metadata["group"])
			secret := (objectAttrs.Metadata["secret"] == "true")
			url := fmt.Sprintf(
				"https://storage.googleapis.com/%s/%s",
				bucketName,
				objectAttrs.Name,
			)
			//to implement: groups rather than boolean secret
			//if group == groupName {
			if (groupName == "Secret" && secret) || (groupName == "SGC" && !secret) {
				songs = append(songs, song{
					Song:       songName,
					Part:       part,
					Recordable: recordable,
					Url:        url,
				})
			}
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
	recordable bool,
	secret bool,
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
	writer.Metadata = map[string]string{
		"recordable": fmt.Sprintf("%v", recordable),
		"secret":     fmt.Sprintf("%v", secret),
	}
	if _, err = io.Copy(writer, f); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}
	log.Printf("Blob %v uploaded.\n", destFileName)
	return nil
}
