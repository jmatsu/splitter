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
			expectedValidness: true,
		},
		"query token": {
			definition: CustomAuthDefinition{
				StyleFormat: QueryAssignFormatPrefix + "test",
				ValueFormat: "%s",
			},
			expectedValidness: true,
		},
		"form token": {
			definition: CustomAuthDefinition{
				StyleFormat: FormParamsAssignFormatPrefix + "test",
				ValueFormat: "%s",
			},
			expectedValidness: true,
		},
		"decorated value format": {
			definition: CustomAuthDefinition{
				StyleFormat: HeadersAssignFormatPrefix + "Authorization",
				ValueFormat: "Bearer %s",
			},
			expectedValidness: true,
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
		// AuthValue only cares about the style format. validate covers the value format.
		"too many %s in value format": {
			definition: CustomAuthDefinition{
				StyleFormat: FormParamsAssignFormatPrefix + "test",
				ValueFormat: "%s %s",
			},
			expectedPrefix: FormParamsAssignFormatPrefix,
			expectedValue:  "test",
		},
		"missing %s in value format": {
			definition: CustomAuthDefinition{
				StyleFormat: FormParamsAssignFormatPrefix + "test",
				ValueFormat: "hello",
			},
			expectedPrefix: FormParamsAssignFormatPrefix,
			expectedValue:  "test",
		},
		"zero": {definition: CustomAuthDefinition{}},
	}

	for name, c := range cases {
		name, c := name, c
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			prefix, value, err := c.definition.AuthValue()

			if c.expectedPrefix == "" {
				if err == nil {
					t.Errorf("%s case is expected to be failure but %s, %s", name, prefix, value)
				}

				return
			}

			if err != nil {
				t.Fatalf("couldn't get a type from valid definition: %v", err)
			}

			if prefix != c.expectedPrefix || value != c.expectedValue {
				t.Errorf("%s case is expected to be %s, %s but %s, %s", name, c.expectedPrefix, c.expectedValue, prefix, value)
			}
		})
	}
}

func Test_CustomServiceDefinition_validate(t *testing.T) {
	t.Parallel()

	validAuth := CustomAuthDefinition{
		StyleFormat: HeadersAssignFormatPrefix + "Authorization",
		ValueFormat: "Bearer %s",
	}

	cases := map[string]struct {
		definition        CustomServiceDefinition
		expectedValidness bool
	}{
		"source file as a form param": {
			definition: CustomServiceDefinition{
				Endpoint:         "https://example.com/path/to/upload",
				SourceFileFormat: FormParamsAssignFormatPrefix + "file",
				AuthDefinition:   validAuth,
			},
			expectedValidness: true,
		},
		"source file as a request body": {
			definition: CustomServiceDefinition{
				Endpoint:         "https://example.com/path/to/upload",
				SourceFileFormat: RequestBodyAssignFormat,
				AuthDefinition:   validAuth,
			},
			expectedValidness: true,
		},
		"with default request values": {
			definition: CustomServiceDefinition{
				Endpoint:         "https://example.com/path/to/upload",
				SourceFileFormat: FormParamsAssignFormatPrefix + "file",
				AuthDefinition:   validAuth,
				DefaultRequestDefinition: DefaultRequestDefinition{
					Headers:    map[string]string{"X-Header": "value"},
					Queries:    map[string][]string{"query": {"value"}},
					FormParams: map[string]string{"param": "value"},
				},
			},
			expectedValidness: true,
		},
		"source file format is a prefix only": {
			definition: CustomServiceDefinition{
				Endpoint:         "https://example.com/path/to/upload",
				SourceFileFormat: FormParamsAssignFormatPrefix,
				AuthDefinition:   validAuth,
			},
		},
		"source file as a query param is not supported": {
			definition: CustomServiceDefinition{
				Endpoint:         "https://example.com/path/to/upload",
				SourceFileFormat: QueryAssignFormatPrefix + "file",
				AuthDefinition:   validAuth,
			},
		},
		"unknown source file format": {
			definition: CustomServiceDefinition{
				Endpoint:         "https://example.com/path/to/upload",
				SourceFileFormat: "obababa",
				AuthDefinition:   validAuth,
			},
		},
		"invalid auth definition": {
			definition: CustomServiceDefinition{
				Endpoint:         "https://example.com/path/to/upload",
				SourceFileFormat: FormParamsAssignFormatPrefix + "file",
				AuthDefinition: CustomAuthDefinition{
					StyleFormat: "obababa",
					ValueFormat: "%s",
				},
			},
		},
		"invalid default request definition": {
			definition: CustomServiceDefinition{
				Endpoint:         "https://example.com/path/to/upload",
				SourceFileFormat: FormParamsAssignFormatPrefix + "file",
				AuthDefinition:   validAuth,
				DefaultRequestDefinition: DefaultRequestDefinition{
					Headers: map[string]string{"": "value"},
				},
			},
		},
		"zero": {definition: CustomServiceDefinition{}},
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

func Test_CustomServiceDefinition_SourceFile(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		sourceFileFormat valueAssignFormat
		expectedFormat   string
		expectedName     string
		expectedSuccess  bool
	}{
		"form param": {
			sourceFileFormat: FormParamsAssignFormatPrefix + "file",
			expectedFormat:   FormParamsAssignFormatPrefix,
			expectedName:     "file",
			expectedSuccess:  true,
		},
		"request body": {
			sourceFileFormat: RequestBodyAssignFormat,
			expectedFormat:   RequestBodyAssignFormat,
			expectedSuccess:  true,
		},
		"form param without a name": {
			sourceFileFormat: FormParamsAssignFormatPrefix,
		},
		"unknown format": {
			sourceFileFormat: "obababa",
		},
		"zero": {},
	}

	for name, c := range cases {
		name, c := name, c
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			definition := CustomServiceDefinition{SourceFileFormat: c.sourceFileFormat}

			format, fieldName, err := definition.SourceFile()

			if !c.expectedSuccess {
				if err == nil {
					t.Errorf("%s case is expected to be failure but %s, %s", name, format, fieldName)
				}

				return
			}

			if err != nil {
				t.Fatalf("%s case is expected to be success but not: %v", name, err)
			}

			if format != c.expectedFormat || fieldName != c.expectedName {
				t.Errorf("%s case is expected to be %s, %s but %s, %s", name, c.expectedFormat, c.expectedName, format, fieldName)
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
