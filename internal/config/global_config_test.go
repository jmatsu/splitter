package config

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

// serviceEnvNames are the environment variables that loadServiceConfig reads. Tests must not be affected by them.
var serviceEnvNames = []string{
	"DEPLOYGATE_APP_OWNER_NAME",
	"DEPLOYGATE_API_TOKEN",
	"DEPLOYGATE_DISTRIBUTION_KEY",
	"DEPLOYGATE_DISTRIBUTION_NAME",
	"GOOGLE_APPLICATION_CREDENTIALS",
}

func unsetServiceEnv(t *testing.T) {
	t.Helper()

	for _, name := range serviceEnvNames {
		name := name

		if value, found := os.LookupEnv(name); found {
			value := value
			t.Cleanup(func() {
				_ = os.Setenv(name, value)
			})
		}

		_ = os.Unsetenv(name)
	}
}

func (c *GlobalConfig) assertEquals(other GlobalConfig) error {
	if c.FormatStyle() != other.FormatStyle() {
		return errors.New(fmt.Sprintf("%v does not equal to %v due to #FormatStyle", c.FormatStyle(), other.FormatStyle()))
	}

	if c.NetworkTimeout() != other.NetworkTimeout() {
		return errors.New(fmt.Sprintf("%v does not equal to %v due to #NetworkTimeout", c.NetworkTimeout(), other.NetworkTimeout()))
	}

	if c.WaitTimeout() != other.WaitTimeout() {
		return errors.New(fmt.Sprintf("%v does not equal to %v due to #WaitTimeout", c.WaitTimeout(), other.WaitTimeout()))
	}

	if len(c.deployments) != len(other.deployments) {
		return errors.New(fmt.Sprintf("%v does not equal to %v due to #deployments", c.deployments, other.deployments))
	}

	for name, v := range c.deployments {
		if !reflect.DeepEqual(v, other.deployments[name]) {
			return errors.New(fmt.Sprintf("%s: %#v does not equal to %#v", name, v, other.deployments[name]))
		}
	}

	if len(c.services) != len(other.services) {
		return errors.New(fmt.Sprintf("%v does not equal to %v due to #services", c.services, other.services))
	}

	for name, v := range c.services {
		if !reflect.DeepEqual(v, other.services[name]) {
			return errors.New(fmt.Sprintf("%s: %#v does not equal to %#v", name, v, other.services[name]))
		}
	}

	return nil
}

// A definition that satisfies CustomServiceDefinition#validate.
var customServiceDefinitionValues = map[string]interface{}{
	"endpoint":           "https://example.com/path/to/upload",
	"source-file-format": FormParamsAssignFormatPrefix + "file",
	"auth": map[string]interface{}{
		"style-format": HeadersAssignFormatPrefix + "Authorization",
		"value-format": "Bearer %s",
	},
}

var customServiceDefinition = CustomServiceDefinition{
	Endpoint:         "https://example.com/path/to/upload",
	SourceFileFormat: FormParamsAssignFormatPrefix + "file",
	AuthDefinition: CustomAuthDefinition{
		StyleFormat: HeadersAssignFormatPrefix + "Authorization",
		ValueFormat: "Bearer %s",
	},
}

