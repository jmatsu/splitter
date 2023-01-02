package util

import "testing"

func Test_CutEndpoint(t *testing.T) {
	cases := map[string]struct {
		value string

		expectedBaseUrl string
		expectedPath    string
	}{
		"url with path without port":        {value: "http://localhost/path1/path2", expectedBaseUrl: "http://localhost", expectedPath: "path1/path2"},
		"url with path with port":           {value: "http://localhost:3000/path1/path2", expectedBaseUrl: "http://localhost:3000", expectedPath: "path1/path2"},
		"url with path with trailing slash": {value: "http://localhost/path1/path2/", expectedBaseUrl: "http://localhost", expectedPath: "path1/path2/"},
		"url without port":                  {value: "http://localhost", expectedBaseUrl: "http://localhost", expectedPath: ""},
		"url with port":                     {value: "http://localhost:3000", expectedBaseUrl: "http://localhost:3000", expectedPath: ""},
		"url with trailing slash":           {value: "http://localhost/", expectedBaseUrl: "http://localhost", expectedPath: ""},
		"empty":                             {value: "", expectedBaseUrl: "", expectedPath: ""},
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			baseUrl, path := CutEndpoint(c.value)

			if baseUrl == c.expectedBaseUrl && path == c.expectedPath {
				return
			}

			t.Fatalf("baseUrl = %s, path = %s", baseUrl, path)
		})
	}
}
