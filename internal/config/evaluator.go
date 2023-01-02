package config

import (
	"fmt"
	"github.com/jmatsu/splitter/internal/logger"
	"github.com/pkg/errors"
	"os"
	"reflect"
	"strings"
)

// Evaluate the styled format for the embedded variables.
func evaluateValues(v any) error {
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
