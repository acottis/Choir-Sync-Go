package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"choir-sync/cloudstorage"
	"choir-sync/discord"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

const PROJECTNUMBER string = "845779598570"
const PROJECTNAME string = "choir-sync-go"

// Store our password in memory that we fetch from google secret manager
// Storing as global so we can access from our handlers
var sgcPassword string
var secretPassword string
var sgcPasswordAdmin string
var secretPasswordAdmin string
var discordEndpoint string
var groupName string

// Our json response schema
type response struct {
	Message string `json:"message"`
}

type password struct {
	Password string `json:"password"`
	Admin    bool   `json:"admin"`
}

func init() {
	var err error
	sgcPassword, err = getSecretPayload("sgc-password", "1")
	if err != nil {
		panic(err)
	}
	secretPassword, err = getSecretPayload("secret-password", "1")
	if err != nil {
		panic(err)
	}
	sgcPasswordAdmin, err = getSecretPayload("sgc-password-admin", "1")
	if err != nil {
		panic(err)
	}
	secretPasswordAdmin, err = getSecretPayload("secret-password-admin", "1")
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
	http.HandleFunc("/api/v1/getsongs", getSongsHandler)
	// using a wrapper function for my wrapper function so I can pass an extra parameter into my wrapper function
	http.HandleFunc("/api/v1/uploadfile", func(w http.ResponseWriter, r *http.Request) { fileHandler(w, r, "upload") })
	http.HandleFunc("/api/v1/deletefile", func(w http.ResponseWriter, r *http.Request) { fileHandler(w, r, "delete") })
	http.HandleFunc("/api/v1/renamefile", func(w http.ResponseWriter, r *http.Request) { fileHandler(w, r, "rename") })
	http.HandleFunc("/api/v1/uploadrecording", func(w http.ResponseWriter, r *http.Request) { fileHandler(w, r, "sendrec") })

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

	// Decode the body
	decoder := json.NewDecoder(req.Body)
	var password password
	err := decoder.Decode(&password)
	if err != nil {
		// JSON does not meet our schema
		log.Print(err)
		resW.WriteHeader(401)
		res = response{Message: "auth_error: failed to parse authentication request"}
	} else {
		// Check if password is correct
		if password.Admin {
			err = authenticateAdmin(password.Password)
		} else {
			err = authenticateUser(password.Password)
		}
		if err != nil {
			// Password not correct
			resW.WriteHeader(401)
			res = response{Message: err.Error()}
		} else {
			// Good password
			res = response{Message: "Successfully authenticated"}
		}
	}

	// Convert our response object to JSON
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

// Wrapper for handlers
func fileHandler(resW http.ResponseWriter, req *http.Request, requesttype string) {
	var res response

	switch requesttype {
	case "upload":
		res = uploadFileHandler(resW, req)
	case "delete":
		res = deleteFileHandler(resW, req)
	case "rename":
		res = renameFileHandler(resW, req)
	case "sendrec":
		res = sendRecordingHandler(resW, req)
	}

	bytes, err := json.Marshal(res)
	if err != nil {
		log.Print(err)
	} else {
		resW.Header().Set("Content-Type", "application/json")
		fmt.Fprint(resW, string(bytes))
	}
}

// Http Handler for uploading a file
func uploadFileHandler(resW http.ResponseWriter, req *http.Request) response {
	bucketName := PROJECTNAME + ".appspot.com"

	temp_file_name := "tmp/tempfile.mp3"

	err := req.ParseMultipartForm(32 << 20)
	if err != nil {
		log.Print(err)
		return response{Message: "upload_error: failed to parse upload request"}
	}
	password := req.PostFormValue("password")

	// Check if password is correct
	err = authenticateAdmin(password)
	if err != nil {
		// Password not correct
		resW.WriteHeader(401)
		return response{Message: err.Error()}
	}

	song_name := req.PostFormValue("song_name")
	track_name := req.PostFormValue("track_name")
	recordable := (req.PostFormValue("recordable") == "true")

	new_file_name := song_name + "_" + track_name + ".mp3"

	new_file, _, err := req.FormFile("new_file")
	if err != nil {
		log.Print(err)
		return response{Message: "upload_error: failed to parse uploaded file"}
	}

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, new_file); err != nil {
		log.Print(err)
		return response{"upload_error: failed to write file to bytes"}
	}
	err = os.WriteFile(temp_file_name, buf.Bytes(), 0644)
	if err != nil {
		log.Print(err)
		return response{"upload_error: failed to write temporary file"}
	}

	if err := cloudstorage.UploadFileToGoogle(bucketName, temp_file_name, new_file_name, recordable, false); err != nil {
		log.Print(err)
		return response{"upload_error: failed to upload file"}
	}

	err = os.Remove(temp_file_name)
	if err != nil {
		log.Print(err)
		return response{"upload_error: failed to delete temporary file"}
	}

	return response{Message: "Upload Endpoint"}
}

