package firebase_app_distribution

import (
	"github.com/jedib0t/go-pretty/v6/table"
)

type DistributionResult struct {
	v1UploadReleaseResponse
	aabInfo *aabInfoResponse
	RawJson string
}

type uploadResponse struct {
	OperationName string `json:"name"`
}

var TableBuilder = func(w table.Writer, v any) {
	resp := v.(DistributionResult)

	w.AppendRows([]table.Row{
		{"Firebase App Distribution Property", ""},
	})
	w.AppendSeparator()
	w.AppendRows([]table.Row{
		{"Release Response", resp.Result},
		{"First Uploaded At", resp.Release.CreatedAt},
	})

	// it's okay to use aabInfo != nil as *if android*
	if aabInfo := resp.aabInfo; aabInfo != nil {
		w.AppendSeparator()
		w.AppendRows([]table.Row{
			{"App Bundles", ""},
		})
		w.AppendSeparator()
		w.AppendRows([]table.Row{
			{"Play Store Integration", aabInfo.IntegrationState},
		})

		if certificate := aabInfo.TestCertificate; certificate != nil {
			w.AppendSeparator()
			w.AppendRows([]table.Row{
				{"App Bundle Certificate", ""},
			})
			w.AppendSeparator()
			w.AppendRows([]table.Row{
				{"MD5", certificate.Md5},
				{"SHA1", certificate.Sha1},
				{"SHA256", certificate.Sha256},
			})
		}
	}

	w.AppendSeparator()
	w.AppendRows([]table.Row{
		{"App Property", ""},
	})
	w.AppendSeparator()
	w.AppendRows([]table.Row{
		{"Version Code", resp.Release.BuildVersion},
		{"Version Name", resp.Release.DisplayVersion},
	})

	w.SetCaption("The official document is available at %s", "https://firebase.google.com/docs/app-distribution/android/distribute-fastlane")
}
