package config

import (
	"encoding/json"
	"fmt"
	"github.com/jmatsu/splitter/internal/logger"
	"gopkg.in/yaml.v3"
	"os"
	"reflect"
	"strings"

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
	envPrefix = "SPLITTER_"

	DefaultConfigName = "splitter.yml"

	distributionsKey = "distributions"

	DeploygateService = "deploygate"
	LocalService      = "local"
)

func ToEnvName(name string) string {
	return fmt.Sprintf("%s%s", envPrefix, strings.ToUpper(name))
}

type ServiceConfig interface {
	testConfig | DeployGateConfig | LocalConfig
}

type Config struct {
	rawConfig rawConfig
	services  map[string]*Distribution
}

type rawConfig struct {
	Distributions map[string]interface{} `yaml:"distributions"`
	FormatStyle   string                 `yaml:"format-style,omitempty"`
}

type serviceNameHolder struct {
	ServiceName string `json:"service"`
}

type Distribution struct {
	ServiceName   string
	ServiceConfig any
}

var config Config

func NewConfig() Config {
	return Config{}
}

func GetConfig() Config {
	return config
}

func LoadConfig(path *string) error {
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
		return fmt.Errorf("failed to read a config file: %v", err)
	}

	config = Config{
		rawConfig: rawConfig{
			Distributions: viper.GetStringMap(distributionsKey),
			FormatStyle:   viper.GetString("format-style"),
		},
	}

	if err := config.configure(); err != nil {
		return fmt.Errorf("your config file may not contain some of required values or they are invalid: %v", err)
	}

	return nil
}

func (c *Config) configure() error {
	for name, values := range c.rawConfig.Distributions {
		logger.Logger.Debug().Msgf("Configuring %s", name)

		values, correct := values.(map[string]interface{})

		if !correct {
			return fmt.Errorf("%s must be Mapping", name)
		}

		holder := serviceNameHolder{}

		if bytes, err := json.Marshal(values); err != nil {
			return fmt.Errorf("cannot load %s config: %v", name, err)
		} else if err := json.Unmarshal(bytes, &holder); err != nil {
			return fmt.Errorf("cannot load %s config: %v", name, err)
		}

		switch holder.ServiceName {
		case DeploygateService:
			deploygate := DeployGateConfig{}

			if err := loadServiceConfig(&deploygate, values); err != nil {
				return fmt.Errorf("cannot load %s config: %v", name, err)
			}

			if c.services == nil {
				c.services = map[string]*Distribution{}
			}

			c.services[name] = &Distribution{
				ServiceName:   DeploygateService,
				ServiceConfig: &deploygate,
			}
		default:
			return fmt.Errorf("%s of %s is an unknown service", holder.ServiceName, name)
		}
	}

	return nil
}

func (c *Config) FormatStyle() string {
	return c.rawConfig.FormatStyle
}

func (c *Config) SetFormatStyle(style string) {
	c.rawConfig.FormatStyle = style
}

func (c *Config) Dump(path string) error {
	if bytes, err := yaml.Marshal(c.rawConfig); err != nil {
		return fmt.Errorf("failed to parse a config file to %s: %v", path, err)
	} else if err := os.WriteFile(path, bytes, 0644); err != nil {
		return fmt.Errorf("failed to dump a config file to %s: %v", path, err)
	}

	return nil
}

func (c *Config) GetDistribution(name string) (*Distribution, error) {
	if d := c.services[name]; d != nil {
		switch d.ServiceName {
		case DeploygateService:
			config := d.ServiceConfig.(*DeployGateConfig)

			if err := evaluateAndValidate(config); err != nil {
				return nil, err
			}
		}

		return d, nil
	} else {
		return nil, fmt.Errorf("%s distribution is not found", name)
	}
}

func loadServiceConfig[T ServiceConfig](v *T, values map[string]interface{}) error {
	// 1. Get a service config and read the name first
	// 2. Set values from the config file
	// 3. Overwrite them by the environment variables

	if bytes, err := json.Marshal(values); err != nil {
		panic(err)
	} else if err := json.Unmarshal(bytes, v); err != nil {
		panic(err)
	} else if err := env.Parse(v); err != nil {
		panic(err)
	}

	return nil
}

func evaluateAndValidate[T ServiceConfig](v *T) error {
	if err := evaluateValues(v); err != nil {
		return err
	} else if err := validateMissingValues(v); err != nil {
		return err
	}

	return nil
}

func evaluateValues[T ServiceConfig](v *T) error {
	vRef := reflect.ValueOf(v).Elem()

	if vRef.Kind() != reflect.Struct {
		return fmt.Errorf("%v is not a struct", v)
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

func validateMissingValues[T ServiceConfig](v *T) error {
	var missingKeys []string

	vRef := reflect.ValueOf(v).Elem()

	if vRef.Kind() != reflect.Struct {
		return fmt.Errorf("%v is not a struct", v)
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
		return fmt.Errorf("%d keys lacked or their values are empty: %s", num, strings.Join(missingKeys, ","))
	} else {
		return nil
	}
}