// Http Handler for deleting a file
func deleteFileHandler(resW http.ResponseWriter, req *http.Request) response {
	bucketName := PROJECTNAME + ".appspot.com"

	err := req.ParseMultipartForm(32 << 20)
	if err != nil {
		log.Print(err)
		return response{"delete_error: failed to parse delete request"}
	}
	password := req.PostFormValue("password")

	// Check if password is correct
	err = authenticateAdmin(password)
	if err != nil {
		// Password not correct
		resW.WriteHeader(401)
		return response{Message: err.Error()}
	}

	song_name := req.PostFormValue("song_name")
	track_name := req.PostFormValue("track_name")

	del_file_name := song_name + "_" + track_name + ".mp3"

	if err := cloudstorage.DeleteFileFromGoogle(bucketName, del_file_name); err != nil {
		log.Print(err)
		return response{Message: "delete_error: failed to delete file"}
	}

	return response{Message: "Delete Endpoint"}
}

// Http Handler for renaming a file
func renameFileHandler(resW http.ResponseWriter, req *http.Request) response {
	bucketName := PROJECTNAME + ".appspot.com"

	err := req.ParseMultipartForm(32 << 20)
	if err != nil {
		log.Print(err)
		return response{"delete_error: failed to parse rename request"}
	}
	password := req.PostFormValue("password")

	// Check if password is correct
	err = authenticateAdmin(password)
	if err != nil {
		// Password not correct
		resW.WriteHeader(401)
		return response{Message: err.Error()}
	}

	orig_song_name := req.PostFormValue("orig_song_name")
	orig_track_name := req.PostFormValue("orig_track_name")
	new_song_name := req.PostFormValue("new_song_name")
	new_track_name := req.PostFormValue("new_track_name")

	orig_file_name := orig_song_name + "_" + orig_track_name + ".mp3"
	new_file_name := new_song_name + "_" + new_track_name + ".mp3"

	if err := cloudstorage.RenameFileInGoogle(bucketName, orig_file_name, new_file_name); err != nil {
		log.Print(err)
		return response{Message: "rename_error: failed to rename file"}
	}

	return response{Message: "Rename Endpoint"}
}

// Http Handler for sending a recording to discord
func sendRecordingHandler(resW http.ResponseWriter, req *http.Request) response {

	err := req.ParseMultipartForm(32 << 20)
	if err != nil {
		log.Print(err)
		return response{"sendrec_error: failed to parse upload request"}
	}

	password := req.PostFormValue("password")
	// Check if password is correct
	err = authenticateUser(password)
	if err != nil {
		// Password not correct
		resW.WriteHeader(401)
		return response{Message: err.Error()}
	}

	message := req.PostFormValue("message")
	singer := req.PostFormValue("singer_name")
	file_name := req.PostFormValue("file_name")
	temp_file_name := "tmp/" + file_name
	log.Print(temp_file_name)

	recording, _, err := req.FormFile("recording")
	if err != nil {
		log.Print(err)
		return response{Message: "upload_error: failed to parse uploaded file"}
	}
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, recording); err != nil {
		log.Print(err)
		return response{"upload_error: failed to write file to bytes"}
	}
	//err = os.WriteFile(temp_file_name, buf.Bytes(), 0644)
	err = os.WriteFile("tmp/tempfile.mp3", buf.Bytes(), 0644)
	if err != nil {
		log.Print(err)
		return response{"upload_error: failed to write temporary file"}
	}

	if err := discord.UploadFile(discordEndpoint, temp_file_name, singer, message); err != nil {
		log.Print(err)
		return response{Message: err.Error()}
	}

	return response{Message: "Send Recording Endpoint"}
}

// Get files from google storage
func getSongsHandler(resW http.ResponseWriter, req *http.Request) {
	// The bucket we are using for storage for this app
	var bucketName = PROJECTNAME + ".appspot.com"
	var res response

	// Decode the body
	decoder := json.NewDecoder(req.Body)
	var password password
	err := decoder.Decode(&password)
	if err != nil {
		// If the JSON does not meet our schema
		log.Print(err)
		resW.WriteHeader(401)
		res = response{Message: "auth_error: failed to parse authentication request"}
	}
	// Check if password is correct
	if password.Admin {
		err = authenticateAdmin(password.Password)
	} else {
		err = authenticateUser(password.Password)
	}
	if err != nil {
		resW.WriteHeader(401)
		res = response{Message: err.Error()}
	} else {
		// Good password
		res = response{Message: "Successfully authenticated"}
	}

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
func authenticateUser(password string) error {

	if password == sgcPassword { // Good Passwords
		groupName = "SGC"
	} else if password == secretPassword {
		groupName = "Secret"
	} else { // Bad password
		log.Printf("Bad password")
		return fmt.Errorf("auth_error: bad password")
	}

	return nil
}

// Check an incoming message has the correct password
func authenticateAdmin(password string) error {

	if password == sgcPasswordAdmin { // Good Passwords
		groupName = "SGC"
	} else if password == secretPasswordAdmin {
		groupName = "Secret"
	} else { // Bad password
		log.Printf("Bad password")
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
