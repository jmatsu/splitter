package config

import (
	"fmt"
	"github.com/jmatsu/splitter/internal/logger"
	"github.com/pkg/errors"
	"reflect"
	"strings"
)

// Collect errors via the reflection. Target fields must have yaml tag. Required fields must not be zero values.
func validateMissingValues(v any) error {
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
