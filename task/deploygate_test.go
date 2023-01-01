package task

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jmatsu/splitter/service"
	"testing"
)

func Test_deployGateTableBuilder(t *testing.T) {
	cases := map[string]struct {
		result service.DeployGateDistributionResult
	}{
		"zero 1": {
			result: service.DeployGateDistributionResult{},
		},
		"zero 2": {
			result: service.DeployGateDistributionResult{
				DeployGateUploadResponse: service.DeployGateUploadResponse{
					Results: service.DeployGateBinaryFragment{
						Distribution: &service.DeployGateDistributionFragment{},
					},
				},
			},
		},
		"regular with distribution": {
			result: service.DeployGateDistributionResult{
				DeployGateUploadResponse: service.DeployGateUploadResponse{
					Results: service.DeployGateBinaryFragment{
						OsName:           "Android",
						Name:             "App Name",
						PackageName:      "com.example",
						Revision:         1,
						VersionCode:      "1",
						VersionName:      "1.0",
						SdkVersion:       15,
						RawSdkVersion:    "30",
						TargetSdkVersion: 30,
						DownloadUrl:      "http://example.com",
						Distribution: &service.DeployGateDistributionFragment{
							AccessKey:   "access key",
							Title:       "title",
							ReleaseNote: "sample",
							Url:         "http://example.com",
						},
						User: service.DeployGateUserFragment{
							Name: "name",
						},
					},
				},
			},
		},
		"regular without distribution": {
			result: service.DeployGateDistributionResult{
				DeployGateUploadResponse: service.DeployGateUploadResponse{
					Results: service.DeployGateBinaryFragment{
						OsName:           "Android",
						Name:             "App Name",
						PackageName:      "com.example",
						Revision:         1,
						VersionCode:      "1",
						VersionName:      "1.0",
						SdkVersion:       15,
						RawSdkVersion:    "30",
						TargetSdkVersion: 30,
						DownloadUrl:      "http://example.com",
						Distribution:     nil,
						User: service.DeployGateUserFragment{
							Name: "name",
						},
					},
				},
			},
		},
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			w := table.NewWriter()

			// no panic is ok
			deployGateTableBuilder(w, c.result)
		})
	}
}
