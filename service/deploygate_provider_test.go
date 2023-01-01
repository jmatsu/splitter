package service

import (
	"github.com/jmatsu/splitter/internal/net"
	"reflect"
	"testing"
)

func Test_DeployGateUploadAppRequest_toForm(t *testing.T) {
	t.Parallel()

	sampleMessage1 := "sample1"

	cases := map[string]struct {
		request  DeployGateUploadAppRequest
		expected net.Form
	}{
		"with fully ios options": {
			request: DeployGateUploadAppRequest{
				filePath:   "path/to/file",
				iOSOptions: struct{ DisableNotification bool }{DisableNotification: true},
			},
			expected: net.Form{
				Fields: []net.ValueField{
					net.FileField("file", "path/to/file"),
					net.BooleanField("disable_notify", true),
				},
			},
		},
		"with too much distribution options": {
			request: DeployGateUploadAppRequest{
				filePath: "path/to/file",
				distributionOptions: &deployGateDistributionOptions{
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
			request: DeployGateUploadAppRequest{
				filePath: "path/to/file",
				distributionOptions: &deployGateDistributionOptions{
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
			request: DeployGateUploadAppRequest{
				filePath: "path/to/file",
				message:  &sampleMessage1,
				distributionOptions: &deployGateDistributionOptions{
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
			request: DeployGateUploadAppRequest{
				filePath: "path/to/file",
			},
			expected: net.Form{
				Fields: []net.ValueField{
					net.FileField("file", "path/to/file"),
				},
			},
		},
		"zero": {
			request: DeployGateUploadAppRequest{},
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

			form := c.request.toForm()

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
