package service

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io"
	"os"
)

const (
	scope = "https://www.googleapis.com/auth/cloud-platform"
)

// FirebaseToken fetches a new token from a credentials file. Currently, only non-interactive way is supported.
func FirebaseToken(ctx context.Context, credentialsPath string) (*oauth2.Token, error) {
	var jsonContent string

	if credentialsPath != "" {
		f, err := os.Open(credentialsPath)

		if err != nil {
			return nil, errors.Wrapf(err, "failed to open %s", credentialsPath)
		}

		defer func() {
			_ = f.Close()
		}()

		bytes, err := io.ReadAll(f)

		if err != nil {
			return nil, errors.Wrapf(err, "failed to read %s", credentialsPath)
		}

		jsonContent = string(bytes)
	}

	if c, err := findGoogleCredentials(ctx, jsonContent); err != nil {
		return nil, errors.Wrap(err, "failed to create credentials")
	} else if t, err := c.TokenSource.Token(); err != nil {
		return nil, errors.Wrap(err, "failed to fetch a token")
	} else {
		return t, nil
	}
}

func findGoogleCredentials(ctx context.Context, jsonContent string) (*google.Credentials, error) {
	params := google.CredentialsParams{
		Scopes: []string{scope},
		State:  "state",
	}

	if jsonContent == "" {
		return google.FindDefaultCredentialsWithParams(ctx, params)
	}

	credentialsType, err := credentialsTypeOf([]byte(jsonContent))

	if err != nil {
		return nil, err
	}

	return google.CredentialsFromJSONWithTypeAndParams(ctx, []byte(jsonContent), credentialsType, params)
}

// credentialsTypeOf reads the type a credentials file declares for itself. splitter accepts
// whichever type a user has configured, so the expected type comes from the file rather than
// being fixed to one of google's constants.
func credentialsTypeOf(jsonContent []byte) (google.CredentialsType, error) {
	var credentials struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal(jsonContent, &credentials); err != nil {
		return "", errors.Wrap(err, "failed to parse the credentials")
	}

	return google.CredentialsType(credentials.Type), nil
}