func Test_Config_configure(t *testing.T) {
	unsetServiceEnv(t)

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
						"service":      FirebaseAppDistributionService,
						"app-id":       "1:123456:android:xxxxx",
						"access-token": "def2-token",
					},
					"def3": map[string]interface{}{
						"service":          LocalService,
						"destination-path": "def3-destination-path",
					},
					"def4": map[string]interface{}{
						"service":  TestFlightService,
						"apple-id": "def4-apple-id",
						"password": "def4-password",
					},
				},
			},
			expected: &GlobalConfig{
				rawConfig: rawConfig{
					FormatStyle:    DefaultFormat,
					NetworkTimeout: DefaultNetworkTimeout,
					WaitTimeout:    DefaultWaitTimeout,
				},
				deployments: map[string]Deployment{
					"def1": {
						ServiceName: DeploygateService,
						ServiceConfig: DeployGateConfig{
							serviceNameHolder: serviceNameHolder{Name: DeploygateService},
							AppOwnerName:      "def1-owner",
							ApiToken:          "def1-token",
						},
					},
					"def2": {
						ServiceName: FirebaseAppDistributionService,
						ServiceConfig: FirebaseAppDistributionConfig{
							serviceNameHolder: serviceNameHolder{Name: FirebaseAppDistributionService},
							AppId:             "1:123456:android:xxxxx",
							AccessToken:       "def2-token",
						},
					},
					"def3": {
						ServiceName: LocalService,
						ServiceConfig: LocalConfig{
							serviceNameHolder: serviceNameHolder{Name: LocalService},
							DestinationPath:   "def3-destination-path",
						},
					},
					"def4": {
						ServiceName: TestFlightService,
						ServiceConfig: TestFlightConfig{
							serviceNameHolder: serviceNameHolder{Name: TestFlightService},
							AppleID:           "def4-apple-id",
							Password:          "def4-password",
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
					"def4": map[string]interface{}{
						"service": TestFlightService,
					},
				},
			},
			expected: &GlobalConfig{
				rawConfig: rawConfig{
					FormatStyle:    DefaultFormat,
					NetworkTimeout: DefaultNetworkTimeout,
					WaitTimeout:    DefaultWaitTimeout,
				},
				deployments: map[string]Deployment{
					"def1": {
						ServiceName: DeploygateService,
						ServiceConfig: DeployGateConfig{
							serviceNameHolder: serviceNameHolder{Name: DeploygateService},
						},
					},
					"def2": {
						ServiceName: FirebaseAppDistributionService,
						ServiceConfig: FirebaseAppDistributionConfig{
							serviceNameHolder: serviceNameHolder{Name: FirebaseAppDistributionService},
						},
					},
					"def3": {
						ServiceName: LocalService,
						ServiceConfig: LocalConfig{
							serviceNameHolder: serviceNameHolder{Name: LocalService},
						},
					},
					"def4": {
						ServiceName: TestFlightService,
						ServiceConfig: TestFlightConfig{
							serviceNameHolder: serviceNameHolder{Name: TestFlightService},
						},
					},
				},
			},
		},
		"lifecycle steps": {
			rawConfig: rawConfig{
				Deployments: map[string]interface{}{
					"def1": map[string]interface{}{
						"service":          LocalService,
						"destination-path": "def1-destination-path",
						"pre-steps":        []interface{}{[]interface{}{"echo", "pre"}},
						"post-steps":       []interface{}{[]interface{}{"echo", "post"}},
					},
				},
			},
			expected: &GlobalConfig{
				rawConfig: rawConfig{
					FormatStyle:    DefaultFormat,
					NetworkTimeout: DefaultNetworkTimeout,
					WaitTimeout:    DefaultWaitTimeout,
				},
				deployments: map[string]Deployment{
					"def1": {
						ServiceName: LocalService,
						ServiceConfig: LocalConfig{
							serviceNameHolder: serviceNameHolder{Name: LocalService},
							ExecutionConfig: ExecutionConfig{
								PreSteps:  [][]string{{"echo", "pre"}},
								PostSteps: [][]string{{"echo", "post"}},
							},
							DestinationPath: "def1-destination-path",
						},
						Lifecycle: ExecutionConfig{
							PreSteps:  [][]string{{"echo", "pre"}},
							PostSteps: [][]string{{"echo", "post"}},
						},
					},
				},
			},
		},
		"custom service": {
			rawConfig: rawConfig{
				Services: map[string]interface{}{
					"custom1": customServiceDefinitionValues,
				},
				Deployments: map[string]interface{}{
					"def1": map[string]interface{}{
						"service":    "custom1",
						"auth-token": "def1-auth-token",
					},
				},
			},
			expected: &GlobalConfig{
				rawConfig: rawConfig{
					FormatStyle:    DefaultFormat,
					NetworkTimeout: DefaultNetworkTimeout,
					WaitTimeout:    DefaultWaitTimeout,
				},
				services: map[string]CustomServiceDefinition{
					"custom1": customServiceDefinition,
				},
				deployments: map[string]Deployment{
					"def1": {
						ServiceName: "custom1",
						ServiceConfig: CustomServiceConfig{
							serviceNameHolder: serviceNameHolder{Name: "custom1"},
							AuthToken:         "def1-auth-token",
						},
					},
				},
			},
		},
		"non-default global values": {
			rawConfig: rawConfig{
				FormatStyle:    MarkdownFormat,
				NetworkTimeout: "1m",
				WaitTimeout:    "30s",
			},
			expected: &GlobalConfig{
				rawConfig: rawConfig{
					FormatStyle:    MarkdownFormat,
					NetworkTimeout: "1m",
					WaitTimeout:    "30s",
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
		"unknown service": {
			rawConfig: rawConfig{
				Deployments: map[string]interface{}{
					"def1": map[string]interface{}{
						"service": "unknown",
					},
				},
			},
		},
		"reserved service name": {
			rawConfig: rawConfig{
				Services: map[string]interface{}{
					DeploygateService: customServiceDefinitionValues,
				},
			},
		},
		"invalid service definition": {
			rawConfig: rawConfig{
				Services: map[string]interface{}{
					"custom1": map[string]interface{}{
						"endpoint":           "https://example.com/path/to/upload",
						"source-file-format": "unsupported",
					},
				},
			},
		},
		"deployment is not a mapping": {
			rawConfig: rawConfig{
				Deployments: map[string]interface{}{
					"def1": "not a mapping",
				},
			},
		},
		"service definition is not a mapping": {
			rawConfig: rawConfig{
				Services: map[string]interface{}{
					"custom1": "not a mapping",
				},
			},
		},
		"invalid format style": {
			rawConfig: rawConfig{
				FormatStyle: "unknown",
			},
		},
		"invalid network timeout": {
			rawConfig: rawConfig{
				NetworkTimeout: "10 minutes",
			},
		},
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			config := GlobalConfig{
				rawConfig: c.rawConfig,
			}

			err := config.configure()

			if c.expected == nil {
				if err == nil {
					t.Errorf("%s case is expected to be failure but not", name)
				}

				return
			}

			if err != nil {
				t.Errorf("%s case is expected to be success but not: %v", name, err)
				return
			}

			if err := c.expected.assertEquals(config); err != nil {
				t.Errorf("%v", err)
			}
		})
	}
}

