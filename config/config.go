package config

import "cassie/leaderless_replication_v3"

type Configurations struct {
	Cluster leaderless_replication_v3.Network
	Timeout int
}
