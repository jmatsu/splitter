package deploygate

import (
	"context"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/internal/net"
	"reflect"
	"testing"
)

func Test_Provider_toForm(t *testing.T) {
	t.Parallel()

	provider := Provider{
		DeployGateConfig: config.DeployGateConfig{
			ApiToken:     "ApiToken",
			AppOwnerName: "AppOwnerName",
		},
		ctx: context.TODO(),
	}

	sampleMessage1 := "sample1"

	cases := map[string]struct {
		request  UploadRequest
		expected net.Form
	}{
		"with fully ios options": {
			request: UploadRequest{
				FilePath:   "path/to/file",
				IOSOptions: struct{ DisableNotification bool }{DisableNotification: true},
			},
			expected: net.Form{
				Fields: []net.ValueField{
					net.FileField("file", "path/to/file"),
					net.BooleanField("disable_notify", true),
				},
			},
		},
		"with too much distribution options": {
			request: UploadRequest{
				FilePath: "path/to/file",
				DistributionOptions: struct {
					Name        string
					AccessKey   string
					ReleaseNote *string
				}{
					AccessKey: "dist_key",
					Name:      "dist_name",
				},
			},
			expected: net.Form{
				Fields: []net.ValueField{
					net.FileField("file", "path/to/file"),
					net.StringField("distribution_key", "dist_key"),
				},
			},
		},
		"with fully distribution options": {
			request: UploadRequest{
				FilePath: "path/to/file",
				DistributionOptions: struct {
					Name        string
					AccessKey   string
					ReleaseNote *string
				}{
					AccessKey:   "dist_key",
					ReleaseNote: &sampleMessage1,
				},
			},
			expected: net.Form{
				Fields: []net.ValueField{
					net.FileField("file", "path/to/file"),
					net.StringField("distribution_key", "dist_key"),
					net.StringField("release_note", "sample1"),
				},
			},
		},
		"with partial distribution options": {
			request: UploadRequest{
				FilePath: "path/to/file",
				Message:  &sampleMessage1,
				DistributionOptions: struct {
					Name        string
					AccessKey   string
					ReleaseNote *string
				}{
					Name: "dist_name1",
				},
			},
			expected: net.Form{
				Fields: []net.ValueField{
					net.FileField("file", "path/to/file"),
					net.StringField("message", "sample1"),
					net.StringField("distribution_name", "dist_name1"),
					net.StringField("release_note", "sample1"),
				},
			},
		},
		"minimum": {
			request: UploadRequest{
				FilePath: "path/to/file",
			},
			expected: net.Form{
				Fields: []net.ValueField{
					net.FileField("file", "path/to/file"),
				},
			},
		},
		"zero": {
			request: UploadRequest{},
			expected: net.Form{
				Fields: []net.ValueField{
					net.FileField("file", ""),
				},
			},
		},
	}

	for name, c := range cases {
		name, c := name, c
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			form := provider.toForm(&c.request)

			if len(form.Fields) != len(c.expected.Fields) {
				t.Errorf("actual length is %d but expected %d", len(form.Fields), len(c.expected.Fields))
			}

			for _, ef := range c.expected.Fields {
				af := findField(form.Fields, ef.FieldName)

				if af == nil {
					t.Errorf("%s is not found in a form", ef.FieldName)
				} else if !reflect.DeepEqual(*af, ef) {
					t.Errorf("%v is not equal to %v", *af, ef)
				}
			}
		})
	}
}

func findField(fields []net.ValueField, name string) *net.ValueField {
	for _, field := range fields {
		if field.FieldName == name {
			return &field
		}
	}

	return nil
}
