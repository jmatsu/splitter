package firebase_app_distribution

import (
	"github.com/jedib0t/go-pretty/v6/table"
)

type DistributionResult struct {
	*v1UploadReleaseResponse
	aabInfo *aabInfoResponse
	RawJson string
}

type uploadResponse struct {
	OperationName string `json:"name"`
}

var TableBuilder = func(w table.Writer, v any) {
	resp := v.(DistributionResult)

	w.AppendHeader(table.Row{
		"Key", "Value",
	})

	if resp.v1UploadReleaseResponse == nil {
		w.SetCaption("In async mode, only a few information is available.")
	}

	w.AppendRows([]table.Row{
		{"Firebase App Distribution", ""},
	})
	w.AppendSeparator()

	if resp.v1UploadReleaseResponse != nil {
		w.AppendRows([]table.Row{
			{"Processed State", resp.Result},
			{"First Uploaded At", resp.Release.CreatedAt},
		})
	}

	// it's okay to use aabInfo != nil as *if android*
	if aabInfo := resp.aabInfo; aabInfo != nil {
		w.AppendRows([]table.Row{
			{"App Bundle Available", aabInfo.IntegrationState == aabIntegrationIntegrated},
			{"Play Store Integration", aabInfo.IntegrationState},
		})

		if certificate := aabInfo.TestCertificate; certificate != nil {
			w.AppendRows([]table.Row{
				{"App Bundle Certificate MD5", certificate.Md5},
				{"App Bundle Certificate SHA1", certificate.Sha1},
				{"App Bundle Certificate SHA256", certificate.Sha256},
			})
		}
	}

	if resp.v1UploadReleaseResponse != nil {
		w.AppendSeparator()
		w.AppendRows([]table.Row{
			{"App Property", ""},
		})
		w.AppendSeparator()
		w.AppendRows([]table.Row{
			{"App Version Code", resp.Release.BuildVersion},
			{"App Version Name", resp.Release.DisplayVersion},
		})
	}
}
