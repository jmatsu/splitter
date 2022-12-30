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

func NewProvider(ctx context.Context, config config.DeployGateConfig) *Provider {
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
		} else if errorResponse.Message != "" {
			return nil, fmt.Errorf("failed to upload your app to DeployGate due to '%s'", errorResponse.Message)
		} else {
			return nil, fmt.Errorf("failed to upload your app to DeployGate due to '%s'", string(bytes))
		}
	}
}

func (p *Provider) toForm(request *UploadRequest) *net.Form {
	form := net.Form{}

	form.Set(net.FileField("file", request.FilePath))

	if request.Message != nil {
		logger.Debug().Msgf("message option was found")

		form.Set(net.StringField("message", *request.Message))
	}

	var distributionOptionFound = request.DistributionOptions.AccessKey != "" || request.DistributionOptions.Name != ""

	if request.DistributionOptions.AccessKey != "" && request.DistributionOptions.Name != "" {
		logger.Warn().Msgf("the both of distribution's access key and name are specified so this provider prioritizes access key")
	}

	if distributionOptionFound {
		if request.DistributionOptions.AccessKey != "" {
			form.Set(net.StringField("distribution_key", request.DistributionOptions.AccessKey))
		} else if request.DistributionOptions.Name != "" {
			form.Set(net.StringField("distribution_name", request.DistributionOptions.Name))
		}

		if request.DistributionOptions.ReleaseNote != nil {
			form.Set(net.StringField("release_note", *request.DistributionOptions.ReleaseNote))
		} else if request.Message != nil {
			logger.Debug().Msgf("set message as release note as a fallback")
			form.Set(net.StringField("release_note", *request.Message))
		}
	} else {
		logger.Debug().Msgf("distribution options were empty")
	}

	var iosOptionFound = request.IOSOptions.DisableNotification

	if iosOptionFound {
		form.Set(net.BooleanField("disable_notify", request.IOSOptions.DisableNotification))
	}

	return &form
}
