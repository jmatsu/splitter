package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/spf13/viper"
	"github.com/caarlos0/env/v6"
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
		Logger.Debug().Msgf("Loading a config file on %s", path)
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

func getValueFromEnv(name string) *string {
		if v, exist := os.LookupEnv(name); exist {
			return &v
		} else {
			return nil
	}
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

func (c *Config) readServiceConfig(name string) (any, error) {
	// 1. Get a service config and read the name first
	// 2. Set values from environment variables
	// 3. Overwrite them by the config file
	// 4. Validate the values

	serviceName, service, err := c.getService(name)

	if err != nil {
		return nil, err
	}

	v := (func(name string) any {
		switch name {
		case deploygateService:
			return DeployGateConfig{}
		}

		return nil
	})(serviceName)

	if v == nil {
		return nil, fmt.Errorf("%s is an unknown service", serviceName)
	}

	if err := env.Parse(&v); err != nil {
		return nil, fmt.Errorf("Failed to load environment variables for %s: %v", name, err)
	} else if byte, err := json.Marshal(service); err != nil {
		return nil, fmt.Errorf("Failed to marshel your %s config: %v", name, err)
	} else if err := json.Unmarshal(byte, &v); err != nil {
		return nil, fmt.Errorf("Failed to load your %s config: %v", name, err)
	} else if err := validateServiceConfig(&v); err != nil {
		return nil, err
	}

	return v, nil
}

func validateServiceConfig(v any) error {
	missingKeys := []string{}

	property := reflect.ValueOf(v).Elem()

	for i := 0; i < property.NumField(); i++ {
		value := property.Field(i)
		field := property.Type().Field(i)
		tag := property.Type().Field(i).Tag

		Logger.Debug().Msgf("%v = %s: json:\"%s\"", field.Name, value, tag.Get("json"))

		if t, found := tag.Lookup("json"); found {
			key := strings.SplitN(t, ",", 2)[0]

			b, found := tag.Lookup("required")

			if value.IsNil()  {
				if b == "true" || found {
					Logger.Error().Msgf("%s is required but not found", key)
					missingKeys = append(missingKeys, key)
				} else {
					Logger.Debug().Msgf("%s is nil but ok", key)
				}
			} else {
				Logger.Debug().Msgf("%s is set", key)
			}
		}
	}

	if num := len(missingKeys); num > 0 {
		return fmt.Errorf("%d keys lacked", num)
	} else {
		return nil
	}
}