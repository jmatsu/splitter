package service

import (
	"context"
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

		defer f.Close()

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

	if jsonContent != "" {
		return google.CredentialsFromJSONWithParams(ctx, []byte(jsonContent), params)
	} else {
		return google.FindDefaultCredentialsWithParams(ctx, params)
	}
}
