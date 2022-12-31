package firebase_app_distribution

import (
	bytes2 "bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
)

type release struct {
	Name           string `json:"name"`
	DisplayVersion string `json:"displayVersion"`
	BuildVersion   string `json:"buildVersion"`
	CreatedAt      string `json:"createTime"`
	ReleaseNote    *struct {
		Text string `json:"text"`
	} `json:"releaseNotes,omitempty"`
}

type updateReleaseRequest struct {
	ReleaseName string `json:"name"`
	ReleaseNote struct {
		Text string `json:"text"`
	} `json:"releaseNotes"`
}

type updateReleaseResponse struct {
	release
}

func newUpdateReleaseRequest(release release, releaseNote string) *updateReleaseRequest {
	return &updateReleaseRequest{
		ReleaseName: release.Name,
		ReleaseNote: struct {
			Text string `json:"text"`
		}{
			Text: releaseNote,
		},
	}
}

func (p *Provider) updateReleaseNote(request *updateReleaseRequest) (*updateReleaseResponse, error) {
	path := fmt.Sprintf("/v1/%s", request.ReleaseName)

	client := baseClient.WithHeaders(map[string][]string{
		"Authorization": {fmt.Sprintf("Bearer %s", p.AccessToken)},
	})

	bytes, err := json.Marshal(*request)

	if err != nil {
		return nil, errors.Wrap(err, "cannot marshal the update release request")
	}

	code, bytes, err := client.DoPatch(p.ctx, []string{path}, map[string]string{
		"updateMask": "release_notes.text",
	}, "application/json", bytes2.NewBuffer(bytes))

	if err != nil {
		return nil, errors.Wrap(err, "failed to get a response from operation state api")
	}

	var response updateReleaseResponse

	if 200 <= code && code < 300 {
		if err := json.Unmarshal(bytes, &response); err != nil {
			return nil, errors.Wrap(err, "cannot unmarshal update release response")
		} else {
			return &response, nil
		}
	} else {
		return nil, errors.New(fmt.Sprintf("got %d response: %s", code, string(bytes)))
	}
}
