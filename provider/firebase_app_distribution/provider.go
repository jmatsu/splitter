package firebase_app_distribution

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jmatsu/splitter/internal/config"
	logger2 "github.com/jmatsu/splitter/internal/logger"
	"github.com/jmatsu/splitter/internal/net"
	"github.com/rs/zerolog"
	"path/filepath"
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
	osName      string
	packageName string
	filePath    string
	releaseNote *string
}

func (r *UploadRequest) SetReleaseNote(value string) {
	if value != "" {
		r.releaseNote = &value
	} else {
		r.releaseNote = nil
	}
}

func (p *Provider) Distribute(filePath string, builder func(req *UploadRequest)) (*DistributionResult, error) {
	request := &UploadRequest{
		osName:      p.OsName,
		packageName: p.PackageName,
		filePath:    filePath,
	}

	builder(request)

	logger.Debug().Msgf("the request has been built: %v", *request)

	var response uploadResponse

	if bytes, err := p.distribute(request); err != nil {
		return nil, err
	} else if err := json.Unmarshal(bytes, &response); err != nil {
		return nil, fmt.Errorf("failed to parse the response of your app to Firebase App Distribution but succeeded to upload: %v", err)
	} else {
		return &DistributionResult{
			uploadResponse: response,
			RawJson:        string(bytes),
		}, nil
	}
}

// https://firebase.google.com/docs/reference/app-distribution/rest/v1/upload.v1.projects.apps.releases/upload
// required: firebaseappdistro.releases.update
func (p *Provider) distribute(request *UploadRequest) ([]byte, error) {
	path := fmt.Sprintf("/upload/v1/projects/%s/apps/%s:%s/releases:upload", p.ProjectNumber, request.osName, request.packageName)

	client := baseClient.WithHeaders(map[string][]string{
		"Authorization":           {fmt.Sprintf("Bearer %s", p.AccessToken)},
		"X-Goog-Upload-File-Name": {filepath.Base(request.filePath)},
		"X-Goog-Upload-Protocol":  {"raw"},
	})

	code, bytes, err := client.DoPostFileBody(p.ctx, []string{path}, request.filePath)

	if err != nil {
		return nil, fmt.Errorf("failed to distribute to Firebase App Distribution: %v", err)
	}

	if code <= 200 && code < 300 {
		return bytes, nil
	} else {
		return nil, fmt.Errorf("failed to upload your app to Firebase App Distribution due to '%s'", string(bytes))
	}
}
