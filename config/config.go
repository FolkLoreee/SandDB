package config

import (
	"sanddb/utils"
)

type Configurations struct {
	// For some reason viper uses mapstructure instead of the yaml tag: https://github.com/spf13/viper/issues/385
	Ring                   utils.Ring `mapstructure:"ring"`
	RepairTimeout          int        `mapstructure:"repair_timeout"`
	InternalRequestTimeout int        `mapstructure:"internal_request_timeout"`
	ReplicationFactor      int        `mapstructure:"replication_factor"`
	GCGraceSeconds         int        `mapstructure:"gc_grace_seconds"`
	Timeout                int        `mapstructure:"timeout"`
}
