package config

import (
	"fmt"
	"golang.org/x/exp/slices"
	"strings"
)

var osNames = []string{"android", "ios"}

type FirebaseAppDistributionConfig struct {
	AccessToken   string `json:"access-token,omitempty" required:"true"`
	ProjectNumber string `json:"project-number,omitempty" required:"true"`
	OsName        string `json:"os,omitempty" required:"true"`
	PackageName   string `json:"package-name,omitempty" required:"true"`
}

func (c *FirebaseAppDistributionConfig) Validate() error {
	c.OsName = strings.ToLower(c.OsName)

	if err := validateMissingValues(c); err != nil {
		return err
	} else if !slices.Contains(osNames, c.OsName) {
		return fmt.Errorf("%s is not acceptable os name", c.OsName)
	}

	return nil
}
