package service

import (
	"testing"
)

func assertValue(t *testing.T, name string, actual *string, expected string) {
	t.Helper()

	if expected == "" {
		if actual != nil {
			t.Errorf("%s is expected to be null but %s", name, *actual)
		}

		return
	}

	if actual == nil {
		t.Errorf("%s is expected to be %s but null", name, expected)
	} else if *actual != expected {
		t.Errorf("%s is expected to be %s but %s", name, expected, *actual)
	}
}

func Test_DeployGateDeployResult_NormalizedResponse(t *testing.T) {
	cases := map[string]struct {
		result              DeployGateDeployResult
		expectedOs          string
		expectedInstallUrl  string
		expectedReleaseNote string
	}{
		"zero": {
			result: DeployGateDeployResult{},
		},
		"without a distribution": {
			result: DeployGateDeployResult{
				DeployGateUploadResponse: DeployGateUploadResponse{
					Results: DeployGateBinaryFragment{OsName: "Android"},
				},
			},
			expectedOs: "android",
		},
		"with a distribution": {
			result: DeployGateDeployResult{
				DeployGateUploadResponse: DeployGateUploadResponse{
					Results: DeployGateBinaryFragment{
						OsName: "iOS",
						Distribution: &DeployGateDistributionFragment{
							Url:         "https://example.com/distributions/key",
							ReleaseNote: "a release note",
						},
					},
				},
			},
			expectedOs:          "ios",
			expectedInstallUrl:  "https://example.com/distributions/key",
			expectedReleaseNote: "a release note",
		},
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			normalized := c.result.NormalizedResponse()

			assertValue(t, "os", normalized.App.Os, c.expectedOs)
			assertValue(t, "install url", normalized.Release.InstallUrl, c.expectedInstallUrl)
			assertValue(t, "release note", normalized.Release.ReleaseNote, c.expectedReleaseNote)
		})
	}
}

func Test_FirebaseAppDistributionDeployResult_NormalizedResponse(t *testing.T) {
	cases := map[string]struct {
		result              FirebaseAppDistributionDeployResult
		expectedOs          string
		expectedVersionName string
		expectedReleaseNote string
	}{
		"zero": {
			result: FirebaseAppDistributionDeployResult{},
		},
		"an app bundle without a response": {
			result: FirebaseAppDistributionDeployResult{
				AabInfo: &FirebaseAppDistributionAabInfoResponse{},
			},
			expectedOs: "android",
		},
		"with a response": {
			result: FirebaseAppDistributionDeployResult{
				FirebaseAppDistributionGetOperationStateResponse: FirebaseAppDistributionGetOperationStateResponse{
					Response: &FirebaseAppDistributionV1UploadReleaseResponse{
						Release: FirebaseAppDistributionReleaseFragment{
							DisplayVersion: "1.0",
							ReleaseNote:    &FirebaseAppDistributionReleaseNoteFragment{Text: "a release note"},
						},
					},
				},
			},
			expectedVersionName: "1.0",
			expectedReleaseNote: "a release note",
		},
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			normalized := c.result.NormalizedResponse()

			assertValue(t, "os", normalized.App.Os, c.expectedOs)
			assertValue(t, "version name", normalized.App.VersionName, c.expectedVersionName)
			assertValue(t, "release note", normalized.Release.ReleaseNote, c.expectedReleaseNote)
		})
	}
}

func Test_LocalDeployResult_NormalizedResponse(t *testing.T) {
	result := LocalDeployResult{
		LocalMoveResponse: LocalMoveResponse{DestinationFilePath: "dist/app.apk"},
	}

	assertValue(t, "destination path", result.NormalizedResponse().Release.DestinationPath, "dist/app.apk")
	assertValue(t, "destination path", (&LocalDeployResult{}).NormalizedResponse().Release.DestinationPath, "")
}