func Test_GlobalConfig_Deployment(t *testing.T) {
	unsetServiceEnv(t)

	t.Setenv("SPLITTER_TEST_OWNER_NAME", "evaluated-owner")

	cases := map[string]struct {
		rawConfig rawConfig

		name string

		expectedServiceName   string
		expectedServiceConfig any
		expectedDefinition    CustomServiceDefinition
		expectedSuccess       bool
	}{
		"deploygate": {
			rawConfig: rawConfig{
				Deployments: map[string]interface{}{
					"def1": map[string]interface{}{
						"service":        DeploygateService,
						"app-owner-name": "def1-owner",
						"api-token":      "def1-token",
					},
				},
			},
			name:                "def1",
			expectedServiceName: DeploygateService,
			expectedServiceConfig: DeployGateConfig{
				serviceNameHolder: serviceNameHolder{Name: DeploygateService},
				AppOwnerName:      "def1-owner",
				ApiToken:          "def1-token",
			},
			expectedSuccess: true,
		},
		"deploygate with an embedded variable": {
			rawConfig: rawConfig{
				Deployments: map[string]interface{}{
					"def1": map[string]interface{}{
						"service":        DeploygateService,
						"app-owner-name": "format:${SPLITTER_TEST_OWNER_NAME}",
						"api-token":      "def1-token",
					},
				},
			},
			name:                "def1",
			expectedServiceName: DeploygateService,
			expectedServiceConfig: DeployGateConfig{
				serviceNameHolder: serviceNameHolder{Name: DeploygateService},
				AppOwnerName:      "evaluated-owner",
				ApiToken:          "def1-token",
			},
			expectedSuccess: true,
		},
		"firebase app distribution": {
			rawConfig: rawConfig{
				Deployments: map[string]interface{}{
					"def1": map[string]interface{}{
						"service":      FirebaseAppDistributionService,
						"app-id":       "1:123456:android:xxxxx",
						"access-token": "def1-token",
					},
				},
			},
			name:                "def1",
			expectedServiceName: FirebaseAppDistributionService,
			expectedServiceConfig: FirebaseAppDistributionConfig{
				serviceNameHolder: serviceNameHolder{Name: FirebaseAppDistributionService},
				AppId:             "1:123456:android:xxxxx",
				AccessToken:       "def1-token",
			},
			expectedSuccess: true,
		},
		"local": {
			rawConfig: rawConfig{
				Deployments: map[string]interface{}{
					"def1": map[string]interface{}{
						"service":          LocalService,
						"destination-path": "def1-destination-path",
					},
				},
			},
			name:                "def1",
			expectedServiceName: LocalService,
			expectedServiceConfig: LocalConfig{
				serviceNameHolder: serviceNameHolder{Name: LocalService},
				DestinationPath:   "def1-destination-path",
			},
			expectedSuccess: true,
		},
		"test flight": {
			rawConfig: rawConfig{
				Deployments: map[string]interface{}{
					"def1": map[string]interface{}{
						"service":  TestFlightService,
						"apple-id": "def1-apple-id",
						"password": "def1-password",
					},
				},
			},
			name:                "def1",
			expectedServiceName: TestFlightService,
			expectedServiceConfig: TestFlightConfig{
				serviceNameHolder: serviceNameHolder{Name: TestFlightService},
				AppleID:           "def1-apple-id",
				Password:          "def1-password",
			},
			expectedSuccess: true,
		},
		"custom service": {
			rawConfig: rawConfig{
				Services: map[string]interface{}{
					"custom1": customServiceDefinitionValues,
				},
				Deployments: map[string]interface{}{
					"def1": map[string]interface{}{
						"service":    "custom1",
						"auth-token": "def1-auth-token",
					},
				},
			},
			name:                "def1",
			expectedServiceName: "custom1",
			expectedServiceConfig: CustomServiceConfig{
				serviceNameHolder: serviceNameHolder{Name: "custom1"},
				AuthToken:         "def1-auth-token",
			},
			expectedDefinition: customServiceDefinition,
			expectedSuccess:    true,
		},
		"missing required values": {
			rawConfig: rawConfig{
				Deployments: map[string]interface{}{
					"def1": map[string]interface{}{
						"service": DeploygateService,
					},
				},
			},
			name: "def1",
		},
		"unknown deployment": {
			rawConfig: rawConfig{
				Deployments: map[string]interface{}{
					"def1": map[string]interface{}{
						"service":          LocalService,
						"destination-path": "def1-destination-path",
					},
				},
			},
			name: "def2",
		},
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			config := GlobalConfig{rawConfig: c.rawConfig}

			if err := config.configure(); err != nil {
				t.Fatalf("failed to configure: %v", err)
			}

			deployment, definition, err := config.Deployment(c.name)

			if !c.expectedSuccess {
				if err == nil {
					t.Errorf("%s case is expected to be failure but not", name)
				}

				return
			}

			if err != nil {
				t.Fatalf("%s case is expected to be success but not: %v", name, err)
			}

			if deployment.ServiceName != c.expectedServiceName {
				t.Errorf("service name is expected to be %s but %s", c.expectedServiceName, deployment.ServiceName)
			}

			if !reflect.DeepEqual(deployment.ServiceConfig, c.expectedServiceConfig) {
				t.Errorf("service config is expected to be %#v but %#v", c.expectedServiceConfig, deployment.ServiceConfig)
			}

			if !reflect.DeepEqual(definition, c.expectedDefinition) {
				t.Errorf("definition is expected to be %#v but %#v", c.expectedDefinition, definition)
			}
		})
	}
}

