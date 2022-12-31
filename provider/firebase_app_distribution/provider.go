package firebase_app_distribution

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

const (
	endpoint = "https://firebaseappdistribution.googleapis.com"
)

var baseClient *net.HttpClient
var logger zerolog.Logger

func init() {
	logger = logger2.Logger.With().Str("provider", "firebase app distribution").Logger()
	baseClient = net.GetHttpClient(endpoint)
}

type Provider struct {
	config.FirebaseAppDistributionConfig
	ctx context.Context
}

func NewProvider(ctx context.Context, config *config.FirebaseAppDistributionConfig) *Provider {
	return &Provider{
		FirebaseAppDistributionConfig: *config,
		ctx:                           ctx,
	}
}

type UploadRequest struct {
	projectNumber string
	appId         string
	filePath      string
	releaseNote   *string
}

func (r *UploadRequest) SetReleaseNote(value string) {
	if value != "" {
		r.releaseNote = &value
	} else {
		r.releaseNote = nil
	}
}

func (r *UploadRequest) OsName() string {
	return strings.SplitN(r.appId, ":", 4)[2]
}

func (r *UploadRequest) fileType() string {
	if s, ext, found := strings.Cut(filepath.Ext(r.filePath), "."); found {
		return strings.ToLower(ext)
	} else {
		return strings.ToLower(s)
	}
}

func (p *Provider) Distribute(filePath string, builder func(req *UploadRequest)) (*DistributionResult, error) {
	logger.Info().Msg("preparing to upload...")

	request := &UploadRequest{
		projectNumber: p.ProjectNumber(),
		appId:         p.AppId,
		filePath:      filePath,
	}

	builder(request)

	logger.Debug().Msgf("the request has been built: %v", *request)

	var aabInfo *aabInfoResponse

	if request.OsName() == "android" {
		aabInfo, _ = p.getAabInfo(&aabInfoRequest{
			appId:         request.appId,
			projectNumber: request.projectNumber,
		})

		if request.fileType() == "aab" {
			if err := checkIntegrationState(aabInfo.IntegrationState); err != nil {
				return nil, err
			}
		}
	}

	var response uploadResponse

	if bytes, err := p.distribute(request); err != nil {
		return nil, err
	} else if err := json.Unmarshal(bytes, &response); err != nil {
		return nil, errors.Wrap(err, "failed to parse the response of your app to Firebase App Distribution but succeeded to upload")
	} else if config.GetGlobalConfig().Async {
		return &DistributionResult{
			aabInfo: aabInfo,
			RawJson: string(bytes),
		}, nil
	} else {
		if doneResp, err := p.waitForOperationDone(&getOperationStateRequest{
			operationName: response.OperationName,
		}); err != nil {
			return nil, err
		} else {
			return &DistributionResult{
				v1UploadReleaseResponse: doneResp.Response,
				aabInfo:                 aabInfo,
				RawJson:                 string(bytes),
			}, nil
		}
	}
}

// https://firebase.google.com/docs/reference/app-distribution/rest/v1/upload.v1.projects.apps.releases/upload
// required: firebaseappdistro.releases.update
func (p *Provider) distribute(request *UploadRequest) ([]byte, error) {
	path := fmt.Sprintf("/upload/v1/projects/%s/apps/%s/releases:upload", request.projectNumber, request.appId)

	client := baseClient.WithHeaders(map[string][]string{
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
