package service

import (
	"fmt"
	"github.com/jmatsu/splitter/internal/net"
	"github.com/pkg/errors"
	"path/filepath"
)

type firebaseAppDistributionUploadResponse struct {
	OperationName string `json:"name"`

	RawResponse *net.HttpResponse `json:"-"`
}

var _ net.TypedHttpResponse = &firebaseAppDistributionUploadResponse{}

func (r *firebaseAppDistributionUploadResponse) Set(v *net.HttpResponse) {
	r.RawResponse = v
}

type FirebaseAppDistributionUploadAppRequest struct {
	projectNumber string
	appId         string
	filePath      string
}

// https://firebase.google.com/docs/reference/app-distribution/rest/v1/upload.v1.projects.apps.releases/upload
// required: firebaseappdistro.releases.update
func (p *FirebaseAppDistributionProvider) upload(request *FirebaseAppDistributionUploadAppRequest) (*firebaseAppDistributionUploadResponse, error) {
	path := fmt.Sprintf("/upload/v1/projects/%s/apps/%s/releases:upload", request.projectNumber, request.appId)

	client := p.client.WithHeaders(map[string][]string{
		"Authorization":           {fmt.Sprintf("Bearer %s", p.AccessToken)},
		"X-Goog-Upload-File-Name": {filepath.Base(request.filePath)},
		"X-Goog-Upload-Protocol":  {"raw"},
	})

	resp, err := client.DoPostFileBody(p.ctx, []string{path}, nil, request.filePath)

	if err != nil {
		return nil, errors.Wrap(err, "failed to distribute to Firebase App Distribution")
	}

	if resp.Successful() {
		if v, err := resp.ParseJson(&firebaseAppDistributionUploadResponse{}); err != nil {
			return nil, errors.Wrap(err, "succeeded to upload but something went wrong")
		} else {
			return v.(*firebaseAppDistributionUploadResponse), nil
		}
	} else {
		return nil, errors.Wrap(resp.Err(), "failed to upload your app to Firebase App Distribution")
	}
}
