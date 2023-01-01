package service

import (
	"github.com/jedib0t/go-pretty/v6/table"
)

type FirebaseAppDistributionDistributionResult struct {
	Result  string
	release *firebaseAppDistributionRelease
	aabInfo *firebaseAppDistributionAabInfoResponse
	RawJson string
}

type firebaseAppDistributionUploadResponse struct {
	OperationName string `json:"name"`
}

var FirebaseAppDistributionTableBuilder = func(w table.Writer, v any) {
	resp := v.(FirebaseAppDistributionDistributionResult)

	w.AppendHeader(table.Row{
		"Key", "Value",
	})

	if resp.release == nil {
		w.SetCaption("In async mode, only a few information is available.")
	}

	w.AppendRows([]table.Row{
		{"Firebase App Distribution", ""},
	})
	w.AppendSeparator()

	if resp.release != nil {
		w.AppendRows([]table.Row{
			{"Processed State", resp.Result},
			{"First Uploaded At", resp.release.CreatedAt},
			{"First Uploaded At", resp.release.CreatedAt},
		})

		if resp.release.ReleaseNote != nil {
			w.AppendRows([]table.Row{
				{"Release Note", resp.release.ReleaseNote.Text},
			})
		}
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

	if resp.release != nil {
		w.AppendSeparator()
		w.AppendRows([]table.Row{
			{"App Property", ""},
		})
		w.AppendSeparator()
		w.AppendRows([]table.Row{
			{"App Version Code", resp.release.BuildVersion},
			{"App Version Name", resp.release.DisplayVersion},
		})
	}
}
