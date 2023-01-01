package config

type ExecutionConfig struct {
	PreSteps  [][]string `json:"pre-steps,omitempty"`
	PostSteps [][]string `json:"post-steps,omitempty"`
}
