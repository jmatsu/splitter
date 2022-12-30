package deploygate

import "github.com/jedib0t/go-pretty/v6/table"

type UploadResponse struct {
	Results struct {
		Name             string `json:"name"`
		PackageName      string `json:"package_name"`
		Revision         uint32 `json:"revision"`
		VersionCode      string `json:"version_code"`
		VersionName      string `json:"version_name"`
		MinSdkVersion    uint8  `json:"sdk_version"`
		TargetSdkVersion uint8  `json:"target_sdk_version"`
		DownloadUrl      string `json:"file"`
		User             struct {
			Name string
		} `json:"user"`
	} `json:"results"`
}

var TableBuilder = func(w table.Writer, v any) {
	resp := v.(UploadResponse)

	w.AppendRows([]table.Row{
		{"DeployGate Property", ""},
	})
	w.AppendSeparator()
	w.AppendRows([]table.Row{
		{"Owner Name", resp.Results.User.Name},
		{"Revision", resp.Results.Revision},
		{"Download URL", resp.Results.DownloadUrl},
	})
	w.AppendSeparator()
	w.AppendRows([]table.Row{
		{"App Property", ""},
	})
	w.AppendSeparator()
	w.AppendRows([]table.Row{
		{"Label", resp.Results.Name},
		{"Package Name", resp.Results.PackageName},
		{"Version Code", resp.Results.VersionCode},
		{"Version Name", resp.Results.VersionName},
		{"Min SDK Version", resp.Results.MinSdkVersion},
		{"Target SDK Version", resp.Results.TargetSdkVersion},
	})
}
