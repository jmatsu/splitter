package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/caarlos0/env/v6"
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigName("splitter")
	viper.SetConfigType("yml")

	viper.SetEnvPrefix("SPLITTER_")
}

const (
	servicesKey = "services"

	deploygateService = "deploygate"
)

type Config struct {
	rawConfig rawConfig
	services  map[string]any
}

type rawConfig struct {
	Services map[string]interface{}
}

type ServiceNameHolder struct {
	Service string `json:"service"`
}

var config Config

func GetConfig() Config {
	return config
}

func LoadConfig(path *string) error {
	if path != nil {
		viper.SetConfigFile(*path)
		Logger.Debug().Msgf("Loading a config file on %s", *path)
	} else {
		viper.AddConfigPath(".")

		if wd, err := os.Getwd(); err == nil {
			Logger.Debug().Msgf("Loading a config file on %s", wd)
		} else {
			Logger.Debug().Err(err).Msgf("Cannot loading the current working directory")
		}
	}

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("Failed to read a config file: %v", err)
	}

	config = Config{
		rawConfig: rawConfig{
			Services: viper.GetStringMap("services"),
		},
	}

	return nil
}

func (c *Config) getService(name string) (string, any, error) {
	scopes := []string{servicesKey}

	v := c.rawConfig.Services[name]

	if v == nil {
		return "", nil, fmt.Errorf("%s is required", strings.Join(append(scopes, name), "."))
	}

	holder := ServiceNameHolder{}

	if byte, err := json.Marshal(v); err != nil {
		return "", nil, fmt.Errorf("XYZ")
	} else if err := json.Unmarshal(byte, &holder); err != nil {
		return "", nil, err
	} else {
		return holder.Service, v, nil
	}
}

func (c *Config) configure() error {
	for name, values := range c.rawConfig.Services {
		values, correct := values.(map[string]interface{})

		if !correct {
			return fmt.Errorf("%s must be Mapping", name)
		}

		holder := ServiceNameHolder{}

		if byte, err := json.Marshal(values); err != nil {
			return fmt.Errorf("cannot load %s config: %v", name, err)
		} else if err := json.Unmarshal(byte, &holder); err != nil {
			return fmt.Errorf("cannot load %s config: %v", name, err)
		}

		if v, err := provideZeroService(holder.Service); err != nil {
			return fmt.Errorf("cannot load %s config: %v", name, err)
		} else if conf, err := loadServiceConfig(v, values); err != nil {
			return fmt.Errorf("cannot load %s config: %v", name, err)
		} else {
			c.services[name] = conf
		}
	}

	return nil
}

func loadServiceConfig(v Validatable, values map[string]interface{}) (any, error) {
	// 1. Get a service config and read the name first
	// 2. Set values from environment variables
	// 3. Overwrite them by the config file
	// 4. Validate the values

	if err := env.Parse(&v); err != nil {
		panic(err)
	} else if byte, err := json.Marshal(values); err != nil {
		panic(err)
	} else if err := json.Unmarshal(byte, &v); err != nil {
		panic(err)
	}

	if err := v.validate(); err != nil {
		return nil, err
	} else {
		return v, nil
	}
}

type Validatable interface {
	validate() error
}

func validateMissingValues(v any) error {
	var missingKeys []string

	entity := reflect.ValueOf(v).Elem()

	for i := 0; i < entity.NumField(); i++ {
		value := entity.Field(i)
		field := entity.Type().Field(i)
		tag := entity.Type().Field(i).Tag

		Logger.Debug().Msgf("%v = %s: json:\"%s\"", field.Name, value, tag.Get("json"))

		if t, found := tag.Lookup("json"); found {
			key := strings.SplitN(t, ",", 2)[0]

			b, found := tag.Lookup("required")

			if found && b == "true" && value.IsZero() {
				Logger.Error().Msgf("%s is required but not assigned", key)
				missingKeys = append(missingKeys, key)
			} else {
				Logger.Debug().Msgf("%s is set", key)
			}
		}
	}

	if num := len(missingKeys); num > 0 {
		return fmt.Errorf("%d keys lacked or their values are empty", num)
	} else {
		return nil
	}
}

func provideZeroService(name string) (Validatable, error) {
	switch name {
	case deploygateService:
		return DeployGateConfig{}, nil
	default:
		return nil, fmt.Errorf("%s is an unknown service", name)
	}
}
