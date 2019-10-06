package cache

import (
	"github.com/lpicanco/microcache"
	"github.com/lpicanco/microcache/configuration"
)
 
// Cache instance
var Cache = microcache.New(configuration.DefaultConfiguration(1000))
