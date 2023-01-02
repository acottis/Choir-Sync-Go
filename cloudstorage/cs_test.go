package cloudstorage

import (
	"testing"
)

func TestGetSongsInBucket(t *testing.T) {
	var bucketName = "choir-sync-go.appspot.com"
	songs, err := GetSongsInBucket(bucketName, "SGC")
	if err != nil {
		panic(err)
	}
	PrettyPrint(songs)
}
