package firebase_app_distribution

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func Test_Token(t *testing.T) {
	wd, _ := os.Getwd()
	path := filepath.Join(wd, ".fixtures", "service_account_key.json")

	if _, err := os.Stat(path); err != nil {
		t.Skipf("%s is not found", path)
		return
	}

	if v, err := Token(context.TODO(), path); err != nil {
		t.Errorf("failed to get a token: %v", err)
	} else if !v.Valid() {
		t.Error("the token is invalid")
	}
}
