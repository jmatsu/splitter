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
)

var deployGateLogger zerolog.Logger
var deployGateBaseClient *net.HttpClient

func init() {
	deployGateLogger = logger2.Logger.With().Str("service", "deploygate").Logger()
	deployGateBaseClient = net.NewHttpClient("https://deploygate.com")
}

type DeployGateProvider struct {
	config.DeployGateConfig
	ctx context.Context
}

func NewDeployGateProvider(ctx context.Context, config *config.DeployGateConfig) *DeployGateProvider {
	return &DeployGateProvider{
		DeployGateConfig: *config,
		ctx:              ctx,
	}
}

type DeployGateUploadAppRequest struct {
	filePath            string
	message             *string
	distributionOptions *deployGateDistributionOptions
	iOSOptions          deployGateIOSOptions
}

type deployGateDistributionOptions struct {
	Name        string
	AccessKey   string
	ReleaseNote *string
}

type deployGateIOSOptions struct {
	DisableNotification bool
}

func NewDeployGateUploadAppRequest(filePath string) *DeployGateUploadAppRequest {
	return &DeployGateUploadAppRequest{
		filePath:            filePath,
		distributionOptions: &deployGateDistributionOptions{},
	}
}

func (r *DeployGateUploadAppRequest) SetMessage(value string) {
	if value != "" {
		r.message = &value
	} else {
		r.message = nil
	}
}

func (r *DeployGateUploadAppRequest) SetDistributionAccessKey(value string) {
	if value != "" {
		r.getDistributionOptions().AccessKey = value
	} else {
		r.getDistributionOptions().AccessKey = ""
	}
}

func (r *DeployGateUploadAppRequest) SetDistributionName(value string) {
	if value != "" {
		r.getDistributionOptions().Name = value
	} else {
		r.getDistributionOptions().Name = ""
	}
}

func (r *DeployGateUploadAppRequest) SetDistributionReleaseNote(value string) {
	if value != "" {
		r.getDistributionOptions().ReleaseNote = &value
	} else {
		r.getDistributionOptions().ReleaseNote = nil
	}
}

func (r *DeployGateUploadAppRequest) SetIOSDisableNotification(value bool) {
	r.iOSOptions.DisableNotification = value
}

func (r *DeployGateUploadAppRequest) getDistributionOptions() *deployGateDistributionOptions {
	if r.distributionOptions == nil {
		r.distributionOptions = &deployGateDistributionOptions{}
	}

	return r.distributionOptions
}

type errorResponse struct {
	Message string `json:"message"`
}

func (p *DeployGateProvider) Distribute(filePath string, builder func(req *DeployGateUploadAppRequest)) (*DeployGateDistributionResult, error) {
	request := NewDeployGateUploadAppRequest(filePath)

	builder(request)

	deployGateLogger.Debug().Msgf("the request has been built: %v", *request)

	var response deployGateUploadResponse

	if bytes, err := p.distribute(request); err != nil {
		return nil, err
	} else if err := json.Unmarshal(bytes, &response); err != nil {
		return nil, errors.Wrap(err, "failed to parse the response of your app to DeployGate but succeeded to upload")
	} else {
		return &DeployGateDistributionResult{
			deployGateUploadResponse: response,
			RawJson:                  string(bytes),
		}, nil
	}
}

func (p *DeployGateProvider) distribute(request *DeployGateUploadAppRequest) ([]byte, error) {
	client := deployGateBaseClient.WithHeaders(map[string][]string{
		"Authorization": {fmt.Sprintf("Bearer %s", p.ApiToken)},
	})

	code, bytes, err := client.DoPostMultipartForm(p.ctx, []string{"api", "users", p.AppOwnerName, "apps"}, p.toForm(request))

	if err != nil {
		return nil, errors.Wrap(err, "failed to upload your app to DeployGate")
	}

	if 200 <= code && code < 300 {
		return bytes, nil
	} else {
		var errorResponse errorResponse

		if err := json.Unmarshal(bytes, &errorResponse); err != nil {
			return nil, errors.Wrapf(err, "failed to upload your app to DeployGate due to: %s", string(bytes))
		} else if errorResponse.Message != "" {
			return nil, errors.New(fmt.Sprintf("failed to upload your app to DeployGate due to '%s'", errorResponse.Message))
		} else {
			return nil, errors.New(fmt.Sprintf("failed to upload your app to DeployGate due to '%s'", string(bytes)))
		}
	}
}

func (p *DeployGateProvider) toForm(request *DeployGateUploadAppRequest) *net.Form {
	form := net.Form{}

	form.Set(net.FileField("file", request.filePath))

	if request.message != nil {
		deployGateLogger.Debug().Msgf("message option was found")

		form.Set(net.StringField("message", *request.message))
	}

	if request.distributionOptions != nil {
		if request.distributionOptions.AccessKey != "" && request.distributionOptions.Name != "" {
			deployGateLogger.Warn().Msgf("the both of distribution's access key and name are specified so this provider prioritizes access key")
		}

		if request.distributionOptions.AccessKey != "" {
			form.Set(net.StringField("distribution_key", request.distributionOptions.AccessKey))
		} else if request.distributionOptions.Name != "" {
			form.Set(net.StringField("distribution_name", request.distributionOptions.Name))
		}

		if request.distributionOptions.ReleaseNote != nil {
			form.Set(net.StringField("release_note", *request.distributionOptions.ReleaseNote))
		} else if request.message != nil {
			deployGateLogger.Debug().Msgf("set message as release note as a fallback")
			form.Set(net.StringField("release_note", *request.message))
		}
	} else {
		deployGateLogger.Debug().Msgf("distribution options were empty")
	}

	var iosOptionFound = request.iOSOptions.DisableNotification

	if iosOptionFound {
		form.Set(net.BooleanField("disable_notify", request.iOSOptions.DisableNotification))
	}

	return &form
}
