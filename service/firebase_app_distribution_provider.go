package service

import (
	"context"
	"encoding/json"
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

	var response firebaseAppDistributionUploadResponse
	var rawJson string

	if bytes, err := p.distribute(request); err != nil {
		return nil, err
	} else if err := json.Unmarshal(bytes, &response); err != nil {
		return nil, errors.Wrap(err, "failed to parse the response of your app to Firebase App Distribution but succeeded to upload")
	} else {
		rawJson = string(bytes)
	}

	var release firebaseAppDistributionRelease
	var result string

	firebaseAppDistributionLogger.Debug().Msgf("start waiting for %s", response.OperationName)

	if resp, err := p.waitForOperationDone(&firebaseAppDistributionGetOperationStateRequest{
		operationName: response.OperationName,
	}); err != nil {
		return nil, err
	} else {
		release = resp.Response.Release
		result = resp.Response.Result
	}

	if request.releaseNote != nil {
		firebaseAppDistributionLogger.Debug().Msg("start updating the release note")

		req := release.NewUpdateRequest(*request.releaseNote)

		if resp, err := p.updateReleaseNote(req); err != nil {
			firebaseAppDistributionLogger.Warn().Err(err).Msg("failed to update the release note")
		} else {
			release = resp.firebaseAppDistributionRelease
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

		req := release.NewDistributeRequest(testerEmails, groupAliases)

		if err := p.distributeRelease(req); err != nil {
			firebaseAppDistributionLogger.Warn().Err(err).Msg("failed to distribute the release")
		}
	}

	return &FirebaseAppDistributionDistributionResult{
		Release: release,
		Result:  result,
		AabInfo: aabInfo,
		RawJson: rawJson,
	}, nil
}

// https://firebase.google.com/docs/reference/app-distribution/rest/v1/upload.v1.projects.apps.releases/upload
// required: firebaseappdistro.releases.update
func (p *FirebaseAppDistributionProvider) distribute(request *FirebaseAppDistributionUploadAppRequest) ([]byte, error) {
	path := fmt.Sprintf("/upload/v1/projects/%s/apps/%s/releases:upload", request.projectNumber, request.appId)

	client := p.client.WithHeaders(map[string][]string{
		"Authorization":           {fmt.Sprintf("Bearer %s", p.AccessToken)},
		"X-Goog-Upload-File-Name": {filepath.Base(request.filePath)},
		"X-Goog-Upload-Protocol":  {"raw"},
	})

	code, bytes, err := client.DoPostFileBody(p.ctx, []string{path}, request.filePath)

	if err != nil {
		return nil, errors.Wrap(err, "failed to distribute to Firebase App Distribution")
	}

	if code <= 200 && code < 300 {
		return bytes, nil
	} else {
		return nil, errors.New(fmt.Sprintf("failed to upload your app to Firebase App Distribution due to '%s'", string(bytes)))
	}
}
