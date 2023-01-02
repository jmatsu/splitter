package config

import (
	"fmt"
	"github.com/jmatsu/splitter/internal/logger"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

func init() {
	baseName, extName, _ := strings.Cut(DefaultConfigName, ".")

	viper.SetConfigName(baseName)
	viper.SetConfigType(extName)

	viper.SetEnvPrefix(envPrefix)
}

const (
	envPrefix = "SPLITTER_" // The prefix of environment variables used for splitter's global options.

	DefaultConfigName = "splitter.yml"

	deploymentsKey        = "deployments" // deployment definitions' key in the config file.
	serviceDefinitionsKey = "services"    // service definitions' key in the config file.

	DeploygateService              = "deploygate"                // represents DeployGateConfig
	LocalService                   = "local"                     // represents LocalConfig
	FirebaseAppDistributionService = "firebase-app-distribution" // represents FirebaseAppDistributionConfig
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
	config := config // create a shallow copy
	return config
}

func LoadGlobalConfig(path *string) error {
	if path != nil {
		viper.SetConfigFile(*path)
		logger.Logger.Debug().Msgf("Loading a config file on %s", *path)
	} else {
		viper.AddConfigPath(".")

		if wd, err := os.Getwd(); err == nil {
			logger.Logger.Debug().Msgf("Loading a config file on the current directory: %s", wd)
		} else {
			logger.Logger.Debug().Err(err).Msgf("Cannot loading the current working directory")
		}
	}

	if err := viper.ReadInConfig(); path != nil && err != nil {
		return errors.Wrap(err, "failed to read a config file")
	}

	config.rawConfig = rawConfig{
		Deployments:    viper.GetStringMap(deploymentsKey),
		Services:       viper.GetStringMap(serviceDefinitionsKey),
		FormatStyle:    viper.GetString("format-style"),
		WaitTimeout:    viper.GetString("wait-timeout"),
		NetworkTimeout: viper.GetString("network-timeout"),
	}

	if err := config.configure(); err != nil {
		return errors.Wrap(err, "your config file may not contain some of required values or they are invalid")
	}

	return nil
}

func (c *GlobalConfig) configure() error {
	if c.deployments == nil {
		c.deployments = map[string]Deployment{}
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

		if slices.Contains([]string{DeploygateService, FirebaseAppDistributionService, LocalService}, name) {
			return errors.New(fmt.Sprintf("%s is a reserved name", name))
		}

		var definition CustomServiceDefinition

		if bytes, err := yaml.Marshal(values); err != nil {
			return errors.Wrapf(err, "cannot load %s service definition", name)
		} else if err := yaml.Unmarshal(bytes, &definition); err != nil {
			return errors.Wrapf(err, "cannot load %s service definition", name)
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
		default:
			if _, ok := c.services[name]; ok {
				logger.Logger.Debug().Msgf("%s is a custom service", name)

				custom := CustomServiceConfig{}

				if err := loadServiceConfig(&custom, values); err != nil {
					return errors.Wrapf(err, "cannot load %s config", name)
				}

				c.deployments[name] = Deployment{
					ServiceName:   name,
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
		} else if v < 0 {
			return errors.Wrapf(err, "network timeout must be positive")
		} else if v.Minutes() > 30 {
			return errors.Wrapf(err, "network timeout must be equal or less than 30 minutes")
		}
	} else {
		return errors.New("empty network timeout is invalid")
	}

	if c.rawConfig.WaitTimeout != "" {
		if v, err := time.ParseDuration(c.rawConfig.WaitTimeout); err != nil {
			return errors.Wrapf(err, "wait timeout is not valid time format: %s", c.rawConfig.WaitTimeout)
		} else if v < 0 {
			return errors.Wrapf(err, "wait timeout must be positive")
		} else if v.Minutes() > 10 {
			return errors.Wrapf(err, "wait timeout must be equal or less than 10 minutes")
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

func (c *GlobalConfig) Deployment(name string) (Deployment, error) {
	if d, ok := c.deployments[name]; ok {
		switch d.ServiceName {
		case DeploygateService:
			config := d.ServiceConfig.(*DeployGateConfig)

			if err := evaluateAndValidate(config); err != nil {
				return Deployment{}, err
			}
		case FirebaseAppDistributionService:
			config := d.ServiceConfig.(*FirebaseAppDistributionConfig)

			if err := evaluateAndValidate(config); err != nil {
				return Deployment{}, err
			}
		case LocalService:
			config := d.ServiceConfig.(*LocalConfig)

			if err := evaluateAndValidate(config); err != nil {
				return Deployment{}, err
			}
		}

		return d, nil
	} else {
		return Deployment{}, errors.New(fmt.Sprintf("%s deployment is not found", name))
	}
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
