package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"choir-sync/cloudstorage"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

const PROJECTNUMBER string = "845779598570"
const PROJECTNAME string = "choir-sync-go"

// Store our password in memory that we fetch from google secret manager
// storing as global so we can access from our handlers
var standardPassword string
var secretPassword string
var discordEndpoint string
var groupName string

// Our json respone schema
type response struct {
	Message string `json:"message"`
}

type password struct {
	Password string `json:"password"`
}

type upload struct {
	Password   string `json:"password"`
	New_file   string `json:"new_file"`
	Song_name  string `json:"song_name"`
	Track_name string `json:"track_name"`
	Recordable bool   `json:"recordable"`
}

func init() {
	var err error
	standardPassword, err = getSecretPayload("standard-password", "1")
	if err != nil {
		panic(err)
	}
	secretPassword, err = getSecretPayload("secret-password", "1")
	if err != nil {
		panic(err)
	}
	discordEndpoint, err = getSecretPayload("discord-endpoint", "1")
	if err != nil {
		panic(err)
	}
}

func main() {
	// Get the port from the env, or default to port 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s, http://localhost:8080", port)
	}

	// Serve our static files to the root directory
	http.Handle("/", http.FileServer(http.Dir("static")))
	// Our API endpoints
	http.HandleFunc("/api/v1/test", testHandler)
	http.HandleFunc("/api/v1/auth", authHandler)
	http.HandleFunc("/api/v1/uploadfile", uploadFileHandler)
	http.HandleFunc("/api/v1/getsongs", getSongsHandler)

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

// Handler for checking the user password
func authHandler(resW http.ResponseWriter, req *http.Request) {
	var res response

	// Check if password is correct
	err := authenticate(req)
	if err != nil {
		// If the JSON does not meed out schema, return 401
		resW.WriteHeader(401)
		res = response{Message: err.Error()}
	} else {
		// Good password
		res = response{Message: "Successfully authenticated"}
	}

	// Conver our response object to JSON
	bytes, err := json.Marshal(res)
	if err != nil {
		log.Print(err)
		resW.WriteHeader(500)
		fmt.Fprint(resW, "Server error")
	} else {
		resW.Header().Set("Content-Type", "application/json")
		fmt.Fprint(resW, string(bytes))
	}
}

// Http Handler for uploading a file
func uploadFileHandler(resW http.ResponseWriter, req *http.Request) {

	log.Printf("%#v", req.Body)

	decoder := json.NewDecoder(req.Body)
	var upload_info upload
	err := decoder.Decode(&upload_info)
	if err != nil {
		// If the JSON does not meet our schema
		log.Print(err)
		log.Print("upload_error: failed to parse upload request")
	}
	var new_file = upload_info.New_file
	var song_name = upload_info.Song_name
	var track_name = upload_info.Track_name
	var recordable = upload_info.Recordable

	var new_file_name = song_name + "_" + track_name + ".mp3"
	new_file = "testfile.mp3"

	var bucketName = PROJECTNAME + ".appspot.com"
	if err := cloudstorage.UploadFileToGoogle(bucketName, new_file, new_file_name, recordable, false); err != nil {
		log.Print(err)
	}
	res := response{Message: "Upload Endpoint"}
	bytes, err := json.Marshal(res)
	if err != nil {
		log.Print(err)
	} else {
		resW.Header().Set("Content-Type", "application/json")
		fmt.Fprint(resW, string(bytes))
	}
}

// Get files from google storage
func getSongsHandler(resW http.ResponseWriter, req *http.Request) {
	// The bucket we are using for storage for this app
	var bucketName = PROJECTNAME + ".appspot.com"
	var res response

	// Check user is authenticated

	songs, err := cloudstorage.GetSongsInBucket(bucketName, groupName)
	if err != nil {
		log.Print(err)
		res = response{Message: "Failed to get songs"}
		bytes, err := json.Marshal(res)
		if err != nil {
			log.Print(err)
		} else {
			resW.Header().Set("Content-Type", "application/json")
			resW.WriteHeader(404)
			fmt.Fprint(resW, string(bytes))
		}
	} else {
		bytes, err := json.Marshal(songs)
		if err != nil {
			log.Print(err)
		} else {
			resW.Header().Set("Content-Type", "application/json")
			fmt.Fprint(resW, string(bytes))
		}
	}
}

// Check an incoming message has the correct password
func authenticate(req *http.Request) error {

	// Decode the body
	decoder := json.NewDecoder(req.Body)
	var password password
	err := decoder.Decode(&password)
	if err != nil {
		// If the JSON does not meet out schema
		return fmt.Errorf("auth_error: failed to parse authentication request")
	}

	if password.Password == standardPassword { // Good Passwords
		groupName = "SGC"
	} else if password.Password == secretPassword {
		groupName = "Secret"
	} else { // Bad password
		log.Printf("Bad password from: %s", req.RemoteAddr)
		return fmt.Errorf("auth_error: bad password")
	}

	return nil
}

// https://pkg.go.dev/cloud.google.com/go/secretmanager/apiv1/secretmanagerpb#AccessSecretVersionRequest
// Requires the name in
func getSecretPayload(secretName string, secretVerion string) (string, error) {
	ctx := context.Background()
	c, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer c.Close()
	name := fmt.Sprintf(
		"projects/%s/secrets/%s/versions/%s",
		PROJECTNUMBER,
		secretName,
		secretVerion,
	)
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	}
	resp, err := c.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", err
	} else {
		return string(resp.Payload.Data), nil
	}
}
