package config

import (
	"fmt"
	"github.com/pkg/errors"
	"reflect"
	"testing"
)

func (c *GlobalConfig) assertEquals(other GlobalConfig) error {
	if len(c.deployments) != len(other.deployments) {
		return errors.New(fmt.Sprintf("%v does not equal to %v due to #deployments", c.deployments, other.deployments))
	}

	if c.FormatStyle() != other.FormatStyle() {
		return errors.New(fmt.Sprintf("%v does not equal to %v due to #FormatStyle", c.FormatStyle(), other.FormatStyle()))
	}

	if c.NetworkTimeout() != other.NetworkTimeout() {
		return errors.New(fmt.Sprintf("%v does not equal to %v due to #NetworkTimeout", c.NetworkTimeout(), other.NetworkTimeout()))
	}

	if c.WaitTimeout() != other.WaitTimeout() {
		return errors.New(fmt.Sprintf("%v does not equal to %v due to #WaitTimeout", c.WaitTimeout(), other.WaitTimeout()))
	}

	for name, v := range c.deployments {
		if !reflect.DeepEqual(v, other.deployments[name]) {
			return nil
		} else {
			return errors.New(fmt.Sprintf("%v does not equal to %v", v, other.deployments[name]))
		}
	}

	return nil
}

func Test_Config_configure(t *testing.T) {
	cases := map[string]struct {
		rawConfig rawConfig
		expected  *GlobalConfig
	}{
		"fully-written": {
			rawConfig: rawConfig{
				Deployments: map[string]interface{}{
					"def1": map[string]interface{}{
						"service":        DeploygateService,
						"app-owner-name": "def1-owner",
						"api-token":      "def1-token",
					},
					"def2": map[string]interface{}{
						"service":        FirebaseAppDistributionService,
						"project-number": "123456",
						"access-token":   "def2-token",
						"os-name":        "android",
						"package-name":   "com.example.android",
					},
					"def3": map[string]interface{}{
						"service":          LocalService,
						"destination-path": "def3-destination-path",
					},
				},
			},
			expected: &GlobalConfig{
				rawConfig: rawConfig{
					FormatStyle:    DefaultFormat,
					NetworkTimeout: DefaultNetworkTimeout,
					WaitTimeout:    DefaultWaitTimeout,
				},
				deployments: map[string]*Deployment{
					"def1": {
						ServiceName: DeploygateService,
						ServiceConfig: DeployGateConfig{
							AppOwnerName: "def1-owner",
							ApiToken:     "def1-token",
						},
					},
					"def2": {
						ServiceName: FirebaseAppDistributionService,
						ServiceConfig: FirebaseAppDistributionConfig{
							AccessToken: "def2-token",
							AppId:       "1:123456:android:xxxxx",
						},
					},
					"def3": {
						ServiceName: LocalService,
						ServiceConfig: LocalConfig{
							DestinationPath: "def3-destination-path",
						},
					},
				},
			},
		},
		"lacked": {
			rawConfig: rawConfig{
				Deployments: map[string]interface{}{
					"def1": map[string]interface{}{
						"service": DeploygateService,
					},
					"def2": map[string]interface{}{
						"service": FirebaseAppDistributionService,
					},
					"def3": map[string]interface{}{
						"service": LocalService,
					},
				},
			},
			expected: &GlobalConfig{
				rawConfig: rawConfig{
					FormatStyle:    DefaultFormat,
					NetworkTimeout: DefaultNetworkTimeout,
					WaitTimeout:    DefaultWaitTimeout,
				},
				deployments: map[string]*Deployment{
					"def1": {
						ServiceName:   DeploygateService,
						ServiceConfig: DeployGateConfig{},
					},
					"def2": {
						ServiceName:   FirebaseAppDistributionService,
						ServiceConfig: FirebaseAppDistributionConfig{},
					},
					"def3": {
						ServiceName:   LocalService,
						ServiceConfig: LocalConfig{},
					},
				},
			},
		},
		"zero": {
			rawConfig: rawConfig{},
			expected: &GlobalConfig{
				rawConfig: rawConfig{
					FormatStyle:    DefaultFormat,
					NetworkTimeout: DefaultNetworkTimeout,
					WaitTimeout:    DefaultWaitTimeout,
				},
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			config := GlobalConfig{
				rawConfig: c.rawConfig,
			}

			err := config.configure()

			if c.expected == nil && err != nil {
				return
			}

			if c.expected != nil && err == nil {
				if err := c.expected.assertEquals(config); err != nil {
					t.Errorf("%v", err)
				}

				return
			}

			if err != nil {
				t.Errorf("%s case is expected to be success but not: %v", name, err)
			} else {
				t.Errorf("%s case is expected to be failure but not", name)
			}
		})
	}
}
