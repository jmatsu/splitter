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
	return &exec.AltoolCredential{
		Password: r.password,
		IssuerID: r.issuerID,
		ApiKey:   r.apiKey,
	}
}

// TestFlightUploadAppResponse holds what `altool --upload-app --output-format json` reports.
type TestFlightUploadAppResponse struct {
	SuccessMessage string `json:"success-message"`
	ToolVersion    string `json:"tool-version"`
}

func (p *TestFlightProvider) uploadApp(request *TestFlightUploadAppRequest) ([]byte, error) {
	altool := exec.NewAltool(p.ctx)
	return altool.UploadApp(request.filePath, request.appleID, request.NewAltoolCredential())
}
