package deploygate

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jmatsu/splitter/internal"
	internalHttp "github.com/jmatsu/splitter/internal/http"
	"github.com/rs/zerolog"
)

var logger zerolog.Logger
var baseClient *internal.HttpClient

func init() {
	logger = internal.Logger.With().Str("provider", "deploygate").Logger()
	baseClient = internal.GetHttpClient("https://deploygate.com")
}

type Provider struct {
	internal.DeployGateConfig
	ctx context.Context
}

type UploadRequest struct {
	FilePath            string
	Message             *string
	DistributionOptions *struct {
		Name        string
		AccessKey   string
		ReleaseNote *string
	}
	IOSOptions *struct {
		DisableNotification bool
	}
}

type UploadResponse struct {
	Results struct {
		Name             string `json:"name"`
		PackageName      string `json:"package_name"`
		Revision         uint32 `json:"revision"`
		VersionCode      string `json:"version_code"`
		VersionName      string `json:"version_name"`
		MinSdkVersion    string `json:"sdk_version"`
		TargetSdkVersion uint8  `json:"target_sdk_version"`
		Signature        string `json:"signature"`
		FileChecksum     string `json:"md5"`
		DownloadUrl      string `json:"file"`
		User             struct {
			Name string
		} `json:"user"`
	} `json:"results"`
}

type UploadErrorResponse struct {
	Message string `json:"message"`
}

func (p *Provider) distribute(request *UploadRequest) (UploadResponse, error) {
	client := baseClient.WithHeaders(map[string][]string{
		"Authorization": {fmt.Sprintf("Bearer %s", p.ApiToken)},
	})

	code, bytes, err := client.DoPostMultipartForm(p.ctx, []string{"api", "users", p.AppOwnerName, "apps"}, p.toForm(request))

	var response UploadResponse

	if err != nil {
		return response, fmt.Errorf("failed to upload your app to DeployGate with HttpStatus(%d): %v", code, err)
	}

	if 200 <= code && code < 300 {
		if err := json.Unmarshal(bytes, &response); err != nil {
			return response, fmt.Errorf("failed to parse the response of your app to DeployGate but succeeded to upload: %v", err)
		} else {
			return response, nil
		}
	} else {
		var errorResponse UploadErrorResponse

		if err := json.Unmarshal(bytes, &errorResponse); err != nil {
			return response, fmt.Errorf("failed to upload your app to DeployGate due to: %s, %v", string(bytes), err)
		} else {
			return response, fmt.Errorf("failed to upload your app to DeployGate due to '%s'", errorResponse.Message)
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

	if request.DistributionOptions != nil {
		logger.Debug().Msgf("distribution options were found")

		if request.DistributionOptions.AccessKey != "" && request.DistributionOptions.Name != "" {
			logger.Warn().Msgf("the both of distribution's access key and name are specified so this provider prioritizes access key")
		}

		if request.DistributionOptions.AccessKey != "" {
			form.Set(internalHttp.StringField("distribution_key", request.DistributionOptions.AccessKey))
		} else if request.DistributionOptions.Name != "" {
			form.Set(internalHttp.StringField("distribution_name", request.DistributionOptions.Name))
		} else {
			logger.Warn().Msgf("distribution options were empty")
		}

		if request.DistributionOptions.ReleaseNote != nil {
			form.Set(internalHttp.StringField("release_note", *request.DistributionOptions.ReleaseNote))
		} else if request.Message != nil {
			logger.Debug().Msgf("set message as release note as a fallback")
			form.Set(internalHttp.StringField("release_note", *request.Message))
		}
	}

	if request.IOSOptions != nil {
		logger.Debug().Msgf("ios options were found")

		form.Set(internalHttp.BooleanField("disable_notify", request.IOSOptions.DisableNotification))
	}

	return &form
}
