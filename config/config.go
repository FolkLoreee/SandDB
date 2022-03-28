package config

import "sanddb/read_write"

type Configurations struct {
	// For some reason viper uses mapstructure instead of the yaml tag: https://github.com/spf13/viper/issues/385
	Ring              read_write.Ring `mapstructure:"ring"`
	Timeout           int             `mapstructure:"timeout"`
	ReplicationFactor int             `mapstructure:"replication_factor"`
}
