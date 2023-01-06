package discord

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

// Upload a file to discord, takes a `fileName` which is the path to the file on
// the disk and `message` which is sent along side the file to discord
func UploadFile(discordEndpoint string, fileName string, singer string, message string) error {

	// Discord Base URL Constant
	const DISCORD_BASE_URL = "https://discord.com/api/webhooks/"

	// Our webhook endpoint https://discord.com/developers/docs/reference#uploading-files
	// discordEndpoint := os.Getenv("DISCORD_ENDPOINT")
	// if discordEndpoint == "" {
	// 	return errors.New("env: Missing env variable DISCORD_ENDPOINT")
	// }

	var discord_message string
	if message == "" {
		discord_message = singer + "didn't add a comment"
	} else {
		discord_message = singer + "said: " + message
	}

	// Create an empty buffer for body
	body := &bytes.Buffer{}

	// Create a writer to target out body buffer
	multiWriter := multipart.NewWriter(body)

	// Add our message to the formdata
	multiWriter.WriteField("content", discord_message)

	//Add our file to the multipart
	partWriter, err := multiWriter.CreateFormFile("file", fileName)
	if err != nil {
		return err
	}
	// Open our file
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	// Copy the contents of the file to the file section of the body buffer
	_, err = io.Copy(partWriter, file)
	if err != nil {
		return err
	}

	// Close our file and writer to our multipart
	file.Close()
	multiWriter.Close()

	// Create new http client
	client := &http.Client{}
	// Create our request
	req, err := http.NewRequest(
		http.MethodPost,
		DISCORD_BASE_URL+discordEndpoint,
		body,
	)
	// Append the required header for formdata
	req.Header.Add("Content-Type", multiWriter.FormDataContentType())
	if err != nil {
		return err
	}

	// Send the request to the server
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	// Check we succeeded
	if res.StatusCode != 200 {
		return errors.New("http: Failed to send, received non 200 status code")
	}

	return nil
}