func Test_GlobalConfig_AddDeployment(t *testing.T) {
	unsetServiceEnv(t)

	cases := map[string]struct {
		serviceName     string
		expectedSuccess bool
	}{
		"deploygate":                {serviceName: DeploygateService, expectedSuccess: true},
		"firebase app distribution": {serviceName: FirebaseAppDistributionService, expectedSuccess: true},
		"local":                     {serviceName: LocalService, expectedSuccess: true},
		"test flight":               {serviceName: TestFlightService, expectedSuccess: true},
	}

	unknownServiceConfig := GlobalConfig{
		rawConfig: rawConfig{
			Deployments: map[string]interface{}{},
		},
	}

	if err := unknownServiceConfig.configure(); err != nil {
		t.Fatalf("failed to configure: %v", err)
	}

	if err := unknownServiceConfig.AddDeployment("new1", "obababa"); err == nil {
		t.Errorf("an unknown service name is expected to be rejected but not")
	} else if _, ok := unknownServiceConfig.rawConfig.Deployments["new1"]; ok {
		t.Errorf("an unknown service is expected to leave the raw config untouched but not")
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			config := GlobalConfig{
				rawConfig: rawConfig{
					Deployments: map[string]interface{}{},
				},
			}

			if err := config.configure(); err != nil {
				t.Fatalf("failed to configure: %v", err)
			}

			if err := config.AddDeployment("new1", c.serviceName); (err == nil) != c.expectedSuccess {
				t.Fatalf("%s case is expected to be %t but %t", name, c.expectedSuccess, err == nil)
			}

			d, ok := config.deployments["new1"]

			if !ok {
				t.Fatalf("new1 is not added")
			}

			if d.ServiceName != c.serviceName {
				t.Errorf("service name is expected to be %s but %s", c.serviceName, d.ServiceName)
			}

			if _, ok := config.rawConfig.Deployments["new1"]; !ok {
				t.Errorf("new1 is not added to the raw config so it won't be dumped")
			}

			// The same name must not be added twice.
			if err := config.AddDeployment("new1", c.serviceName); err == nil {
				t.Errorf("the duplicated name is expected to be rejected but not")
			}
		})
	}
}

