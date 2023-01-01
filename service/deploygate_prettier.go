package service

type DeployGateDistributionResult struct {
	deployGateUploadResponse
	RawJson string
}

func (r *DeployGateDistributionResult) RawJsonResponse() string {
	return r.RawJson
}

func (r *DeployGateDistributionResult) ValueResponse() any {
	return *r
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
}
