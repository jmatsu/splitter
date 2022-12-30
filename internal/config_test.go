package internal

import (
	"encoding/json"
	"reflect"
	"testing"
)

type TestConfig struct {
	ValueParam           string  `json:"param1" env:"TEST_PARAM1"`
	PointerParam         *string `json:"param2" env:"TEST_PARAM2"`
	RequiredValueParam   string  `json:"param3" env:"TEST_PARAM3" required:"true"`
	RequiredPointerParam *string `json:"param4" env:"TEST_PARAM4" required:"true"`
}

func (c TestConfig) validate() error {
	return validateMissingValues(&c)
}

func Test_validateMissingValues(t *testing.T) {
	t.Parallel()

	sampleValue1 := "Sample1"
	sampleValue2 := "Sample2"

	cases := map[string]struct {
		config            TestConfig
		expectedValidness bool
	}{
		"fully-filled": {
			config: TestConfig{
				ValueParam:           sampleValue1,
				PointerParam:         &sampleValue2,
				RequiredValueParam:   sampleValue1,
				RequiredPointerParam: &sampleValue2,
			},
			expectedValidness: true,
		},
		"pointer-filled": {
			config: TestConfig{
				PointerParam:         &sampleValue2,
				RequiredPointerParam: &sampleValue2,
			},
			expectedValidness: false,
		},
		"pointer-non-filled": {
			config: TestConfig{
				ValueParam:         sampleValue1,
				RequiredValueParam: sampleValue1,
			},
			expectedValidness: false,
		},
		"required-values-filled": {
			config: TestConfig{
				RequiredValueParam:   sampleValue1,
				RequiredPointerParam: &sampleValue2,
			},
			expectedValidness: true,
		},
		"non-required-values-filled": {
			config: TestConfig{
				ValueParam:   sampleValue1,
				PointerParam: &sampleValue2,
			},
			expectedValidness: false,
		},
		"zero": {
			config:            TestConfig{},
			expectedValidness: false,
		},
	}

	for name, c := range cases {
		name, c := name, c
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if err := validateMissingValues(&c.config); (err == nil) != c.expectedValidness {
				t.Errorf("%s case is expected to be %t but %t: %v", name, c.expectedValidness, err == nil, err)
			}
		})
	}
}

func Test_loadServiceConfig(t *testing.T) {
	pointerValue1 := "Sample1"
	pointerValue2 := "Sample2"

	cases := map[string]struct {
		values   map[string]interface{}
		envs     map[string]string
		expected *TestConfig
	}{
		"fully-written": {
			values: map[string]interface{}{
				"param1": "value1",
				"param2": pointerValue1,
				"param3": "value3",
				"param4": pointerValue2,
			},
			expected: &TestConfig{
				ValueParam:           "value1",
				PointerParam:         &pointerValue1,
				RequiredValueParam:   "value3",
				RequiredPointerParam: &pointerValue2,
			},
		},
		"fully-from-env": {
			envs: map[string]string{
				"TEST_PARAM1": "value1",
				"TEST_PARAM2": pointerValue1,
				"TEST_PARAM3": "value3",
				"TEST_PARAM4": pointerValue2,
			},
			expected: &TestConfig{
				ValueParam:           "value1",
				PointerParam:         &pointerValue1,
				RequiredValueParam:   "value3",
				RequiredPointerParam: &pointerValue2,
			},
		},
		"mixed-definitions": {
			values: map[string]interface{}{
				"param2": pointerValue1,
				"param3": "value3",
			},
			envs: map[string]string{
				"TEST_PARAM3": "env.value3",
				"TEST_PARAM4": pointerValue2,
			},
			expected: &TestConfig{
				PointerParam:         &pointerValue1,
				RequiredValueParam:   "env.value3",
				RequiredPointerParam: &pointerValue2,
			},
		},
		"zero": {
			expected: nil,
		},
	}

	deepEqual := func(lhs TestConfig, rhs TestConfig) bool {
		lbytes, _ := json.Marshal(&lhs)
		rbytes, _ := json.Marshal(&rhs)
		return reflect.DeepEqual(lbytes, rbytes)
	}

	for name, c := range cases {
		name, c := name, c
		t.Run(name, func(t *testing.T) {
			for name, value := range c.envs {
				t.Setenv(name, value)
			}

			actual := TestConfig{}

			if err := loadServiceConfig(&actual, c.values); err != nil && c.expected != nil {
				t.Errorf("%s case is expected to be success but %v", name, err)
			} else if c.expected != nil && !deepEqual(actual, *c.expected) {
				t.Errorf("%v does not equal to %v", actual, c.expected)
			}
		})
	}
}