func Test_GlobalConfig_Definition(t *testing.T) {
	config := GlobalConfig{
		rawConfig: rawConfig{
			Services: map[string]interface{}{
				"custom1": customServiceDefinitionValues,
			},
		},
	}

	if err := config.configure(); err != nil {
		t.Fatalf("failed to configure: %v", err)
	}

	if v, err := config.Definition("custom1"); err != nil {
		t.Errorf("custom1 is expected to be found but not: %v", err)
	} else if !reflect.DeepEqual(v, customServiceDefinition) {
		t.Errorf("definition is expected to be %#v but %#v", customServiceDefinition, v)
	}

	if _, err := config.Definition("custom2"); err == nil {
		t.Errorf("custom2 is expected to be missing but found")
	}
}

func Test_GlobalConfig_Dump(t *testing.T) {
	unsetServiceEnv(t)

	config := GlobalConfig{
		rawConfig: rawConfig{
			Deployments: map[string]interface{}{},
		},
	}

	if err := config.configure(); err != nil {
		t.Fatalf("failed to configure: %v", err)
	}

	if err := config.AddDeployment("new1", LocalService); err != nil {
		t.Fatalf("failed to add a deployment: %v", err)
	}

	path := filepath.Join(t.TempDir(), "splitter.yml")

	if err := config.Dump(path); err != nil {
		t.Fatalf("failed to dump: %v", err)
	}

	bytes, err := os.ReadFile(path)

	if err != nil {
		t.Fatalf("failed to read the dumped file: %v", err)
	}

	// The dumped file must be loadable again.
	loaded := GlobalConfig{}

	if err := yaml.Unmarshal(bytes, &loaded.rawConfig); err != nil {
		t.Fatalf("failed to parse the dumped file: %v", err)
	}

	if err := loaded.configure(); err != nil {
		t.Fatalf("the dumped file is not loadable: %v", err)
	}

	if d, ok := loaded.deployments["new1"]; !ok {
		t.Errorf("new1 is not dumped")
	} else if d.ServiceName != LocalService {
		t.Errorf("service name is expected to be %s but %s", LocalService, d.ServiceName)
	}
}

