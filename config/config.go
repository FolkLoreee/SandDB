package config

import "sanddb/strict_quorum"

type Configurations struct {
	Ring    strict_quorum.Ring
	Timeout int
}
