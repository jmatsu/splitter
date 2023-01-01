package config

import "os"

// LocalConfig contains the enough values to use local file system.
type LocalConfig struct {
	serviceNameHolder `yaml:",inline"`
	ExecutionConfig   `yaml:",inline"`

	// A destination file path. Absolute and/or relative paths are supported.
	DestinationPath string `yaml:"destination-path" required:"true"`

	// Specify true if you are okay to overwrite the destination file. Otherwise, this command fails.
	AllowOverwrite bool `yaml:"allow-overwrite,omitempty"`

	// 0644 for example. zero value means keeping the perm mode of the source file
	FileMode os.FileMode `yaml:"file-mode,omitempty"`

	// Specify true if you would like to delete the source file later and the behavior looks *move* then.
	DeleteSource bool `yaml:"delete-source,omitempty"`
}

func (c *LocalConfig) Validate() error {
	return validateMissingValues(c)
}
