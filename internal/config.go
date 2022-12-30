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

type ServiceConfig interface {
	DeployGateConfig | testConfig
}

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
		return fmt.Errorf("failed to read a config file: %v", err)
	}

	config = Config{
		rawConfig: rawConfig{
			Services: viper.GetStringMap("services"),
		},
	}
	if err := config.configure(); err != nil {
		return fmt.Errorf("your config file may not contain some of required values or they are invalid: %v", err)
	}

	return nil
}

func (c *Config) configure() error {
	for name, values := range c.rawConfig.Services {
		values, correct := values.(map[string]interface{})

		if !correct {
			return fmt.Errorf("%s must be Mapping", name)
		}

		holder := ServiceNameHolder{}

		if bytes, err := json.Marshal(values); err != nil {
			return fmt.Errorf("cannot load %s config: %v", name, err)
		} else if err := json.Unmarshal(bytes, &holder); err != nil {
			return fmt.Errorf("cannot load %s config: %v", name, err)
		}

		switch holder.Service {
		case deploygateService:
			deploygate := DeployGateConfig{}

			if err := loadServiceConfig(&deploygate, values); err != nil {
				return fmt.Errorf("cannot load %s config: %v", name, err)
			}

			if c.services == nil {
				c.services = map[string]any{}
			}

			c.services[name] = deploygate
		default:
			return fmt.Errorf("%s is an unknown service", holder.Service)
		}
	}

	return nil
}

func loadServiceConfig[T ServiceConfig](v *T, values map[string]interface{}) error {
	// 1. Get a service config and read the name first
	// 2. Set values from the config file
	// 3. Overwrite them by the environment variables
	// 4. Validate the values

	if byte, err := json.Marshal(values); err != nil {
		panic(err)
	} else if err := json.Unmarshal(byte, v); err != nil {
		panic(err)
	} else if err := env.Parse(v); err != nil {
		panic(err)
	}

	return validateMissingValues(v)
}

func validateMissingValues(v any) error {
	var missingKeys []string

	vRef := reflect.ValueOf(v)

	if vRef.Kind() == reflect.Pointer {
		vRef = vRef.Elem()

		if vRef.Kind() != reflect.Struct {
			return fmt.Errorf("%v is not a struct", v)
		}
	} else if vRef.Kind() != reflect.Struct {
		return fmt.Errorf("%v is not a struct", v)
	}

	for i := 0; i < vRef.NumField(); i++ {
		value := vRef.Field(i)
		field := vRef.Type().Field(i)
		tag := vRef.Type().Field(i).Tag

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
