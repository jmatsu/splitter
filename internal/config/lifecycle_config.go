package config

// ExecutionConfig represents pre-/post-hooks of each config
type ExecutionConfig struct {
	PreSteps  [][]string `json:"pre-steps"`
	PostSteps [][]string `json:"post-steps"`
}
