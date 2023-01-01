package service

import (
	"context"
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
	FirebaseAppDistributionGetOperationStateResponse

	AabInfo *FirebaseAppDistributionAabInfoResponse
}

func (r *FirebaseAppDistributionDistributionResult) RawJsonResponse() string {
	return r.FirebaseAppDistributionGetOperationStateResponse.RawResponse.RawJson()
}

func (r *FirebaseAppDistributionDistributionResult) ValueResponse() any {
	return *r
}

type FirebaseAppDistributionDistributeRequest struct {
	projectNumber string
	appId         string
	filePath      string
	releaseNote   string
	groupAliases  []string
	testerEmails  []string
}

func (r *FirebaseAppDistributionDistributeRequest) SetReleaseNote(value string) {
	r.releaseNote = value
}

func (r *FirebaseAppDistributionDistributeRequest) SetTesterEmails(value []string) {
	if value != nil && len(value) > 0 {
		r.testerEmails = value
	} else {
		r.testerEmails = []string{}
	}
}

func (r *FirebaseAppDistributionDistributeRequest) OsName() string {
	return strings.SplitN(r.appId, ":", 4)[2]
}

func (r *FirebaseAppDistributionDistributeRequest) fileType() string {
	if s, ext, found := strings.Cut(filepath.Ext(r.filePath), "."); found {
		return strings.ToLower(ext)
	} else {
		return strings.ToLower(s)
	}
}

func (r *FirebaseAppDistributionDistributeRequest) NewUploadRequest() *FirebaseAppDistributionUploadAppRequest {
	return &FirebaseAppDistributionUploadAppRequest{
		projectNumber: r.projectNumber,
		appId:         r.appId,
		filePath:      r.filePath,
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

func (p *FirebaseAppDistributionProvider) Distribute(filePath string, builder func(req *FirebaseAppDistributionDistributeRequest)) (*FirebaseAppDistributionDistributionResult, error) {
	firebaseAppDistributionLogger.Info().Msg("preparing to upload...")

	if err := p.fetchToken(); err != nil {
		return nil, errors.Wrap(err, "a valid token is required to make requests")
	}

	request := &FirebaseAppDistributionDistributeRequest{
		projectNumber: p.ProjectNumber(),
		appId:         p.AppId,
		filePath:      filePath,
	}

	if len(p.FirebaseAppDistributionConfig.GroupAliases) > 0 {
		request.groupAliases = p.FirebaseAppDistributionConfig.GroupAliases
	}

	builder(request)

	firebaseAppDistributionLogger.Debug().Msgf("the request has been built: %v", *request)

	var aabInfo *FirebaseAppDistributionAabInfoResponse

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

	if r, err := p.upload(request.NewUploadRequest()); err != nil {
		return nil, err
	} else {
		operation = r.OperationName
	}

	var response *FirebaseAppDistributionGetOperationStateResponse

	firebaseAppDistributionLogger.Debug().Msgf("start waiting for %s", operation)

	if resp, err := p.waitForOperationDone(&firebaseAppDistributionGetOperationStateRequest{
		operationName: operation,
	}); err != nil {
		return nil, err
	} else {
		response = resp
	}

	if request.releaseNote != "" {
		firebaseAppDistributionLogger.Debug().Msg("start updating the release note")

		req := response.Response.Release.NewUpdateRequest(request.releaseNote)

		if resp, err := p.updateReleaseNote(req); err != nil {
			firebaseAppDistributionLogger.Warn().Err(err).Msg("failed to update the release note")
		} else {
			response.Response.Release = resp.FirebaseAppDistributionReleaseFragment
		}
	}

	if len(request.groupAliases) > 0 || len(request.testerEmails) > 0 {
		firebaseAppDistributionLogger.Debug().Msg("start distribution the release")

		var groupAliases []string
		var testerEmails []string

		if len(request.groupAliases) > 0 {
			groupAliases = request.groupAliases
		}

		if len(request.testerEmails) > 0 {
			testerEmails = request.testerEmails
		}

		req := response.Response.Release.NewDistributeRequest(testerEmails, groupAliases)

		if err := p.distributeRelease(req); err != nil {
			firebaseAppDistributionLogger.Warn().Err(err).Msg("failed to distribute the release")
		}
	}

	return &FirebaseAppDistributionDistributionResult{
		FirebaseAppDistributionGetOperationStateResponse: *response,
		AabInfo: aabInfo,
	}, nil
}
