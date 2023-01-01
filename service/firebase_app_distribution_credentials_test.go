package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func Test_FirebaseToken(t *testing.T) {
	wd, _ := os.Getwd()
	path := filepath.Join(wd, ".fixtures", "google_credentials.json")

	if _, err := os.Stat(path); err != nil {
		t.Skipf("%s is not found", path)
		return
	}

	if v, err := FirebaseToken(context.TODO(), path); err != nil {
		t.Errorf("failed to get a token: %v", err)
	} else if !v.Valid() {
		t.Error("the token is invalid")
	}
}
