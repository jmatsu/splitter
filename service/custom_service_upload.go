package service

import (
	"fmt"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/internal/net"
	"github.com/pkg/errors"
)

type CustomServiceUploadAppRequest struct {
	filePath string

	headers map[string][]string
	queries map[string]string
	form    net.Form
}

type CustomServiceUploadResponse struct {
	RawResponse *net.HttpResponse `json:"-"`
}

func (r *CustomServiceUploadResponse) Set(v *net.HttpResponse) {
	r.RawResponse = v
}

var _ net.TypedHttpResponse = &CustomServiceUploadResponse{}

func (p *CustomServiceProvider) upload(request *CustomServiceUploadAppRequest) (*CustomServiceUploadResponse, error) {
	authToken := fmt.Sprintf(p.CustomServiceDefinition.AuthDefinition.ValueFormat, p.CustomServiceConfig.AuthToken)

	if prefix, name, err := p.CustomServiceDefinition.AuthDefinition.AuthValue(); err != nil {
		return nil, errors.Wrap(err, "couldn't get an auth")
	} else {
		switch prefix {
		case config.HeadersAssignFormatPrefix:
			request.headers[name] = []string{authToken}
		case config.FormParamsAssignFormatPrefix:
			request.form.Set(net.StringField(name, authToken))
		case config.QueryAssignFormatPrefix:
			request.queries[name] = authToken
		default:
			panic(fmt.Sprintf("%s is not implemented yet", prefix))
		}
	}

	if format, name, err := p.SourceFile(); err != nil {
		switch format {
		case config.RequestBodyAssignFormat:
			if !request.form.Empty() {
				return nil, errors.New(fmt.Sprintf("%s is not compatible with form requests", format))
			}
		case config.FormParamsAssignFormatPrefix:
			request.form.Set(net.FileField(name, request.filePath))
		default:
			panic(fmt.Sprintf("%s is not implemented yet", format))
		}
	}

	client := p.client.WithHeaders(request.headers)

	var resp *net.HttpResponse
	var err error

	if request.form.Empty() {
		resp, err = client.DoPostFileBody(p.ctx, []string{p.path}, request.queries, request.filePath)
	} else {
		resp, err = client.DoPostMultipartForm(p.ctx, []string{p.path}, request.queries, &request.form)
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
