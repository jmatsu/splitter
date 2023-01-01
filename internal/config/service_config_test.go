package config

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"reflect"
	"testing"
)

func (c testConfig) assertEquals(other testConfig) error {
	lbytes, _ := yaml.Marshal(&c)
	rbytes, _ := yaml.Marshal(&other)
	if reflect.DeepEqual(lbytes, rbytes) {
		return nil
	} else {
		return errors.New(fmt.Sprintf("%v does not equal to %v", string(lbytes), string(rbytes)))
	}
}

func Test_evaluateValues(t *testing.T) {
	sampleValue1 := "Sample1"
	sampleValue2 := "Sample2"

	cases := map[string]struct {
		config   testConfig
		envs     map[string]string
		expected testConfig
	}{
		"no expansion": {
			config: testConfig{
				ValueParam:           sampleValue1,
				PointerParam:         &sampleValue2,
				RequiredValueParam:   sampleValue1,
				RequiredPointerParam: &sampleValue2,
			},
			expected: testConfig{
				ValueParam:           sampleValue1,
				PointerParam:         &sampleValue2,
				RequiredValueParam:   sampleValue1,
				RequiredPointerParam: &sampleValue2,
			},
		},
		"with format and values": {
			config: testConfig{
				ValueParam:           "format:${FROM_ENV_VALUE1}",
				PointerParam:         &sampleValue2,
				RequiredValueParam:   "format:${FROM_ENV_VALUE2}",
				RequiredPointerParam: &sampleValue2,
			},
			envs: map[string]string{
				"FROM_ENV_VALUE1": sampleValue1,
			},
			expected: testConfig{
				ValueParam:           sampleValue1,
				PointerParam:         &sampleValue2,
				RequiredValueParam:   "",
				RequiredPointerParam: &sampleValue2,
			},
		},
	}

	for name, c := range cases {
		name, c := name, c
		t.Run(name, func(t *testing.T) {
			if c.envs != nil {
				for name, value := range c.envs {
					t.Setenv(name, value)
				}
			}

			if err := evaluateValues(&c.config); err != nil {
				t.Errorf("%s case is expected to be success but not: %v", name, err)
			} else if err := c.config.assertEquals(c.expected); err != nil {
				t.Errorf("%v is expected to be equal to %v but not: %v", c.config, c.expected, err)
			}
		})
	}
}

func Test_validateMissingValues(t *testing.T) {
	t.Parallel()

	sampleValue1 := "Sample1"
	sampleValue2 := "Sample2"

	cases := map[string]struct {
		config            testConfig
		expectedValidness bool
	}{
		"fully-filled": {
			config: testConfig{
				ValueParam:           sampleValue1,
				PointerParam:         &sampleValue2,
				RequiredValueParam:   sampleValue1,
				RequiredPointerParam: &sampleValue2,
			},
			expectedValidness: true,
		},
		"pointer-filled": {
			config: testConfig{
				PointerParam:         &sampleValue2,
				RequiredPointerParam: &sampleValue2,
			},
			expectedValidness: false,
		},
		"pointer-non-filled": {
			config: testConfig{
				ValueParam:         sampleValue1,
				RequiredValueParam: sampleValue1,
			},
			expectedValidness: false,
		},
		"required-values-filled": {
			config: testConfig{
				RequiredValueParam:   sampleValue1,
				RequiredPointerParam: &sampleValue2,
			},
			expectedValidness: true,
		},
		"non-required-values-filled": {
			config: testConfig{
				ValueParam:   sampleValue1,
				PointerParam: &sampleValue2,
			},
			expectedValidness: false,
		},
		"zero": {
			config:            testConfig{},
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
		expected *testConfig
	}{
		"fully-written": {
			values: map[string]interface{}{
				"param1": "value1",
				"param2": pointerValue1,
				"param3": "value3",
				"param4": pointerValue2,
			},
			expected: &testConfig{
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
			expected: &testConfig{
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
			expected: &testConfig{
				PointerParam:         &pointerValue1,
				RequiredValueParam:   "env.value3",
				RequiredPointerParam: &pointerValue2,
			},
		},
		"zero": {
			expected: nil,
		},
	}

	for name, c := range cases {
		name, c := name, c
		t.Run(name, func(t *testing.T) {
			for name, value := range c.envs {
				t.Setenv(name, value)
			}

			actual := testConfig{}

			err := loadServiceConfig(&actual, c.values)

			if c.expected == nil && err != nil {
				return
			}

			if c.expected != nil && err == nil {
				if err := actual.assertEquals(*c.expected); err != nil {
					t.Errorf("%v", err)
				}

				return
			}

			if err != nil {
				t.Errorf("%s case is expected to be success but not: %v", name, err)
			}
		})
	}
}
