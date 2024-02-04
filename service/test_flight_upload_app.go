package service

import "github.com/jmatsu/splitter/internal/exec"

type testFlightUploadAppRequest struct {
	appleID  string
	password string
	filePath string
}

type testFlightUploadAppResponse struct {
}

func (p *TestFlightProvider) uploadApp(request *testFlightUploadAppRequest) ([]byte, error) {
	altool := exec.NewAltool(p.ctx)
	return altool.UploadApp(request.filePath, request.appleID, request.password)
}
