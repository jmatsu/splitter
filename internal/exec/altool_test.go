package exec

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"k8s.io/utils/exec"
	testingexec "k8s.io/utils/exec/testing"
)

// newRecordingAltool returns an Altool whose command line records the invocation instead of running it.
func newRecordingAltool(recorded *[]string, err error) *Altool {
	sh := &testingexec.FakeExec{
		CommandScript: []testingexec.FakeCommandAction{
			func(cmd string, args ...string) exec.Cmd {
				*recorded = append([]string{cmd}, args...)

				fakeCmd := &testingexec.FakeCmd{}
				testingexec.InitFakeCmd(fakeCmd, cmd, args...)
				fakeCmd.RunScript = []testingexec.FakeAction{
					func() ([]byte, []byte, error) {
						return []byte("stdout"), []byte("stderr"), err
					},
				}
				return fakeCmd
			},
		},
		ExactOrder: true,
	}

	return &Altool{commandLine: NewCommandLine(context.TODO(), sh)}
}

func Test_Altool_UploadApp(t *testing.T) {
	cases := map[string]struct {
		credential *AltoolCredential
		expected   []string
	}{
		"with a password": {
			credential: &AltoolCredential{Password: "password"},
			expected: []string{
				"xcrun", "altool", "--upload-app",
				"-f", "path/to/app.ipa",
				"-t", "ios",
				"--username", "apple-id",
				"--output-format", "json",
				"--password", "password",
			},
		},
		"with an api key": {
			credential: &AltoolCredential{ApiKey: "api-key", IssuerID: "issuer-id"},
			expected: []string{
				"xcrun", "altool", "--upload-app",
				"-f", "path/to/app.ipa",
				"-t", "ios",
				"--username", "apple-id",
				"--output-format", "json",
				"--apiKey", "api-key", "--apiIssuer", "issuer-id",
			},
		},
		"a password takes priority": {
			credential: &AltoolCredential{Password: "password", ApiKey: "api-key", IssuerID: "issuer-id"},
			expected: []string{
				"xcrun", "altool", "--upload-app",
				"-f", "path/to/app.ipa",
				"-t", "ios",
				"--username", "apple-id",
				"--output-format", "json",
				"--password", "password",
			},
		},
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			var recorded []string

			altool := newRecordingAltool(&recorded, nil)

			stdout, err := altool.UploadApp("path/to/app.ipa", "apple-id", c.credential)

			if err != nil {
				t.Fatalf("%s is expected to be success but not: %v", name, err)
			}

			if string(stdout) != "stdout" {
				t.Errorf("stdout is expected to be stdout but %s", string(stdout))
			}

			if !reflect.DeepEqual(recorded, c.expected) {
				t.Errorf("%s is expected to run %v but %v", name, c.expected, recorded)
			}
		})
	}
}

func Test_Altool_UploadApp_failure(t *testing.T) {
	var recorded []string

	altool := newRecordingAltool(&recorded, errors.New("altool is not available"))

	if _, err := altool.UploadApp("path/to/app.ipa", "apple-id", &AltoolCredential{Password: "password"}); err == nil {
		t.Errorf("a failure of altool is expected to be an error but not")
	}
}

func Test_NewAltool(t *testing.T) {
	if altool := NewAltool(context.TODO()); altool.commandLine == nil {
		t.Errorf("NewAltool must build a command line")
	}
}
