package config

// Only for testing
type testConfig struct {
	ValueParam           string  `yaml:"param1" env:"TEST_PARAM1"`
	PointerParam         *string `yaml:"param2" env:"TEST_PARAM2"`
	RequiredValueParam   string  `yaml:"param3" env:"TEST_PARAM3" required:"true"`
	RequiredPointerParam *string `yaml:"param4" env:"TEST_PARAM4" required:"true"`
}
