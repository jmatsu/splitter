package config

import (
	"fmt"
	"github.com/jmatsu/splitter/internal/logger"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

const (
	envPrefix = "SPLITTER_" // The prefix of environment variables used for splitter's global options.

	DefaultConfigName = "splitter.yml"

	DeploygateService              = "deploygate"                // represents DeployGateConfig
	LocalService                   = "local"                     // represents LocalConfig
	FirebaseAppDistributionService = "firebase-app-distribution" // represents FirebaseAppDistributionConfig
	TestFlightService              = "test-flight"               // represents TestFlightConfig
)

func ToEnvName(name string) string {
	return fmt.Sprintf("%s%s", envPrefix, strings.ToUpper(name))
}

// GlobalConfig is a shared configuration in one command execution.
type GlobalConfig struct {
	rawConfig   rawConfig
	deployments map[string]Deployment
	services    map[string]CustomServiceDefinition
}

type rawConfig struct {
	Deployments    map[string]interface{} `yaml:"deployments"`
	Services       map[string]interface{} `yaml:"services"`
	FormatStyle    string                 `yaml:"format-style,omitempty"`
	NetworkTimeout string                 `yaml:"network-timeout,omitempty"`
	WaitTimeout    string                 `yaml:"wait-timeout,omitempty"`
}

// Deployment holds a service name and its config struct
type Deployment struct {
	ServiceName   string
	ServiceConfig any // See serviceConfig interface
	Lifecycle     ExecutionConfig
}

type FormatStyle = string

const (
	PrettyFormat   FormatStyle = "pretty"
	RawFormat      FormatStyle = "raw"
	MarkdownFormat FormatStyle = "markdown"

	DefaultFormat = PrettyFormat

	DefaultNetworkTimeout = "10m"
	DefaultWaitTimeout    = "5m"
)

// configNameCandidates are looked up on the working directory in this order when --config is absent.
var configNameCandidates = []string{DefaultConfigName, "splitter.yaml"}

var styles = []FormatStyle{
	PrettyFormat,
	RawFormat,
	MarkdownFormat,
}

var config = &GlobalConfig{}

func NewConfig() *GlobalConfig {
	return &GlobalConfig{}
}

func SetGlobalFormatStyle(value string) {
	config.rawConfig.FormatStyle = value
}

func SetGlobalNetworkTimeout(value string) {
	config.rawConfig.NetworkTimeout = value
}

func SetGlobalWaitTimeout(value string) {
	config.rawConfig.WaitTimeout = value
}

func CurrentConfig() *GlobalConfig {
	return config
}

func LoadGlobalConfig(path *string) error {
	filePath, err := resolveConfigPath(path)

	if err != nil {
		return err
	}

	var raw rawConfig

	if filePath != "" {
		logger.Logger.Debug().Msgf("Loading a config file on %s", filePath)

		bytes, err := os.ReadFile(filePath)

		if err != nil {
			return errors.Wrapf(err, "failed to read a config file on %s", filePath)
		}

		// The config file is parsed by yaml directly because deployment and service names are
		// case-sensitive keys that a user refers to by --name.
		if err := yaml.Unmarshal(bytes, &raw); err != nil {
			return errors.Wrapf(err, "failed to parse a config file on %s", filePath)
		}
	}

	config.rawConfig = raw

	if err := config.configure(); err != nil {
		return errors.Wrap(err, "your config file may not contain some of required values or they are invalid")
	}

	return nil
}

// resolveConfigPath returns a path to the config file to load. An empty path means no config file
// is available, which is not an error unless a user has specified one explicitly.
func resolveConfigPath(path *string) (string, error) {
	if path != nil {
		if _, err := os.Stat(*path); err != nil {
			return "", errors.Wrapf(err, "failed to read a config file on %s", *path)
		}

		return *path, nil
	}

	wd, err := os.Getwd()

	if err != nil {
		logger.Logger.Debug().Err(err).Msg("Cannot load the current working directory")
		return "", nil
	}

	for _, name := range configNameCandidates {
		candidate := filepath.Join(wd, name)

		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	logger.Logger.Debug().Msgf("No config file is found on the current directory: %s", wd)

	return "", nil
}

func (c *GlobalConfig) configure() error {
	if c.deployments == nil {
		c.deployments = map[string]Deployment{}
	}

	if c.services == nil {
		c.services = map[string]CustomServiceDefinition{}
	}

	if c.rawConfig.FormatStyle == "" {
		c.rawConfig.FormatStyle = DefaultFormat
	}

	if c.rawConfig.NetworkTimeout == "" {
		c.rawConfig.NetworkTimeout = DefaultNetworkTimeout
	}

	if c.rawConfig.WaitTimeout == "" {
		c.rawConfig.WaitTimeout = DefaultWaitTimeout
	}

	for name, values := range c.rawConfig.Services {
		logger.Logger.Debug().Msgf("Configuring the service of %s", name)

		values, correct := values.(map[string]interface{})

		if !correct {
			return errors.New(fmt.Sprintf("%s must be Mapping", name))
		}

		if slices.Contains([]string{DeploygateService, FirebaseAppDistributionService, LocalService, TestFlightService}, name) {
			return errors.New(fmt.Sprintf("%s is a reserved name", name))
		}

		var definition CustomServiceDefinition

		if bytes, err := yaml.Marshal(values); err != nil {
			return errors.Wrapf(err, "cannot load %s service definition", name)
		} else if err := yaml.Unmarshal(bytes, &definition); err != nil {
			return errors.Wrapf(err, "cannot load %s service definition", name)
		} else if err := definition.validate(); err != nil {
			return errors.Wrapf(err, "%s service definition is invalid", name)
		}

		c.services[name] = definition
	}

	for name, values := range c.rawConfig.Deployments {
		logger.Logger.Debug().Msgf("Configuring the deployment of %s", name)

		values, correct := values.(map[string]interface{})

		if !correct {
			return errors.New(fmt.Sprintf("%s must be Mapping", name))
		}

		var service serviceNameHolder

		if bytes, err := yaml.Marshal(values); err != nil {
			return errors.Wrapf(err, "cannot load %s config", name)
		} else if err := yaml.Unmarshal(bytes, &service); err != nil {
			return errors.Wrapf(err, "cannot load %s config", name)
		}

		switch service.Name {
		case DeploygateService:
			deploygate := DeployGateConfig{}

			if err := loadServiceConfig(&deploygate, values); err != nil {
				return errors.Wrapf(err, "cannot load %s config", name)
			}

			c.deployments[name] = Deployment{
				ServiceName:   deploygate.Name,
				ServiceConfig: deploygate,
				Lifecycle:     deploygate.ExecutionConfig,
			}
		case FirebaseAppDistributionService:
			firebase := FirebaseAppDistributionConfig{}

			if err := loadServiceConfig(&firebase, values); err != nil {
				return errors.Wrapf(err, "cannot load %s config", name)
			}

			c.deployments[name] = Deployment{
				ServiceName:   firebase.Name,
				ServiceConfig: firebase,
				Lifecycle:     firebase.ExecutionConfig,
			}
		case LocalService:
			local := LocalConfig{}

			if err := loadServiceConfig(&local, values); err != nil {
				return errors.Wrapf(err, "cannot load %s config", name)
			}

			c.deployments[name] = Deployment{
				ServiceName:   local.Name,
				ServiceConfig: local,
				Lifecycle:     local.ExecutionConfig,
			}
		case TestFlightService:
			testFlight := TestFlightConfig{}

			if err := loadServiceConfig(&testFlight, values); err != nil {
				return errors.Wrapf(err, "cannot load %s config", name)
			}

			c.deployments[name] = Deployment{
				ServiceName:   testFlight.Name,
				ServiceConfig: testFlight,
				Lifecycle:     testFlight.ExecutionConfig,
			}
		default:
			if _, ok := c.services[service.Name]; ok {
				logger.Logger.Debug().Msgf("%s is a custom service", service.Name)

				custom := CustomServiceConfig{}

				if err := loadServiceConfig(&custom, values); err != nil {
					return errors.Wrapf(err, "cannot load %s config", name)
				}

				c.deployments[name] = Deployment{
					ServiceName:   service.Name,
					ServiceConfig: custom,
					Lifecycle:     custom.ExecutionConfig,
				}
			} else {
				return errors.New(fmt.Sprintf("%s of %s is an unknown service", service.Name, name))
			}
		}
	}

	return c.Validate()
}

func (c *GlobalConfig) FormatStyle() string {
	return c.rawConfig.FormatStyle
}

// NetworkTimeout is a read/connection timeout for requests
func (c *GlobalConfig) NetworkTimeout() time.Duration {
	var value = DefaultNetworkTimeout

	if c.rawConfig.NetworkTimeout != "" {
		value = c.rawConfig.NetworkTimeout
	}

	timeout, _ := time.ParseDuration(value)

	return timeout
}

// WaitTimeout is a timeout for polling service's processing
func (c *GlobalConfig) WaitTimeout() time.Duration {
	var value = DefaultWaitTimeout

	if c.rawConfig.WaitTimeout != "" {
		value = c.rawConfig.WaitTimeout
	}

	timeout, _ := time.ParseDuration(value)

	return timeout
}

func (c *GlobalConfig) Validate() error {
	if c.rawConfig.FormatStyle != "" {
		if !slices.Contains(styles, c.rawConfig.FormatStyle) {
			return errors.New(fmt.Sprintf("%s is unknown format style", c.rawConfig.FormatStyle))
		}
	} else {
		return errors.New("empty format is invalid")
	}

	if c.rawConfig.NetworkTimeout != "" {
		if v, err := time.ParseDuration(c.rawConfig.NetworkTimeout); err != nil {
			return errors.Wrapf(err, "network timeout is not valid time format: %s", c.rawConfig.NetworkTimeout)
		} else if v <= 0 {
			return errors.New("network timeout must be positive")
		} else if v.Minutes() > 30 {
			return errors.New("network timeout must be equal or less than 30 minutes")
		}
	} else {
		return errors.New("empty network timeout is invalid")
	}

	if c.rawConfig.WaitTimeout != "" {
		if v, err := time.ParseDuration(c.rawConfig.WaitTimeout); err != nil {
			return errors.Wrapf(err, "wait timeout is not valid time format: %s", c.rawConfig.WaitTimeout)
		} else if v <= 0 {
			return errors.New("wait timeout must be positive")
		} else if v.Minutes() > 10 {
			return errors.New("wait timeout must be equal or less than 10 minutes")
		}
	} else {
		return errors.New("empty wait timeout is invalid")
	}

	return nil
}

func (c *GlobalConfig) Dump(path string) error {
	if bytes, err := yaml.Marshal(c.rawConfig); err != nil {
		return errors.Wrapf(err, "failed to parse a config file to %s", path)
	} else if err := os.WriteFile(path, bytes, 0644); err != nil {
		return errors.Wrapf(err, "failed to dump a config file to %s", path)
	}

	return nil
}

func (c *GlobalConfig) Definition(name string) (CustomServiceDefinition, error) {
	if s, ok := c.services[name]; ok {
		return s, nil
	} else {
		return CustomServiceDefinition{}, errors.New(fmt.Sprintf("%s is not found in services", name))
	}
}

func (c *GlobalConfig) Deployment(name string) (Deployment, CustomServiceDefinition, error) {
	var definition CustomServiceDefinition

	d, ok := c.deployments[name]

	if !ok {
		return Deployment{}, definition, errors.New(fmt.Sprintf("%s deployment is not found", name))
	}

	// configure stores service configs by value so the evaluated values must be written back.
	switch conf := d.ServiceConfig.(type) {
	case DeployGateConfig:
		if err := evaluateAndValidate(&conf); err != nil {
			return Deployment{}, definition, err
		}

		d.ServiceConfig = conf
	case FirebaseAppDistributionConfig:
		if err := evaluateAndValidate(&conf); err != nil {
			return Deployment{}, definition, err
		}

		d.ServiceConfig = conf
	case LocalConfig:
		if err := evaluateAndValidate(&conf); err != nil {
			return Deployment{}, definition, err
		}

		d.ServiceConfig = conf
	case TestFlightConfig:
		if err := evaluateAndValidate(&conf); err != nil {
			return Deployment{}, definition, err
		}

		d.ServiceConfig = conf
	case CustomServiceConfig:
		if err := evaluateAndValidate(&conf); err != nil {
			return Deployment{}, definition, err
		} else if v, err := c.Definition(conf.Name); err != nil {
			return Deployment{}, definition, err
		} else {
			definition = v
		}

		d.ServiceConfig = conf
	default:
		return Deployment{}, definition, errors.New(fmt.Sprintf("%s deployment has an unknown service config", name))
	}

	return d, definition, nil
}

func (c *GlobalConfig) AddDeployment(name string, serviceName string) error {
	if d, ok := c.deployments[name]; ok {
		return errors.New(fmt.Sprintf("%s (service = %s) already exists in the config.", name, d.ServiceName))
	} else {
		d = Deployment{
			ServiceName: serviceName,
		}

		switch serviceName {
		case DeploygateService:
			d.ServiceConfig = DeployGateConfig{
				serviceNameHolder: serviceNameHolder{
					Name: DeploygateService,
				},
				AppOwnerName: "DeployGate's user name or group name",
				ApiToken:     fmt.Sprintf("format:${%s_DEPLOYGATE_API_TOKEN}", name),
			}
		case FirebaseAppDistributionService:
			d.ServiceConfig = FirebaseAppDistributionConfig{
				serviceNameHolder: serviceNameHolder{
					Name: FirebaseAppDistributionService,
				},
				AppId:                 "App ID e.g. 1:123456789:android:xxxxx",
				GoogleCredentialsPath: "path to Google Credentials JSON",
			}
		case LocalService:
			d.ServiceConfig = LocalConfig{
				serviceNameHolder: serviceNameHolder{
					Name: LocalService,
				},
				DestinationPath: "path to the destination",
			}
		case TestFlightService:
			d.ServiceConfig = TestFlightConfig{
				serviceNameHolder: serviceNameHolder{
					Name: TestFlightService,
				},
				AppleID:  "Your AppleID",
				ApiKey:   fmt.Sprintf("format:${%s_TESTFLIGHT_API_KEY}", name),
				IssuerID: "Issuer ID of ApiKey. You can use app-specific password instead of api key and issuer id",
			}
		default:
			return errors.New(fmt.Sprintf("%s is an unknown service. %s are supported", serviceName, strings.Join([]string{DeploygateService, FirebaseAppDistributionService, LocalService, TestFlightService}, ", ")))
		}

		if c.rawConfig.Deployments == nil {
			c.rawConfig.Deployments = map[string]interface{}{}
		}

		var values map[string]interface{}

		if bytes, err := yaml.Marshal(d.ServiceConfig); err != nil {
			panic(err)
		} else if err := yaml.Unmarshal(bytes, &values); err != nil {
			panic(err)
		}

		c.rawConfig.Deployments[name] = values

		return c.configure()
	}
}

func evaluateAndValidate(v any) error {
	if err := evaluateValues(v); err != nil {
		return err
	} else if err := validateMissingValues(v); err != nil {
		return err
	}

	return nil
}
