package config

import "os"

type LocalConfig struct {
	ExecutionConfig

	DestinationPath string      `json:"destination-path" required:"true"`
	AllowOverwrite  bool        `json:"allow-overwrite"`
	FileMode        os.FileMode `json:"file-mode"`
	DeleteSource    bool        `json:"delete-source"`
}

func (c *LocalConfig) Validate() error {
	return validateMissingValues(c)
}
