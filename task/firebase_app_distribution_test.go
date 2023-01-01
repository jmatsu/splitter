package task

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jmatsu/splitter/service"
	"testing"
)

func Test_firebaseAppDistributionTableBuilder(t *testing.T) {
	cases := map[string]struct {
		result service.FirebaseAppDistributionDistributionResult
	}{
		"zero 1": {
			result: service.FirebaseAppDistributionDistributionResult{},
		},
		"zero 2": {
			result: service.FirebaseAppDistributionDistributionResult{
				FirebaseAppDistributionGetOperationStateResponse: service.FirebaseAppDistributionGetOperationStateResponse{
					Response: &service.FirebaseAppDistributionV1UploadReleaseResponse{},
				},
			},
		},
		"zero 3": {
			result: service.FirebaseAppDistributionDistributionResult{
				FirebaseAppDistributionGetOperationStateResponse: service.FirebaseAppDistributionGetOperationStateResponse{
					Response: &service.FirebaseAppDistributionV1UploadReleaseResponse{
						Release: service.FirebaseAppDistributionReleaseFragment{
							ReleaseNote: &service.FirebaseAppDistributionReleaseNoteFragment{},
						},
					},
				},
			},
		},
		"regular": {
			result: service.FirebaseAppDistributionDistributionResult{
				FirebaseAppDistributionGetOperationStateResponse: service.FirebaseAppDistributionGetOperationStateResponse{
					OperationName: "op",
					Done:          true,
					Response: &service.FirebaseAppDistributionV1UploadReleaseResponse{
						Result: "result",
						Release: service.FirebaseAppDistributionReleaseFragment{
							Name:           "release name",
							DisplayVersion: "1.0",
							BuildVersion:   "1",
							CreatedAt:      "2022-12-31T11:41:17.873594Z",
							ReleaseNote: &service.FirebaseAppDistributionReleaseNoteFragment{
								Text: "release note",
							},
						},
					},
				},
			},
		},
		"regular without release note": {
			result: service.FirebaseAppDistributionDistributionResult{
				FirebaseAppDistributionGetOperationStateResponse: service.FirebaseAppDistributionGetOperationStateResponse{
					OperationName: "op",
					Done:          true,
					Response: &service.FirebaseAppDistributionV1UploadReleaseResponse{
						Result: "result",
						Release: service.FirebaseAppDistributionReleaseFragment{
							Name:           "release name",
							DisplayVersion: "1.0",
							BuildVersion:   "1",
							CreatedAt:      "2022-12-31T11:41:17.873594Z",
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
			firebaseAppDistributionTableBuilder(w, c.result)
		})
	}
}
