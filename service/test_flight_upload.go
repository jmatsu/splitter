package service

import (
	"context"
	"github.com/cidertool/asc-go/asc"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/internal/net"
	"github.com/jmatsu/splitter/service/ios"
	"github.com/pkg/errors"
	"os"
)

type TestFlightProvider struct {
	config.TestFlightConfig
	ctx              context.Context
	developerAccount *ios.AppleDeveloperAccount
}

type TestFlightUploadAppRequest struct {
	filePath string
}

func (p *TestFlightProvider) upload(request TestFlightUploadAppRequest) (*string, error) {
	ditto := ios.NewDitto(p.ctx)

	f, err := os.CreateTemp(os.TempDir(), "testflight-*")

	if err != nil {
		return nil, errors.Wrap(err, "failed to create a temp file")
	}

	defer f.Close()

	if err := ditto.CreateZip(request.filePath, f.Name()); err != nil {
		return nil, errors.Wrap(err, "a zip file is required to upload")
	}

	notarytool := ios.NewNotarytool(p.ctx)
	if uuid, err := notarytool.Submit(p.developerAccount, f.Name(), true); err != nil {
		return nil, nil
	}

	return nil, nil
}

func (p *TestFlightProvider) x() (*string, error) {
	key, _ := os.ReadFile(p.TestFlightConfig.KeyPath)

	auth, err := asc.NewTokenConfig(p.TestFlightConfig.KeyId, p.TestFlightConfig.IssuerId, p.TestFlightConfig.TokenExpiry(), key)

	if err != nil {
		return nil, err
	}

	c := auth.Client()
	c.Timeout = config.CurrentConfig().NetworkTimeout()

	client := asc.NewClient(c)
	client.UserAgent = net.UserAgent()
	_ = client.TestFlight

	return nil, nil
}
