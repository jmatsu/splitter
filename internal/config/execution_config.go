package config

// ExecutionConfig represents pre-/post-hooks of each config
type ExecutionConfig struct {
	PreSteps  [][]string `yaml:"pre-steps,omitempty"`
	PostSteps [][]string `yaml:"post-steps,omitempty"`
}
