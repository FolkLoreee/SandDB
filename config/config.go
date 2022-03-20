package config

import "sanddb/consistent_hashing"

type Configurations struct {
	Ring consistent_hashing.Ring
}