package service

import "github.com/jmatsu/splitter/internal/exec"

type TestFlightUploadAppRequest struct {
	appleID  string
	password string
	issuerID string
	apiKey   string
	filePath string
}

func (r *TestFlightUploadAppRequest) NewAltoolCredential() *exec.AltoolCredential {
	if r.password != "" {
		if r.issuerID != "" || r.apiKey != "" {
			testFlightLogger.Warn().Msg("issuer id and api key will be ignored")
		}
	}

	return &exec.AltoolCredential{
		Password: r.password,
		IssuerID: r.issuerID,
		ApiKey:   r.apiKey,
	}
}

type testFlightUploadAppResponse struct {
}

func (p *TestFlightProvider) uploadApp(request *TestFlightUploadAppRequest) ([]byte, error) {
	altool := exec.NewAltool(p.ctx)
	return altool.UploadApp(request.filePath, request.appleID, request.NewAltoolCredential())
}
