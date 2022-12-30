package deploygate

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jmatsu/splitter/internal"
	internalHttp "github.com/jmatsu/splitter/internal/http"
	logger2 "github.com/jmatsu/splitter/internal/logger"
	"github.com/rs/zerolog"
)

var logger zerolog.Logger
var baseClient *internal.HttpClient

func init() {
	logger = logger2.Logger.With().Str("provider", "deploygate").Logger()
	baseClient = internal.GetHttpClient("https://deploygate.com")
}

type Provider struct {
	internal.DeployGateConfig
	ctx context.Context
}

func NewProvider(ctx context.Context, config internal.DeployGateConfig) *Provider {
	return &Provider{
		DeployGateConfig: config,
		ctx:              ctx,
	}
}

type UploadRequest struct {
	FilePath            string
	Message             *string
	DistributionOptions struct {
		Name        string
		AccessKey   string
		ReleaseNote *string
	}
	IOSOptions struct {
		DisableNotification bool
	}
}

type errorResponse struct {
	Message string `json:"message"`
}

func (p *Provider) Distribute(request *UploadRequest) ([]byte, error) {
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
		} else {
			return nil, fmt.Errorf("failed to upload your app to DeployGate due to '%s'", errorResponse.Message)
		}
	}
}

func (p *Provider) toForm(request *UploadRequest) *internalHttp.Form {
	form := internalHttp.Form{}

	form.Set(internalHttp.FileField("file", request.FilePath))

	if request.Message != nil {
		logger.Debug().Msgf("message option was found")

		form.Set(internalHttp.StringField("message", *request.Message))
	}

	var distributionOptionFound = request.DistributionOptions.AccessKey != "" || request.DistributionOptions.Name != ""

	if request.DistributionOptions.AccessKey != "" && request.DistributionOptions.Name != "" {
		logger.Warn().Msgf("the both of distribution's access key and name are specified so this provider prioritizes access key")
	}

	if distributionOptionFound {
		if request.DistributionOptions.AccessKey != "" {
			form.Set(internalHttp.StringField("distribution_key", request.DistributionOptions.AccessKey))
		} else if request.DistributionOptions.Name != "" {
			form.Set(internalHttp.StringField("distribution_name", request.DistributionOptions.Name))
		}

		if request.DistributionOptions.ReleaseNote != nil {
			form.Set(internalHttp.StringField("release_note", *request.DistributionOptions.ReleaseNote))
		} else if request.Message != nil {
			logger.Debug().Msgf("set message as release note as a fallback")
			form.Set(internalHttp.StringField("release_note", *request.Message))
		}
	} else {
		logger.Debug().Msgf("distribution options were empty")
	}

	var iosOptionFound = request.IOSOptions.DisableNotification

	if iosOptionFound {
		form.Set(internalHttp.BooleanField("disable_notify", request.IOSOptions.DisableNotification))
	}

	return &form
}
