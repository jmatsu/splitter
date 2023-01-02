package config

import (
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"net/url"
	"strings"
)

type valueAssignFormat = string

const (
	RequestBodyAssignFormat      valueAssignFormat = "request_body"
	FormParamsAssignFormatPrefix valueAssignFormat = "form_params."
	HeadersAssignFormatPrefix    valueAssignFormat = "headers."
	QueryAssignFormatPrefix      valueAssignFormat = "query_params."
)

type CustomServiceDefinition struct {
	Endpoint                 string                   `yaml:"endpoint" required:"true"`
	SourceFileFormat         valueAssignFormat        `yaml:"source-file-format" required:"true"`
	AuthDefinition           CustomAuthDefinition     `yaml:"auth" required:"true"`
	DefaultRequestDefinition DefaultRequestDefinition `yaml:"default,omitempty"`
}

func (d *CustomServiceDefinition) validate() error {
	if _, err := url.Parse(d.Endpoint); err != nil {
		return errors.Wrapf(err, "%s is not a URL format", d.Endpoint)
	}

	if d.SourceFileFormat != RequestBodyAssignFormat {
		var valid bool

		for _, prefix := range []string{FormParamsAssignFormatPrefix, QueryAssignFormatPrefix} {
			if strings.HasPrefix(d.SourceFileFormat, prefix) {
				if len(d.SourceFileFormat) > len(prefix) {
					return errors.New(fmt.Sprintf("%s must contain *name*", d.SourceFileFormat))
				}

				valid = true
				break
			}
		}

		if !valid {
			return errors.New(fmt.Sprintf("%s does not follow the correct format", d.SourceFileFormat))
		}
	}

	return nil
}

func (d *CustomServiceDefinition) SourceFile() (string, string, error) {
	if d.SourceFileFormat == RequestBodyAssignFormat {
		return RequestBodyAssignFormat, "", nil
	} else if strings.HasPrefix(d.SourceFileFormat, FormParamsAssignFormatPrefix) {
		name := d.SourceFileFormat[len(FormParamsAssignFormatPrefix):]

		if name == "" {
			return "", "", errors.New(fmt.Sprintf("no name is available in %s", d.SourceFileFormat))
		}

		return FormParamsAssignFormatPrefix, name, nil
	}

	return "", "", errors.New(fmt.Sprintf("no source file format is found in %s", d.SourceFileFormat))
}

type CustomAuthDefinition struct {
	StyleFormat valueAssignFormat `yaml:"style-format"`
	ValueFormat string            `yaml:"value-format"`
}

func (d *CustomAuthDefinition) validate() error {
	var valid bool

	for _, prefix := range []string{FormParamsAssignFormatPrefix, HeadersAssignFormatPrefix, QueryAssignFormatPrefix} {
		if strings.HasPrefix(d.StyleFormat, prefix) {
			if len(d.StyleFormat) > len(prefix) {
				return errors.New(fmt.Sprintf("%s must contain *name*", d.StyleFormat))
			}

			valid = true
			break
		}
	}

	if !valid {
		return errors.New(fmt.Sprintf("%s does not follow the correct format", d.StyleFormat))
	}

	if n := len(strings.SplitN(d.ValueFormat, "%s", 3)); n > 1 {
		return errors.New(fmt.Sprintf("%s contains 2 or more %%s", d.StyleFormat))
	} else if n == 0 {
		return errors.New(fmt.Sprintf("%s must contain %%s", d.StyleFormat))
	}

	return nil
}

func (d *CustomAuthDefinition) AuthValue() (string, string, error) {
	for _, prefix := range []string{FormParamsAssignFormatPrefix, HeadersAssignFormatPrefix, QueryAssignFormatPrefix} {
		if strings.HasPrefix(d.StyleFormat, prefix) {
			name := d.StyleFormat[len(prefix):]

			if name == "" {
				return "", "", errors.New(fmt.Sprintf("no name is available in %s", d.StyleFormat))
			}

			return prefix, name, nil
		}
	}

	return "", "", errors.New(fmt.Sprintf("no authentication method is found in %s", d.StyleFormat))
}

type DefaultRequestDefinition struct {
	Headers    map[string]string   `yaml:"headers,omitempty"`
	Queries    map[string][]string `yaml:"queries,omitempty"`
	FormParams map[string]string   `yaml:"form-params,omitempty"`
}

func (d *DefaultRequestDefinition) validate() error {
	if slices.Contains(maps.Keys(d.Headers), "") {
		return errors.New("headers has at least one empty key")
	}

	if slices.Contains(maps.Keys(d.Queries), "") {
		return errors.New("query has at least one empty key")
	}

	if slices.Contains(maps.Keys(d.FormParams), "") {
		return errors.New("form-params has at least one empty key")
	}

	return nil
}
