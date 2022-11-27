package main

import (
	"testing"
)

func TestGetSecretPayload(t *testing.T) {
	getSecretPayload("standard-password", "1")
}
