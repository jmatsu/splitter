package config

type ExecutionConfig struct {
	PreSteps  [][]string `json:"pre-steps"`
	PostSteps [][]string `json:"post-steps"`
}
