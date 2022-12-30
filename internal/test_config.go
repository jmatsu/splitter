package internal

// Only for testing
type testConfig struct {
	ValueParam           string  `json:"param1" env:"TEST_PARAM1"`
	PointerParam         *string `json:"param2" env:"TEST_PARAM2"`
	RequiredValueParam   string  `json:"param3" env:"TEST_PARAM3" required:"true"`
	RequiredPointerParam *string `json:"param4" env:"TEST_PARAM4" required:"true"`
}
