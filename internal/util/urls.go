package util

import (
	"fmt"
	"strings"
)

func CutEndpoint(v string) (string, string) {
	scheme, t, ok := strings.Cut(v, "://")

	if !ok {
		return "", ""
	}

	hostname, path, ok := strings.Cut(t, "/")

	if !ok {
		return fmt.Sprintf("%s://%s", scheme, hostname), ""
	}

	return fmt.Sprintf("%s://%s", scheme, hostname), path
}
