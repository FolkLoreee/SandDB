package config

import "sanddb/strict_quorum"

type Configurations struct {
	Cluster strict_quorum.Cluster
	Timeout int
}