func Test_GlobalConfig_Validate(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		rawConfig         rawConfig
		expectedValidness bool
	}{
		"defaults": {
			rawConfig: rawConfig{
				FormatStyle:    DefaultFormat,
				NetworkTimeout: DefaultNetworkTimeout,
				WaitTimeout:    DefaultWaitTimeout,
			},
			expectedValidness: true,
		},
		"raw format": {
			rawConfig: rawConfig{
				FormatStyle:    RawFormat,
				NetworkTimeout: "1s",
				WaitTimeout:    "1s",
			},
			expectedValidness: true,
		},
		"markdown format": {
			rawConfig: rawConfig{
				FormatStyle:    MarkdownFormat,
				NetworkTimeout: "30m",
				WaitTimeout:    "10m",
			},
			expectedValidness: true,
		},
		"unknown format style": {
			rawConfig: rawConfig{
				FormatStyle:    "unknown",
				NetworkTimeout: DefaultNetworkTimeout,
				WaitTimeout:    DefaultWaitTimeout,
			},
		},
		"empty format style": {
			rawConfig: rawConfig{
				NetworkTimeout: DefaultNetworkTimeout,
				WaitTimeout:    DefaultWaitTimeout,
			},
		},
		"malformed network timeout": {
			rawConfig: rawConfig{
				FormatStyle:    DefaultFormat,
				NetworkTimeout: "10 minutes",
				WaitTimeout:    DefaultWaitTimeout,
			},
		},
		"too long network timeout": {
			rawConfig: rawConfig{
				FormatStyle:    DefaultFormat,
				NetworkTimeout: "31m",
				WaitTimeout:    DefaultWaitTimeout,
			},
		},
		"empty network timeout": {
			rawConfig: rawConfig{
				FormatStyle: DefaultFormat,
				WaitTimeout: DefaultWaitTimeout,
			},
		},
		"malformed wait timeout": {
			rawConfig: rawConfig{
				FormatStyle:    DefaultFormat,
				NetworkTimeout: DefaultNetworkTimeout,
				WaitTimeout:    "5 minutes",
			},
		},
		"too long wait timeout": {
			rawConfig: rawConfig{
				FormatStyle:    DefaultFormat,
				NetworkTimeout: DefaultNetworkTimeout,
				WaitTimeout:    "11m",
			},
		},
		"empty wait timeout": {
			rawConfig: rawConfig{
				FormatStyle:    DefaultFormat,
				NetworkTimeout: DefaultNetworkTimeout,
			},
		},
		"zero": {
			rawConfig: rawConfig{},
		},
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			config := GlobalConfig{rawConfig: c.rawConfig}

			if err := config.Validate(); (err == nil) != c.expectedValidness {
				t.Errorf("%s case is expected to be %t but %t: %v", name, c.expectedValidness, err == nil, err)
			}
		})
	}
}

func Test_GlobalConfig_Timeouts(t *testing.T) {
	t.Parallel()

	networkTimeout, _ := time.ParseDuration(DefaultNetworkTimeout)
	waitTimeout, _ := time.ParseDuration(DefaultWaitTimeout)

	// Zero values must fall back into the defaults.
	config := GlobalConfig{}

	if v := config.NetworkTimeout(); v != networkTimeout {
		t.Errorf("network timeout is expected to be %s but %s", networkTimeout, v)
	}

	if v := config.WaitTimeout(); v != waitTimeout {
		t.Errorf("wait timeout is expected to be %s but %s", waitTimeout, v)
	}

	config = GlobalConfig{
		rawConfig: rawConfig{
			NetworkTimeout: "1m",
			WaitTimeout:    "30s",
		},
	}

	if v := config.NetworkTimeout(); v != time.Minute {
		t.Errorf("network timeout is expected to be %s but %s", time.Minute, v)
	}

	if v := config.WaitTimeout(); v != 30*time.Second {
		t.Errorf("wait timeout is expected to be %s but %s", 30*time.Second, v)
	}
}

func Test_ToEnvName(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"config_file": "SPLITTER_CONFIG_FILE",
		"FORMAT":      "SPLITTER_FORMAT",
		"":            "SPLITTER_",
	}

	for name, expected := range cases {
		name, expected := name, expected

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if v := ToEnvName(name); v != expected {
				t.Errorf("%s is expected to be %s but %s", name, expected, v)
			}
		})
	}
}

