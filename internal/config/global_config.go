package config

import (
	"encoding/json"
	"fmt"
	"github.com/jmatsu/splitter/internal/logger"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
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

	distributionsKey = "distributions" // distribution definitions' key in the config file.

	DeploygateService              = "deploygate"                // represents DeployGateConfig
	LocalService                   = "local"                     // represents LocalConfig
	FirebaseAppDistributionService = "firebase-app-distribution" // represents FirebaseAppDistributionConfig
)

func ToEnvName(name string) string {
	return fmt.Sprintf("%s%s", envPrefix, strings.ToUpper(name))
}

type ServiceConfig interface {
	testConfig | DeployGateConfig | LocalConfig | FirebaseAppDistributionConfig
}

// GlobalConfig is a shared configuration in one command execution.
type GlobalConfig struct {
	rawConfig rawConfig
	services  map[string]*Distribution

	Async bool
}

type rawConfig struct {
	Distributions  map[string]interface{} `yaml:"distributions"`
	FormatStyle    string                 `yaml:"format-style,omitempty"`
	NetworkTimeout string                 `yaml:"network-timeout,omitempty"`
	WaitTimeout    string                 `yaml:"wait-timeout,omitempty"`
}

type serviceNameHolder struct {
	ServiceName string `json:"service"`
}

// Distribution holds a service name and its config struct
type Distribution struct {
	ServiceName   string
	ServiceConfig any // See ServiceConfig interface
	Lifecycle     *ExecutionConfig
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

func SetGlobalAsync(async bool) {
	config.Async = async
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
		Distributions:  viper.GetStringMap(distributionsKey),
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
	if c.services == nil {
		c.services = map[string]*Distribution{}
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

	for name, values := range c.rawConfig.Distributions {
		logger.Logger.Debug().Msgf("Configuring %s", name)

		values, correct := values.(map[string]interface{})

		if !correct {
			return errors.New(fmt.Sprintf("%s must be Mapping", name))
		}

		holder := serviceNameHolder{}

		if bytes, err := json.Marshal(values); err != nil {
			return errors.Wrapf(err, "cannot load %s config", name)
		} else if err := json.Unmarshal(bytes, &holder); err != nil {
			return errors.Wrapf(err, "cannot load %s config", name)
		}

		switch holder.ServiceName {
		case DeploygateService:
			deploygate := DeployGateConfig{}

			if err := loadServiceConfig(&deploygate, values); err != nil {
				return errors.Wrapf(err, "cannot load %s config", name)
			}

			c.services[name] = &Distribution{
				ServiceName:   holder.ServiceName,
				ServiceConfig: &deploygate,
				Lifecycle:     &deploygate.ExecutionConfig,
			}
		case FirebaseAppDistributionService:
			firebase := FirebaseAppDistributionConfig{}

			if err := loadServiceConfig(&firebase, values); err != nil {
				return errors.Wrapf(err, "cannot load %s config", name)
			}

			c.services[name] = &Distribution{
				ServiceName:   holder.ServiceName,
				ServiceConfig: &firebase,
				Lifecycle:     &firebase.ExecutionConfig,
			}
		case LocalService:
			local := LocalConfig{}

			if err := loadServiceConfig(&local, values); err != nil {
				return errors.Wrapf(err, "cannot load %s config", name)
			}

			c.services[name] = &Distribution{
				ServiceName:   holder.ServiceName,
				ServiceConfig: &local,
				Lifecycle:     &local.ExecutionConfig,
			}
		default:
			return errors.New(fmt.Sprintf("%s of %s is an unknown service", holder.ServiceName, name))
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

func (c *GlobalConfig) Distribution(name string) (*Distribution, error) {
	if d := c.services[name]; d != nil {
		switch d.ServiceName {
		case DeploygateService:
			config := d.ServiceConfig.(*DeployGateConfig)

			if err := evaluateAndValidate(config); err != nil {
				return nil, err
			}
		case FirebaseAppDistributionService:
			config := d.ServiceConfig.(*FirebaseAppDistributionConfig)

			if err := evaluateAndValidate(config); err != nil {
				return nil, err
			}
		case LocalService:
			config := d.ServiceConfig.(*LocalConfig)

			if err := evaluateAndValidate(config); err != nil {
				return nil, err
			}
		}

		return d, nil
	} else {
		return nil, errors.New(fmt.Sprintf("%s distribution is not found", name))
	}
}

func evaluateAndValidate[T ServiceConfig](v *T) error {
	if err := evaluateValues(v); err != nil {
		return err
	} else if err := validateMissingValues(v); err != nil {
		return err
	}

	return nil
}

// Set values to the config. Priority: Environment Variables > Given values
func loadServiceConfig[T ServiceConfig](v *T, values map[string]interface{}) error {
	if bytes, err := json.Marshal(values); err != nil {
		panic(err)
	} else if err := json.Unmarshal(bytes, v); err != nil {
		panic(err)
	} else if err := env.Parse(v); err != nil {
		panic(err)
	}

	return nil
}

// Evaluate the styled format for the embedded variables.
func evaluateValues[T ServiceConfig](v *T) error {
	vRef := reflect.ValueOf(v).Elem()

	if vRef.Kind() != reflect.Struct {
		return errors.New(fmt.Sprintf("%v is not a struct", v))
	}

	for i := 0; i < vRef.NumField(); i++ {
		value := vRef.Field(i)
		field := vRef.Type().Field(i)
		tag := vRef.Type().Field(i).Tag

		if _, found := tag.Lookup("json"); !found {
			continue
		}

		if value.Kind() == reflect.String {
			if prefix, format, ok := strings.Cut(value.String(), ":"); ok && prefix == "format" {
				newValue := os.ExpandEnv(format)
				value.SetString(newValue)

				logger.Logger.Debug().Msgf("%s = %v: is evaluated", field.Name, newValue)
			} else {
				logger.Logger.Debug().Msgf("%s = %v: needn't be evaluated", field.Name, value)
			}
		}
	}

	return nil
}

// Collect errors via the reflection. Target fields must have json tag. Required fields must not be zero values.
func validateMissingValues[T ServiceConfig](v *T) error {
	var missingKeys []string

	vRef := reflect.ValueOf(v).Elem()

	if vRef.Kind() != reflect.Struct {
		return errors.New(fmt.Sprintf("%v is not a struct", v))
	}

	for i := 0; i < vRef.NumField(); i++ {
		value := vRef.Field(i)
		field := vRef.Type().Field(i)
		tag := vRef.Type().Field(i).Tag

		t, found := tag.Lookup("json")

		if !found {
			logger.Logger.Debug().Msgf("%v is ignored", field.Name)
			continue
		}

		logger.Logger.Debug().Msgf("%s = %v: json:\"%s\"", field.Name, value, t)

		key, _, _ := strings.Cut(t, ",")

		if b, found := tag.Lookup("required"); found && b == "true" {
			if value.IsZero() {
				logger.Logger.Error().Msgf("%s is required but not assigned", key)
				missingKeys = append(missingKeys, key)
			} else {
				logger.Logger.Debug().Msgf("%s is set", key)
			}
		} else {
			logger.Logger.Debug().Msgf("%s is optional", key)
		}
	}

	if num := len(missingKeys); num > 0 {
		return errors.New(fmt.Sprintf("%d keys lacked or their values are empty: %s", num, strings.Join(missingKeys, ",")))
	} else {
		return nil
	}
}
