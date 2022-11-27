package cloudstorage

import (
	"os"
	"testing"
)

func TestGetSongsInBucket(t *testing.T) {
	var bucketName = os.Getenv("GOOGLE_CLOUD_PROJECT") + ".appspot.com"
	songs, err := GetSongsInBucket(bucketName)
	if err != nil {
		panic(err)
	}
	PrettyPrint(songs)
}
