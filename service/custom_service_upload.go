package service

import (
	"fmt"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/internal/net"
	"github.com/pkg/errors"
)

type CustomServiceUploadAppRequest struct {
	filePath string
}

type CustomServiceUploadResponse struct {
	RawResponse *net.HttpResponse `json:"-"`
}

func (r *CustomServiceUploadResponse) Set(v *net.HttpResponse) {
	r.RawResponse = v
}

var _ net.TypedHttpResponse = &CustomServiceUploadResponse{}

func (p *CustomServiceProvider) upload(request *CustomServiceUploadAppRequest) (*CustomServiceUploadResponse, error) {
	var form *net.Form
	var queries map[string]string

	client := p.client
	authToken := fmt.Sprintf(p.CustomServiceDefinition.AuthDefinition.ValueFormat, p.CustomServiceConfig.AuthToken)

	if prefix, name, err := p.CustomServiceDefinition.AuthDefinition.AuthValue(); err != nil {
		return nil, errors.Wrap(err, "couldn't get an auth")
	} else {
		switch prefix {
		case config.HeadersAssignFormatPrefix:
			client = p.client.WithHeaders(map[string][]string{
				name: {authToken},
			})
		case config.FormParamsAssignFormatPrefix:
			form = &net.Form{}
			form.Set(net.StringField(name, authToken))
		case config.QueryAssignFormatPrefix:
			queries[name] = authToken
		default:
			panic(fmt.Sprintf("%s is not implemented yet", prefix))
		}
	}

	if format, name, err := p.SourceFile(); err != nil {
		switch format {
		case config.RequestBodyAssignFormat:
			if form != nil {
				return nil, errors.New(fmt.Sprintf("%s is not compatible with form requests", format))
			}

			// no op
		case config.FormParamsAssignFormatPrefix:
			if form == nil {
				form = &net.Form{}
			}

			form.Set(net.StringField(name, request.filePath))
		default:
			panic(fmt.Sprintf("%s is not implemented yet", format))
		}
	}

	var resp *net.HttpResponse
	var err error

	if form != nil {
		resp, err = client.DoPostMultipartForm(p.ctx, []string{p.path}, queries, form)
	} else {
		resp, err = client.DoPostFileBody(p.ctx, []string{p.path}, queries, request.filePath)
	}

	if err != nil {
		return nil, errors.Wrap(err, "failed to upload your app to DeployGate")
	}

	if resp.Successful() {
		if v, err := resp.ParseJson(&CustomServiceUploadResponse{}); err != nil {
			return nil, errors.Wrap(err, "succeeded to upload but something went wrong")
		} else {
			return v.(*CustomServiceUploadResponse), nil
		}
	} else {
		return nil, errors.Wrap(resp.Err(), "failed to upload your app to custom service")
	}
}
