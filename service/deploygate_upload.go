package service

import (
	"fmt"
	"github.com/jmatsu/splitter/internal/net"
	"github.com/pkg/errors"
)

func NewDeployGateUploadAppRequest(filePath string) *DeployGateUploadAppRequest {
	return &DeployGateUploadAppRequest{
		filePath:            filePath,
		distributionOptions: &deployGateDistributionOptions{},
	}
}

type DeployGateUploadAppRequest struct {
	filePath            string
	message             *string
	distributionOptions *deployGateDistributionOptions
	iOSOptions          deployGateIOSOptions
}

type deployGateUploadResponse struct {
	Results struct {
		OsName string `json:"os_name"`

		Name             string  `json:"name"`
		PackageName      string  `json:"package_name"`
		Revision         uint32  `json:"revision"`
		VersionCode      string  `json:"version_code"`
		VersionName      string  `json:"version_name"`
		SdkVersion       uint16  `json:"sdk_version"`
		RawSdkVersion    *string `json:"raw_sdk_version"`
		TargetSdkVersion *uint16 `json:"target_sdk_version"`
		DownloadUrl      string  `json:"file"`
		User             struct {
			Name string
		} `json:"user"`
		Distribution *struct {
			AccessKey   string `json:"access_key"`
			Title       string `json:"title"`
			ReleaseNote string `json:"release_note"`
			Url         string `json:"url"`
		} `json:"distribution"`
	} `json:"results"`

	RawResponse *net.HttpResponse `json:"-"`
}

func (r *deployGateUploadResponse) Set(v *net.HttpResponse) {
	r.RawResponse = v
}

type deployGateDistributionOptions struct {
	Name        string
	AccessKey   string
	ReleaseNote *string
}

type deployGateIOSOptions struct {
	DisableNotification bool
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

func (r *DeployGateUploadAppRequest) toForm() *net.Form {
	form := net.Form{}

	form.Set(net.FileField("file", r.filePath))

	if r.message != nil {
		deployGateLogger.Debug().Msgf("message option was found")

		form.Set(net.StringField("message", *r.message))
	}

	if r.distributionOptions != nil {
		if r.distributionOptions.AccessKey != "" && r.distributionOptions.Name != "" {
			deployGateLogger.Warn().Msgf("the both of distribution's access key and name are specified so this provider prioritizes access key")
		}

		if r.distributionOptions.AccessKey != "" {
			form.Set(net.StringField("distribution_key", r.distributionOptions.AccessKey))
		} else if r.distributionOptions.Name != "" {
			form.Set(net.StringField("distribution_name", r.distributionOptions.Name))
		}

		if r.distributionOptions.ReleaseNote != nil {
			form.Set(net.StringField("release_note", *r.distributionOptions.ReleaseNote))
		} else if r.message != nil {
			deployGateLogger.Debug().Msgf("set message as release note as a fallback")
			form.Set(net.StringField("release_note", *r.message))
		}
	} else {
		deployGateLogger.Debug().Msgf("distribution options were empty")
	}

	var iosOptionFound = r.iOSOptions.DisableNotification

	if iosOptionFound {
		form.Set(net.BooleanField("disable_notify", r.iOSOptions.DisableNotification))
	}

	return &form
}

func (p *DeployGateProvider) upload(request *DeployGateUploadAppRequest) (*deployGateUploadResponse, error) {
	client := p.client.WithHeaders(map[string][]string{
		"Authorization": {fmt.Sprintf("Bearer %s", p.ApiToken)},
	})

	resp, err := client.DoPostMultipartForm(p.ctx, []string{"api", "users", p.AppOwnerName, "apps"}, request.toForm())

	if err != nil {
		return nil, errors.Wrap(err, "failed to upload your app to DeployGate")
	}

	if resp.Successful() {
		if v, err := resp.ParseJson(&deployGateUploadResponse{}); err != nil {
			return nil, errors.Wrap(err, "succeeded to upload but something went wrong")
		} else {
			return v.(*deployGateUploadResponse), nil
		}
	} else {
		return nil, errors.Wrap(resp.Err(), "failed to upload your app to DeployGate")
	}
}
