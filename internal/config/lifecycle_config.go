package config

type LifecycleConfig struct {
	PreSteps  [][]string `json:"pre-steps,omitempty"`
	PostSteps [][]string `json:"post-steps,omitempty"`
}
