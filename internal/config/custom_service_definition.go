package config

import (
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"strings"
)

type valueAssignFormat = string

const (
	requestBodyAssignFormat      valueAssignFormat = "request_body"
	formParamsAssignFormatPrefix valueAssignFormat = "form_params."
	headersAssignFormatPrefix    valueAssignFormat = "headers."
	queryAssignFormatPrefix      valueAssignFormat = "query."
)

type CustomServiceDefinition struct {
	Endpoint                 string                   `yaml:"endpoint"`
	SourceFileFormat         valueAssignFormat        `yaml:"source-file-format"`
	AuthDefinition           CustomAuthDefinition     `yaml:"auth"`
	DefaultRequestDefinition DefaultRequestDefinition `yaml:"default,omitempty"`
}

func (d *CustomServiceDefinition) validate() error {

	if d.SourceFileFormat != requestBodyAssignFormat {
		var valid bool

		for _, prefix := range []string{formParamsAssignFormatPrefix, queryAssignFormatPrefix} {
			if !strings.HasPrefix(d.SourceFileFormat, prefix) {
				valid = false
				break
			}
		}

		if !valid {
			return errors.New(fmt.Sprintf("%s does not follow the correct format", d.SourceFileFormat))
		}
	}

	return nil
}

type CustomAuthDefinition struct {
	StyleFormat valueAssignFormat `yaml:"style-format"`
	ValueFormat string            `yaml:"value-format"`
}

func (d *CustomAuthDefinition) validate() error {
	var valid bool

	for _, prefix := range []string{formParamsAssignFormatPrefix, headersAssignFormatPrefix, queryAssignFormatPrefix} {
		if !strings.HasPrefix(d.StyleFormat, prefix) {
			valid = false
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

type DefaultRequestDefinition struct {
	Headers    map[string]string `yaml:"headers,omitempty"`
	Query      map[string]string `yaml:"query,omitempty"`
	FormParams map[string]string `yaml:"form-params,omitempty"`
}

func (d *DefaultRequestDefinition) validate() error {
	if slices.Contains(maps.Keys(d.Headers), "") {
		return errors.New("headers has at least one empty key")
	}

	if slices.Contains(maps.Keys(d.Query), "") {
		return errors.New("query has at least one empty key")
	}

	if slices.Contains(maps.Keys(d.FormParams), "") {
		return errors.New("form-params has at least one empty key")
	}

	return nil
}
