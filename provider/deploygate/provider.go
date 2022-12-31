package deploygate

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jmatsu/splitter/internal/config"
	logger2 "github.com/jmatsu/splitter/internal/logger"
	"github.com/jmatsu/splitter/internal/net"
	"github.com/rs/zerolog"
)

var logger zerolog.Logger
var baseClient *net.HttpClient

func init() {
	logger = logger2.Logger.With().Str("provider", "deploygate").Logger()
	baseClient = net.GetHttpClient("https://deploygate.com")
}

type Provider struct {
	config.DeployGateConfig
	ctx context.Context
}

func NewProvider(ctx context.Context, config *config.DeployGateConfig) *Provider {
	return &Provider{
		DeployGateConfig: *config,
		ctx:              ctx,
	}
}

type UploadRequest struct {
	filePath            string
	message             *string
	distributionOptions *distributionOptions
	iOSOptions          iOSOptions
}

type distributionOptions struct {
	Name        string
	AccessKey   string
	ReleaseNote *string
}

type iOSOptions struct {
	DisableNotification bool
}

func NewUploadRequest(filePath string) *UploadRequest {
	return &UploadRequest{
		filePath:            filePath,
		distributionOptions: &distributionOptions{},
	}
}

func (r *UploadRequest) SetMessage(value string) {
	if value != "" {
		r.message = &value
	} else {
		r.message = nil
	}
}

func (r *UploadRequest) SetDistributionAccessKey(value string) {
	if value != "" {
		r.getDistributionOptions().AccessKey = value
	} else {
		r.getDistributionOptions().AccessKey = ""
	}
}

func (r *UploadRequest) SetDistributionName(value string) {
	if value != "" {
		r.getDistributionOptions().Name = value
	} else {
		r.getDistributionOptions().Name = ""
	}
}

func (r *UploadRequest) SetDistributionReleaseNote(value string) {
	if value != "" {
		r.getDistributionOptions().ReleaseNote = &value
	} else {
		r.getDistributionOptions().ReleaseNote = nil
	}
}

func (r *UploadRequest) SetIOSDisableNotification(value bool) {
	r.iOSOptions.DisableNotification = value
}

func (r *UploadRequest) getDistributionOptions() *distributionOptions {
	if r.distributionOptions == nil {
		r.distributionOptions = &distributionOptions{}
	}

	return r.distributionOptions
}

type errorResponse struct {
	Message string `json:"message"`
}

func (p *Provider) Distribute(filePath string, builder func(req *UploadRequest)) (*DistributionResult, error) {
	request := NewUploadRequest(filePath)

	builder(request)

	logger.Debug().Msgf("the request has been built: %v", *request)

	var response uploadResponse

	if bytes, err := p.distribute(request); err != nil {
		return nil, err
	} else if err := json.Unmarshal(bytes, &response); err != nil {
		return nil, fmt.Errorf("failed to parse the response of your app to DeployGate but succeeded to upload: %v", err)
	} else {
		return &DistributionResult{
			uploadResponse: response,
			RawJson:        string(bytes),
		}, nil
	}
}

func (p *Provider) distribute(request *UploadRequest) ([]byte, error) {
	client := baseClient.WithHeaders(map[string][]string{
		"Authorization": {fmt.Sprintf("Bearer %s", p.ApiToken)},
	})

	code, bytes, err := client.DoPostMultipartForm(p.ctx, []string{"api", "users", p.AppOwnerName, "apps"}, p.toForm(request))

	if err != nil {
		return nil, fmt.Errorf("failed to upload your app to DeployGate: %v", err)
	}

	if 200 <= code && code < 300 {
		return bytes, nil
	} else {
		var errorResponse errorResponse

		if err := json.Unmarshal(bytes, &errorResponse); err != nil {
			return nil, fmt.Errorf("failed to upload your app to DeployGate due to: %s, %v", string(bytes), err)
		} else if errorResponse.Message != "" {
			return nil, fmt.Errorf("failed to upload your app to DeployGate due to '%s'", errorResponse.Message)
		} else {
			return nil, fmt.Errorf("failed to upload your app to DeployGate due to '%s'", string(bytes))
		}
	}
}

func (p *Provider) toForm(request *UploadRequest) *net.Form {
	form := net.Form{}

	form.Set(net.FileField("file", request.filePath))

	if request.message != nil {
		logger.Debug().Msgf("message option was found")

		form.Set(net.StringField("message", *request.message))
	}

	if request.distributionOptions != nil {
		if request.distributionOptions.AccessKey != "" && request.distributionOptions.Name != "" {
			logger.Warn().Msgf("the both of distribution's access key and name are specified so this provider prioritizes access key")
		}

		if request.distributionOptions.AccessKey != "" {
			form.Set(net.StringField("distribution_key", request.distributionOptions.AccessKey))
		} else if request.distributionOptions.Name != "" {
			form.Set(net.StringField("distribution_name", request.distributionOptions.Name))
		}

		if request.distributionOptions.ReleaseNote != nil {
			form.Set(net.StringField("release_note", *request.distributionOptions.ReleaseNote))
		} else if request.message != nil {
			logger.Debug().Msgf("set message as release note as a fallback")
			form.Set(net.StringField("release_note", *request.message))
		}
	} else {
		logger.Debug().Msgf("distribution options were empty")
	}

	var iosOptionFound = request.iOSOptions.DisableNotification

	if iosOptionFound {
		form.Set(net.BooleanField("disable_notify", request.iOSOptions.DisableNotification))
	}

	return &form
}
