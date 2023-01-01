package service

import (
	bytes2 "bytes"
	"encoding/json"
	"fmt"
	"github.com/jmatsu/splitter/internal/net"
	"github.com/pkg/errors"
)

type FirebaseAppDistributionReleaseFragment struct {
	Name           string                                      `json:"name"`
	DisplayVersion string                                      `json:"displayVersion"`
	BuildVersion   string                                      `json:"buildVersion"`
	CreatedAt      string                                      `json:"createTime"`
	ReleaseNote    *FirebaseAppDistributionReleaseNoteFragment `json:"releaseNotes"`
}

type FirebaseAppDistributionReleaseNoteFragment struct {
	Text string `json:"text"`
}

type firebaseAppDistributionUpdateReleaseRequest struct {
	ReleaseName string `json:"name"`
	ReleaseNote struct {
		Text string `json:"text"`
	} `json:"releaseNotes"`
}

type firebaseAppDistributionUpdateReleaseResponse struct {
	FirebaseAppDistributionReleaseFragment

	RawResponse *net.HttpResponse `json:"-"`
}

func (r *firebaseAppDistributionUpdateReleaseResponse) Set(v *net.HttpResponse) {
	r.RawResponse = v
}

type firebaseAppDistributionDistributeReleaseRequest struct {
	ReleaseName  string   `json:"-"`
	TesterEmails []string `json:"testerEmails"`
	GroupAliases []string `json:"groupAliases"`
}

func (r FirebaseAppDistributionReleaseFragment) NewUpdateRequest(releaseNote string) *firebaseAppDistributionUpdateReleaseRequest {
	return &firebaseAppDistributionUpdateReleaseRequest{
		ReleaseName: r.Name,
		ReleaseNote: struct {
			Text string `json:"text"`
		}{
			Text: releaseNote,
		},
	}
}

func (r FirebaseAppDistributionReleaseFragment) NewDistributeRequest(testerEmails []string, groupAliases []string) *firebaseAppDistributionDistributeReleaseRequest {
	return &firebaseAppDistributionDistributeReleaseRequest{
		ReleaseName:  r.Name,
		TesterEmails: testerEmails,
		GroupAliases: groupAliases,
	}
}

func (p *FirebaseAppDistributionProvider) updateReleaseNote(request *firebaseAppDistributionUpdateReleaseRequest) (*firebaseAppDistributionUpdateReleaseResponse, error) {
	path := fmt.Sprintf("/v1/%s", request.ReleaseName)

	client := p.client.WithHeaders(map[string][]string{
		"Authorization": {fmt.Sprintf("Bearer %s", p.AccessToken)},
	})

	bytes, err := json.Marshal(*request)

	if err != nil {
		return nil, errors.Wrap(err, "cannot marshal the update release request")
	}

	resp, err := client.DoPatch(p.ctx, []string{path}, map[string]string{
		"updateMask": "release_notes.text",
	}, "application/json", bytes2.NewBuffer(bytes))

	if err != nil {
		return nil, errors.Wrap(err, "failed to get a response from update release api")
	}

	if resp.Successful() {
		if v, err := resp.ParseJson(&firebaseAppDistributionUpdateReleaseResponse{}); err != nil {
			return nil, errors.Wrap(err, "succeeded to upload the release but something went wrong")
		} else {
			return v.(*firebaseAppDistributionUpdateReleaseResponse), nil
		}
	} else {
		return nil, errors.Wrap(resp.Err(), "failed to upload the release")
	}
}

func (p *FirebaseAppDistributionProvider) distributeRelease(request *firebaseAppDistributionDistributeReleaseRequest) error {
	path := fmt.Sprintf("/v1/%s:distribute", request.ReleaseName)

	client := p.client.WithHeaders(map[string][]string{
		"Authorization": {fmt.Sprintf("Bearer %s", p.AccessToken)},
	})

	bytes, err := json.Marshal(*request)

	if err != nil {
		return errors.Wrap(err, "cannot marshal the distribute release request")
	}

	resp, err := client.DoPost(p.ctx, []string{path}, nil, "application/json", bytes2.NewBuffer(bytes))

	if err != nil {
		return errors.Wrap(err, "failed to get a response from distribute api")
	}

	if resp.Successful() {
		return nil
	} else {
		return errors.Wrap(resp.Err(), "failed to distribute the release to testers")
	}
}
