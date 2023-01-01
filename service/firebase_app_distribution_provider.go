package service

import (
	"context"
	"fmt"
	"github.com/jmatsu/splitter/internal/config"
	logger2 "github.com/jmatsu/splitter/internal/logger"
	"github.com/jmatsu/splitter/internal/net"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"path/filepath"
	"strings"
)

var firebaseAppDistributionLogger zerolog.Logger

func init() {
	firebaseAppDistributionLogger = logger2.Logger.With().Str("service", "firebase app distribution").Logger()
}

func NewFirebaseAppDistributionProvider(ctx context.Context, config *config.FirebaseAppDistributionConfig) *FirebaseAppDistributionProvider {
	return &FirebaseAppDistributionProvider{
		FirebaseAppDistributionConfig: *config,
		ctx:                           ctx,
		client:                        net.NewHttpClient("https://firebaseappdistribution.googleapis.com"),
	}
}

type FirebaseAppDistributionProvider struct {
	config.FirebaseAppDistributionConfig
	ctx    context.Context
	client *net.HttpClient
}

type FirebaseAppDistributionDistributionResult struct {
	firebaseAppDistributionGetOperationStateResponse

	AabInfo *firebaseAppDistributionAabInfoResponse
}

func (r *FirebaseAppDistributionDistributionResult) RawJsonResponse() string {
	return r.firebaseAppDistributionGetOperationStateResponse.RawResponse.RawJson()
}

func (r *FirebaseAppDistributionDistributionResult) ValueResponse() any {
	return *r
}

type firebaseAppDistributionUploadResponse struct {
	OperationName string `json:"name"`

	RawResponse *net.HttpResponse `json:"-"`
}

func (r *firebaseAppDistributionUploadResponse) Set(v *net.HttpResponse) {
	r.RawResponse = v
}

type FirebaseAppDistributionUploadAppRequest struct {
	projectNumber string
	appId         string
	filePath      string
	releaseNote   *string
	groupAliases  *[]string
	testerEmails  *[]string
}

func (r *FirebaseAppDistributionUploadAppRequest) SetReleaseNote(value string) {
	if value != "" {
		r.releaseNote = &value
	} else {
		r.releaseNote = nil
	}
}

func (r *FirebaseAppDistributionUploadAppRequest) SetTesterEmails(value []string) {
	if len(value) > 0 {
		r.testerEmails = &value
	} else {
		r.testerEmails = nil
	}
}

func (r *FirebaseAppDistributionUploadAppRequest) OsName() string {
	return strings.SplitN(r.appId, ":", 4)[2]
}

func (r *FirebaseAppDistributionUploadAppRequest) fileType() string {
	if s, ext, found := strings.Cut(filepath.Ext(r.filePath), "."); found {
		return strings.ToLower(ext)
	} else {
		return strings.ToLower(s)
	}
}

func (p *FirebaseAppDistributionProvider) fetchToken() error {
	if p.AccessToken == "" && p.GoogleCredentialsPath != "" {
		if t, err := FirebaseToken(p.ctx, p.GoogleCredentialsPath); err != nil {
			return errors.Wrap(err, "cannot fetch a token")
		} else {
			p.AccessToken = t.AccessToken
			p.GoogleCredentialsPath = ""
		}
	}

	return nil
}

func (p *FirebaseAppDistributionProvider) Distribute(filePath string, builder func(req *FirebaseAppDistributionUploadAppRequest)) (*FirebaseAppDistributionDistributionResult, error) {
	firebaseAppDistributionLogger.Info().Msg("preparing to upload...")

	if err := p.fetchToken(); err != nil {
		return nil, errors.Wrap(err, "a valid token is required to make requests")
	}

	request := &FirebaseAppDistributionUploadAppRequest{
		projectNumber: p.ProjectNumber(),
		appId:         p.AppId,
		filePath:      filePath,
	}

	if len(p.FirebaseAppDistributionConfig.GroupAliases) > 0 {
		request.groupAliases = &p.FirebaseAppDistributionConfig.GroupAliases
	}

	builder(request)

	firebaseAppDistributionLogger.Debug().Msgf("the request has been built: %v", *request)

	var aabInfo *firebaseAppDistributionAabInfoResponse

	if request.OsName() == "android" {
		aabInfo, _ = p.getAabInfo(&firebaseAppDistributionAabInfoRequest{
			appId:         request.appId,
			projectNumber: request.projectNumber,
		})

		if request.fileType() == "aab" {
			if err := checkAppBundleIntegrationState(aabInfo.IntegrationState); err != nil {
				return nil, err
			}
		}
	}

	var operation string

	if r, err := p.upload(request); err != nil {
		return nil, err
	} else {
		operation = r.OperationName
	}

	var response *firebaseAppDistributionGetOperationStateResponse

	firebaseAppDistributionLogger.Debug().Msgf("start waiting for %s", operation)

	if resp, err := p.waitForOperationDone(&firebaseAppDistributionGetOperationStateRequest{
		operationName: operation,
	}); err != nil {
		return nil, err
	} else {
		response = resp
	}

	if request.releaseNote != nil {
		firebaseAppDistributionLogger.Debug().Msg("start updating the release note")

		req := response.Response.Release.NewUpdateRequest(*request.releaseNote)

		if resp, err := p.updateReleaseNote(req); err != nil {
			firebaseAppDistributionLogger.Warn().Err(err).Msg("failed to update the release note")
		} else {
			response.Response.Release = resp.firebaseAppDistributionRelease
		}
	}

	if request.groupAliases != nil || request.testerEmails != nil {
		firebaseAppDistributionLogger.Debug().Msg("start distribution the release")

		var groupAliases []string
		var testerEmails []string

		if request.groupAliases != nil {
			groupAliases = *request.groupAliases
		}

		if request.testerEmails != nil {
			testerEmails = *request.testerEmails
		}

		req := response.Response.Release.NewDistributeRequest(testerEmails, groupAliases)

		if err := p.distributeRelease(req); err != nil {
			firebaseAppDistributionLogger.Warn().Err(err).Msg("failed to distribute the release")
		}
	}

	return &FirebaseAppDistributionDistributionResult{
		firebaseAppDistributionGetOperationStateResponse: *response,
		AabInfo: aabInfo,
	}, nil
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

	resp, err := client.DoPostFileBody(p.ctx, []string{path}, request.filePath)

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
