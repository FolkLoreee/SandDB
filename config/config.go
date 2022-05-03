package config

import "sanddb/anti_entropy"

type Configurations struct {
	// For some reason viper uses mapstructure instead of the yaml tag: https://github.com/spf13/viper/issues/385
	Ring                   anti_entropy.Ring `mapstructure:"ring"`
	RepairTimeout          int               `mapstructure:"repair_timeout"`
	InternalRequestTimeout int               `mapstructure:"internal_request_timeout"`
	ReplicationFactor      int               `mapstructure:"replication_factor"`
	GCGraceSeconds         int               `mapstructure:"gc_grace_seconds"`
}