func Test_LoadGlobalConfig(t *testing.T) {
	unsetServiceEnv(t)

	// LoadGlobalConfig mutates the package level state so it must be restored.
	original := config
	t.Cleanup(func() {
		config = original
		viper.Reset()
	})

	content := `
format-style: markdown
network-timeout: 1m
wait-timeout: 30s
services:
  custom1:
    endpoint: https://example.com/path/to/upload
    source-file-format: form_params.file
    auth:
      style-format: headers.Authorization
      value-format: "Bearer %s"
deployments:
  local1:
    service: local
    destination-path: ./dist/app.apk
  custom-deployment:
    service: custom1
    auth-token: custom-auth-token
`

	path := filepath.Join(t.TempDir(), DefaultConfigName)

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write a config file: %v", err)
	}

	config = &GlobalConfig{}

	if err := LoadGlobalConfig(&path); err != nil {
		t.Fatalf("failed to load the config file: %v", err)
	}

	loaded := CurrentConfig()

	if v := loaded.FormatStyle(); v != MarkdownFormat {
		t.Errorf("format style is expected to be %s but %s", MarkdownFormat, v)
	}

	if v := loaded.NetworkTimeout(); v != time.Minute {
		t.Errorf("network timeout is expected to be %s but %s", time.Minute, v)
	}

	if v := loaded.WaitTimeout(); v != 30*time.Second {
		t.Errorf("wait timeout is expected to be %s but %s", 30*time.Second, v)
	}

	if d, _, err := loaded.Deployment("local1"); err != nil {
		t.Errorf("local1 is expected to be found but not: %v", err)
	} else if conf, ok := d.ServiceConfig.(LocalConfig); !ok {
		t.Errorf("local1 is expected to hold LocalConfig but %T", d.ServiceConfig)
	} else if conf.DestinationPath != "./dist/app.apk" {
		t.Errorf("destination path is expected to be ./dist/app.apk but %s", conf.DestinationPath)
	}

	if d, definition, err := loaded.Deployment("custom-deployment"); err != nil {
		t.Errorf("custom-deployment is expected to be found but not: %v", err)
	} else if d.ServiceName != "custom1" {
		t.Errorf("service name is expected to be custom1 but %s", d.ServiceName)
	} else if definition.Endpoint != "https://example.com/path/to/upload" {
		t.Errorf("the service definition is not resolved: %#v", definition)
	}

	// The global setters must affect the loaded config.
	SetGlobalFormatStyle(RawFormat)
	SetGlobalNetworkTimeout("2m")
	SetGlobalWaitTimeout("1m")

	loaded = CurrentConfig()

	if v := loaded.FormatStyle(); v != RawFormat {
		t.Errorf("format style is expected to be %s but %s", RawFormat, v)
	}

	if v := loaded.NetworkTimeout(); v != 2*time.Minute {
		t.Errorf("network timeout is expected to be %s but %s", 2*time.Minute, v)
	}

	if v := loaded.WaitTimeout(); v != time.Minute {
		t.Errorf("wait timeout is expected to be %s but %s", time.Minute, v)
	}

	if err := loaded.Validate(); err != nil {
		t.Errorf("the overridden config is expected to be valid but not: %v", err)
	}
}

func Test_LoadGlobalConfig_missingFile(t *testing.T) {
	original := config
	t.Cleanup(func() {
		config = original
		viper.Reset()
	})

	path := filepath.Join(t.TempDir(), "not-found.yml")

	config = &GlobalConfig{}

	if err := LoadGlobalConfig(&path); err == nil {
		t.Errorf("a missing config file is expected to be rejected but not")
	}
}

func Test_NewConfig(t *testing.T) {
	t.Parallel()

	// A new config must be dumpable as a boilerplate.
	conf := NewConfig()

	if conf == nil {
		t.Fatal("NewConfig must not return nil")
	}

	if err := conf.Dump(filepath.Join(t.TempDir(), DefaultConfigName)); err != nil {
		t.Errorf("failed to dump a new config: %v", err)
	}
}
