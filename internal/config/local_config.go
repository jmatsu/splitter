package config

import "os"

type LocalConfig struct {
	LifecycleConfig

	DestinationPath string      `json:"destination-path,omitempty" required:"true"`
	AllowOverwrite  bool        `json:"allow-overwrite,omitempty"`
	FileMode        os.FileMode `json:"file-mode,omitempty"`
	DeleteSource    bool        `json:"delete-source,omitempty"`
}

func (c *LocalConfig) Validate() error {
	return validateMissingValues(c)
}
