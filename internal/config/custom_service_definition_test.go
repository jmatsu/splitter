package config

import "testing"

func Test_CustomAuthDefinition_validate(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		definition        CustomAuthDefinition
		expectedValidness bool
	}{
		"header token": {
			definition: CustomAuthDefinition{
				StyleFormat: HeadersAssignFormatPrefix + "test",
				ValueFormat: "%s",
			},
			expectedValidness: false,
		},
		"query token": {
			definition: CustomAuthDefinition{
				StyleFormat: QueryAssignFormatPrefix + "test",
				ValueFormat: "%s",
			},
			expectedValidness: false,
		},
		"form token": {
			definition: CustomAuthDefinition{
				StyleFormat: FormParamsAssignFormatPrefix + "test",
				ValueFormat: "%s",
			},
			expectedValidness: false,
		},
		"style format is a prefix only": {
			definition: CustomAuthDefinition{
				StyleFormat: FormParamsAssignFormatPrefix,
				ValueFormat: "%s",
			},
			expectedValidness: false,
		},
		"unknown style format": {
			definition: CustomAuthDefinition{
				StyleFormat: "obababa",
				ValueFormat: "%s",
			},
			expectedValidness: false,
		},
		"too many %s in value format": {
			definition: CustomAuthDefinition{
				StyleFormat: FormParamsAssignFormatPrefix + "test",
				ValueFormat: "%s %s",
			},
			expectedValidness: false,
		},
		"missing %s in value format": {
			definition: CustomAuthDefinition{
				StyleFormat: FormParamsAssignFormatPrefix + "test",
				ValueFormat: "hello",
			},
			expectedValidness: false,
		},
		"zero": {definition: CustomAuthDefinition{}, expectedValidness: false},
	}

	for name, c := range cases {
		name, c := name, c
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if err := c.definition.validate(); (err == nil) != c.expectedValidness {
				t.Errorf("%s case is expected to be %t but %t", name, c.expectedValidness, err == nil)
			}
		})
	}
}

func Test_CustomAuthDefinition_AuthType(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		definition     CustomAuthDefinition
		expectedPrefix string
		expectedValue  string
	}{
		"header token": {
			definition: CustomAuthDefinition{
				StyleFormat: HeadersAssignFormatPrefix + "test",
				ValueFormat: "%s",
			},
			expectedPrefix: HeadersAssignFormatPrefix,
			expectedValue:  "test",
		},
		"query token": {
			definition: CustomAuthDefinition{
				StyleFormat: QueryAssignFormatPrefix + "test",
				ValueFormat: "%s",
			},
			expectedPrefix: QueryAssignFormatPrefix,
			expectedValue:  "test",
		},
		"form token": {
			definition: CustomAuthDefinition{
				StyleFormat: FormParamsAssignFormatPrefix + "test",
				ValueFormat: "%s",
			},
			expectedPrefix: FormParamsAssignFormatPrefix,
			expectedValue:  "test",
		},
		"style format is a prefix only": {
			definition: CustomAuthDefinition{
				StyleFormat: FormParamsAssignFormatPrefix,
				ValueFormat: "%s",
			},
		},
		"unknown style format": {
			definition: CustomAuthDefinition{
				StyleFormat: "obababa",
				ValueFormat: "%s",
			},
		},
		"too many %s in value format": {
			definition: CustomAuthDefinition{
				StyleFormat: FormParamsAssignFormatPrefix + "test",
				ValueFormat: "%s %s",
			},
		},
		"missing %s in value format": {
			definition: CustomAuthDefinition{
				StyleFormat: FormParamsAssignFormatPrefix + "test",
				ValueFormat: "hello",
			},
		},
		"zero": {definition: CustomAuthDefinition{}},
	}

	for name, c := range cases {
		name, c := name, c
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if err := c.definition.validate(); err != nil {
				return // expected or don't need to consider
			} else if prefix, value, err := c.definition.AuthValue(); err != nil {
				t.Fatalf("couldn't get a type from valid definition: %v", err)
			} else if prefix != c.expectedPrefix || value != c.expectedValue {
				t.Errorf("%s case is expected to be %s, %s but %s, %s", name, c.expectedPrefix, c.expectedValue, prefix, value)
			}
		})
	}
}

func Test_DefaultRequestDefinition_validate(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		definition        DefaultRequestDefinition
		expectedValidness bool
	}{
		"empty in queries": {
			definition: DefaultRequestDefinition{
				Queries: map[string][]string{
					"key1": {},
				},
			},
			expectedValidness: true,
		},
		"empty key in headers": {
			definition: DefaultRequestDefinition{
				Headers: map[string]string{
					"": "value",
				},
			},
			expectedValidness: false,
		},
		"empty key in queries": {
			definition: DefaultRequestDefinition{
				Queries: map[string][]string{
					"": {"value"},
				},
			},
			expectedValidness: false,
		},
		"empty key in form params": {
			definition: DefaultRequestDefinition{
				FormParams: map[string]string{
					"": "value",
				},
			},
			expectedValidness: false,
		},
		"empty structs": {
			definition: DefaultRequestDefinition{
				Headers:    map[string]string{},
				Queries:    map[string][]string{},
				FormParams: map[string]string{},
			},
			expectedValidness: true,
		},
		"zero": {
			definition:        DefaultRequestDefinition{},
			expectedValidness: true,
		},
	}

	for name, c := range cases {
		name, c := name, c
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if err := c.definition.validate(); (err == nil) != c.expectedValidness {
				t.Errorf("%s case is expected to be %t but %t", name, c.expectedValidness, err == nil)
			}
		})
	}
}
