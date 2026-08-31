package service

// NormalizedResult is a service-agnostic view of a deployment result. Every field is nullable
// because no service exposes all of them.
type NormalizedResult struct {
	App     NormalizedApp     `json:"app"`
	Release NormalizedRelease `json:"release"`
}

type NormalizedApp struct {
	Name        *string `json:"name"`
	Identifier  *string `json:"identifier"`
	Os          *string `json:"os"`
	VersionName *string `json:"version_name"`
	VersionCode *string `json:"version_code"`
}

type NormalizedRelease struct {
	InstallUrl      *string `json:"install_url"`
	DownloadUrl     *string `json:"download_url"`
	DestinationPath *string `json:"destination_path"`
	ReleaseNote     *string `json:"release_note"`
}

// nullable turns an unset value into a null field instead of an empty string.
func nullable(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}
