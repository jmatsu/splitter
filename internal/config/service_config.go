package config

import (
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/jmatsu/splitter/internal/logger"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"os"
	"reflect"
	"strings"
)

type serviceNameHolder struct {
	Name string `yaml:"service"`
}

type serviceConfig interface {
	testConfig | DeployGateConfig | LocalConfig | FirebaseAppDistributionConfig
}

func evaluateAndValidate[T serviceConfig](v *T) error {
	if err := evaluateValues(v); err != nil {
		return err
	} else if err := validateMissingValues(v); err != nil {
		return err
	}

	return nil
}

// Set values to the config. Priority: Environment Variables > Given values
func loadServiceConfig[T serviceConfig](v *T, values map[string]interface{}) error {
	if bytes, err := yaml.Marshal(values); err != nil {
		panic(err)
	} else if err := yaml.Unmarshal(bytes, v); err != nil {
		panic(err)
	} else if err := env.Parse(v); err != nil {
		panic(err)
	}

	return nil
}

// Evaluate the styled format for the embedded variables.
func evaluateValues[T serviceConfig](v *T) error {
	vRef := reflect.ValueOf(v).Elem()

	if vRef.Kind() != reflect.Struct {
		return errors.New(fmt.Sprintf("%v is not a struct", v))
	}

	for i := 0; i < vRef.NumField(); i++ {
		value := vRef.Field(i)
		field := vRef.Type().Field(i)
		tag := vRef.Type().Field(i).Tag

		if _, found := tag.Lookup("yaml"); !found {
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

// Collect errors via the reflection. Target fields must have yaml tag. Required fields must not be zero values.
func validateMissingValues[T serviceConfig](v *T) error {
	var missingKeys []string

	vRef := reflect.ValueOf(v).Elem()

	if vRef.Kind() != reflect.Struct {
		return errors.New(fmt.Sprintf("%v is not a struct", v))
	}

	for i := 0; i < vRef.NumField(); i++ {
		value := vRef.Field(i)
		field := vRef.Type().Field(i)
		tag := vRef.Type().Field(i).Tag

		t, found := tag.Lookup("yaml")

		if !found {
			logger.Logger.Debug().Msgf("%v is ignored", field.Name)
			continue
		}

		logger.Logger.Debug().Msgf("%s = %v: yaml:\"%s\"", field.Name, value, t)

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
