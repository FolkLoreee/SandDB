package config

import "sanddb/read_write"

type Configurations struct {
	Ring    read_write.Ring
	Timeout int
}
