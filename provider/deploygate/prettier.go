package deploygate

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"strings"
)

type UploadResponse struct {
	Results struct {
		OsName string `json:"os_name"`

		Name             string  `json:"name"`
		PackageName      string  `json:"package_name"`
		Revision         uint32  `json:"revision"`
		VersionCode      string  `json:"version_code"`
		VersionName      string  `json:"version_name"`
		SdkVersion       uint16  `json:"sdk_version"`
		RawSdkVersion    *string `json:"raw_sdk_version,omitempty"`
		TargetSdkVersion *uint16 `json:"target_sdk_version,omitempty"`
		DownloadUrl      string  `json:"file"`
		User             struct {
			Name string
		} `json:"user"`
		Distribution *struct {
			AccessKey   string `json:"access_key"`
			Title       string `json:"title"`
			ReleaseNote string `json:"release_note"`
			Url         string `json:"url"`
		} `json:"distribution,omitempty"`
	} `json:"results"`
}

var TableBuilder = func(w table.Writer, v any) {
	resp := v.(UploadResponse).Results

	w.AppendRows([]table.Row{
		{"DeployGate Property", ""},
	})
	w.AppendSeparator()
	w.AppendRows([]table.Row{
		{"Owner Name", resp.User.Name},
		{"Revision", resp.Revision},
	})

	switch strings.ToLower(resp.OsName) {
	case "android":
		w.AppendRows([]table.Row{
			{"APK Download URL", resp.DownloadUrl},
		})
	case "ios":
		w.AppendRows([]table.Row{
			{"IPA Download URL", resp.DownloadUrl},
		})
	}

	if resp.Distribution != nil {
		w.AppendRows([]table.Row{
			{"Distribution Name", resp.Distribution.Title},
			{"Distribution AccessKey", resp.Distribution.AccessKey},
			{"Distribution URL", resp.Distribution.Url},
		})
	}

	w.AppendSeparator()
	w.AppendRows([]table.Row{
		{"App Property", ""},
	})
	w.AppendSeparator()

	switch strings.ToLower(resp.OsName) {
	case "android":
		w.AppendRows([]table.Row{
			{"Display Name", resp.Name},
			{"Package Name", resp.PackageName},
			{"Version Code", resp.VersionCode},
			{"Version Name", resp.VersionName},
			{"Min SDK Version", resp.SdkVersion},
			{"Target SDK Version", *resp.TargetSdkVersion},
		})

	case "ios":
		w.AppendRows([]table.Row{
			{"Display Name", resp.Name},
			{"Bundle Identifier", resp.PackageName},
			{"Version Code", resp.VersionCode},
			{"Version Name", resp.VersionName},
			{"Build SDK Version", *resp.RawSdkVersion},
		})
	}
}
